package model

import (
	"time"
)

// AIConversation AI 对话记录
type AIConversation struct {
	BaseModel

	// 用户信息
	UserID uint  `gorm:"not null;index"    json:"user_id"`
	User   *User `gorm:"foreignKey:UserID" json:"user,omitempty"`

	// 对话基本信息
	Title     string `gorm:"type:varchar(255);not null"             json:"title"`      // 对话标题
	Provider  string `gorm:"type:varchar(50);not null;index"        json:"provider"`   // AI 提供商
	Model     string `gorm:"type:varchar(100);not null"             json:"model"`      // 使用的模型
	SessionID string `gorm:"type:varchar(100);not null;uniqueIndex" json:"session_id"` // 会话ID

	// 对话状态
	Status   string `gorm:"type:varchar(20);not null;default:'active'" json:"status"`    // 状态：active, archived, deleted
	IsPublic bool   `gorm:"default:false"                              json:"is_public"` // 是否公开

	// 统计信息
	MessageCount int     `gorm:"default:0"                    json:"message_count"` // 消息数量
	TotalTokens  int     `gorm:"default:0"                    json:"total_tokens"`  // 总token数
	TotalCost    float64 `gorm:"type:decimal(10,6);default:0" json:"total_cost"`    // 总费用

	// 配置信息
	Temperature float32 `gorm:"type:decimal(3,2);default:0.7" json:"temperature"` // 温度参数
	MaxTokens   int     `gorm:"default:4096"                  json:"max_tokens"`  // 最大token数

	// 时间信息
	LastMessageAt *time.Time `gorm:"index" json:"last_message_at"` // 最后消息时间

	// 关联数据
	Messages []AIMessage `gorm:"foreignKey:ConversationID" json:"messages,omitempty"`
}

// AIMessage AI 对话消息
type AIMessage struct {
	BaseModel

	// 关联信息
	ConversationID uint            `gorm:"not null;index"            json:"conversation_id"`
	Conversation   *AIConversation `gorm:"foreignKey:ConversationID" json:"conversation,omitempty"`

	// 消息基本信息
	Role        string `gorm:"type:varchar(20);not null"       json:"role"`         // 角色：user, assistant, system
	Content     string `gorm:"type:longtext;not null"          json:"content"`      // 消息内容
	ContentType string `gorm:"type:varchar(20);default:'text'" json:"content_type"` // 内容类型：text, image, file

	// 消息状态
	Status string `gorm:"type:varchar(20);default:'sent'" json:"status"` // 状态：sent, received, error

	// Token 统计
	PromptTokens     int `gorm:"default:0" json:"prompt_tokens"`     // 输入token数
	CompletionTokens int `gorm:"default:0" json:"completion_tokens"` // 输出token数
	TotalTokens      int `gorm:"default:0" json:"total_tokens"`      // 总token数

	// 费用信息
	Cost float64 `gorm:"type:decimal(10,6);default:0" json:"cost"` // 本条消息费用

	// 技术信息
	Provider     string  `gorm:"type:varchar(50)"  json:"provider"`      // AI 提供商
	Model        string  `gorm:"type:varchar(100)" json:"model"`         // 使用的模型
	Temperature  float32 `gorm:"type:decimal(3,2)" json:"temperature"`   // 温度参数
	FinishReason string  `gorm:"type:varchar(50)"  json:"finish_reason"` // 完成原因

	// 响应时间
	ResponseTime int `gorm:"default:0" json:"response_time"` // 响应时间（毫秒）

	// 元数据
	Metadata string `gorm:"type:json" json:"metadata,omitempty"` // 额外元数据（JSON格式）
}

// AIUsageStats AI 使用统计
type AIUsageStats struct {
	BaseModel

	// 统计维度
	UserID   uint   `gorm:"not null;index"                   json:"user_id"`
	Provider string `gorm:"type:varchar(50);not null;index"  json:"provider"`
	Model    string `gorm:"type:varchar(100);not null;index" json:"model"`
	Date     string `gorm:"type:date;not null;index"         json:"date"` // 统计日期 YYYY-MM-DD

	// 统计数据
	RequestCount     int     `gorm:"default:0"                    json:"request_count"`     // 请求次数
	MessageCount     int     `gorm:"default:0"                    json:"message_count"`     // 消息数量
	PromptTokens     int     `gorm:"default:0"                    json:"prompt_tokens"`     // 输入token数
	CompletionTokens int     `gorm:"default:0"                    json:"completion_tokens"` // 输出token数
	TotalTokens      int     `gorm:"default:0"                    json:"total_tokens"`      // 总token数
	TotalCost        float64 `gorm:"type:decimal(10,6);default:0" json:"total_cost"`        // 总费用

	// 性能数据
	AvgResponseTime int `gorm:"default:0" json:"avg_response_time"` // 平均响应时间（毫秒）
	ErrorCount      int `gorm:"default:0" json:"error_count"`       // 错误次数

	// 组合索引
	_ struct{} `gorm:"uniqueIndex:idx_usage_unique,user_id,provider,model,date"`
}

