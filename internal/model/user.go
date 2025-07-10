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

// User 用户模型
type User struct {
	BaseModel
	Username string     `gorm:"type:varchar(50);uniqueIndex;not null" json:"username" validate:"required,min=3,max=50"`
	Email    string     `gorm:"type:varchar(100);uniqueIndex;not null" json:"email" validate:"required,email"`
	Password string     `gorm:"type:varchar(255);not null" json:"-"`
	Nickname string     `gorm:"type:varchar(50)" json:"nickname"`
	Avatar   string     `gorm:"type:varchar(255)" json:"avatar"`
	Status   int        `gorm:"type:tinyint;default:1;comment:状态 1:正常 0:禁用" json:"status"`
	LastIP   string     `gorm:"type:varchar(45)" json:"last_ip"`
	LastTime *time.Time `json:"last_time"`
}

// TableName 表名
func (User) TableName() string {
	return "users"
}

// UserProfile 用户资料模型
type UserProfile struct {
	BaseModel
	UserID   uint       `gorm:"index;not null" json:"user_id"`
	RealName string     `gorm:"type:varchar(20)" json:"real_name"`
	Phone    string     `gorm:"type:varchar(20)" json:"phone"`
	Gender   int        `gorm:"type:tinyint;default:0;comment:性别 0:未知 1:男 2:女" json:"gender"`
	Birthday *time.Time `json:"birthday"`
	Address  string     `gorm:"type:varchar(255)" json:"address"`
	Bio      string     `gorm:"type:text" json:"bio"`
}

// TableName 表名
func (UserProfile) TableName() string {
	return "user_profiles"
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=50"`
	Nickname string `json:"nickname" validate:"max=50"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Nickname string `json:"nickname" validate:"max=50"`
	Avatar   string `json:"avatar" validate:"url"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=6,max=50"`
}

// UserResponse 用户响应
type UserResponse struct {
	ID        uint       `json:"id"`
	Username  string     `json:"username"`
	Email     string     `json:"email"`
	Nickname  string     `json:"nickname"`
	Avatar    string     `json:"avatar"`
	Status    int        `json:"status"`
	LastIP    string     `json:"last_ip"`
	LastTime  *time.Time `json:"last_time"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}
