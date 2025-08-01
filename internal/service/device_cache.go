package service

import (
	"ai-svc/internal/config"
	"ai-svc/internal/model"
	"ai-svc/pkg/logger"
	"fmt"
	"sync"
	"time"
)

// deviceCacheService 内存设备缓存服务实现（占位符，后续可替换为Redis）
type deviceCacheService struct {
	mutex         sync.RWMutex
	onlineDevices map[string]time.Time         // deviceID -> 在线到期时间
	userDevices   map[uint]map[string]bool     // userID -> deviceID -> true
	deviceInfo    map[string]*model.UserDevice // deviceID -> 设备信息
	config        *config.DeviceCacheConfig
}

// NewDeviceCacheService 创建设备缓存服务实例
func NewDeviceCacheService() DeviceCacheService {
	return &deviceCacheService{
		onlineDevices: make(map[string]time.Time),
		userDevices:   make(map[uint]map[string]bool),
		deviceInfo:    make(map[string]*model.UserDevice),
		config:        &config.AppConfig.Device.Cache,
	}
}

// SetDeviceOnline 设置设备在线状态
func (c *deviceCacheService) SetDeviceOnline(userID uint, deviceID string, ttl time.Duration) error {
	if !c.config.Enabled {
		return nil
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 设置设备在线状态，记录过期时间
	c.onlineDevices[deviceID] = time.Now().Add(ttl)

	// 添加到用户设备映射
	if c.userDevices[userID] == nil {
		c.userDevices[userID] = make(map[string]bool)
	}
	c.userDevices[userID][deviceID] = true

	return nil
}

// SetDeviceOffline 设置设备离线状态
func (c *deviceCacheService) SetDeviceOffline(userID uint, deviceID string) error {
	if !c.config.Enabled {
		return nil
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 移除设备在线状态
	delete(c.onlineDevices, deviceID)

	// 从用户设备映射中移除
	if userDevices, exists := c.userDevices[userID]; exists {
		delete(userDevices, deviceID)
		if len(userDevices) == 0 {
			delete(c.userDevices, userID)
		}
	}

	return nil
}

// IsDeviceOnline 检查设备是否在线
func (c *deviceCacheService) IsDeviceOnline(deviceID string) bool {
	if !c.config.Enabled {
		return false
	}

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	expireTime, exists := c.onlineDevices[deviceID]
	if !exists {
		return false
	}

	// 检查是否过期
	if time.Now().After(expireTime) {
		// 过期了，异步清理
		go func() {
			c.mutex.Lock()
			delete(c.onlineDevices, deviceID)
			c.mutex.Unlock()
		}()
		return false
	}

	return true
}

// GetUserOnlineDevices 获取用户在线设备列表
func (c *deviceCacheService) GetUserOnlineDevices(userID uint) ([]string, error) {
	if !c.config.Enabled {
		return []string{}, nil
	}

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	userDeviceMap, exists := c.userDevices[userID]
	if !exists {
		return []string{}, nil
	}

	var onlineDevices []string
	now := time.Now()

	for deviceID := range userDeviceMap {
		if expireTime, exists := c.onlineDevices[deviceID]; exists && now.Before(expireTime) {
			onlineDevices = append(onlineDevices, deviceID)
		}
	}

	return onlineDevices, nil
}

// CacheDeviceInfo 缓存设备信息
func (c *deviceCacheService) CacheDeviceInfo(device *model.UserDevice) error {
	if !c.config.Enabled {
		return nil
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 深拷贝设备信息
	deviceCopy := *device
	c.deviceInfo[device.DeviceID] = &deviceCopy

	return nil
}

// GetCachedDeviceInfo 获取缓存的设备信息
func (c *deviceCacheService) GetCachedDeviceInfo(deviceID string) (*model.UserDevice, error) {
	if !c.config.Enabled {
		return nil, fmt.Errorf("缓存未启用")
	}

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	device, exists := c.deviceInfo[deviceID]
	if !exists {
		return nil, fmt.Errorf("设备信息不存在于缓存中")
	}

	// 返回拷贝，避免外部修改
	deviceCopy := *device
	return &deviceCopy, nil
}

// RemoveDeviceInfo 移除设备信息缓存
func (c *deviceCacheService) RemoveDeviceInfo(deviceID string) error {
	if !c.config.Enabled {
		return nil
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 删除设备信息
	delete(c.deviceInfo, deviceID)
	// 删除在线状态
	delete(c.onlineDevices, deviceID)

	// 从用户设备映射中移除
	for userID, userDeviceMap := range c.userDevices {
		if _, exists := userDeviceMap[deviceID]; exists {
			delete(userDeviceMap, deviceID)
			if len(userDeviceMap) == 0 {
				delete(c.userDevices, userID)
			}
			break
		}
	}

	return nil
}

// CleanExpiredDevices 清理过期的设备缓存
func (c *deviceCacheService) CleanExpiredDevices() error {
	if !c.config.Enabled {
		return nil
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	expiredDevices := make([]string, 0)

	// 找出过期的设备
	for deviceID, expireTime := range c.onlineDevices {
		if now.After(expireTime) {
			expiredDevices = append(expiredDevices, deviceID)
		}
	}

	// 清理过期设备
	for _, deviceID := range expiredDevices {
		delete(c.onlineDevices, deviceID)
		delete(c.deviceInfo, deviceID)

		// 从用户设备映射中移除
		for userID, userDeviceMap := range c.userDevices {
			if _, exists := userDeviceMap[deviceID]; exists {
				delete(userDeviceMap, deviceID)
				if len(userDeviceMap) == 0 {
					delete(c.userDevices, userID)
				}
			}
		}
	}

	if len(expiredDevices) > 0 {
		logger.Info("清理过期设备缓存", map[string]any{
			"expired_count": len(expiredDevices),
		})
	}

	return nil
}

// GetCacheStats 获取缓存统计信息
func (c *deviceCacheService) GetCacheStats() map[string]any {
	if !c.config.Enabled {
		return map[string]any{
			"enabled": false,
		}
	}

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return map[string]any{
		"enabled":        true,
		"type":           "memory", // 内存缓存
		"online_devices": len(c.onlineDevices),
		"cached_devices": len(c.deviceInfo),
		"users_count":    len(c.userDevices),
	}
}
