package repository

import (
	"time"

	"ai-svc/internal/model"
	"ai-svc/pkg/database"

	"gorm.io/gorm"
)

// SMSRepository 短信仓储接口.
type SMSRepository interface {
	CreateSMSCode(smsCode *model.SMSVerificationCode) error
	GetLatestSMSCode(phone, purpose string) (*model.SMSVerificationCode, error)
	UpdateSMSCode(smsCode *model.SMSVerificationCode) error
	CountSMSInDuration(phone, purpose string, duration time.Duration) (int64, error)
	CountSMSByIPInDuration(clientIP string, duration time.Duration) (int64, error)
	CleanExpiredSMSCodes() error
}

// smsRepository 短信仓储实现.
type smsRepository struct {
	db *gorm.DB
}

// NewSMSRepository 创建短信仓储实例.
func NewSMSRepository() SMSRepository {
	return &smsRepository{
		db: database.GetDB(),
	}
}

// CreateSMSCode 创建短信验证码.
func (r *smsRepository) CreateSMSCode(smsCode *model.SMSVerificationCode) error {
	return r.db.Create(smsCode).Error
}

// GetLatestSMSCode 根据手机号和用途获取最新的验证码.
func (r *smsRepository) GetLatestSMSCode(phone, purpose string) (*model.SMSVerificationCode, error) {
	var smsCode model.SMSVerificationCode
	err := r.db.Where("phone = ? AND purpose = ?", phone, purpose).
		Order("created_at DESC").First(&smsCode).Error
	if err != nil {
		return nil, err
	}
	return &smsCode, nil
}

// UpdateSMSCode 更新短信验证码.
func (r *smsRepository) UpdateSMSCode(smsCode *model.SMSVerificationCode) error {
	return r.db.Save(smsCode).Error
}

// CountSMSInDuration 统计指定时间段内的短信发送次数.
func (r *smsRepository) CountSMSInDuration(phone, purpose string, duration time.Duration) (int64, error) {
	var count int64
	query := r.db.Model(&model.SMSVerificationCode{}).
		Where("phone = ? AND created_at > ?", phone, time.Now().Add(-duration))

	if purpose != "" {
		query = query.Where("purpose = ?", purpose)
	}

	err := query.Count(&count).Error
	return count, err
}

// CountSMSByIPInDuration 统计指定时间段内指定IP的短信发送次数.
func (r *smsRepository) CountSMSByIPInDuration(clientIP string, duration time.Duration) (int64, error) {
	var count int64
	err := r.db.Model(&model.SMSVerificationCode{}).
		Where("client_ip = ? AND created_at > ?", clientIP, time.Now().Add(-duration)).
		Count(&count).Error
	return count, err
}

// CleanExpiredSMSCodes 清理过期的短信验证码.
func (r *smsRepository) CleanExpiredSMSCodes() error {
	return r.db.Where("expired_at < ?", time.Now()).Delete(&model.SMSVerificationCode{}).Error
}
