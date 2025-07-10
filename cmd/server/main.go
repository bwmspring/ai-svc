package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ai-svc/internal/config"
	"ai-svc/internal/model"
	"ai-svc/internal/routes"
	"ai-svc/pkg/database"
	"ai-svc/pkg/logger"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	if err := config.LoadConfig("./configs/config.yaml"); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化日志
	if err := logger.Init(
		config.AppConfig.Log.Level,
		config.AppConfig.Log.Format,
		config.AppConfig.Log.Output,
	); err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}

	logger.Info("开始启动服务", map[string]interface{}{
		"version": "1.0.0",
		"mode":    config.AppConfig.Server.Mode,
	})

	// 连接数据库
	// if err := database.Connect(); err != nil {
	// 	logger.Fatal("数据库连接失败", map[string]interface{}{"error": err.Error()})
	// }
	// defer database.Close()

	// // 自动迁移数据库表
	// if err := migrateDatabase(); err != nil {
	// 	logger.Fatal("数据库迁移失败", map[string]interface{}{"error": err.Error()})
	// }

	// 设置Gin模式
	gin.SetMode(config.AppConfig.Server.Mode)

	// 设置路由
	router := routes.SetupRoutes()

	// 创建HTTP服务器
	server := &http.Server{
		Addr:         ":" + config.AppConfig.Server.Port,
		Handler:      router,
		ReadTimeout:  time.Duration(config.AppConfig.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(config.AppConfig.Server.WriteTimeout) * time.Second,
	}

	// 启动服务器（优雅启动）
	go func() {
		logger.Info("服务器启动", map[string]interface{}{
			"port": config.AppConfig.Server.Port,
			"mode": config.AppConfig.Server.Mode,
		})
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("服务器启动失败", map[string]interface{}{"error": err.Error()})
		}
	}()

	// 等待中断信号以优雅关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("正在关闭服务器...", nil)

	// 优雅关闭服务器，等待现有连接完成
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("服务器强制关闭", map[string]interface{}{"error": err.Error()})
	} else {
		logger.Info("服务器已优雅关闭", nil)
	}
}

// migrateDatabase 数据库迁移
func migrateDatabase() error {
	db := database.GetDB()

	logger.Info("开始数据库迁移", nil)

	// 自动迁移表结构
	if err := db.AutoMigrate(
		&model.User{},
		&model.UserProfile{},
	); err != nil {
		return fmt.Errorf("数据库迁移失败: %w", err)
	}

	logger.Info("数据库迁移完成", nil)
	return nil
}
