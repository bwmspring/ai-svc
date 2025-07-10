package middleware

import (
	"net/http"
	"time"

	"ai-svc/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestID is a middleware that injects a 'X-Request-ID' into the context and request/response header of each request.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查请求头中是否已有 X-Request-ID
		requestID := c.GetHeader("X-Request-ID")

		// 如果没有，则生成一个新的
		if requestID == "" {
			requestID = generateRequestID()
		}

		// 将 Request ID 添加到上下文中
		c.Set("request_id", requestID)

		// 将 Request ID 添加到响应头中
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// generateRequestID 生成唯一的请求ID
func generateRequestID() string {
	return uuid.New().String()
}

// GetRequestID 从上下文中获取请求ID
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		return requestID.(string)
	}
	return ""
}

// CORS 跨域中间件
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin")

		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization, Cache-Control, X-File-Name")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")

		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		c.Next()
	}
}

// Logger 日志中间件
func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// 获取请求ID
		requestID := param.Request.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = "unknown"
		}

		logger.Info("HTTP请求", map[string]interface{}{
			"request_id": requestID,
			"timestamp":  param.TimeStamp.Format(time.RFC3339),
			"status":     param.StatusCode,
			"method":     param.Method,
			"path":       param.Path,
			"query":      param.Request.URL.RawQuery,
			"ip":         param.ClientIP,
			"user_agent": param.Request.UserAgent(),
			"latency":    param.Latency,
			"error":      param.ErrorMessage,
		})
		return ""
	})
}

// Recovery 恢复中间件
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		requestID := GetRequestID(c)
		if requestID == "" {
			requestID = "unknown"
		}

		logger.Error("服务器内部错误", map[string]interface{}{
			"request_id": requestID,
			"error":      recovered,
			"path":       c.Request.URL.Path,
			"method":     c.Request.Method,
			"ip":         c.ClientIP(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":       500,
			"message":    "服务器内部错误",
			"request_id": requestID,
		})
	})
}

// RateLimit 限流中间件（简单实现）
func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 这里可以实现基于IP的限流逻辑
		// 可以使用 Redis 或内存存储来记录请求频率
		c.Next()
	}
}
