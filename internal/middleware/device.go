package middleware

import (
	"ai-svc/pkg/logger"
	"ai-svc/pkg/response"
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

// 设备验证相关常量.
const (
	// 错误消息.
	ErrMsgUserOrDeviceInfoMissing       = "用户或设备信息缺失"
	ErrMsgDeviceSessionValidationFailed = "设备会话验证失败"
	ErrMsgDeviceSessionInvalid          = "设备会话无效，请重新登录"
	ErrMsgUpdateDeviceActivityFailed    = "更新设备活跃时间失败"

	// 上下文键.
	ContextKeyDeviceValidation = "device_validation_enabled"

	// 超时设置.
	DefaultDeviceValidationTimeout = 5 * time.Second
)

// DeviceValidationConfig 设备验证配置.
type DeviceValidationConfig struct {
	// 是否启用设备验证
	Enabled bool
	// 是否必须验证会话ID
	RequireSessionID bool
	// 验证超时时间
	Timeout time.Duration
	// 是否启用设备活跃时间更新
	UpdateActivity bool
}

// DefaultDeviceValidationConfig 默认设备验证配置.
func DefaultDeviceValidationConfig() *DeviceValidationConfig {
	return &DeviceValidationConfig{
		Enabled:          true,
		RequireSessionID: false,
		Timeout:          DefaultDeviceValidationTimeout,
		UpdateActivity:   true,
	}
}

// DeviceService 设备服务接口（避免循环依赖）.
type DeviceService interface {
	ValidateDeviceSession(userID uint, deviceID, sessionID string) (bool, error)
	UpdateDeviceActivity(deviceID string) error
}

// DeviceInfo 设备信息结构体.
type DeviceInfo struct {
	RequestID string
	UserID    uint
	DeviceID  string
	SessionID string
}

// extractDeviceInfo 提取设备信息.
func extractDeviceInfo(c *gin.Context) *DeviceInfo {
	return &DeviceInfo{
		RequestID: GetRequestID(c),
		UserID:    GetCurrentUserID(c),
		DeviceID:  GetCurrentDeviceID(c),
		SessionID: GetCurrentSessionID(c),
	}
}

// logFields 构建日志字段.
func (di *DeviceInfo) logFields() map[string]any {
	fields := map[string]any{
		"request_id": di.RequestID,
		"user_id":    di.UserID,
		"device_id":  di.DeviceID,
	}
	if di.SessionID != "" {
		fields["session_id"] = di.SessionID
	}
	return fields
}

// validateBasicInfo 验证基本信息.
func (di *DeviceInfo) validateBasicInfo() bool {
	return di.UserID > 0 && di.DeviceID != ""
}

// 独立于JWT认证，专门处理设备会话验证.
func DeviceValidationMiddleware(deviceService DeviceService) gin.HandlerFunc {
	return DeviceValidationMiddlewareWithConfig(deviceService, DefaultDeviceValidationConfig())
}

// DeviceValidationMiddlewareWithConfig 带配置的设备验证中间件.
func DeviceValidationMiddlewareWithConfig(deviceService DeviceService, config *DeviceValidationConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否启用设备验证
		if !config.Enabled {
			c.Next()
			return
		}

		// 提取设备信息
		deviceInfo := extractDeviceInfo(c)

		// 基本信息检查
		if !deviceInfo.validateBasicInfo() {
			logger.Warn("设备验证失败：用户或设备信息缺失", deviceInfo.logFields())
			response.Error(c, response.UNAUTHORIZED, ErrMsgUserOrDeviceInfoMissing)
			c.Abort()
			return
		}

		// 验证设备会话
		if err := validateDeviceSession(c, deviceService, deviceInfo, config); err != nil {
			c.Abort()
			return
		}

		// 更新设备活跃时间
		if config.UpdateActivity {
			updateDeviceActivity(deviceService, deviceInfo)
		}

		c.Next()
	}
}