// AIProviderConfig AI 提供商配置（运行时配置缓存）
type AIProviderConfig struct {
	BaseModel

	// 基本信息
	Name        string `gorm:"type:varchar(50);not null;uniqueIndex" json:"name"`
	DisplayName string `gorm:"type:varchar(100);not null"            json:"display_name"`
	Enabled     bool   `gorm:"default:true"                          json:"enabled"`

	// 配置信息
	BaseURL      string `gorm:"type:varchar(255)" json:"base_url"`
	APIKey       string `gorm:"type:varchar(255)" json:"api_key"`
	Organization string `gorm:"type:varchar(255)" json:"organization,omitempty"`
	SecretKey    string `gorm:"type:varchar(255)" json:"secret_key,omitempty"`
	Version      string `gorm:"type:varchar(50)"  json:"version,omitempty"`
	Region       string `gorm:"type:varchar(50)"  json:"region,omitempty"`

	// 限制配置
	MaxTokens   int     `gorm:"default:4096"                  json:"max_tokens"`
	Temperature float32 `gorm:"type:decimal(3,2);default:0.7" json:"temperature"`
	RateLimit   int     `gorm:"default:60"                    json:"rate_limit"` // 每分钟请求限制

	// 定价信息
	InputPrice  float64 `gorm:"type:decimal(10,6);default:0" json:"input_price"`  // 输入价格（每1K tokens）
	OutputPrice float64 `gorm:"type:decimal(10,6);default:0" json:"output_price"` // 输出价格（每1K tokens）

	// 状态信息
	LastUsedAt   *time.Time `json:"last_used_at,omitempty"`
	ErrorCount   int        `json:"error_count"            gorm:"default:0"`
	SuccessCount int        `json:"success_count"          gorm:"default:0"`
}

// ConversationStatus 对话状态常量
const (
	ConversationStatusActive   = "active"
	ConversationStatusArchived = "archived"
	ConversationStatusDeleted  = "deleted"
)

// MessageRole 消息角色常量
const (
	MessageRoleUser      = "user"
	MessageRoleAssistant = "assistant"
	MessageRoleSystem    = "system"
)

// MessageStatus 消息状态常量
const (
	MessageStatusSent     = "sent"
	MessageStatusReceived = "received"
	MessageStatusError    = "error"
)

// ContentType 内容类型常量
const (
	ContentTypeText  = "text"
	ContentTypeImage = "image"
	ContentTypeFile  = "file"
)

// TableName 指定表名
func (AIConversation) TableName() string {
	return "ai_conversations"
}

func (AIMessage) TableName() string {
	return "ai_messages"
}

func (AIUsageStats) TableName() string {
	return "ai_usage_stats"
}

func (AIProviderConfig) TableName() string {
	return "ai_provider_configs"
}

// 模型方法

// AddMessage 添加消息到对话
func (c *AIConversation) AddMessage(role, content string) *AIMessage {
	return &AIMessage{
		ConversationID: c.ID,
		Role:           role,
		Content:        content,
		Status:         MessageStatusSent,
		Provider:       c.Provider,
		Model:          c.Model,
		Temperature:    c.Temperature,
	}
}

// UpdateStats 更新对话统计
func (c *AIConversation) UpdateStats(tokenUsage TokenUsage, cost float64) {
	c.MessageCount++
	c.TotalTokens += tokenUsage.TotalTokens
	c.TotalCost += cost
	now := time.Now()
	c.LastMessageAt = &now
}

// TokenUsage Token 使用量结构
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// CalculateCost 计算费用
func (t *TokenUsage) CalculateCost(inputPrice, outputPrice float64) float64 {
	inputCost := float64(t.PromptTokens) / 1000.0 * inputPrice
	outputCost := float64(t.CompletionTokens) / 1000.0 * outputPrice
	return inputCost + outputCost
}

// IsExpired 检查配置是否过期（超过24小时未使用）
func (p *AIProviderConfig) IsExpired() bool {
	if p.LastUsedAt == nil {
		return false
	}
	return time.Since(*p.LastUsedAt) > 24*time.Hour
}

// UpdateUsage 更新使用统计
func (p *AIProviderConfig) UpdateUsage(success bool) {
	now := time.Now()
	p.LastUsedAt = &now
	if success {
		p.SuccessCount++
	} else {
		p.ErrorCount++
	}
}
