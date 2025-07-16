package service

import (
	"ai-svc/internal/middleware"
	"ai-svc/internal/model"
	"ai-svc/pkg/logger"
	"errors"
)

// JWTService JWT服务接口.
type JWTService interface {
	GenerateToken(user *model.User, device *model.UserDevice, sessionToken string) (string, error)
	ValidateToken(tokenString string) (*middleware.JWTClaims, error)
}

// jwtService JWT服务实现.
type jwtService struct{}

// NewJWTService 创建JWT服务实例.
func NewJWTService() JWTService {
	return &jwtService{}
}

// GenerateToken 生成JWT令牌.
func (s *jwtService) GenerateToken(user *model.User, device *model.UserDevice, sessionToken string) (string, error) {
	if user == nil || device == nil {
		return "", errors.New("用户或设备信息不能为空")
	}

	// 生成JWT Token
	token, err := middleware.GenerateToken(
		user.ID,
		user.Phone,
		device.DeviceID,
		device.DeviceType,
		sessionToken,
	)
	if err != nil {
		logger.Error("生成JWT Token失败", map[string]any{
			"error":       err.Error(),
			"user_id":     user.ID,
			"device_id":   device.DeviceID,
			"device_type": device.DeviceType,
		})
		return "", err
	}

	logger.Info("JWT Token生成成功", map[string]any{
		"user_id":     user.ID,
		"phone":       user.Phone,
		"device_id":   device.DeviceID,
		"device_type": device.DeviceType,
	})

	return token, nil
}

// ValidateToken 验证Token基本有效性.
func (s *jwtService) ValidateToken(tokenString string) (*middleware.JWTClaims, error) {
	return middleware.ParseToken(tokenString)
}
