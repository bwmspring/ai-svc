package service

import (
	"ai-svc/internal/model"
	"ai-svc/internal/repository"
	"ai-svc/pkg/logger"
	"crypto/md5"
	"errors"
	"fmt"
	"sort"
	"time"

	"gorm.io/gorm"
)

// DeviceService 设备管理服务接口.
type DeviceService interface {
	// 设备管理
	RegisterDevice(userID uint, deviceInfo *model.DeviceInfo, clientIP, userAgent string) (*model.UserDevice, error)
	GetUserDevices(userID uint) (*model.UserDevicesResponse, error)
	KickDevices(userID uint, deviceIDs []string) error
	UpdateDeviceActivity(deviceID string) error

	// 会话管理
	CreateSession(userID uint, deviceID, jwtToken string) (*model.UserSession, error)
	GetSessionByToken(token string) (*model.UserSession, error)
	ValidateSession(token string) (*model.UserSession, error)
	ValidateDeviceSession(userID uint, deviceID, sessionID string) (bool, error)
	UpdateSession(session *model.UserSession) error
	DeleteSession(token string) error

	// 设备限制管理
	CheckDeviceLimit(userID uint) error
	KickOldestDevice(userID uint) error
	CleanupExpiredSessions() error
	CleanupOfflineDevices() error
}

// deviceService 设备管理服务实现.
type deviceService struct {
	deviceRepo repository.DeviceRepository
	maxDevices int // 用户最大设备数量限制
}

// NewDeviceService 创建设备管理服务实例.
func NewDeviceService(deviceRepo repository.DeviceRepository) DeviceService {
	return &deviceService{
		deviceRepo: deviceRepo,
		maxDevices: 5, // 默认最大5台设备
	}
}

// NewDeviceServiceWithLimit 创建带设备限制的设备管理服务实例.
func NewDeviceServiceWithLimit(deviceRepo repository.DeviceRepository, maxDevices int) DeviceService {
	return &deviceService{
		deviceRepo: deviceRepo,
		maxDevices: maxDevices,
	}
}

