package model

import "time"

// DeviceType 设备类型常量
const (
	DeviceTypePC          = "pc"
	DeviceTypeIOS         = "ios"
	DeviceTypeAndroid     = "android"
	DeviceTypeMiniprogram = "miniprogram"
	DeviceTypeWeb         = "web"
)

// UserDevice 用户设备模型
type UserDevice struct {
	ID                uint       `gorm:"primarykey"                                    json:"id"`
	UserID            uint       `gorm:"index;not null"                                json:"user_id"`
	DeviceID          string     `gorm:"type:varchar(64);not null;uniqueIndex"        json:"device_id"`          // 服务端生成的唯一ID
	DeviceFingerprint string     `gorm:"type:varchar(128);not null;index"             json:"device_fingerprint"` // 客户端指纹
	DeviceType        string     `gorm:"type:varchar(20);not null"                     json:"device_type"`
	DeviceName        string     `gorm:"type:varchar(100)"                             json:"device_name"`
	AppVersion        string     `gorm:"type:varchar(20)"                              json:"app_version"`
	OSVersion         string     `gorm:"type:varchar(50)"                              json:"os_version"`
	Platform          string     `gorm:"type:varchar(50)"                              json:"platform"`
	ClientIP          string     `gorm:"type:varchar(45)"                              json:"client_ip"`
	UserAgent         string     `gorm:"type:varchar(500)"                             json:"user_agent"`
	Status            int        `gorm:"type:tinyint;default:1"                        json:"status"`         // 设备状态
	SecurityLevel     int        `gorm:"type:tinyint;default:2"                        json:"security_level"` // 安全级别
	TrustScore        float32    `gorm:"type:decimal(3,2);default:1.00"                json:"trust_score"`    // 信任分数
	LoginAt           time.Time  `gorm:"not null"                                      json:"login_at"`
	LastActiveAt      time.Time  `gorm:"not null"                                      json:"last_active_at"`
	ExpiresAt         *time.Time `gorm:"index"                                        json:"expires_at"` // 设备过期时间
	CreatedAt         time.Time  `                                                     json:"created_at"`
	UpdatedAt         time.Time  `                                                     json:"updated_at"`
}

// TableName 表名
func (UserDevice) TableName() string {
	return "user_devices"
}

// IsOnline 检查设备是否在线
func (d *UserDevice) IsOnline() bool {
	return d.Status == 1 && time.Since(d.LastActiveAt) < 30*time.Minute
}

// DeviceInfo 设备信息
type DeviceInfo struct {
	DeviceID   string `json:"device_id"   validate:"required,max=100"`                              // 设备唯一标识
	DeviceType string `json:"device_type" validate:"required,oneof=pc ios android miniprogram web"` // 设备类型
	DeviceName string `json:"device_name" validate:"max=100"`                                       // 设备名称
	AppVersion string `json:"app_version" validate:"max=20"`                                        // 应用版本
	OSVersion  string `json:"os_version"  validate:"max=50"`                                        // 操作系统版本
}

// DeviceResponse 设备响应
type DeviceResponse struct {
	ID           uint      `json:"id"`
	DeviceID     string    `json:"device_id"`
	DeviceType   string    `json:"device_type"`
	DeviceName   string    `json:"device_name"`
	AppVersion   string    `json:"app_version"`
	OSVersion    string    `json:"os_version"`
	ClientIP     string    `json:"client_ip"`
	Status       int       `json:"status"`
	LoginAt      time.Time `json:"login_at"`
	LastActiveAt time.Time `json:"last_active_at"`
	IsOnline     bool      `json:"is_online"`
	IsCurrent    bool      `json:"is_current"` // 是否为当前设备
	CreatedAt    time.Time `json:"created_at"`
}

// DeviceListResponse 设备列表响应
type DeviceListResponse struct {
	Devices []DeviceResponse `json:"devices"`
	Summary DeviceSummary    `json:"summary"`
	Limits  DeviceLimits     `json:"limits"`
}

// DeviceSummary 设备摘要
type DeviceSummary struct {
	Total            int `json:"total"`
	Online           int `json:"online"`
	MobileCount      int `json:"mobile_count"` // ios + android
	PCCount          int `json:"pc_count"`
	WebCount         int `json:"web_count"`
	MiniprogramCount int `json:"miniprogram_count"`
}

// DeviceLimits 设备限制
type DeviceLimits struct {
	Mobile      int `json:"mobile"`
	PC          int `json:"pc"`
	Web         int `json:"web"`
	Miniprogram int `json:"miniprogram"`
}

// UserDevicesResponse 用户设备列表响应.
type UserDevicesResponse struct {
	Devices     []*DeviceResponse `json:"devices"`
	TotalCount  int               `json:"total_count"`
	OnlineCount int               `json:"online_count"`
	MaxDevices  int               `json:"max_devices"`
}

// KickDeviceRequest 踢出设备请求.
type KickDeviceRequest struct {
	DeviceIDs []string `json:"device_ids" validate:"required,min=1"` // 要踢出的设备ID列表
}
