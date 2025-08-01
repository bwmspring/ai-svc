package service

import (
	"ai-svc/internal/config"
	"ai-svc/internal/model"
	"ai-svc/internal/repository"
	"ai-svc/pkg/logger"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// DeviceCacheService 设备缓存服务接口
type DeviceCacheService interface {
	// 设备在线状态
	SetDeviceOnline(userID uint, deviceID string, ttl time.Duration) error
	SetDeviceOffline(userID uint, deviceID string) error
	IsDeviceOnline(deviceID string) bool
	GetUserOnlineDevices(userID uint) ([]string, error)

	// 设备信息缓存
	CacheDeviceInfo(device *model.UserDevice) error
	GetCachedDeviceInfo(deviceID string) (*model.UserDevice, error)
	RemoveDeviceInfo(deviceID string) error
}

// DeviceService 设备管理服务接口
type DeviceService interface {
	// 核心设备管理
	HandleDeviceLogin(
		userID uint,
		deviceInfo *model.DeviceRegistrationRequest,
		clientIP, userAgent string,
	) (*model.UserDevice, error)
	GetUserDevices(userID uint, currentDeviceID string) (*model.DeviceListResponse, error)
	KickDevices(userID uint, deviceIDs []string) error
	KickOtherDevices(userID uint, currentDeviceID string) error
	UpdateDeviceActivity(deviceID string) error

	// 设备查询
	GetDeviceByID(deviceID string) (*model.UserDevice, error)
	CountUserDevicesByType(userID uint, deviceType string) (int, error)

	// 清理维护
	CleanOfflineDevices() error
	StartCleanupScheduler()
	StopCleanupScheduler()
}

// deviceService 设备管理服务实现
type deviceService struct {
	deviceRepo    repository.DeviceRepository
	cacheService  DeviceCacheService
	config        *config.DeviceConfig
	cleanupTicker *time.Ticker
	cleanupStop   chan bool
}

// NewDeviceService 创建设备管理服务实例
func NewDeviceService(deviceRepo repository.DeviceRepository) DeviceService {
	return &deviceService{
		deviceRepo:   deviceRepo,
		cacheService: NewDeviceCacheService(), // 直接创建缓存服务
		config:       &config.AppConfig.Device,
		cleanupStop:  make(chan bool),
	}
}

// HandleDeviceLogin 处理设备登录
func (s *deviceService) HandleDeviceLogin(
	userID uint,
	deviceInfo *model.DeviceRegistrationRequest,
	clientIP, userAgent string,
) (*model.UserDevice, error) {
	// 1. 通过设备指纹检查设备是否已存在
	existingDevice, err := s.deviceRepo.GetDeviceByFingerprint(deviceInfo.DeviceFingerprint)
	if err == nil && existingDevice != nil {
		// 设备已存在，检查归属
		if existingDevice.UserID != userID {
			// 设备指纹被其他用户占用，这是安全问题
			logger.Error("设备指纹冲突：设备已被其他用户使用", map[string]any{
				"device_fingerprint": deviceInfo.DeviceFingerprint,
				"current_user_id":    userID,
				"existing_user_id":   existingDevice.UserID,
				"client_ip":          clientIP,
			})
			return nil, fmt.Errorf("设备已被其他用户使用，请联系技术支持")
		}

		// 设备属于当前用户，更新现有设备活跃时间
		return s.updateDeviceActivity(existingDevice, clientIP, userAgent)
	}

	// 2. 新设备需要检查数量限制
	if s.config.Kickout.Enabled {
		if err := s.checkAndHandleDeviceLimit(userID, deviceInfo.DeviceType); err != nil {
			return nil, err
		}
	}

	// 3. 注册新设备
	return s.registerNewDevice(userID, deviceInfo, clientIP, userAgent)
}

// GetUserDevices 获取用户设备列表
func (s *deviceService) GetUserDevices(userID uint, currentDeviceID string) (*model.DeviceListResponse, error) {
	devices, err := s.deviceRepo.GetUserDevices(userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户设备列表失败: %w", err)
	}

	response := &model.DeviceListResponse{
		Devices: make([]model.DeviceResponse, len(devices)),
		Summary: model.DeviceSummary{},
		Limits: model.DeviceLimits{
			Mobile:      s.config.Limits.MobileDevices,
			PC:          s.config.Limits.PCDevices,
			Web:         s.config.Limits.WebDevices,
			Miniprogram: s.config.Limits.MiniprogramDevices,
		},
	}

	// 转换设备数据并统计
	for i, device := range devices {
		isOnline := device.IsOnline()
		isCurrent := device.DeviceID == currentDeviceID

		response.Devices[i] = model.DeviceResponse{
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
			IsOnline:     isOnline,
			IsCurrent:    isCurrent,
			CreatedAt:    device.CreatedAt,
		}

		// 统计数据
		response.Summary.Total++
		if isOnline {
			response.Summary.Online++
		}

		// 按设备类型统计
		switch device.DeviceType {
		case model.DeviceTypeIOS, model.DeviceTypeAndroid:
			response.Summary.MobileCount++
		case model.DeviceTypePC:
			response.Summary.PCCount++
		case model.DeviceTypeWeb:
			response.Summary.WebCount++
		case model.DeviceTypeMiniprogram:
			response.Summary.MiniprogramCount++
		}
	}

	return response, nil
}

// KickDevices 踢出指定设备
func (s *deviceService) KickDevices(userID uint, deviceIDs []string) error {
	for _, deviceID := range deviceIDs {
		// 1. 验证设备是否属于该用户
		device, err := s.deviceRepo.GetDeviceByDeviceID(deviceID)
		if err != nil {
			logger.Warn("踢出设备时未找到设备", map[string]any{
				"user_id":   userID,
				"device_id": deviceID,
				"error":     err.Error(),
			})
			continue
		}

		if device.UserID != userID {
			logger.Warn("用户尝试踢出不属于自己的设备", map[string]any{
				"user_id":      userID,
				"device_id":    deviceID,
				"device_owner": device.UserID,
			})
			continue
		}

		// 2. 从缓存中移除设备
		if s.cacheService != nil {
			s.cacheService.SetDeviceOffline(userID, deviceID)
		}

		// 3. 从数据库删除设备
		if err := s.deviceRepo.DeleteDeviceByDeviceID(deviceID); err != nil {
			logger.Error("删除设备失败", map[string]any{
				"user_id":   userID,
				"device_id": deviceID,
				"error":     err.Error(),
			})
			return fmt.Errorf("删除设备 %s 失败: %w", deviceID, err)
		}

		logger.Info("设备已被踢出", map[string]any{
			"user_id":     userID,
			"device_id":   deviceID,
			"device_type": device.DeviceType,
		})
	}

	return nil
}

// KickOtherDevices 踢出其他所有设备（保留当前设备）
func (s *deviceService) KickOtherDevices(userID uint, currentDeviceID string) error {
	devices, err := s.deviceRepo.GetUserDevices(userID)
	if err != nil {
		return fmt.Errorf("获取用户设备列表失败: %w", err)
	}

	var deviceIDs []string
	for _, device := range devices {
		if device.DeviceID != currentDeviceID {
			deviceIDs = append(deviceIDs, device.DeviceID)
		}
	}

	if len(deviceIDs) > 0 {
		return s.KickDevices(userID, deviceIDs)
	}

	return nil
}

// UpdateDeviceActivity 更新设备活跃时间
func (s *deviceService) UpdateDeviceActivity(deviceID string) error {
	// 1. 更新缓存
	if s.cacheService != nil {
		// 这里使用较短的TTL来表示设备在线状态
		ttl := time.Duration(s.config.Activity.OnlineTimeoutMinutes) * time.Minute
		if err := s.cacheService.SetDeviceOnline(0, deviceID, ttl); err != nil {
			logger.Warn("更新设备缓存活跃状态失败", map[string]any{
				"device_id": deviceID,
				"error":     err.Error(),
			})
		}
	}

	// 2. 更新数据库
	return s.deviceRepo.UpdateDeviceActivity(deviceID)
}

// GetDeviceByID 根据设备ID获取设备信息
func (s *deviceService) GetDeviceByID(deviceID string) (*model.UserDevice, error) {
	return s.deviceRepo.GetDeviceByDeviceID(deviceID)
}

// CountUserDevicesByType 统计用户指定类型的设备数量
func (s *deviceService) CountUserDevicesByType(userID uint, deviceType string) (int, error) {
	devices, err := s.deviceRepo.GetUserDevices(userID)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, device := range devices {
		// 对于移动设备类型，包括iOS和Android
		if s.IsMobileDevice(deviceType) && s.IsMobileDevice(device.DeviceType) {
			count++
		} else if device.DeviceType == deviceType {
			count++
		}
	}

	return count, nil
}

// CleanOfflineDevices 清理离线设备
func (s *deviceService) CleanOfflineDevices() error {
	return s.deviceRepo.CleanOfflineDevices()
}

// StartCleanupScheduler 启动清理调度器
func (s *deviceService) StartCleanupScheduler() {
	interval := time.Duration(s.config.Activity.CleanupIntervalHours) * time.Hour
	s.cleanupTicker = time.NewTicker(interval)

	go func() {
		for {
			select {
			case <-s.cleanupTicker.C:
				if err := s.CleanOfflineDevices(); err != nil {
					logger.Error("清理离线设备失败", map[string]any{
						"error": err.Error(),
					})
				}
			case <-s.cleanupStop:
				return
			}
		}
	}()

	logger.Info("设备清理调度器已启动", map[string]any{
		"interval_hours": s.config.Activity.CleanupIntervalHours,
	})
}

// StopCleanupScheduler 停止清理调度器
func (s *deviceService) StopCleanupScheduler() {
	if s.cleanupTicker != nil {
		s.cleanupTicker.Stop()
	}

	select {
	case s.cleanupStop <- true:
	default:
	}

	logger.Info("设备清理调度器已停止", map[string]any{})
}

// 私有方法

// getDeviceLimit 根据设备类型获取限制数量
func (s *deviceService) getDeviceLimit(deviceType string) int {
	switch deviceType {
	case model.DeviceTypeIOS, model.DeviceTypeAndroid:
		return s.config.Limits.MobileDevices // 10台
	case model.DeviceTypePC:
		return s.config.Limits.PCDevices // 3台
	case model.DeviceTypeMiniprogram:
		return s.config.Limits.MiniprogramDevices // 5台
	case model.DeviceTypeWeb:
		return s.config.Limits.WebDevices // 3台
	default:
		return 5 // 默认限制
	}
}

// IsMobileDevice 判断是否为移动设备
func (s *deviceService) IsMobileDevice(deviceType string) bool {
	return deviceType == model.DeviceTypeIOS || deviceType == model.DeviceTypeAndroid
}

// checkAndHandleDeviceLimit 检查并处理设备数量限制
func (s *deviceService) checkAndHandleDeviceLimit(userID uint, deviceType string) error {
	limit := s.getDeviceLimit(deviceType)

	// 统计同类型设备数量
	count, err := s.CountUserDevicesByType(userID, deviceType)
	if err != nil {
		return err
	}

	// 如果超过限制，踢出最旧设备
	if count >= limit {
		return s.kickOldestDeviceByType(userID, deviceType)
	}

	return nil
}

// kickOldestDeviceByType 踢出指定类型的最旧设备
func (s *deviceService) kickOldestDeviceByType(userID uint, deviceType string) error {
	devices, err := s.deviceRepo.GetUserDevices(userID)
	if err != nil {
		return err
	}

	// 过滤出指定类型的设备
	var targetDevices []*model.UserDevice
	for _, device := range devices {
		if s.IsMobileDevice(deviceType) && s.IsMobileDevice(device.DeviceType) {
			targetDevices = append(targetDevices, device)
		} else if device.DeviceType == deviceType {
			targetDevices = append(targetDevices, device)
		}
	}

	if len(targetDevices) == 0 {
		return nil
	}

	// 找到最旧的设备（已经按last_active_at DESC排序，所以最后一个是最旧的）
	oldestDevice := targetDevices[len(targetDevices)-1]

	// 踢出设备
	return s.KickDevices(userID, []string{oldestDevice.DeviceID})
}

// updateDeviceActivity 更新现有设备活跃时间
func (s *deviceService) updateDeviceActivity(
	device *model.UserDevice,
	clientIP, userAgent string,
) (*model.UserDevice, error) {
	now := time.Now()
	device.LastActiveAt = now
	device.ClientIP = clientIP
	device.UserAgent = userAgent
	device.Status = 1 // 设置为在线

	if err := s.deviceRepo.UpdateDevice(device); err != nil {
		return nil, fmt.Errorf("更新设备活跃时间失败: %w", err)
	}

	// 更新缓存
	if s.cacheService != nil {
		ttl := time.Duration(s.config.Activity.OnlineTimeoutMinutes) * time.Minute
		s.cacheService.SetDeviceOnline(device.UserID, device.DeviceID, ttl)
	}

	logger.Info("设备活跃时间已更新", map[string]any{
		"user_id":   device.UserID,
		"device_id": device.DeviceID,
		"client_ip": clientIP,
	})

	return device, nil
}

// generateDeviceID 生成设备ID
func generateDeviceID(deviceType string) (string, error) {
	// 根据设备类型确定前缀
	var prefix string
	switch deviceType {
	case "pc":
		prefix = "pc_"
	case "ios":
		prefix = "ios_"
	case "android":
		prefix = "and_"
	case "web":
		prefix = "web_"
	case "miniprogram":
		prefix = "mp_"
	default:
		prefix = "dev_"
	}

	// 生成16字节随机数
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	// 转换为十六进制字符串
	randomStr := hex.EncodeToString(randomBytes)

	return prefix + randomStr, nil
}

// registerNewDevice 注册新设备
func (s *deviceService) registerNewDevice(
	userID uint,
	deviceInfo *model.DeviceRegistrationRequest,
	clientIP, userAgent string,
) (*model.UserDevice, error) {
	// 生成设备ID
	deviceID, err := generateDeviceID(deviceInfo.DeviceType)
	if err != nil {
		return nil, fmt.Errorf("生成设备ID失败: %w", err)
	}

	now := time.Now()
	device := &model.UserDevice{
		UserID:            userID,
		DeviceID:          deviceID,
		DeviceFingerprint: deviceInfo.DeviceFingerprint,
		DeviceType:        deviceInfo.DeviceType,
		DeviceName:        deviceInfo.DeviceName,
		AppVersion:        deviceInfo.AppVersion,
		OSVersion:         deviceInfo.OSVersion,
		Platform:          deviceInfo.Platform,
		ClientIP:          clientIP,
		UserAgent:         userAgent,
		Status:            1,
		LoginAt:           now,
		LastActiveAt:      now,
	}

	// 数据库事务保护
	err = s.deviceRepo.CreateDevice(device)
	if err != nil {
		return nil, fmt.Errorf("创建设备失败: %w", err)
	}

	// 更新缓存
	if s.cacheService != nil {
		ttl := time.Duration(s.config.Activity.OnlineTimeoutMinutes) * time.Minute
		s.cacheService.SetDeviceOnline(userID, device.DeviceID, ttl)
	}

	logger.Info("新设备已注册", map[string]any{
		"user_id":            userID,
		"device_id":          deviceID,
		"device_fingerprint": deviceInfo.DeviceFingerprint,
		"device_type":        deviceInfo.DeviceType,
		"device_name":        deviceInfo.DeviceName,
		"client_ip":          clientIP,
	})

	return device, nil
}
