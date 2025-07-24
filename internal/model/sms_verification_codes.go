package model

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"ai-svc/pkg/utils"
)

// SMSVerificationCode 短信验证码模型.
type SMSVerificationCode struct {
	ID        uint       `gorm:"primarykey"                      json:"id"`
	Phone     string     `gorm:"type:varchar(20);not null;index" json:"phone"      validate:"required"`
	Code      string     `gorm:"type:varchar(10);not null"       json:"-"`
	Token     string     `gorm:"type:varchar(64);not null;uniqueIndex" json:"token"` // 唯一验证token
	Purpose   string     `gorm:"type:varchar(20);not null"       json:"purpose"`     // login, register, reset等
	ClientIP  string     `gorm:"type:varchar(45);not null"       json:"client_ip"`   // 客户端IP地址
	UserAgent string     `gorm:"type:varchar(500)"               json:"user_agent"`  // 用户代理
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

// GenerateToken 生成唯一验证token
func (s *SMSVerificationCode) GenerateToken() error {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return err
	}
	s.Token = hex.EncodeToString(bytes)
	return nil
}

// GetMaskedPhone 获取脱敏后的手机号
func (s *SMSVerificationCode) GetMaskedPhone() string {
	return utils.MaskPhone(s.Phone)
}

// VerificationCodePurpose 验证码用途枚举
const (
	PurposeLogin    = "login"    // 登录
	PurposeRegister = "register" // 注册
	PurposeReset    = "reset"    // 重置密码
	PurposeChange   = "change"   // 变更个人信息
	PurposePayment  = "payment"  // 支付验证
	PurposeWithdraw = "withdraw" // 提现验证
	PurposeSecurity = "security" // 安全设置变更
	PurposeDevice   = "device"   // 设备绑定
)

// IsHighSecurityPurpose 检查是否为高安全级别用途
func IsHighSecurityPurpose(purpose string) bool {
	highSecurityPurposes := []string{
		PurposeChange,
		PurposePayment,
		PurposeWithdraw,
		PurposeSecurity,
	}

	for _, p := range highSecurityPurposes {
		if p == purpose {
			return true
		}
	}
	return false
}

// GetPurposeDescription 获取用途描述
func GetPurposeDescription(purpose string) string {
	descriptions := map[string]string{
		PurposeLogin:    "登录验证",
		PurposeRegister: "注册验证",
		PurposeReset:    "重置密码",
		PurposeChange:   "变更个人信息",
		PurposePayment:  "支付验证",
		PurposeWithdraw: "提现验证",
		PurposeSecurity: "安全设置变更",
		PurposeDevice:   "设备绑定",
	}

	if desc, exists := descriptions[purpose]; exists {
		return desc
	}
	return "未知用途"
}
