package middleware

import (
	"ai-svc/internal/config"
	"ai-svc/pkg/logger"
	"ai-svc/pkg/response"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims JWT载荷.
type JWTClaims struct {
	UserID     uint   `json:"user_id"`
	Phone      string `json:"phone"`
	DeviceID   string `json:"device_id"`
	DeviceType string `json:"device_type"`
	SessionID  string `json:"session_id,omitempty"` // 可选的会话ID
	jwt.RegisteredClaims
}

// JWTAuth 基础JWT认证中间件.
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := GetRequestID(c)

		// 1. 提取Token
		token := c.GetHeader("Authorization")
		if token == "" {
			logger.Warn("JWT认证失败：未提供认证令牌", map[string]any{
				"request_id": requestID,
				"path":       c.Request.URL.Path,
			})
			response.Error(c, response.UNAUTHORIZED, "未提供认证令牌")
			c.Abort()
			return
		}

		// 2. 检查Token格式
		if !strings.HasPrefix(token, "Bearer ") {
			logger.Warn("JWT认证失败：令牌格式错误", map[string]any{
				"request_id": requestID,
			})
			response.Error(c, response.UNAUTHORIZED, "认证令牌格式错误")
			c.Abort()
			return
		}

		// 3. 解析Token
		tokenString := strings.TrimPrefix(token, "Bearer ")
		claims, err := ParseToken(tokenString)
		if err != nil {
			logger.Warn("JWT认证失败：令牌解析失败", map[string]any{
				"request_id": requestID,
				"error":      err.Error(),
			})
			response.Error(c, response.UNAUTHORIZED, "认证令牌无效")
			c.Abort()
			return
		}

		// 4. 检查Token是否过期
		if time.Now().Unix() > claims.ExpiresAt.Unix() {
			logger.Warn("JWT认证失败：令牌已过期", map[string]any{
				"request_id": requestID,
				"user_id":    claims.UserID,
				"device_id":  claims.DeviceID,
				"expired_at": claims.ExpiresAt.Unix(),
			})
			response.Error(c, response.UNAUTHORIZED, "认证令牌已过期")
			c.Abort()
			return
		}

		// 5. 将用户信息存入上下文
		c.Set("user_id", claims.UserID)
		c.Set("phone", claims.Phone)
		c.Set("device_id", claims.DeviceID)
		c.Set("device_type", claims.DeviceType)
		c.Set("session_id", claims.SessionID)

		// 记录成功认证日志
		logger.Info("JWT认证成功", map[string]any{
			"request_id":  requestID,
			"user_id":     claims.UserID,
			"phone":       claims.Phone,
			"device_id":   claims.DeviceID,
			"device_type": claims.DeviceType,
			"path":        c.Request.URL.Path,
		})

		c.Next()
	}
}

// GenerateToken 生成JWT令牌.
func GenerateToken(userID uint, phone, deviceID, deviceType, sessionID string) (string, error) {
	expireTime := time.Now().Add(time.Duration(config.AppConfig.JWT.ExpireTime) * time.Second)

	claims := JWTClaims{
		UserID:     userID,
		Phone:      phone,
		DeviceID:   deviceID,
		DeviceType: deviceType,
		SessionID:  sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "ai-svc",
			Subject:   phone,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.AppConfig.JWT.Secret))
}

// ParseToken 解析JWT令牌.
func ParseToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (any, error) {
		return []byte(config.AppConfig.JWT.Secret), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrInvalidKey
}

// GetCurrentUserID 获取当前用户ID.
func GetCurrentUserID(c *gin.Context) uint {
	if userID, exists := c.Get("user_id"); exists {
		return userID.(uint)
	}
	return 0
}

// GetCurrentPhone 获取当前用户手机号.
func GetCurrentPhone(c *gin.Context) string {
	if phone, exists := c.Get("phone"); exists {
		return phone.(string)
	}
	return ""
}

// GetCurrentDeviceID 获取当前设备ID.
func GetCurrentDeviceID(c *gin.Context) string {
	if deviceID, exists := c.Get("device_id"); exists {
		return deviceID.(string)
	}
	return ""
}

// GetCurrentDeviceType 获取当前设备类型.
func GetCurrentDeviceType(c *gin.Context) string {
	if deviceType, exists := c.Get("device_type"); exists {
		return deviceType.(string)
	}
	return ""
}

// GetCurrentSessionID 获取当前会话ID.
func GetCurrentSessionID(c *gin.Context) string {
	if sessionID, exists := c.Get("session_id"); exists {
		return sessionID.(string)
	}
	return ""
}

// IsValidUser 检查当前用户是否有效.
func IsValidUser(c *gin.Context) bool {
	return GetCurrentUserID(c) > 0
}

// RequireValidUser 要求有效用户的中间件.
func RequireValidUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !IsValidUser(c) {
			response.Error(c, response.UNAUTHORIZED, "无效用户")
			c.Abort()
			return
		}
		c.Next()
	}
}

// DeviceTypeMiddleware 设备类型验证中间件.
func DeviceTypeMiddleware(allowedTypes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		currentDeviceType := GetCurrentDeviceType(c)

		// 检查设备类型是否在允许列表中
		allowed := false
		for _, deviceType := range allowedTypes {
			if currentDeviceType == deviceType {
				allowed = true
				break
			}
		}

		if !allowed {
			requestID := GetRequestID(c)
			logger.Warn("设备类型验证失败", map[string]any{
				"request_id":    requestID,
				"current_type":  currentDeviceType,
				"allowed_types": allowedTypes,
				"user_id":       GetCurrentUserID(c),
			})
			response.Error(c, response.FORBIDDEN, "设备类型不被允许")
			c.Abort()
			return
		}

		c.Next()
	}
}
