package routes

import (
	"ai-svc/internal/controller"
	"ai-svc/internal/middleware"
	"ai-svc/internal/repository"
	"ai-svc/internal/service"
	"ai-svc/pkg/response"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置路由
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
	smsService := service.NewSMSService(smsRepo)
	userService := service.NewUserService(userRepo, smsService)
	userController := controller.NewUserController(userService, smsService)
	smsController := controller.NewSMSController(smsService)

	// 创建频率限制器
	rateLimiter := middleware.NewRateLimiter()

	// API路由组
	api := router.Group("/api/v1")
	{
		// 公开接口（无需认证）
		api.POST("/sms/send", middleware.SMSRateLimit(rateLimiter), smsController.SendSMS) // 发送短信验证码（带频率限制）
		api.POST("/auth/login", userController.LoginWithSMS)                               // 手机号+验证码登录（同时完成注册）

		// 需要认证的接口
		auth := api.Group("/users")
		auth.Use(middleware.JWTAuth())
		{
			// 当前用户相关接口
			auth.GET("/profile", userController.GetProfile)
			auth.PUT("/profile", userController.UpdateProfile)

			// 用户管理接口（可以根据需要添加权限控制）
			auth.GET("/list", userController.GetUserList)
			auth.GET("/search", userController.SearchUsers)
			auth.GET("/:id", userController.GetUserByID)
			auth.DELETE("/:id", userController.DeleteUser)
		}
	}

	return router
}
