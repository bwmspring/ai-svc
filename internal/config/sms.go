package config

// SMSConfig 短信服务配置
type SMSConfig struct {
	Provider string           `mapstructure:"provider" yaml:"provider" json:"provider"` // 服务提供商: aliyun, tencent, mock
	Aliyun   AliyunSMSConfig  `mapstructure:"aliyun"   yaml:"aliyun"   json:"aliyun"`
	Tencent  TencentSMSConfig `mapstructure:"tencent"  yaml:"tencent"  json:"tencent"`
}

// AliyunSMSConfig 阿里云短信配置
type AliyunSMSConfig struct {
	AccessKeyID     string `mapstructure:"access_key_id"     yaml:"access_key_id"     json:"access_key_id"`
	AccessKeySecret string `mapstructure:"access_key_secret" yaml:"access_key_secret" json:"access_key_secret"`
	SignName        string `mapstructure:"sign_name"         yaml:"sign_name"         json:"sign_name"`
	TemplateCode    string `mapstructure:"template_code"     yaml:"template_code"     json:"template_code"`
}

// TencentSMSConfig 腾讯云短信配置
type TencentSMSConfig struct {
	SecretID   string `mapstructure:"secret_id"   yaml:"secret_id"   json:"secret_id"`
	SecretKey  string `mapstructure:"secret_key"  yaml:"secret_key"  json:"secret_key"`
	AppID      string `mapstructure:"app_id"      yaml:"app_id"      json:"app_id"`
	Sign       string `mapstructure:"sign"        yaml:"sign"        json:"sign"`
	TemplateID string `mapstructure:"template_id" yaml:"template_id" json:"template_id"`
}

// RateLimitConfig 频率限制配置（保持向后兼容）
type RateLimitConfig struct {
	PerMinute int `yaml:"per_minute" json:"per_minute"` // 每分钟最多发送次数
	PerHour   int `yaml:"per_hour"   json:"per_hour"`   // 每小时最多发送次数
	PerDay    int `yaml:"per_day"    json:"per_day"`    // 每天最多发送次数
}
