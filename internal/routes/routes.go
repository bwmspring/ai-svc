package routes

import (
	"time"

	"ai-svc/internal/controller"
	"ai-svc/internal/middleware"
	"ai-svc/internal/repository"
	"ai-svc/internal/service"
	"ai-svc/pkg/response"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置路由.
func SetupRoutes() *gin.Engine {
	// 创建Gin引擎
	router := gin.New()

	// 应用全局中间件（注意顺序）
	router.Use(middleware.RequestID()) // 首先添加请求ID
	router.Use(middleware.Logger())    // 然后记录日志（包含请求ID）
	router.Use(middleware.Recovery())  // 错误恢复
	router.Use(middleware.CORS())      // 跨域支持

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		response.Success(c, gin.H{
			"status":     "ok",
			"message":    "服务运行正常",
			"request_id": middleware.GetRequestID(c),
		})
	})

	// 初始化依赖
	userRepo := repository.NewUserRepository()
	smsRepo := repository.NewSMSRepository()
	deviceRepo := repository.NewDeviceRepository()
	smsService := service.NewSMSService(smsRepo)
	deviceService := service.NewDeviceService(deviceRepo)
	userService := service.NewUserService(userRepo, smsService, deviceService)
	userController := controller.NewUserController(userService, smsService)
	smsController := controller.NewSMSController(smsService)

	// 创建频率限制器
	rateLimiter := middleware.NewRateLimiter()

	// API路由组
	api := router.Group("/api/v1")
	{
		// 公开接口（无需认证）
		// 使用SMS限流配置
		api.POST(
			"/sms/send",
			middleware.SMSRateLimit(rateLimiter),
			smsController.SendSMS,
		)
		// 登录接口使用登录专用的限流配置
		api.POST(
			"/auth/login",
			middleware.LoginRateLimit(rateLimiter),
			userController.LoginWithSMS,
		)
		// Token刷新接口（公开，使用登录限流）
		api.POST(
			"/auth/refresh",
			middleware.LoginRateLimit(rateLimiter),
			userController.RefreshToken,
		)

		// 需要认证的接口 - 统一使用设备验证
		auth := api.Group("/users")
		// 使用增强的JWT+设备认证，确保设备有效性
		auth.Use(middleware.JWTWithDeviceAuth())
		{
			// 当前用户相关接口（使用一般API限流）
			auth.GET(
				"/profile",
				middleware.APIRateLimit(rateLimiter),
				userController.GetProfile,
			)
			auth.PUT(
				"/profile",
				middleware.APIRateLimit(rateLimiter),
				userController.UpdateProfile,
			)

			// 设备管理接口
			auth.GET(
				"/devices",
				middleware.APIRateLimit(rateLimiter),
				userController.GetUserDevices,
			)

			// 设备踢出接口（敏感操作，使用严格限流）
			auth.POST(
				"/devices/kick",
				middleware.ConfigRateLimit(rateLimiter, "login"), // 使用登录限流作为严格限流
				userController.KickDevices,
			)
		}

		// 设备管理接口（使用增强认证）
		devices := api.Group("/devices")
		devices.Use(middleware.JWTWithDeviceAuth())
		{
			// 设备心跳上报
			devices.POST(
				"/heartbeat",
				middleware.APIRateLimit(rateLimiter),
				func(c *gin.Context) {
					response.Success(c, gin.H{"message": "heartbeat received"})
				},
			)
		}

		// 管理接口 - 统一使用设备验证
		admin := api.Group("/users")
		admin.Use(middleware.JWTWithDeviceAuth())
		{
			// 用户管理接口（查询操作使用API限流）
			admin.GET(
				"/list",
				middleware.APIRateLimit(rateLimiter),
				userController.GetUserList,
			)
			admin.GET(
				"/search",
				middleware.APIRateLimit(rateLimiter),
				userController.SearchUsers,
			)
			admin.GET(
				"/:id",
				middleware.APIRateLimit(rateLimiter),
				userController.GetUserByID,
			)
			admin.DELETE(
				"/:id",
				middleware.ConfigRateLimit(rateLimiter, "login"), // 使用登录限流作为严格限流
				userController.DeleteUser,
			)
		}

		// 移动端专用接口（设备验证+设备类型限制）
		mobile := api.Group("/mobile")
		mobile.Use(middleware.JWTWithDeviceAuth())                      // 先进行设备验证
		mobile.Use(middleware.DeviceTypeMiddleware("mobile", "tablet")) // 再限制设备类型
		{
			mobile.GET(
				"/config",
				middleware.APIRateLimit(rateLimiter),
				func(c *gin.Context) {
					response.Success(c, gin.H{
						"message":     "移动端配置",
						"device_type": middleware.GetCurrentDeviceType(c),
					})
				},
			)
		}

		// 管理后台接口（设备验证+设备类型限制）
		adminPanel := api.Group("/admin")
		adminPanel.Use(middleware.JWTWithDeviceAuth())         // 先进行设备验证
		adminPanel.Use(middleware.DeviceTypeMiddleware("web")) // 再限制设备类型
		{
			adminPanel.GET(
				"/dashboard",
				middleware.APIRateLimit(rateLimiter),
				func(c *gin.Context) {
					response.Success(c, gin.H{
						"message":     "管理后台仪表板",
						"user_id":     middleware.GetCurrentUserID(c),
						"device_id":   middleware.GetCurrentDeviceID(c),
						"device_type": middleware.GetCurrentDeviceType(c),
					})
				},
			)
		}

		// 自定义配置示例：创建一个极严格的限流配置
		veryStrictConfig := middleware.RateLimitConfig{
			Capacity:       1,
			RefillRate:     1,
			RefillInterval: time.Minute,
			ErrorMsg:       "此操作每分钟只能执行1次，请稍后再试",
		}

		// 使用自定义配置的示例接口
		api.POST(
			"/dangerous-operation",
			middleware.CustomRateLimit(rateLimiter, veryStrictConfig),
			func(c *gin.Context) {
				response.Success(c, gin.H{"message": "危险操作执行成功"})
			},
		)
	}

	return router
}
