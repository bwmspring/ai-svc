package routes

import (
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
		// 使用预定义的SMS限流配置
		api.POST(
			"/sms/send",
			middleware.CustomRateLimit(rateLimiter, middleware.SMSRateLimitConfig),
			smsController.SendSMS,
		)
		// 登录接口使用登录专用的限流配置
		api.POST(
			"/auth/login",
			middleware.CustomRateLimit(rateLimiter, middleware.LoginRateLimitConfig),
			userController.LoginWithSMS,
		)

		// 需要认证的接口
		auth := api.Group("/users")
		// 使用基础JWT认证
		auth.Use(middleware.JWTAuth())
		// 添加设备验证中间件
		auth.Use(middleware.DeviceValidationMiddleware(deviceService))
		{
			// 当前用户相关接口（使用一般API限流）
			auth.GET(
				"/profile",
				middleware.CustomRateLimit(rateLimiter, middleware.APIRateLimitConfig),
				userController.GetProfile,
			)
			auth.PUT(
				"/profile",
				middleware.CustomRateLimit(rateLimiter, middleware.APIRateLimitConfig),
				userController.UpdateProfile,
			)

			// 设备管理接口（敏感操作，使用严格限流）
			auth.GET(
				"/devices",
				middleware.CustomRateLimit(rateLimiter, middleware.APIRateLimitConfig),
				userController.GetUserDevices,
			)
			auth.POST(
				"/devices/kick",
				middleware.CustomRateLimit(rateLimiter, middleware.StrictRateLimitConfig),
				userController.KickDevices,
			)

			// 用户管理接口（查询操作使用宽松限流，删除操作使用严格限流）
			auth.GET(
				"/list",
				middleware.CustomRateLimit(rateLimiter, middleware.LaxRateLimitConfig),
				userController.GetUserList,
			)
			auth.GET(
				"/search",
				middleware.CustomRateLimit(rateLimiter, middleware.LaxRateLimitConfig),
				userController.SearchUsers,
			)
			auth.GET(
				"/:id",
				middleware.CustomRateLimit(rateLimiter, middleware.APIRateLimitConfig),
				userController.GetUserByID,
			)
			auth.DELETE(
				"/:id",
				middleware.CustomRateLimit(rateLimiter, middleware.StrictRateLimitConfig),
				userController.DeleteUser,
			)
		}

		// 移动端专用接口（只允许mobile和tablet设备类型）
		mobile := api.Group("/mobile")
		mobile.Use(middleware.JWTAuth())
		mobile.Use(middleware.DeviceTypeMiddleware("mobile", "tablet"))
		{
			mobile.GET(
				"/config",
				middleware.CustomRateLimit(rateLimiter, middleware.APIRateLimitConfig),
				func(c *gin.Context) {
					response.Success(c, gin.H{
						"message":     "移动端配置",
						"device_type": middleware.GetCurrentDeviceType(c),
					})
				},
			)
		}

		// 管理后台接口（只允许web设备类型）
		admin := api.Group("/admin")
		admin.Use(middleware.JWTAuth())
		admin.Use(middleware.DeviceTypeMiddleware("web"))
		{
			admin.GET(
				"/dashboard",
				middleware.CustomRateLimit(rateLimiter, middleware.APIRateLimitConfig),
				func(c *gin.Context) {
					response.Success(c, gin.H{
						"message":     "管理后台",
						"device_type": middleware.GetCurrentDeviceType(c),
					})
				},
			)
		}

		// 自定义配置示例：创建一个极严格的限流配置
		veryStrictConfig := middleware.RateLimitConfig{
			Capacity:   1,
			RefillRate: 1,
			ErrorMsg:   "此操作每分钟只能执行1次，请稍后再试",
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
