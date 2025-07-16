package model

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel 基础模型.
type BaseModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `                  json:"created_at"`
	UpdatedAt time.Time      `                  json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"      json:"-"`
}

// User 用户模型 - 采用手机号+验证码登录模式.
type User struct {
	BaseModel
	// 基础认证信息（手机号作为主要登录凭证）
	Phone    string `gorm:"type:varchar(20);uniqueIndex;not null" json:"phone"    validate:"required,min=11,max=20"`
	Username string `gorm:"type:varchar(50);uniqueIndex"          json:"username" validate:"min=3,max=50"`
	Email    string `gorm:"type:varchar(100);uniqueIndex"         json:"email"    validate:"email"`

	// 用户资料信息（常用查询字段放在主表）
	Nickname string `gorm:"type:varchar(50)"                               json:"nickname"`
	Avatar   string `gorm:"type:varchar(255)"                              json:"avatar"`
	RealName string `gorm:"type:varchar(50)"                               json:"real_name"`
	Gender   int    `gorm:"type:tinyint;default:0;comment:性别 0:未知 1:男 2:女" json:"gender"`

	// VIP 相关字段
	VIPLevel    int        `gorm:"type:tinyint;default:0;comment:VIP等级 0:普通用户 1:VIP1 2:VIP2 3:VIP3" json:"vip_level"`
	VIPExpireAt *time.Time `gorm:"comment:VIP过期时间"                                                  json:"vip_expire_at"`

	// 状态和行为字段
	Status      int        `gorm:"type:tinyint;default:1;comment:状态 1:正常 0:禁用 -1:删除" json:"status"`
	LastLoginIP string     `gorm:"type:varchar(45)"                                  json:"last_login_ip"`
	LastLoginAt *time.Time `gorm:"comment:最后登录时间"                                    json:"last_login_at"`
	LoginCount  int        `gorm:"type:int;default:0;comment:登录次数"                   json:"login_count"`

	// 扩展字段（不常用的详细信息）
	Birthday *time.Time `json:"birthday"`
	Address  string     `json:"address"  gorm:"type:varchar(255)"`
	Bio      string     `json:"bio"      gorm:"type:text"`
}

// TableName 表名.
func (User) TableName() string {
	return "users"
}

// IsVIP 检查是否为VIP用户.
func (u *User) IsVIP() bool {
	return u.VIPLevel > 0 && (u.VIPExpireAt == nil || u.VIPExpireAt.After(time.Now()))
}

// IsActive 检查用户是否激活.
func (u *User) IsActive() bool {
	return u.Status == 1
}

// UserBehaviorLog 用户行为日志（频繁变更的数据独立存储）.
type UserBehaviorLog struct {
	ID        uint      `gorm:"primarykey"                json:"id"`
	UserID    uint      `gorm:"index;not null"            json:"user_id"`
	Action    string    `gorm:"type:varchar(50);not null" json:"action"`   // login, logout, view, purchase等
	Resource  string    `gorm:"type:varchar(100)"         json:"resource"` // 操作的资源
	IP        string    `gorm:"type:varchar(45)"          json:"ip"`
	UserAgent string    `gorm:"type:varchar(500)"         json:"user_agent"`
	CreatedAt time.Time `                                 json:"created_at"`
}

// TableName 表名.
func (UserBehaviorLog) TableName() string {
	return "user_behavior_logs"
}

// SendSMSRequest 发送短信验证码请求.
type SendSMSRequest struct {
	Phone   string `json:"phone"   validate:"required,min=11,max=20"`
	Purpose string `json:"purpose" validate:"required,oneof=login register reset"` // login, register, reset
}

// LoginWithSMSRequest 手机号+验证码登录请求（扩展设备信息）.
type LoginWithSMSRequest struct {
	Phone      string      `json:"phone"       validate:"required,min=11,max=20"`
	Code       string      `json:"code"        validate:"required,min=4,max=10"`
	DeviceInfo *DeviceInfo `json:"device_info" validate:"required"` // 设备信息
}

// UpdateUserRequest 更新用户请求.
type UpdateUserRequest struct {
	Username string     `json:"username"  validate:"min=3,max=50"`
	Email    string     `json:"email"     validate:"email"`
	Nickname string     `json:"nickname"  validate:"max=50"`
	Avatar   string     `json:"avatar"    validate:"url"`
	RealName string     `json:"real_name" validate:"max=50"`
	Gender   int        `json:"gender"    validate:"min=0,max=2"`
	Birthday *time.Time `json:"birthday"`
	Address  string     `json:"address"   validate:"max=255"`
	Bio      string     `json:"bio"       validate:"max=1000"`
}

