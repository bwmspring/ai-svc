package config

import (
	"time"
)

// AIConfig AI 服务总配置
type AIConfig struct {
	// 默认提供商
	DefaultProvider string `mapstructure:"default_provider" yaml:"default_provider"`

	// 请求超时时间
	Timeout time.Duration `mapstructure:"timeout" yaml:"timeout"`

	// 最大重试次数
	MaxRetries int `mapstructure:"max_retries" yaml:"max_retries"`

	// 提供商配置
	Providers map[string]ProviderConfig `mapstructure:"providers" yaml:"providers"`

	// 功能配置
	Features FeatureConfig `mapstructure:"features" yaml:"features"`
}

// ProviderConfig AI 提供商配置
type ProviderConfig struct {
	// 是否启用
	Enabled bool `mapstructure:"enabled" yaml:"enabled"`

	// 提供商名称
	Name string `mapstructure:"name" yaml:"name"`

	// API 基础URL
	BaseURL string `mapstructure:"base_url" yaml:"base_url"`

	// API 密钥
	APIKey string `mapstructure:"api_key" yaml:"api_key"`

	// 组织ID（OpenAI专用）
	Organization string `mapstructure:"organization" yaml:"organization"`

	// 密钥（百度、腾讯专用）
	SecretKey string `mapstructure:"secret_key" yaml:"secret_key"`

	// Secret ID（腾讯专用）
	SecretID string `mapstructure:"secret_id" yaml:"secret_id"`

	// 版本（Claude专用）
	Version string `mapstructure:"version" yaml:"version"`

	// 区域（腾讯专用）
	Region string `mapstructure:"region" yaml:"region"`

	// 支持的模型列表
	Models []ModelConfig `mapstructure:"models" yaml:"models"`
}

// ModelConfig 模型配置
type ModelConfig struct {
	// 模型名称
	Name string `mapstructure:"name" yaml:"name"`

	// 最大token数
	MaxTokens int `mapstructure:"max_tokens" yaml:"max_tokens"`

	// 温度参数
	Temperature float32 `mapstructure:"temperature" yaml:"temperature"`

	// 定价信息
	Pricing PricingConfig `mapstructure:"pricing" yaml:"pricing"`
}

// PricingConfig 定价配置
type PricingConfig struct {
	// 输入价格（每1K tokens）
	Input float64 `mapstructure:"input" yaml:"input"`

	// 输出价格（每1K tokens）
	Output float64 `mapstructure:"output" yaml:"output"`
}

// FeatureConfig 功能配置
type FeatureConfig struct {
	// 流式响应
	Streaming bool `mapstructure:"streaming" yaml:"streaming"`

	// 对话历史配置
	History HistoryConfig `mapstructure:"history" yaml:"history"`

	// 内容过滤配置
	ContentFilter ContentFilterConfig `mapstructure:"content_filter" yaml:"content_filter"`

	// 使用统计配置
	UsageTracking UsageTrackingConfig `mapstructure:"usage_tracking" yaml:"usage_tracking"`

	// 缓存配置
	Cache CacheConfig `mapstructure:"cache" yaml:"cache"`
}

// HistoryConfig 对话历史配置
type HistoryConfig struct {
	// 是否启用
	Enabled bool `mapstructure:"enabled" yaml:"enabled"`

	// 最大保留消息数
	MaxMessages int `mapstructure:"max_messages" yaml:"max_messages"`

	// 最大token数
	MaxTokens int `mapstructure:"max_tokens" yaml:"max_tokens"`
}

// ContentFilterConfig 内容过滤配置
type ContentFilterConfig struct {
	// 是否启用
	Enabled bool `mapstructure:"enabled" yaml:"enabled"`

	// 敏感词列表
	Keywords []string `mapstructure:"keywords" yaml:"keywords"`
}

// UsageTrackingConfig 使用统计配置
type UsageTrackingConfig struct {
	// 是否启用
	Enabled bool `mapstructure:"enabled" yaml:"enabled"`

	// 是否保存对话记录
	SaveConversations bool `mapstructure:"save_conversations" yaml:"save_conversations"`
}

// CacheConfig 缓存配置
type CacheConfig struct {
	// 是否启用
	Enabled bool `mapstructure:"enabled" yaml:"enabled"`

	// 缓存时间（秒）
	TTL int `mapstructure:"ttl" yaml:"ttl"`
}

// GetProvider 获取指定提供商配置
func (c *AIConfig) GetProvider(name string) (ProviderConfig, bool) {
	provider, exists := c.Providers[name]
	return provider, exists && provider.Enabled
}

// GetDefaultProvider 获取默认提供商配置
func (c *AIConfig) GetDefaultProvider() (ProviderConfig, bool) {
	return c.GetProvider(c.DefaultProvider)
}

// GetModel 获取指定模型配置
func (p *ProviderConfig) GetModel(name string) (ModelConfig, bool) {
	for _, model := range p.Models {
		if model.Name == name {
			return model, true
		}
	}
	return ModelConfig{}, false
}

// GetDefaultModel 获取默认模型（第一个模型）
func (p *ProviderConfig) GetDefaultModel() (ModelConfig, bool) {
	if len(p.Models) > 0 {
		return p.Models[0], true
	}
	return ModelConfig{}, false
}

// IsValidProvider 检查提供商是否有效且启用
func (c *AIConfig) IsValidProvider(name string) bool {
	provider, exists := c.Providers[name]
	return exists && provider.Enabled && provider.APIKey != ""
}

// ListEnabledProviders 列出所有启用的提供商
func (c *AIConfig) ListEnabledProviders() []string {
	var providers []string
	for name, config := range c.Providers {
		if config.Enabled && config.APIKey != "" {
			providers = append(providers, name)
		}
	}
	return providers
}

// ListProviderModels 列出指定提供商的所有模型
func (c *AIConfig) ListProviderModels(providerName string) []string {
	provider, exists := c.GetProvider(providerName)
	if !exists {
		return nil
	}

	var models []string
	for _, model := range provider.Models {
		models = append(models, model.Name)
	}
	return models
}

// CalculatePrice 计算指定模型和token数的价格
func (m *ModelConfig) CalculatePrice(inputTokens, outputTokens int) float64 {
	inputPrice := float64(inputTokens) / 1000.0 * m.Pricing.Input
	outputPrice := float64(outputTokens) / 1000.0 * m.Pricing.Output
	return inputPrice + outputPrice
}
