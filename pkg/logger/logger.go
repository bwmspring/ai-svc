package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

// Init 初始化日志
func Init(level, format, output string) error {
	Log = logrus.New()

	// 设置日志级别
	switch level {
	case "debug":
		Log.SetLevel(logrus.DebugLevel)
	case "info":
		Log.SetLevel(logrus.InfoLevel)
	case "warn":
		Log.SetLevel(logrus.WarnLevel)
	case "error":
		Log.SetLevel(logrus.ErrorLevel)
	default:
		Log.SetLevel(logrus.InfoLevel)
	}

	// 设置日志格式
	switch format {
	case "json":
		Log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	case "text":
		Log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	default:
		Log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}

	// 设置输出
	switch output {
	case "stdout":
		Log.SetOutput(os.Stdout)
	case "file":
		// TODO: 实现文件输出逻辑，可以结合 lumberjack 库
		Log.SetOutput(os.Stdout)
	default:
		Log.SetOutput(os.Stdout)
	}

	return nil
}

// Debug 调试日志
func Debug(message string, fields map[string]interface{}) {
	if fields != nil {
		Log.WithFields(logrus.Fields(fields)).Debug(message)
	} else {
		Log.Debug(message)
	}
}

// Info 信息日志
func Info(message string, fields map[string]interface{}) {
	if fields != nil {
		Log.WithFields(logrus.Fields(fields)).Info(message)
	} else {
		Log.Info(message)
	}
}

// Warn 警告日志
func Warn(message string, fields map[string]interface{}) {
	if fields != nil {
		Log.WithFields(logrus.Fields(fields)).Warn(message)
	} else {
		Log.Warn(message)
	}
}

// Error 错误日志
func Error(message string, fields map[string]interface{}) {
	if fields != nil {
		Log.WithFields(logrus.Fields(fields)).Error(message)
	} else {
		Log.Error(message)
	}
}

// Fatal 致命错误日志
func Fatal(message string, fields map[string]interface{}) {
	if fields != nil {
		Log.WithFields(logrus.Fields(fields)).Fatal(message)
	} else {
		Log.Fatal(message)
	}
}
