package model

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel 基础模型
type BaseModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// User 用户模型 - 采用手机号+验证码登录模式
type User struct {
	BaseModel
	// 基础认证信息（手机号作为主要登录凭证）
	Phone    string `gorm:"type:varchar(20);uniqueIndex;not null" json:"phone" validate:"required,min=11,max=20"`
	Username string `gorm:"type:varchar(50);uniqueIndex" json:"username" validate:"min=3,max=50"`
	Email    string `gorm:"type:varchar(100);uniqueIndex" json:"email" validate:"email"`

	// 用户资料信息（常用查询字段放在主表）
	Nickname string `gorm:"type:varchar(50)" json:"nickname"`
	Avatar   string `gorm:"type:varchar(255)" json:"avatar"`
	RealName string `gorm:"type:varchar(50)" json:"real_name"`
	Gender   int    `gorm:"type:tinyint;default:0;comment:性别 0:未知 1:男 2:女" json:"gender"`

	// VIP 相关字段
	VIPLevel    int        `gorm:"type:tinyint;default:0;comment:VIP等级 0:普通用户 1:VIP1 2:VIP2 3:VIP3" json:"vip_level"`
	VIPExpireAt *time.Time `gorm:"comment:VIP过期时间" json:"vip_expire_at"`

	// 状态和行为字段
	Status      int        `gorm:"type:tinyint;default:1;comment:状态 1:正常 0:禁用 -1:删除" json:"status"`
	LastLoginIP string     `gorm:"type:varchar(45)" json:"last_login_ip"`
	LastLoginAt *time.Time `gorm:"comment:最后登录时间" json:"last_login_at"`
	LoginCount  int        `gorm:"type:int;default:0;comment:登录次数" json:"login_count"`

	// 扩展字段（不常用的详细信息）
	Birthday *time.Time `json:"birthday"`
	Address  string     `gorm:"type:varchar(255)" json:"address"`
	Bio      string     `gorm:"type:text" json:"bio"`
}

// TableName 表名
func (User) TableName() string {
	return "users"
}

// IsVIP 检查是否为VIP用户
func (u *User) IsVIP() bool {
	return u.VIPLevel > 0 && (u.VIPExpireAt == nil || u.VIPExpireAt.After(time.Now()))
}

// IsActive 检查用户是否激活
func (u *User) IsActive() bool {
	return u.Status == 1
}

// UserBehaviorLog 用户行为日志（频繁变更的数据独立存储）
type UserBehaviorLog struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	Action    string    `gorm:"type:varchar(50);not null" json:"action"` // login, logout, view, purchase等
	Resource  string    `gorm:"type:varchar(100)" json:"resource"`       // 操作的资源
	IP        string    `gorm:"type:varchar(45)" json:"ip"`
	UserAgent string    `gorm:"type:varchar(500)" json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName 表名
func (UserBehaviorLog) TableName() string {
	return "user_behavior_logs"
}

// SendSMSRequest 发送短信验证码请求
type SendSMSRequest struct {
	Phone   string `json:"phone" validate:"required,min=11,max=20"`
	Purpose string `json:"purpose" validate:"required,oneof=login register reset"` // login, register, reset
}

// LoginWithSMSRequest 手机号+验证码登录请求
type LoginWithSMSRequest struct {
	Phone string `json:"phone" validate:"required,min=11,max=20"`
	Code  string `json:"code" validate:"required,min=4,max=10"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Username string     `json:"username" validate:"min=3,max=50"`
	Email    string     `json:"email" validate:"email"`
	Nickname string     `json:"nickname" validate:"max=50"`
	Avatar   string     `json:"avatar" validate:"url"`
	RealName string     `json:"real_name" validate:"max=50"`
	Gender   int        `json:"gender" validate:"min=0,max=2"`
	Birthday *time.Time `json:"birthday"`
	Address  string     `json:"address" validate:"max=255"`
	Bio      string     `json:"bio" validate:"max=1000"`
}

// UserResponse 用户响应
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

// UserListResponse 用户列表响应（简化版，用于列表页面）
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

// LoginResponse 登录响应
type LoginResponse struct {
	User  *UserResponse `json:"user"`
	Token string        `json:"token"`
}

// SMSVerificationCode 短信验证码模型
type SMSVerificationCode struct {
	ID        uint       `gorm:"primarykey" json:"id"`
	Phone     string     `gorm:"type:varchar(20);not null;index" json:"phone" validate:"required"`
	Code      string     `gorm:"type:varchar(10);not null" json:"-"`
	Purpose   string     `gorm:"type:varchar(20);not null" json:"purpose"`   // login, register, reset等
	ClientIP  string     `gorm:"type:varchar(45);not null" json:"client_ip"` // 客户端IP地址
	UsedAt    *time.Time `gorm:"comment:使用时间" json:"used_at"`
	ExpiredAt time.Time  `gorm:"not null" json:"expired_at"`
	CreatedAt time.Time  `json:"created_at"`
}

// TableName 表名
func (SMSVerificationCode) TableName() string {
	return "sms_verification_codes"
}

// IsValid 检查验证码是否有效
func (s *SMSVerificationCode) IsValid() bool {
	return s.UsedAt == nil && s.ExpiredAt.After(time.Now())
}

// MarkAsUsed 标记验证码为已使用
func (s *SMSVerificationCode) MarkAsUsed() {
	now := time.Now()
	s.UsedAt = &now
}
