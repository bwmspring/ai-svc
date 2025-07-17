package model

import "time"

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
