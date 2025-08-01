package model

import (
	"ai-svc/pkg/utils"
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

// GetMaskedPhone 获取脱敏后的手机号
func (u *User) GetMaskedPhone() string {
	return utils.MaskPhone(u.Phone)
}

// GetMaskedEmail 获取脱敏后的邮箱
func (u *User) GetMaskedEmail() string {
	return utils.MaskEmail(u.Email)
}

// GetMaskedRealName 获取脱敏后的真实姓名
func (u *User) GetMaskedRealName() string {
	return utils.MaskName(u.RealName)
}

// SendSMSRequest 发送短信验证码请求.
type SendSMSRequest struct {
	Phone   string `json:"phone"           validate:"required,min=11,max=20"`
	Purpose string `json:"purpose"         validate:"required,oneof=login register reset change payment withdraw security device"` // 扩展验证码用途
	Token   string `json:"token,omitempty"`                                                                                        // 可选token，用于高安全级别操作
}

// LoginWithSMSRequest 手机号+验证码登录请求（扩展设备信息）.
type LoginWithSMSRequest struct {
	Phone      string                     `json:"phone"           validate:"required,min=11,max=20"`
	Code       string                     `json:"code"            validate:"required,min=4,max=10"`
	Token      string                     `json:"token,omitempty"`                     // 验证token
	DeviceInfo *DeviceRegistrationRequest `json:"device_info"     validate:"required"` // 设备注册信息（客户端只传设备指纹）
}

// ValidateSMSRequest 验证码验证请求
type ValidateSMSRequest struct {
	Phone   string `json:"phone"           validate:"required,min=11,max=20"`
	Code    string `json:"code"            validate:"required,min=4,max=10"`
	Purpose string `json:"purpose"         validate:"required"`
	Token   string `json:"token,omitempty"` // 验证token
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
	Username    *string    `json:"username"`
	Email       *string    `json:"email"`
	Nickname    *string    `json:"nickname"`
	Avatar      *string    `json:"avatar"`
	RealName    *string    `json:"real_name"`
	Gender      int        `json:"gender"`    // 0:未知 1:男 2:女
	VIPLevel    int        `json:"vip_level"` // 0:普通用户 1:VIP1 2:VIP2 3:VIP3
	VIPExpireAt *time.Time `json:"vip_expire_at"`
	Status      int        `json:"status"` // 0:禁用 1:正常
	LastLoginIP *string    `json:"last_login_ip"`
	LastLoginAt *time.Time `json:"last_login_at"`
	LoginCount  int        `json:"login_count"`
	Birthday    *time.Time `json:"birthday"`
	Address     *string    `json:"address"`
	Bio         *string    `json:"bio"`
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
	User         *UserResponse `json:"user"`
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	ExpiresIn    int64         `json:"expires_in"` // access token过期时间（秒）
	TokenType    string        `json:"token_type"` // Bearer
}

// TokenPair Token对.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // access token过期时间（秒）
	TokenType    string `json:"token_type"` // Bearer
}

// RefreshTokenRequest 刷新Token请求.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}
