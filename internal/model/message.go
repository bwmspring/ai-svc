package model

import (
	"time"
)

// MessageDefinition 消息定义表 - 存储消息的通用信息，所有用户共享
type MessageDefinition struct {
	BaseModel
	// 消息基本信息
	Title       string `gorm:"type:varchar(200);not null"                                       json:"title"`
	Content     string `gorm:"type:text;not null"                                               json:"content"`
	MessageType int    `gorm:"type:tinyint;default:1;comment:消息类型 1:系统通知 2:用户消息 3:管理员消息 4:业务通知" json:"message_type"`
	Priority    int    `gorm:"type:tinyint;default:1;comment:优先级 1:普通 2:重要 3:紧急"                json:"priority"`

	// 发送者信息
	SenderID   uint `gorm:"not null;comment:发送者ID，0表示系统"                                json:"sender_id"`
	SenderType int  `gorm:"type:tinyint;default:1;comment:发送者类型 1:用户 2:系统 3:管理员 4:业务系统" json:"sender_type"`

	// 消息属性
	IsBroadcast bool   `gorm:"default:false;comment:是否为广播消息"       json:"is_broadcast"`
	TargetUsers []uint `gorm:"type:json;comment:目标用户ID列表，为空表示所有用户" json:"target_users"`

	// 扩展字段
	ExtraData string     `gorm:"type:json;comment:扩展数据" json:"extra_data"`
	ExpireAt  *time.Time `gorm:"comment:过期时间"           json:"expire_at"`

	// 统计信息
	TotalRecipients int `gorm:"default:0;comment:总接收者数量" json:"total_recipients"`
	ReadCount       int `gorm:"default:0;comment:已读数量"   json:"read_count"`
}

// TableName 表名
func (MessageDefinition) TableName() string {
	return "message_definitions"
}

// UserMessage 用户消息表 - 存储用户个人的消息状态
type UserMessage struct {
	BaseModel
	// 关联消息定义
	MessageDefinitionID uint `gorm:"not null;index;comment:消息定义ID" json:"message_definition_id"`

	// 接收者信息
	RecipientID uint `gorm:"not null;index;comment:接收者ID" json:"recipient_id"`

	// 消息状态
	IsRead    bool       `gorm:"default:false;index;comment:是否已读"  json:"is_read"`
	IsDeleted bool       `gorm:"default:false;index;comment:是否已删除" json:"is_deleted"`
	ReadAt    *time.Time `gorm:"comment:阅读时间"                      json:"read_at"`

	// 关联查询
	MessageDefinition MessageDefinition `gorm:"foreignKey:MessageDefinitionID" json:"message_definition,omitempty"`
}

// TableName 表名
func (UserMessage) TableName() string {
	return "user_messages"
}

// SendMessageRequest 发送消息请求
type SendMessageRequest struct {
	Title       string     `json:"title"        validate:"required,max=200"`
	Content     string     `json:"content"      validate:"required"`
	MessageType int        `json:"message_type" validate:"min=1,max=4"`
	Priority    int        `json:"priority"     validate:"min=1,max=3"`
	RecipientID uint       `json:"recipient_id" validate:"required"`
	ExtraData   string     `json:"extra_data"`
	ExpireAt    *time.Time `json:"expire_at"`

	// 内部字段，由服务层设置
	SenderID   uint `json:"-"`
	SenderType int  `json:"-"`
}

// SendBroadcastMessageRequest 发送广播消息请求
type SendBroadcastMessageRequest struct {
	Title       string     `json:"title"        validate:"required,max=200"`
	Content     string     `json:"content"      validate:"required"`
	MessageType int        `json:"message_type" validate:"min=1,max=4"`
	Priority    int        `json:"priority"     validate:"min=1,max=3"`
	TargetUsers []uint     `json:"target_users" validate:"omitempty,min=1"` // 为空表示所有用户
	ExtraData   string     `json:"extra_data"`
	ExpireAt    *time.Time `json:"expire_at"`

	// 内部字段，由服务层设置
	SenderID   uint `json:"-"`
	SenderType int  `json:"-"`
}

// MessageResponse 消息响应
type MessageResponse struct {
	ID                  uint       `json:"id"`
	MessageDefinitionID uint       `json:"message_definition_id"`
	Title               string     `json:"title"`
	Content             string     `json:"content"`
	MessageType         int        `json:"message_type"`
	Priority            int        `json:"priority"`
	SenderID            uint       `json:"sender_id"`
	SenderType          int        `json:"sender_type"`
	RecipientID         uint       `json:"recipient_id"`
	IsRead              bool       `json:"is_read"`
	IsDeleted           bool       `json:"is_deleted"`
	ReadAt              *time.Time `json:"read_at"`
	ExtraData           string     `json:"extra_data"`
	ExpireAt            *time.Time `json:"expire_at"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// MessageListResponse 消息列表响应
type MessageListResponse struct {
	Messages   []MessageResponse `json:"messages"`
	Pagination PaginationInfo    `json:"pagination"`
}

// PaginationInfo 分页信息
type PaginationInfo struct {
	Page       int `json:"page"`
	Size       int `json:"size"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// UnreadCountResponse 未读消息数量响应
type UnreadCountResponse struct {
	UnreadCount int `json:"unread_count"`
}

// BatchReadRequest 批量标记已读请求
type BatchReadRequest struct {
	MessageIDs []uint `json:"message_ids" validate:"required,min=1"`
}

// MessageQueryParams 消息查询参数
type MessageQueryParams struct {
	Page        int   `form:"page"     binding:"min=1"`
	Size        int   `form:"size"     binding:"min=1,max=100"`
	MessageType *int  `form:"type"     binding:"min=1,max=4"`
	IsRead      *bool `form:"is_read"`
	Priority    *int  `form:"priority" binding:"min=1,max=3"`
}

// BroadcastMessageResponse 广播消息响应
type BroadcastMessageResponse struct {
	MessageDefinitionID uint   `json:"message_definition_id"`
	Status              string `json:"status"`
	TotalRecipients     int    `json:"total_recipients"`
	EstimatedTime       string `json:"estimated_time"`
}
