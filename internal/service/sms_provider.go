package service

import (
	"errors"
	"fmt"

	"ai-svc/pkg/logger"
)

// SMSProvider 短信服务提供商接口.
type SMSProvider interface {
	SendSMS(phone, code, purpose string) error
	GetProviderName() string
}

// AliyunSMSProvider 阿里云短信服务提供商.
type AliyunSMSProvider struct {
	AccessKeyID     string
	AccessKeySecret string
	SignName        string
	TemplateCode    string
}

// NewAliyunSMSProvider 创建阿里云短信服务提供商.
func NewAliyunSMSProvider(accessKeyID, accessKeySecret, signName, templateCode string) SMSProvider {
	return &AliyunSMSProvider{
		AccessKeyID:     accessKeyID,
		AccessKeySecret: accessKeySecret,
		SignName:        signName,
		TemplateCode:    templateCode,
	}
}

// SendSMS 发送短信.
func (p *AliyunSMSProvider) SendSMS(phone, code, purpose string) error {
	// 这里应该调用阿里云的短信 API
	// 示例代码（需要引入阿里云SDK）:

	/*
		client, err := dysmsapi.NewClientWithAccessKey("cn-hangzhou", p.AccessKeyID, p.AccessKeySecret)
		if err != nil {
			return err
		}

		request := dysmsapi.CreateSendSmsRequest()
		request.PhoneNumbers = phone
		request.SignName = p.SignName
		request.TemplateCode = p.TemplateCode
		request.TemplateParam = fmt.Sprintf(`{"code":"%s"}`, code)

		response, err := client.SendSms(request)
		if err != nil {
			return err
		}

		if response.Code != "OK" {
			return errors.New(response.Message)
		}
	*/

	// 当前使用模拟实现
	logger.Info("阿里云短信发送", map[string]any{
		"phone":   phone,
		"code":    code,
		"purpose": purpose,
	})

	return nil
}

// GetProviderName 获取服务提供商名称.
func (p *AliyunSMSProvider) GetProviderName() string {
	return "阿里云短信"
}

// TencentSMSProvider 腾讯云短信服务提供商.
type TencentSMSProvider struct {
	SecretID   string
	SecretKey  string
	AppID      string
	SignName   string
	TemplateID string
}

// NewTencentSMSProvider 创建腾讯云短信服务提供商.
func NewTencentSMSProvider(secretID, secretKey, appID, signName, templateID string) SMSProvider {
	return &TencentSMSProvider{
		SecretID:   secretID,
		SecretKey:  secretKey,
		AppID:      appID,
		SignName:   signName,
		TemplateID: templateID,
	}
}

// SendSMS 发送短信.
func (p *TencentSMSProvider) SendSMS(phone, code, purpose string) error {
	// 这里应该调用腾讯云的短信 API
	// 示例代码（需要引入腾讯云SDK）:

	/*
		credential := common.NewCredential(p.SecretID, p.SecretKey)
		client, err := sms.NewClient(credential, "ap-beijing", profile.NewClientProfile())
		if err != nil {
			return err
		}

		request := sms.NewSendSmsRequest()
		request.SmsSdkAppId = &p.AppID
		request.SignName = &p.SignName
		request.TemplateId = &p.TemplateID
		request.PhoneNumberSet = []*string{&phone}
		request.TemplateParamSet = []*string{&code}

		response, err := client.SendSms(request)
		if err != nil {
			return err
		}

		if *response.SendStatusSet[0].Code != "Ok" {
			return errors.New(*response.SendStatusSet[0].Message)
		}
	*/

	// 当前使用模拟实现
	logger.Info("腾讯云短信发送", map[string]any{
		"phone":   phone,
		"code":    code,
		"purpose": purpose,
	})

	return nil
}

// GetProviderName 获取服务提供商名称.
func (p *TencentSMSProvider) GetProviderName() string {
	return "腾讯云短信"
}

// MockSMSProvider 模拟短信服务提供商（用于开发测试）.
type MockSMSProvider struct{}

// NewMockSMSProvider 创建模拟短信服务提供商.
func NewMockSMSProvider() SMSProvider {
	return &MockSMSProvider{}
}

// SendSMS 发送短信（模拟）.
func (p *MockSMSProvider) SendSMS(phone, code, purpose string) error {
	// 模拟发送过程
	logger.Info("模拟短信发送", map[string]any{
		"phone":   phone,
		"code":    code,
		"purpose": purpose,
		"message": fmt.Sprintf("【AI服务】您的验证码是%s，5分钟内有效。", code),
	})

	// 模拟可能的发送失败情况
	if phone == "10000000000" {
		return errors.New("模拟发送失败")
	}

	return nil
}

// GetProviderName 获取服务提供商名称.
func (p *MockSMSProvider) GetProviderName() string {
	return "模拟短信服务"
}
