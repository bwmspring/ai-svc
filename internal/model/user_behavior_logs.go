package model

import "time"

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