// RegisterDevice 注册设备.
func (s *deviceService) RegisterDevice(
	userID uint,
	deviceInfo *model.DeviceInfo,
	clientIP, userAgent string,
) (*model.UserDevice, error) {
	// 检查是否已存在该设备
	existingDevice, err := s.deviceRepo.GetDeviceByDeviceID(deviceInfo.DeviceID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	now := time.Now()

	if existingDevice != nil {
		// 更新现有设备信息
		existingDevice.DeviceName = deviceInfo.DeviceName
		existingDevice.AppVersion = deviceInfo.AppVersion
		existingDevice.OSVersion = deviceInfo.OSVersion
		existingDevice.ClientIP = clientIP
		existingDevice.UserAgent = userAgent
		existingDevice.Status = 1
		existingDevice.LoginAt = now
		existingDevice.LastActiveAt = now

		if err := s.deviceRepo.UpdateDevice(existingDevice); err != nil {
			return nil, err
		}
		return existingDevice, nil
	}

	// 检查设备数量限制
	if err := s.CheckDeviceLimit(userID); err != nil {
		return nil, err
	}

	// 创建新设备
	device := &model.UserDevice{
		UserID:       userID,
		DeviceID:     deviceInfo.DeviceID,
		DeviceType:   deviceInfo.DeviceType,
		DeviceName:   deviceInfo.DeviceName,
		AppVersion:   deviceInfo.AppVersion,
		OSVersion:    deviceInfo.OSVersion,
		ClientIP:     clientIP,
		UserAgent:    userAgent,
		Status:       1,
		LoginAt:      now,
		LastActiveAt: now,
	}

	if err := s.deviceRepo.CreateDevice(device); err != nil {
		return nil, err
	}

	return device, nil
}

// GetUserDevices 获取用户设备列表.
func (s *deviceService) GetUserDevices(userID uint) (*model.UserDevicesResponse, error) {
	devices, err := s.deviceRepo.GetUserDevices(userID)
	if err != nil {
		return nil, err
	}

	onlineCount, err := s.deviceRepo.CountUserOnlineDevices(userID)
	if err != nil {
		logger.Error("获取用户在线设备数量失败", map[string]any{"error": err.Error()})
		onlineCount = 0
	}

	deviceResponses := make([]*model.DeviceResponse, len(devices))
	for i, device := range devices {
		deviceResponses[i] = &model.DeviceResponse{
			ID:           device.ID,
			DeviceID:     device.DeviceID,
			DeviceType:   device.DeviceType,
			DeviceName:   device.DeviceName,
			AppVersion:   device.AppVersion,
			OSVersion:    device.OSVersion,
			ClientIP:     device.ClientIP,
			Status:       device.Status,
			LoginAt:      device.LoginAt,
			LastActiveAt: device.LastActiveAt,
			IsOnline:     device.IsOnline(),
			CreatedAt:    device.CreatedAt,
		}
	}

	return &model.UserDevicesResponse{
		Devices:     deviceResponses,
		TotalCount:  len(devices),
		OnlineCount: int(onlineCount),
		MaxDevices:  s.maxDevices,
	}, nil
}

// KickDevices 踢出指定设备.
func (s *deviceService) KickDevices(userID uint, deviceIDs []string) error {
	for _, deviceID := range deviceIDs {
		// 验证设备是否属于该用户
		device, err := s.deviceRepo.GetDeviceByDeviceID(deviceID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue // 设备不存在，跳过
			}
			return err
		}

		if device.UserID != userID {
			return errors.New("无权限操作该设备")
		}

		// 删除设备的所有会话
		if err := s.deviceRepo.DeleteDeviceSessions(deviceID); err != nil {
			logger.Error("删除设备会话失败", map[string]any{"error": err.Error(), "device_id": deviceID})
		}

		// 标记设备为离线
		if err := s.deviceRepo.MarkDeviceOffline(deviceID); err != nil {
			logger.Error("标记设备离线失败", map[string]any{"error": err.Error(), "device_id": deviceID})
		}
	}

	return nil
}

// UpdateDeviceActivity 更新设备活跃时间.
func (s *deviceService) UpdateDeviceActivity(deviceID string) error {
	return s.deviceRepo.UpdateDeviceActivity(deviceID)
}

// CreateSession 创建会话.
func (s *deviceService) CreateSession(userID uint, deviceID, jwtToken string) (*model.UserSession, error) {
	// 生成会话Token
	sessionToken := s.generateSessionToken(userID, deviceID)

	session := &model.UserSession{
		UserID:       userID,
		DeviceID:     deviceID,
		SessionToken: sessionToken,
		JWTToken:     jwtToken,
		ExpiresAt:    time.Now().Add(24 * time.Hour), // 24小时过期
	}

	if err := s.deviceRepo.CreateSession(session); err != nil {
		return nil, err
	}

	return session, nil
}

// GetSessionByToken 根据Token获取会话.
func (s *deviceService) GetSessionByToken(token string) (*model.UserSession, error) {
	return s.deviceRepo.GetSessionByToken(token)
}

// ValidateSession 验证会话.
func (s *deviceService) ValidateSession(token string) (*model.UserSession, error) {
	session, err := s.deviceRepo.GetSessionByToken(token)
	if err != nil {
		return nil, err
	}

	if session.IsExpired() {
		// 清理过期会话
		s.deviceRepo.DeleteSession(token)
		return nil, errors.New("会话已过期")
	}

	return session, nil
}