// validateDeviceSession 验证设备会话.
func validateDeviceSession(
	c *gin.Context,
	deviceService DeviceService,
	deviceInfo *DeviceInfo,
	config *DeviceValidationConfig,
) error {
	// 如果没有会话ID且不要求会话ID，则跳过验证
	if deviceInfo.SessionID == "" && !config.RequireSessionID {
		return nil
	}

	// 如果要求会话ID但没有提供，则返回错误
	if deviceInfo.SessionID == "" && config.RequireSessionID {
		logger.Warn("设备验证失败：缺少会话ID", deviceInfo.logFields())
		response.Error(c, response.UNAUTHORIZED, ErrMsgDeviceSessionInvalid)
		return &DeviceValidationError{Message: "缺少会话ID"}
	}

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	// 在goroutine中执行验证以支持超时
	resultChan := make(chan validationResult, 1)
	go func() {
		isValid, err := deviceService.ValidateDeviceSession(
			deviceInfo.UserID,
			deviceInfo.DeviceID,
			deviceInfo.SessionID,
		)
		resultChan <- validationResult{isValid: isValid, err: err}
	}()

	// 等待结果或超时
	select {
	case result := <-resultChan:
		if result.err != nil {
			logger.Error("设备会话验证失败", mergeMap(deviceInfo.logFields(), map[string]any{
				"error": result.err.Error(),
			}))
			response.Error(c, response.UNAUTHORIZED, ErrMsgDeviceSessionValidationFailed)
			return result.err
		}
		if !result.isValid {
			logger.Warn("设备会话无效", deviceInfo.logFields())
			response.Error(c, response.UNAUTHORIZED, ErrMsgDeviceSessionInvalid)
			return &DeviceValidationError{Message: "设备会话无效"}
		}
	case <-ctx.Done():
		logger.Error("设备会话验证超时", deviceInfo.logFields())
		response.Error(c, response.UNAUTHORIZED, ErrMsgDeviceSessionValidationFailed)
		return &DeviceValidationError{Message: "验证超时"}
	}

	return nil
}

// updateDeviceActivity 更新设备活跃时间.
func updateDeviceActivity(deviceService DeviceService, deviceInfo *DeviceInfo) {
	if err := deviceService.UpdateDeviceActivity(deviceInfo.DeviceID); err != nil {
		// 使用异步日志记录，不阻塞主流程
		go func() {
			logger.Error(ErrMsgUpdateDeviceActivityFailed, mergeMap(deviceInfo.logFields(), map[string]any{
				"error": err.Error(),
			}))
		}()
	}
}

// AuthWithDeviceValidation 认证+设备验证的组合中间件.
func AuthWithDeviceValidation(deviceService DeviceService) gin.HandlerFunc {
	return AuthWithDeviceValidationWithConfig(deviceService, DefaultDeviceValidationConfig())
}

// AuthWithDeviceValidationWithConfig 带配置的认证+设备验证组合中间件.
func AuthWithDeviceValidationWithConfig(deviceService DeviceService, config *DeviceValidationConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先进行JWT认证
		JWTAuth()(c)

		// 如果认证失败，直接返回
		if c.IsAborted() {
			return
		}

		// 然后进行设备验证
		DeviceValidationMiddlewareWithConfig(deviceService, config)(c)
	}
}

// validationResult 验证结果结构体.
type validationResult struct {
	isValid bool
	err     error
}

// DeviceValidationError 设备验证错误.
type DeviceValidationError struct {
	Message string
}

func (e *DeviceValidationError) Error() string {
	return e.Message
}

// mergeMap 合并map.
func mergeMap(map1, map2 map[string]any) map[string]any {
	result := make(map[string]any)
	for k, v := range map1 {
		result[k] = v
	}
	for k, v := range map2 {
		result[k] = v
	}
	return result
}

// SetDeviceValidationEnabled 设置设备验证开关（用于特定路由）.
func SetDeviceValidationEnabled(enabled bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(ContextKeyDeviceValidation, enabled)
		c.Next()
	}
}

// IsDeviceValidationEnabled 检查是否启用设备验证.
func IsDeviceValidationEnabled(c *gin.Context) bool {
	if enabled, exists := c.Get(ContextKeyDeviceValidation); exists {
		return enabled.(bool)
	}
	return true // 默认启用
}
