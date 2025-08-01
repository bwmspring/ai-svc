package database

import (
	"ai-svc/internal/config"
	"ai-svc/pkg/logger"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

var DB *gorm.DB

// Connect 连接数据库.
func Connect() error {
	dbConfig := config.AppConfig.Database
	dsn := dbConfig.GetDSN()

	// 配置GORM日志
	var logLevel gormLogger.LogLevel
	switch config.AppConfig.Logger.Level {
	case "debug":
		logLevel = gormLogger.Info
	case "info":
		logLevel = gormLogger.Warn
	default:
		logLevel = gormLogger.Error
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormLogger.Default.LogMode(logLevel),
	})
	if err != nil {
		logger.Error("数据库连接失败", map[string]any{
			"error": err.Error(),
			"dsn":   dsn,
		})
		return fmt.Errorf("数据库连接失败: %w", err)
	}

	// 获取底层sql.DB对象
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取数据库实例失败: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(dbConfig.MaxIdleConns)
	sqlDB.SetMaxOpenConns(dbConfig.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("数据库连接测试失败: %w", err)
	}

	DB = db
	logger.Info("数据库连接成功", map[string]any{
		"host":     dbConfig.Host,
		"port":     dbConfig.Port,
		"database": dbConfig.Database,
	})

	return nil
}

// Close 关闭数据库连接.
func Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// GetDB 获取数据库实例.
func GetDB() *gorm.DB {
	return DB
}
