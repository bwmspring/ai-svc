package service

import (
	"ai-svc/internal/model"
	"ai-svc/internal/repository"
	"ai-svc/pkg/logger"
	"ai-svc/pkg/utils"
	"crypto/rand"
	"errors"
	"math/big"
	"time"

	"gorm.io/gorm"
)

// SMSService 短信服务接口.
type SMSService interface {
	SendVerificationCode(req *model.SendSMSRequest, clientIP, userAgent string, userID *uint) error
	ValidateVerificationCode(phone, code, purpose, token string) error
	ValidateVerificationCodeWithUser(phone, code, purpose, token string, userID uint) error
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
func (s *smsService) SendVerificationCode(req *model.SendSMSRequest, clientIP, userAgent string, userID *uint) error {
	// 1. 高安全级别操作需要用户身份验证
	if model.IsHighSecurityPurpose(req.Purpose) {
		if userID == nil {
			return errors.New("高安全级别操作需要用户登录")
		}
	}

	// 2. 防刷校验：检查发送频率限制
	if err := s.checkSendFrequency(req.Phone, clientIP); err != nil {
		return err
	}

	// 3. 检查是否已有有效验证码
	if err := s.checkExistingCode(req.Phone, req.Purpose); err != nil {
		return err
	}

	// 4. 生成6位随机验证码
	code, err := s.generateVerificationCode()
	if err != nil {
		logger.Error("生成验证码失败", map[string]any{"error": err.Error()})
		return errors.New("生成验证码失败")
	}

	// 5. 创建验证码记录
	smsCode := &model.SMSVerificationCode{
		Phone:     req.Phone,
		Code:      code,
		Purpose:   req.Purpose,
		ExpiredAt: time.Now().Add(5 * time.Minute), // 5分钟过期
		ClientIP:  clientIP,
		UserAgent: userAgent,
	}

	// 6. 生成唯一验证token
	if err := smsCode.GenerateToken(); err != nil {
		logger.Error("生成验证token失败", map[string]any{"error": err.Error()})
		return errors.New("生成验证token失败")
	}

	// 7. 保存验证码
	if err := s.smsRepo.CreateSMSCode(smsCode); err != nil {
		logger.Error("保存验证码失败", map[string]any{"error": err.Error()})
		return errors.New("保存验证码失败")
	}

	// 8. 调用短信服务商发送短信
	if err := s.sendSMSToProvider(req.Phone, code, req.Purpose); err != nil {
		logger.Error("发送短信失败", map[string]any{"error": err.Error(), "phone": req.Phone})
		return errors.New("发送短信失败")
	}

	// 9. 记录发送日志
	s.logSMSSend(req.Phone, req.Purpose, clientIP, userID)

	return nil
}

// ValidateVerificationCode 验证短信验证码.
func (s *smsService) ValidateVerificationCode(phone, code, purpose, token string) error {
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

	// 验证token（如果提供）
	if token != "" && smsCode.Token != token {
		logger.Warn("验证码token不匹配", map[string]any{
			"phone":   phone,
			"purpose": purpose,
			"ip":      smsCode.ClientIP,
		})
		return errors.New("验证码token无效")
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

// ValidateVerificationCodeWithUser 带用户身份验证的验证码验证
func (s *smsService) ValidateVerificationCodeWithUser(phone, code, purpose, token string, userID uint) error {
	// 高安全级别操作需要验证用户身份
	if model.IsHighSecurityPurpose(purpose) {
		// 这里可以添加额外的用户身份验证逻辑
		// 例如检查用户是否登录、检查用户权限等
		logger.Info("高安全级别验证码验证", map[string]any{
			"phone":   phone,
			"purpose": purpose,
			"user_id": userID,
		})
	}

	return s.ValidateVerificationCode(phone, code, purpose, token)
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
func (s *smsService) logSMSSend(phone, purpose, clientIP string, userID *uint) {
	maskedPhone := utils.MaskPhone(phone)
	logData := map[string]any{
		"phone":     maskedPhone, // 使用脱敏手机号
		"purpose":   purpose,
		"client_ip": clientIP,
		"timestamp": time.Now().Unix(),
	}

	if userID != nil {
		logData["user_id"] = *userID
	}

	logger.Info("短信发送成功", logData)
}
