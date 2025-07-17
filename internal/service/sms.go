package service

import (
	"crypto/rand"
	"errors"
	"math/big"
	"time"

	"ai-svc/internal/model"
	"ai-svc/internal/repository"
	"ai-svc/pkg/logger"

	"gorm.io/gorm"
)

// SMSService 短信服务接口.
type SMSService interface {
	SendVerificationCode(req *model.SendSMSRequest, clientIP string) error
	ValidateVerificationCode(phone, code, purpose string) error
}

// smsService 短信服务实现.
type smsService struct {
	smsRepo     repository.SMSRepository
	smsProvider SMSProvider
}

// NewSMSService 创建短信服务实例.
func NewSMSService(smsRepo repository.SMSRepository) SMSService {
	return &smsService{
		smsRepo:     smsRepo,
		smsProvider: NewMockSMSProvider(), // 默认使用模拟提供商，生产环境应该使用真实的
	}
}

// NewSMSServiceWithProvider 创建带指定提供商的短信服务实例.
func NewSMSServiceWithProvider(smsRepo repository.SMSRepository, provider SMSProvider) SMSService {
	return &smsService{
		smsRepo:     smsRepo,
		smsProvider: provider,
	}
}

// SendVerificationCode 发送短信验证码.
func (s *smsService) SendVerificationCode(req *model.SendSMSRequest, clientIP string) error {
	// 1. 防刷校验：检查发送频率限制
	if err := s.checkSendFrequency(req.Phone, clientIP); err != nil {
		return err
	}

	// 2. 检查是否已有有效验证码
	if err := s.checkExistingCode(req.Phone, req.Purpose); err != nil {
		return err
	}

	// 3. 生成6位随机验证码
	code, err := s.generateVerificationCode()
	if err != nil {
		logger.Error("生成验证码失败", map[string]any{"error": err.Error()})
		return errors.New("生成验证码失败")
	}

	// 4. 创建验证码记录
	smsCode := &model.SMSVerificationCode{
		Phone:     req.Phone,
		Code:      code,
		Purpose:   req.Purpose,
		ExpiredAt: time.Now().Add(5 * time.Minute), // 5分钟过期
		ClientIP:  clientIP,
	}

	// 5. 保存验证码
	if err := s.smsRepo.CreateSMSCode(smsCode); err != nil {
		logger.Error("保存验证码失败", map[string]any{"error": err.Error()})
		return errors.New("保存验证码失败")
	}

	// 6. 调用短信服务商发送短信
	if err := s.sendSMSToProvider(req.Phone, code, req.Purpose); err != nil {
		logger.Error("发送短信失败", map[string]any{"error": err.Error(), "phone": req.Phone})
		return errors.New("发送短信失败")
	}

	// 7. 记录发送日志
	s.logSMSSend(req.Phone, req.Purpose, clientIP)

	return nil
}

// ValidateVerificationCode 验证短信验证码.
func (s *smsService) ValidateVerificationCode(phone, code, purpose string) error {
	smsCode, err := s.smsRepo.GetLatestSMSCode(phone, purpose)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("验证码不存在或已过期")
		}
		return errors.New("验证码验证失败")
	}

	if !smsCode.IsValid() {
		return errors.New("验证码不存在或已过期")
	}

	if smsCode.Code != code {
		// 记录验证失败日志
		logger.Warn("验证码验证失败", map[string]any{
			"phone":   phone,
			"purpose": purpose,
			"ip":      smsCode.ClientIP,
		})
		return errors.New("验证码错误")
	}

	// 标记验证码为已使用
	smsCode.MarkAsUsed()
	if err := s.smsRepo.UpdateSMSCode(smsCode); err != nil {
		logger.Error("更新验证码状态失败", map[string]any{"error": err.Error()})
	}

	return nil
}

// checkSendFrequency 检查发送频率限制.
func (s *smsService) checkSendFrequency(phone, clientIP string) error {
	// 同一手机号1分钟内只能发送1次
	if count, err := s.smsRepo.CountSMSInDuration(phone, "", 1*time.Minute); err == nil && count > 0 {
		return errors.New("发送过于频繁，请稍后再试")
	}

	// 同一手机号1小时内最多发送5次
	if count, err := s.smsRepo.CountSMSInDuration(phone, "", 1*time.Hour); err == nil && count >= 5 {
		return errors.New("今日发送次数已达上限")
	}

	// 同一IP地址1小时内最多发送20次
	if count, err := s.smsRepo.CountSMSByIPInDuration(clientIP, 1*time.Hour); err == nil && count >= 20 {
		return errors.New("发送过于频繁，请稍后再试")
	}

	return nil
}

// checkExistingCode 检查是否已有有效验证码.
func (s *smsService) checkExistingCode(phone, purpose string) error {
	smsCode, err := s.smsRepo.GetLatestSMSCode(phone, purpose)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil // 没有验证码，可以发送
		}
		return nil // 查询失败，继续发送
	}

	// 如果有有效的验证码，检查是否还在冷却期内
	if smsCode.IsValid() {
		// 如果验证码还有效且创建时间在10分钟内，不允许重复发送
		if time.Since(smsCode.CreatedAt) < 10*time.Minute {
			return errors.New("验证码已发送，请稍后再试")
		}
	}

	return nil
}

// generateVerificationCode 生成6位随机验证码.
func (s *smsService) generateVerificationCode() (string, error) {
	const chars = "0123456789"
	code := make([]byte, 6)
	for i := range code {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", err
		}
		code[i] = chars[num.Int64()]
	}
	return string(code), nil
}

// sendSMSToProvider 调用短信服务商发送短信.
func (s *smsService) sendSMSToProvider(phone, code, purpose string) error {
	return s.smsProvider.SendSMS(phone, code, purpose)
}

// logSMSSend 记录短信发送日志.
func (s *smsService) logSMSSend(phone, purpose, clientIP string) {
	logger.Info("短信发送成功", map[string]any{
		"phone":     phone,
		"purpose":   purpose,
		"client_ip": clientIP,
		"timestamp": time.Now().Unix(),
	})
}
