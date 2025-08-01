package cmd

import (
	"ai-svc/internal/config"
	"ai-svc/internal/model"
	"ai-svc/internal/routes"
	"ai-svc/pkg/database"
	"ai-svc/pkg/logger"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// 服务器启动相关的命令行参数.
	serverPort    string // 服务器端口
	serverMode    string // 运行模式：debug, release, test
	enableProfile bool   // 是否启用性能分析
)

// 这是应用程序的主要命令，用于启动 HTTP 服务器.
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "启动 AI 服务 HTTP 服务器",
	Long: `启动 AI 服务的 HTTP 服务器。

服务器提供以下功能：
• RESTful API 接口服务
• 用户认证与授权
• 短信验证码发送
• 设备管理与安全控制
• 智能限流保护
• 健康检查接口

示例用法：
  ai-svc server                          # 使用默认配置启动
  ai-svc server --port 8080              # 指定端口启动
  ai-svc server --mode release           # 生产模式启动
  ai-svc server --config custom.yaml    # 使用自定义配置文件
  ai-svc server --verbose               # 启用详细日志输出`,

	// PreRunE 在主要逻辑执行前运行，用于验证参数和初始化
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// 验证运行模式参数
		if serverMode != "" && serverMode != "debug" && serverMode != "release" && serverMode != "test" {
			return fmt.Errorf("无效的运行模式: %s (支持: debug, release, test)", serverMode)
		}
		return nil
	},

	// RunE 是命令的主要执行函数，返回错误类型
	RunE: func(cmd *cobra.Command, args []string) error {
		return runServer()
	},
}

// init 函数用于初始化 server 命令的参数.
func init() {
	// 将 server 命令添加到根命令
	rootCmd.AddCommand(serverCmd)

	// 定义 server 命令专用的标志

	// --port 标志：指定服务器监听端口
	serverCmd.Flags().StringVarP(&serverPort, "port", "p", "",
		"服务器监听端口 (默认: 8080)")

	// --mode 标志：指定运行模式
	serverCmd.Flags().StringVarP(&serverMode, "mode", "m", "",
		"运行模式 (debug|release|test, 默认: debug)")

	// --profile 标志：启用性能分析
	serverCmd.Flags().BoolVar(&enableProfile, "profile", false,
		"启用 pprof 性能分析接口")

	// 将标志绑定到 Viper，这样可以通过配置文件或环境变量覆盖
	viper.BindPFlag("server.port", serverCmd.Flags().Lookup("port"))
	viper.BindPFlag("server.mode", serverCmd.Flags().Lookup("mode"))
	viper.BindPFlag("server.profile", serverCmd.Flags().Lookup("profile"))
}

// 包含完整的启动流程：配置加载、日志初始化、服务器启动、优雅关闭.
func runServer() error {
	fmt.Println("🚀 正在启动 AI 服务...")

	// 第一步：加载配置文件
	if err := loadConfiguration(); err != nil {
		return fmt.Errorf("配置加载失败: %w", err)
	}

	// 第二步：初始化日志系统
	if err := initializeLogger(); err != nil {
		return fmt.Errorf("日志初始化失败: %w", err)
	}

	// 第三步：应用命令行参数覆盖配置
	applyCommandLineOverrides()

	// 第四步：连接数据库
	if err := connectDatabase(); err != nil {
		return fmt.Errorf("数据库连接失败: %w", err)
	}

	// 第五步：设置 Gin 框架模式
	setupGinMode()

	// 第六步：初始化路由和中间件
	router := routes.SetupRoutes()

	// 第六步：配置 HTTP 服务器
	server := configureHTTPServer(router)

	// 第七步：启动服务器（异步）
	startServer(server)

	// 第八步：等待关闭信号并优雅关闭
	return gracefulShutdown(server)
}

// loadConfiguration 加载应用配置.
func loadConfiguration() error {
	// 确定配置文件路径
	configPath := "./configs/config.yaml"
	if cfgFile != "" {
		configPath = cfgFile
	} else if viper.ConfigFileUsed() != "" {
		configPath = viper.ConfigFileUsed()
	}

	// 加载配置文件
	if err := config.LoadConfig(configPath); err != nil {
		return err
	}

	if verbose {
		fmt.Printf("✅ 配置文件加载成功: %s\n", configPath)
	}

	return nil
}

// initializeLogger 初始化日志系统.
func initializeLogger() error {
	// 从配置中读取日志相关参数
	logLevel := config.AppConfig.Logger.Level
	logFormat := config.AppConfig.Logger.Format
	logOutput := config.AppConfig.Logger.Output

	// 如果启用了详细模式，强制设置为 debug 级别
	if verbose {
		logLevel = "debug"
	}

	// 初始化日志系统
	if err := logger.Init(logLevel, logFormat, logOutput); err != nil {
		return err
	}

	logger.Info("日志系统初始化成功", map[string]any{
		"level":  logLevel,
		"format": logFormat,
		"output": logOutput,
	})

	return nil
}

