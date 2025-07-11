package config

// SMSConfig 短信服务配置
type SMSConfig struct {
	Provider  string           `yaml:"provider" json:"provider"` // 服务提供商: aliyun, tencent, mock
	Aliyun    AliyunSMSConfig  `yaml:"aliyun" json:"aliyun"`
	Tencent   TencentSMSConfig `yaml:"tencent" json:"tencent"`
	RateLimit RateLimitConfig  `yaml:"rate_limit" json:"rate_limit"`
}

// AliyunSMSConfig 阿里云短信配置
type AliyunSMSConfig struct {
	AccessKeyID     string `yaml:"access_key_id" json:"access_key_id"`
	AccessKeySecret string `yaml:"access_key_secret" json:"access_key_secret"`
	SignName        string `yaml:"sign_name" json:"sign_name"`
	TemplateCode    string `yaml:"template_code" json:"template_code"`
}

// TencentSMSConfig 腾讯云短信配置
type TencentSMSConfig struct {
	SecretID   string `yaml:"secret_id" json:"secret_id"`
	SecretKey  string `yaml:"secret_key" json:"secret_key"`
	AppID      string `yaml:"app_id" json:"app_id"`
	SignName   string `yaml:"sign_name" json:"sign_name"`
	TemplateID string `yaml:"template_id" json:"template_id"`
}

// RateLimitConfig 频率限制配置
type RateLimitConfig struct {
	PerMinute int `yaml:"per_minute" json:"per_minute"` // 每分钟最多发送次数
	PerHour   int `yaml:"per_hour" json:"per_hour"`     // 每小时最多发送次数
	PerDay    int `yaml:"per_day" json:"per_day"`       // 每天最多发送次数
}