// UserResponse 用户响应.
type UserResponse struct {
	ID          uint       `json:"id"`
	Phone       string     `json:"phone"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	Nickname    string     `json:"nickname"`
	Avatar      string     `json:"avatar"`
	RealName    string     `json:"real_name"`
	Gender      int        `json:"gender"`
	VIPLevel    int        `json:"vip_level"`
	VIPExpireAt *time.Time `json:"vip_expire_at"`
	Status      int        `json:"status"`
	LastLoginIP string     `json:"last_login_ip"`
	LastLoginAt *time.Time `json:"last_login_at"`
	LoginCount  int        `json:"login_count"`
	Birthday    *time.Time `json:"birthday"`
	Address     string     `json:"address"`
	Bio         string     `json:"bio"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// UserListResponse 用户列表响应（简化版，用于列表页面）.
type UserListResponse struct {
	ID          uint       `json:"id"`
	Phone       string     `json:"phone"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	Nickname    string     `json:"nickname"`
	Avatar      string     `json:"avatar"`
	VIPLevel    int        `json:"vip_level"`
	Status      int        `json:"status"`
	LastLoginAt *time.Time `json:"last_login_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

// LoginResponse 登录响应.
type LoginResponse struct {
	User  *UserResponse `json:"user"`
	Token string        `json:"token"`
}

// SMSVerificationCode 短信验证码模型.
type SMSVerificationCode struct {
	ID        uint       `gorm:"primarykey"                      json:"id"`
	Phone     string     `gorm:"type:varchar(20);not null;index" json:"phone"      validate:"required"`
	Code      string     `gorm:"type:varchar(10);not null"       json:"-"`
	Purpose   string     `gorm:"type:varchar(20);not null"       json:"purpose"`   // login, register, reset等
	ClientIP  string     `gorm:"type:varchar(45);not null"       json:"client_ip"` // 客户端IP地址
	UsedAt    *time.Time `gorm:"comment:使用时间"                    json:"used_at"`
	ExpiredAt time.Time  `gorm:"not null"                        json:"expired_at"`
	CreatedAt time.Time  `                                       json:"created_at"`
}

// TableName 表名.
func (SMSVerificationCode) TableName() string {
	return "sms_verification_codes"
}

// IsValid 检查验证码是否有效.
func (s *SMSVerificationCode) IsValid() bool {
	return s.UsedAt == nil && s.ExpiredAt.After(time.Now())
}

// MarkAsUsed 标记验证码为已使用.
func (s *SMSVerificationCode) MarkAsUsed() {
	now := time.Now()
	s.UsedAt = &now
}

// UserDevice 用户设备模型.
type UserDevice struct {
	ID           uint      `gorm:"primarykey"                                  json:"id"`
	UserID       uint      `gorm:"index;not null"                              json:"user_id"`
	DeviceID     string    `gorm:"type:varchar(100);not null;index"            json:"device_id"`   // 设备唯一标识
	DeviceType   string    `gorm:"type:varchar(20);not null"                   json:"device_type"` // pc, ios, android, miniprogram
	DeviceName   string    `gorm:"type:varchar(100)"                           json:"device_name"` // 设备名称
	AppVersion   string    `gorm:"type:varchar(20)"                            json:"app_version"` // 应用版本
	OSVersion    string    `gorm:"type:varchar(50)"                            json:"os_version"`  // 操作系统版本
	ClientIP     string    `gorm:"type:varchar(45)"                            json:"client_ip"`   // 客户端IP
	UserAgent    string    `gorm:"type:varchar(500)"                           json:"user_agent"`  // 用户代理
	Status       int       `gorm:"type:tinyint;default:1;comment:状态 1:在线 0:离线" json:"status"`
	LoginAt      time.Time `gorm:"not null"                                    json:"login_at"`       // 登录时间
	LastActiveAt time.Time `gorm:"not null"                                    json:"last_active_at"` // 最后活跃时间
	CreatedAt    time.Time `                                                   json:"created_at"`
	UpdatedAt    time.Time `                                                   json:"updated_at"`
}

// TableName 表名.
func (UserDevice) TableName() string {
	return "user_devices"
}

// IsOnline 检查设备是否在线.
func (d *UserDevice) IsOnline() bool {
	return d.Status == 1 && time.Since(d.LastActiveAt) < 30*time.Minute
}

// DeviceType 设备类型常量.
const (
	DeviceTypePC          = "pc"
	DeviceTypeIOS         = "ios"
	DeviceTypeAndroid     = "android"
	DeviceTypeMiniprogram = "miniprogram"
	DeviceTypeWeb         = "web"
)

// UserSession 用户会话模型.
type UserSession struct {
	ID           uint      `gorm:"primarykey"                             json:"id"`
	UserID       uint      `gorm:"index;not null"                         json:"user_id"`
	DeviceID     string    `gorm:"type:varchar(100);not null;index"       json:"device_id"`
	SessionToken string    `gorm:"type:varchar(255);not null;uniqueIndex" json:"session_token"`
	JWTToken     string    `gorm:"type:text;not null"                     json:"jwt_token"`
	ExpiresAt    time.Time `gorm:"not null"                               json:"expires_at"`
	CreatedAt    time.Time `                                              json:"created_at"`
	UpdatedAt    time.Time `                                              json:"updated_at"`
}

// TableName 表名.
func (UserSession) TableName() string {
	return "user_sessions"
}

// IsExpired 检查会话是否过期.
func (s *UserSession) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// DeviceInfo 设备信息请求.
type DeviceInfo struct {
	DeviceID   string `json:"device_id"   validate:"required,max=100"`                              // 设备唯一标识
	DeviceType string `json:"device_type" validate:"required,oneof=pc ios android miniprogram web"` // 设备类型
	DeviceName string `json:"device_name" validate:"max=100"`                                       // 设备名称
	AppVersion string `json:"app_version" validate:"max=20"`                                        // 应用版本
	OSVersion  string `json:"os_version"  validate:"max=50"`                                        // 操作系统版本
}

// DeviceResponse 设备响应.
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
	CreatedAt    time.Time `json:"created_at"`
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