// applyCommandLineOverrides 应用命令行参数覆盖配置文件设置.
func applyCommandLineOverrides() {
	// 如果命令行指定了端口，覆盖配置文件中的端口设置
	if serverPort != "" {
		if port, err := strconv.Atoi(serverPort); err == nil {
			config.AppConfig.Server.Port = port
			logger.Info("端口被命令行参数覆盖", map[string]any{
				"port": port,
			})
		} else {
			logger.Warn("无效的端口号，使用配置文件中的端口", map[string]any{
				"invalid_port": serverPort,
				"config_port":  config.AppConfig.Server.Port,
			})
		}
	}

	// 如果命令行指定了运行模式，覆盖配置文件中的模式设置
	if serverMode != "" {
		config.AppConfig.Server.Mode = serverMode
		logger.Info("运行模式被命令行参数覆盖", map[string]any{
			"mode": serverMode,
		})
	}

	// 如果启用了性能分析，记录日志
	if enableProfile {
		logger.Info("性能分析已启用", map[string]any{
			"profile_enabled": true,
		})
	}
}

// setupGinMode 设置 Gin 框架的运行模式.
func setupGinMode() {
	gin.SetMode(config.AppConfig.Server.Mode)

	logger.Info("Gin 框架模式设置完成", map[string]any{
		"gin_mode": config.AppConfig.Server.Mode,
	})
}

// configureHTTPServer 配置 HTTP 服务器参数.
func configureHTTPServer(router *gin.Engine) *http.Server {
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.AppConfig.Server.Port),
		Handler:      router,
		ReadTimeout:  config.AppConfig.Server.ReadTimeout,
		WriteTimeout: config.AppConfig.Server.WriteTimeout,
		// 设置最大请求头大小（1MB）
		MaxHeaderBytes: 1 << 20,
	}

	logger.Info("HTTP 服务器配置完成", map[string]any{
		"addr":          server.Addr,
		"read_timeout":  config.AppConfig.Server.ReadTimeout,
		"write_timeout": config.AppConfig.Server.WriteTimeout,
	})

	return server
}

// startServer 异步启动 HTTP 服务器.
func startServer(server *http.Server) {
	go func() {
		fmt.Printf("🌟 服务器启动成功，监听端口: %d\n", config.AppConfig.Server.Port)
		fmt.Printf("📊 运行模式: %s\n", config.AppConfig.Server.Mode)
		fmt.Printf("🔗 访问地址: http://localhost:%d\n", config.AppConfig.Server.Port)
		fmt.Printf("💚 健康检查: http://localhost:%d/health\n", config.AppConfig.Server.Port)

		logger.Info("服务器启动", map[string]any{
			"port":    config.AppConfig.Server.Port,
			"mode":    config.AppConfig.Server.Mode,
			"version": "1.0.0",
			"pid":     os.Getpid(),
		})

		// 启动服务器，如果失败则记录错误
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("服务器启动失败", map[string]any{
				"error": err.Error(),
			})
		}
	}()
}

// connectDatabase 连接数据库并执行迁移
func connectDatabase() error {
	// 连接数据库
	if err := database.Connect(); err != nil {
		return fmt.Errorf("数据库连接失败: %w", err)
	}

	// 执行数据库迁移
	db := database.GetDB()
	if err := db.AutoMigrate(
		&model.User{},
		&model.MessageDefinition{},
		&model.UserMessage{},
	); err != nil {
		return fmt.Errorf("数据库迁移失败: %w", err)
	}

	logger.Info("数据库连接和迁移成功", map[string]any{})
	return nil
}

// 等待系统信号，然后优雅地关闭服务器，确保正在处理的请求能够完成.
func gracefulShutdown(server *http.Server) error {
	// 创建信号通道，监听系统中断信号
	quit := make(chan os.Signal, 1)

	// 注册要监听的信号：SIGINT (Ctrl+C) 和 SIGTERM (终止信号)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 阻塞等待信号
	sig := <-quit

	fmt.Printf("\n🛑 收到关闭信号: %v\n", sig)
	logger.Info("收到关闭信号，开始优雅关闭", map[string]any{
		"signal": sig.String(),
	})

	// 设置关闭超时时间（30秒）
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 优雅关闭服务器
	fmt.Println("⏳ 正在等待现有连接完成...")
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("服务器强制关闭", map[string]any{
			"error": err.Error(),
		})
		return fmt.Errorf("服务器强制关闭: %w", err)
	}

	fmt.Println("✅ 服务器已优雅关闭")
	logger.Info("服务器已优雅关闭", nil)

	return nil
}
