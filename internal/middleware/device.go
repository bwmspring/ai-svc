package middleware

import (
	"ai-svc/pkg/logger"
	"ai-svc/pkg/response"

	"github.com/gin-gonic/gin"
)

// DeviceService 设备服务接口（避免循环依赖）
type DeviceService interface {
	ValidateDeviceSession(userID uint, deviceID, sessionID string) (bool, error)
	UpdateDeviceActivity(deviceID string) error
}

// DeviceValidationMiddleware 设备验证中间件
// 独立于JWT认证，专门处理设备会话验证
func DeviceValidationMiddleware(deviceService DeviceService) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := GetRequestID(c)
		userID := GetCurrentUserID(c)
		deviceID := GetCurrentDeviceID(c)
		sessionID := GetCurrentSessionID(c)

		// 基本信息检查
		if userID == 0 || deviceID == "" {
			logger.Warn("设备验证失败：用户或设备信息缺失", map[string]interface{}{
				"request_id": requestID,
				"user_id":    userID,
				"device_id":  deviceID,
			})
			response.Error(c, response.UNAUTHORIZED, "用户或设备信息缺失")
			c.Abort()
			return
		}

		// 验证设备会话（如果有会话ID）
		if sessionID != "" {
			isValid, err := deviceService.ValidateDeviceSession(userID, deviceID, sessionID)
			if err != nil {
				logger.Error("设备会话验证失败", map[string]interface{}{
					"request_id": requestID,
					"user_id":    userID,
					"device_id":  deviceID,
					"session_id": sessionID,
					"error":      err.Error(),
				})
				response.Error(c, response.UNAUTHORIZED, "设备会话验证失败")
				c.Abort()
				return
			}
			if !isValid {
				logger.Warn("设备会话无效", map[string]interface{}{
					"request_id": requestID,
					"user_id":    userID,
					"device_id":  deviceID,
					"session_id": sessionID,
				})
				response.Error(c, response.UNAUTHORIZED, "设备会话无效，请重新登录")
				c.Abort()
				return
			}
		}

		// 更新设备活跃时间
		if err := deviceService.UpdateDeviceActivity(deviceID); err != nil {
			logger.Error("更新设备活跃时间失败", map[string]interface{}{
				"request_id": requestID,
				"device_id":  deviceID,
				"error":      err.Error(),
			})
			// 不中断请求处理，只记录错误
		}

		c.Next()
	}
}

// AuthWithDeviceValidation 认证+设备验证的组合中间件
func AuthWithDeviceValidation(deviceService DeviceService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先进行JWT认证
		JWTAuth()(c)

		// 如果认证失败，直接返回
		if c.IsAborted() {
			return
		}

		// 然后进行设备验证
		DeviceValidationMiddleware(deviceService)(c)
	}
}
