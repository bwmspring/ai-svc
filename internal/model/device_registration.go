package model

import (
	"time"
)

// DeviceRegistrationRequest 设备注册请求
type DeviceRegistrationRequest struct {
	DeviceFingerprint string `json:"device_fingerprint" validate:"required,min=16,max=128"` // 设备指纹（客户端生成）
	DeviceType        string `json:"device_type"        validate:"required,oneof=pc ios android miniprogram web"`
	DeviceName        string `json:"device_name"        validate:"max=100"` // 设备名称
	AppVersion        string `json:"app_version"        validate:"max=20"`  // 应用版本
	OSVersion         string `json:"os_version"         validate:"max=50"`  // 操作系统版本
	Platform          string `json:"platform"           validate:"max=50"`  // 平台信息
	ClientInfo        string `json:"client_info"        validate:"max=200"` // 客户端额外信息
}

// DeviceRegistrationResponse 设备注册响应
type DeviceRegistrationResponse struct {
	DeviceID    string    `json:"device_id"`    // 服务端生成的唯一设备ID
	DeviceToken string    `json:"device_token"` // 设备认证token（可选）
	ExpiresAt   time.Time `json:"expires_at"`   // 设备ID过期时间（可选）
	IsNew       bool      `json:"is_new"`       // 是否为新注册设备
}

// DeviceFingerprint 设备指纹信息（用于检测重复注册）
type DeviceFingerprint struct {
	ID                uint      `gorm:"primarykey"                                    json:"id"`
	Fingerprint       string    `gorm:"type:varchar(128);not null;uniqueIndex"       json:"fingerprint"`         // 设备指纹（唯一）
	DeviceID          string    `gorm:"type:varchar(64);not null;index"              json:"device_id"`           // 关联的设备ID
	UserID            uint      `gorm:"index;not null"                                json:"user_id"`            // 用户ID
	FirstSeenAt       time.Time `gorm:"not null"                                      json:"first_seen_at"`      // 首次见到时间
	LastSeenAt        time.Time `gorm:"not null"                                      json:"last_seen_at"`       // 最后见到时间
	RegistrationCount int       `gorm:"default:1"                                     json:"registration_count"` // 注册次数（检测异常）
	CreatedAt         time.Time `                                                     json:"created_at"`
	UpdatedAt         time.Time `                                                     json:"updated_at"`
}

// TableName 表名
func (DeviceFingerprint) TableName() string {
	return "device_fingerprints"
}

// DeviceIDGenerator 设备ID生成器配置
type DeviceIDGenerator struct {
	Prefix    string `json:"prefix"`    // ID前缀（如 "dev_", "mob_", "web_"）
	Length    int    `json:"length"`    // ID长度
	Timestamp bool   `json:"timestamp"` // 是否包含时间戳
	Random    bool   `json:"random"`    // 是否包含随机数
	Checksum  bool   `json:"checksum"`  // 是否包含校验和
}

// Enhanced DeviceInfo with server-generated ID
type DeviceInfoV2 struct {
	DeviceID          string `json:"device_id"`          // 服务端生成的设备ID
	DeviceFingerprint string `json:"device_fingerprint"` // 客户端生成的设备指纹
	DeviceType        string `json:"device_type"`
	DeviceName        string `json:"device_name"`
	AppVersion        string `json:"app_version"`
	OSVersion         string `json:"os_version"`
	Platform          string `json:"platform"`
	IsRegistered      bool   `json:"is_registered"` // 设备是否已注册
}

// LoginWithDeviceRequest 带设备注册的登录请求
type LoginWithDeviceRequest struct {
	Phone            string                     `json:"phone"        validate:"required,len=11"`
	Code             string                     `json:"code"         validate:"required,len=6"`
	DeviceInfo       *DeviceRegistrationRequest `json:"device_info"  validate:"required"` // 设备注册信息
	ExistingDeviceID string                     `json:"existing_device_id,omitempty"`     // 已有设备ID（可选）
}

// DeviceStatus 设备状态枚举
const (
	DeviceStatusActive    = 1 // 活跃
	DeviceStatusInactive  = 0 // 非活跃
	DeviceStatusSuspended = 2 // 暂停
	DeviceStatusBanned    = 3 // 封禁
)

// DeviceSecurityLevel 设备安全级别
const (
	SecurityLevelLow    = 1 // 低安全级别
	SecurityLevelNormal = 2 // 普通安全级别
	SecurityLevelHigh   = 3 // 高安全级别
)