// ValidateDeviceSession 验证设备会话并检查设备限制.
func (s *deviceService) ValidateDeviceSession(userID uint, deviceID, sessionID string) (bool, error) {
	// 1. 检查会话是否存在且有效
	session, err := s.deviceRepo.GetSessionByToken(sessionID)
	if err != nil {
		return false, err
	}

	if session.IsExpired() {
		// 清理过期会话
		s.deviceRepo.DeleteSession(sessionID)
		return false, errors.New("会话已过期")
	}

	// 2. 验证会话是否匹配设备和用户
	if session.DeviceID != deviceID || session.UserID != userID {
		return false, errors.New("会话不匹配")
	}

	// 3. 检查用户设备总数是否超限
	devices, err := s.deviceRepo.GetUserDevices(userID)
	if err != nil {
		return false, err
	}

	// 4. 如果设备数量超限，检查当前设备是否在允许列表中
	if len(devices) > s.maxDevices {
		// 按活跃时间排序，保留最新的设备
		sort.Slice(devices, func(i, j int) bool {
			return devices[i].LastActiveAt.After(devices[j].LastActiveAt)
		})

		// 检查当前设备是否在前N个设备中
		isAllowed := false
		for i := 0; i < s.maxDevices && i < len(devices); i++ {
			if devices[i].DeviceID == deviceID {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			// 当前设备不在允许的设备列表中，踢出
			s.deviceRepo.DeleteDeviceByDeviceID(deviceID)
			s.deviceRepo.DeleteSession(sessionID)
			return false, errors.New("设备已被踢出，请重新登录")
		}

		// 踢出超限的旧设备
		for i := s.maxDevices; i < len(devices); i++ {
			s.deviceRepo.DeleteDeviceByDeviceID(devices[i].DeviceID)
			// 删除该设备的所有会话
			s.deviceRepo.DeleteDeviceSessions(devices[i].DeviceID)
		}
	}

	return true, nil
}

// UpdateSession 更新会话.
func (s *deviceService) UpdateSession(session *model.UserSession) error {
	return s.deviceRepo.UpdateSession(session)
}

// DeleteSession 删除会话.
func (s *deviceService) DeleteSession(token string) error {
	return s.deviceRepo.DeleteSession(token)
}

// CheckDeviceLimit 检查设备数量限制.
func (s *deviceService) CheckDeviceLimit(userID uint) error {
	count, err := s.deviceRepo.CountUserDevices(userID)
	if err != nil {
		return err
	}

	if int(count) >= s.maxDevices {
		// 踢出最旧的设备
		if err := s.KickOldestDevice(userID); err != nil {
			return errors.New("设备数量已达上限，且无法踢出旧设备")
		}
	}

	return nil
}

// KickOldestDevice 踢出最旧的设备.
func (s *deviceService) KickOldestDevice(userID uint) error {
	devices, err := s.deviceRepo.GetUserDevices(userID)
	if err != nil {
		return err
	}

	if len(devices) == 0 {
		return nil
	}

	// 找到最旧的设备（按最后活跃时间排序）
	oldestDevice := devices[len(devices)-1]

	// 删除设备的所有会话
	if err := s.deviceRepo.DeleteDeviceSessions(oldestDevice.DeviceID); err != nil {
		logger.Error("删除最旧设备会话失败", map[string]any{"error": err.Error(), "device_id": oldestDevice.DeviceID})
	}

	// 删除设备记录
	if err := s.deviceRepo.DeleteDevice(oldestDevice.ID); err != nil {
		return err
	}

	logger.Info("踢出最旧设备", map[string]any{
		"user_id":     userID,
		"device_id":   oldestDevice.DeviceID,
		"device_name": oldestDevice.DeviceName,
	})

	return nil
}

// CleanupExpiredSessions 清理过期会话.
func (s *deviceService) CleanupExpiredSessions() error {
	return s.deviceRepo.CleanExpiredSessions()
}

// CleanupOfflineDevices 清理离线设备.
func (s *deviceService) CleanupOfflineDevices() error {
	return s.deviceRepo.CleanOfflineDevices()
}

// generateSessionToken 生成会话Token.
func (s *deviceService) generateSessionToken(userID uint, deviceID string) string {
	data := fmt.Sprintf("%d:%s:%d", userID, deviceID, time.Now().Unix())
	return fmt.Sprintf("%x", md5.Sum([]byte(data)))
}
