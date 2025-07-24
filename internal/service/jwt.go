package service

import (
	"errors"

	"ai-svc/internal/middleware"
	"ai-svc/internal/model"
	"ai-svc/pkg/logger"
)

// JWTService JWT服务接口.
type JWTService interface {
	GenerateToken(user *model.User, device *model.UserDevice, deviceID string) (string, error)
	ValidateToken(tokenString string) (*middleware.JWTClaims, error)
	GenerateRefreshToken(user *model.User, device *model.UserDevice) (string, error)
	RefreshToken(refreshToken string) (*model.TokenPair, error)
	ValidateRefreshToken(refreshToken string) (*middleware.JWTClaims, error)
}

// jwtService JWT服务实现.
type jwtService struct {
	deviceService DeviceService // 添加设备服务依赖
}

// NewJWTService 创建JWT服务实例.
func NewJWTService() JWTService {
	return &jwtService{}
}

// NewJWTServiceWithDeviceService 创建带设备服务的JWT服务实例.
func NewJWTServiceWithDeviceService(deviceService DeviceService) JWTService {
	return &jwtService{
		deviceService: deviceService,
	}
}

// GenerateToken 生成JWT令牌.
func (s *jwtService) GenerateToken(user *model.User, device *model.UserDevice, deviceID string) (string, error) {
	if user == nil || device == nil {
		return "", errors.New("用户或设备信息不能为空")
	}

	// 生成JWT Token
	token, err := middleware.GenerateToken(
		user.ID,
		user.Phone,
		device.DeviceID,
		device.DeviceType,
		deviceID,
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

// GenerateRefreshToken 生成刷新令牌.
func (s *jwtService) GenerateRefreshToken(user *model.User, device *model.UserDevice) (string, error) {
	if user == nil || device == nil {
		return "", errors.New("用户或设备信息不能为空")
	}

	// 生成长期有效的refresh token (7天)
	refreshToken, err := middleware.GenerateRefreshToken(
		user.ID,
		user.Phone,
		device.DeviceID,
		device.DeviceType,
	)
	if err != nil {
		logger.Error("生成Refresh Token失败", map[string]any{
			"error":     err.Error(),
			"user_id":   user.ID,
			"device_id": device.DeviceID,
		})
		return "", err
	}

	logger.Info("Refresh Token生成成功", map[string]any{
		"user_id":   user.ID,
		"device_id": device.DeviceID,
	})

	return refreshToken, nil
}

// RefreshToken 使用refresh token生成新的access token
func (s *jwtService) RefreshToken(refreshToken string) (*model.TokenPair, error) {
	// 1. 验证refresh token
	claims, err := s.ValidateRefreshToken(refreshToken)
	if err != nil {
		logger.Warn("Refresh Token验证失败", map[string]any{
			"error": err.Error(),
		})
		return nil, errors.New("refresh token无效")
	}

	// 2. **关键安全检查：验证设备是否仍然有效**
	if s.deviceService != nil {
		// 检查设备是否存在且有效
		device, err := s.deviceService.GetDeviceByID(claims.DeviceID)
		if err != nil {
			logger.Warn("Refresh Token失败：设备不存在或已被踢出", map[string]any{
				"user_id":   claims.UserID,
				"device_id": claims.DeviceID,
				"error":     err.Error(),
			})
			return nil, errors.New("设备已被踢出，请重新登录")
		}

		// 检查设备是否属于token中的用户
		if device.UserID != claims.UserID {
			logger.Warn("Refresh Token失败：设备归属验证失败", map[string]any{
				"token_user_id":  claims.UserID,
				"device_user_id": device.UserID,
				"device_id":      claims.DeviceID,
			})
			return nil, errors.New("设备归属验证失败")
		}

		// 检查设备是否在线
		if !device.IsOnline() {
			logger.Warn("Refresh Token失败：设备已离线", map[string]any{
				"user_id":     claims.UserID,
				"device_id":   claims.DeviceID,
				"last_active": device.LastActiveAt,
			})
			return nil, errors.New("设备已离线，请重新登录")
		}

		logger.Info("Refresh Token设备验证通过", map[string]any{
			"user_id":   claims.UserID,
			"device_id": claims.DeviceID,
		})
	} else {
		logger.Warn("设备服务未初始化，跳过设备验证", map[string]any{
			"user_id":   claims.UserID,
			"device_id": claims.DeviceID,
		})
	}

	// 3. 生成新的access token
	newAccessToken, err := middleware.GenerateToken(
		claims.UserID,
		claims.Phone,
		claims.DeviceID,
		claims.DeviceType,
		"", // session_id可以为空或重新生成
	)
	if err != nil {
		logger.Error("生成新Access Token失败", map[string]any{
			"error": err.Error(),
		})
		return nil, errors.New("生成新的access token失败")
	}

	// 4. 生成新的refresh token（轮换机制，提高安全性）
	newRefreshToken, err := middleware.GenerateRefreshToken(
		claims.UserID,
		claims.Phone,
		claims.DeviceID,
		claims.DeviceType,
	)
	if err != nil {
		logger.Error("生成新Refresh Token失败", map[string]any{
			"error": err.Error(),
		})
		return nil, errors.New("生成新的refresh token失败")
	}

	logger.Info("Token刷新成功", map[string]any{
		"user_id":   claims.UserID,
		"device_id": claims.DeviceID,
	})

	return &model.TokenPair{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    24 * 3600, // 24小时，单位秒
		TokenType:    "Bearer",
	}, nil
}

// ValidateRefreshToken 验证刷新令牌.
func (s *jwtService) ValidateRefreshToken(refreshToken string) (*middleware.JWTClaims, error) {
	return middleware.ParseRefreshToken(refreshToken)
}
