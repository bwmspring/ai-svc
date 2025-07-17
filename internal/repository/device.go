package repository

import (
	"time"

	"ai-svc/internal/model"
	"ai-svc/pkg/database"

	"gorm.io/gorm"
)

// DeviceRepository 设备仓储接口.
type DeviceRepository interface {
	// 设备管理
	CreateDevice(device *model.UserDevice) error
	GetDeviceByID(id uint) (*model.UserDevice, error)
	GetDeviceByDeviceID(deviceID string) (*model.UserDevice, error)
	GetDeviceByFingerprint(fingerprint string) (*model.UserDevice, error)
	GetUserDevices(userID uint) ([]*model.UserDevice, error)
	GetUserOnlineDevices(userID uint) ([]*model.UserDevice, error)
	UpdateDevice(device *model.UserDevice) error
	DeleteDevice(id uint) error
	DeleteDeviceByDeviceID(deviceID string) error
	CountUserDevices(userID uint) (int64, error)
	CountUserOnlineDevices(userID uint) (int64, error)

	// 设备在线状态管理
	UpdateDeviceActivity(deviceID string) error
	MarkDeviceOffline(deviceID string) error
	CleanOfflineDevices() error
}

// deviceRepository 设备仓储实现.
type deviceRepository struct {
	db *gorm.DB
}

// NewDeviceRepository 创建设备仓储实例.
func NewDeviceRepository() DeviceRepository {
	return &deviceRepository{
		db: database.GetDB(),
	}
}

// CreateDevice 创建设备.
func (r *deviceRepository) CreateDevice(device *model.UserDevice) error {
	return r.db.Create(device).Error
}

// GetDeviceByID 根据ID获取设备.
func (r *deviceRepository) GetDeviceByID(id uint) (*model.UserDevice, error) {
	var device model.UserDevice
	err := r.db.First(&device, id).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

// GetDeviceByDeviceID 根据设备ID获取设备.
func (r *deviceRepository) GetDeviceByDeviceID(deviceID string) (*model.UserDevice, error) {
	var device model.UserDevice
	err := r.db.Where("device_id = ?", deviceID).First(&device).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

// GetDeviceByFingerprint 根据设备指纹获取设备.
func (r *deviceRepository) GetDeviceByFingerprint(fingerprint string) (*model.UserDevice, error) {
	var device model.UserDevice
	err := r.db.Where("device_fingerprint = ?", fingerprint).First(&device).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

// GetUserDevices 获取用户所有设备.
func (r *deviceRepository) GetUserDevices(userID uint) ([]*model.UserDevice, error) {
	var devices []*model.UserDevice
	err := r.db.Where("user_id = ?", userID).Order("last_active_at DESC").Find(&devices).Error
	return devices, err
}

// GetUserOnlineDevices 获取用户在线设备.
func (r *deviceRepository) GetUserOnlineDevices(userID uint) ([]*model.UserDevice, error) {
	var devices []*model.UserDevice
	thirtyMinutesAgo := time.Now().Add(-30 * time.Minute)
	err := r.db.Where("user_id = ? AND status = 1 AND last_active_at > ?", userID, thirtyMinutesAgo).
		Order("last_active_at DESC").Find(&devices).Error
	return devices, err
}

// UpdateDevice 更新设备.
func (r *deviceRepository) UpdateDevice(device *model.UserDevice) error {
	return r.db.Save(device).Error
}

// DeleteDevice 删除设备.
func (r *deviceRepository) DeleteDevice(id uint) error {
	return r.db.Delete(&model.UserDevice{}, id).Error
}

// DeleteDeviceByDeviceID 根据设备ID删除设备.
func (r *deviceRepository) DeleteDeviceByDeviceID(deviceID string) error {
	return r.db.Where("device_id = ?", deviceID).Delete(&model.UserDevice{}).Error
}

// CountUserDevices 统计用户设备数量.
func (r *deviceRepository) CountUserDevices(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.UserDevice{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

// CountUserOnlineDevices 统计用户在线设备数量.
func (r *deviceRepository) CountUserOnlineDevices(userID uint) (int64, error) {
	var count int64
	thirtyMinutesAgo := time.Now().Add(-30 * time.Minute)
	err := r.db.Model(&model.UserDevice{}).
		Where("user_id = ? AND status = 1 AND last_active_at > ?", userID, thirtyMinutesAgo).
		Count(&count).Error
	return count, err
}

// UpdateDeviceActivity 更新设备活跃时间.
func (r *deviceRepository) UpdateDeviceActivity(deviceID string) error {
	return r.db.Model(&model.UserDevice{}).
		Where("device_id = ?", deviceID).
		Updates(map[string]interface{}{
			"last_active_at": time.Now(),
			"status":         1, // 设置为在线
		}).Error
}

// MarkDeviceOffline 将设备标记为离线.
func (r *deviceRepository) MarkDeviceOffline(deviceID string) error {
	return r.db.Model(&model.UserDevice{}).
		Where("device_id = ?", deviceID).
		Update("status", 0).Error
}

// CleanOfflineDevices 清理长时间离线的设备.
func (r *deviceRepository) CleanOfflineDevices() error {
	// 删除30天前最后活跃的设备
	thirtyDaysAgo := time.Now().Add(-30 * 24 * time.Hour)
	return r.db.Where("last_active_at < ? AND status = 0", thirtyDaysAgo).
		Delete(&model.UserDevice{}).Error
}
