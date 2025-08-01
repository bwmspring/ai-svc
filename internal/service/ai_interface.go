package service

import (
	"ai-svc/internal/model"
	"context"
	"fmt"
	"io"
)

// AIProvider AI 提供商接口 - 各个提供商的具体实现
type AIProvider interface {
	// GetName 获取提供商名称
	GetName() string

	// Chat 发送聊天请求
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)

	// ChatStream 发送流式聊天请求
	ChatStream(ctx context.Context, req *ChatRequest) (<-chan *ChatStreamResponse, error)

	// ListModels 列出可用模型
	ListModels(ctx context.Context) ([]ModelInfo, error)

	// ValidateConfig 验证配置是否有效
	ValidateConfig() error

	// Close 关闭连接
	Close() error
}

// AIService AI 服务接口 - 业务层服务
type AIService interface {
	// 对话管理
	CreateConversation(ctx context.Context, userID uint, title, provider, model string) (*model.AIConversation, error)
	GetConversation(ctx context.Context, userID uint, sessionID string) (*model.AIConversation, error)
	ListConversations(ctx context.Context, userID uint, page, size int) ([]*model.AIConversation, int64, error)
	UpdateConversation(ctx context.Context, userID uint, sessionID string, updates map[string]interface{}) error
	DeleteConversation(ctx context.Context, userID uint, sessionID string) error

	// 消息管理
	SendMessage(
		ctx context.Context,
		userID uint,
		sessionID string,
		content string,
		options *ChatOptions,
	) (*model.AIMessage, error)
	SendMessageStream(
		ctx context.Context,
		userID uint,
		sessionID string,
		content string,
		options *ChatOptions,
	) (<-chan *ChatStreamResponse, error)
	GetMessages(ctx context.Context, userID uint, sessionID string, page, size int) ([]*model.AIMessage, error)

	// 提供商管理
	ListProviders(ctx context.Context) ([]ProviderInfo, error)
	GetProvider(ctx context.Context, name string) (ProviderInfo, error)

	// 统计信息
	GetUsageStats(ctx context.Context, userID uint, startDate, endDate string) ([]*model.AIUsageStats, error)
	GetConversationStats(ctx context.Context, userID uint) (*ConversationStats, error)
}

// ChatRequest 聊天请求
type ChatRequest struct {
	// 基本信息
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`

	// 参数配置
	Temperature      *float32 `json:"temperature,omitempty"`
	MaxTokens        *int     `json:"max_tokens,omitempty"`
	TopP             *float32 `json:"top_p,omitempty"`
	FrequencyPenalty *float32 `json:"frequency_penalty,omitempty"`
	PresencePenalty  *float32 `json:"presence_penalty,omitempty"`

	// 功能开关
	Stream bool `json:"stream"`

	// 停止词
	Stop []string `json:"stop,omitempty"`

	// 用户ID（用于追踪）
	User string `json:"user,omitempty"`
}

// Message 消息结构
type Message struct {
	Role    string `json:"role"` // user, assistant, system
	Content string `json:"content"`
	Name    string `json:"name,omitempty"` // 可选的用户名
}

// ChatResponse 聊天响应
type ChatResponse struct {
	// 基本信息
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`

	// 选择结果
	Choices []Choice `json:"choices"`

	// 使用统计
	Usage model.TokenUsage `json:"usage"`

	// 提供商特定字段
	Provider string                 `json:"provider"`
	Extra    map[string]interface{} `json:"extra,omitempty"`
}

// Choice 响应选择
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// ChatStreamResponse 流式聊天响应
type ChatStreamResponse struct {
	// 基本信息
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`

	// 流式选择
	Choices []StreamChoice `json:"choices"`

	// 使用统计（仅在最后一条消息中）
	Usage *model.TokenUsage `json:"usage,omitempty"`

	// 流式控制
	Done bool `json:"done"`

	// 错误信息
	Error *APIError `json:"error,omitempty"`

	// 提供商信息
	Provider string `json:"provider"`
}

// StreamChoice 流式响应选择
type StreamChoice struct {
	Index int `json:"index"`
	Delta struct {
		Role    string `json:"role,omitempty"`
		Content string `json:"content,omitempty"`
	} `json:"delta"`
	FinishReason *string `json:"finish_reason"`
}

// ModelInfo 模型信息
type ModelInfo struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description,omitempty"`
	MaxTokens    int      `json:"max_tokens"`
	InputPrice   float64  `json:"input_price"`  // 每1K tokens价格
	OutputPrice  float64  `json:"output_price"` // 每1K tokens价格
	Provider     string   `json:"provider"`
	Capabilities []string `json:"capabilities,omitempty"`
}

// ProviderInfo 提供商信息
type ProviderInfo struct {
	Name        string      `json:"name"`
	DisplayName string      `json:"display_name"`
	Enabled     bool        `json:"enabled"`
	Models      []ModelInfo `json:"models"`
	Status      string      `json:"status"` // available, error, disabled
	LastError   string      `json:"last_error,omitempty"`
}

// ChatOptions 聊天选项
type ChatOptions struct {
	Provider     string   `json:"provider,omitempty"`
	Model        string   `json:"model,omitempty"`
	Temperature  *float32 `json:"temperature,omitempty"`
	MaxTokens    *int     `json:"max_tokens,omitempty"`
	Stream       bool     `json:"stream"`
	SystemPrompt string   `json:"system_prompt,omitempty"`
}

// ConversationStats 对话统计
type ConversationStats struct {
	TotalConversations int                       `json:"total_conversations"`
	TotalMessages      int                       `json:"total_messages"`
	TotalTokens        int                       `json:"total_tokens"`
	TotalCost          float64                   `json:"total_cost"`
	ProviderStats      map[string]*ProviderStats `json:"provider_stats"`
}

// ProviderStats 提供商统计
type ProviderStats struct {
	Conversations int     `json:"conversations"`
	Messages      int     `json:"messages"`
	Tokens        int     `json:"tokens"`
	Cost          float64 `json:"cost"`
}

// APIError API错误
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Type    string `json:"type,omitempty"`
}

// Error 实现error接口
func (e *APIError) Error() string {
	return e.Message
}

// 响应类型常量
const (
	ObjectChatCompletion      = "chat.completion"
	ObjectChatCompletionChunk = "chat.completion.chunk"
)

// 完成原因常量
const (
	FinishReasonStop          = "stop"
	FinishReasonLength        = "length"
	FinishReasonContentFilter = "content_filter"
	FinishReasonToolCalls     = "tool_calls"
	FinishReasonError         = "error"
)

// 错误代码常量
const (
	ErrorCodeInvalidRequest      = "invalid_request"
	ErrorCodeInvalidModel        = "invalid_model"
	ErrorCodeInvalidProvider     = "invalid_provider"
	ErrorCodeQuotaExceeded       = "quota_exceeded"
	ErrorCodeRateLimitExceeded   = "rate_limit_exceeded"
	ErrorCodeProviderError       = "provider_error"
	ErrorCodeNetworkError        = "network_error"
	ErrorCodeAuthenticationError = "authentication_error"
)

// 提供商状态常量
const (
	ProviderStatusAvailable = "available"
	ProviderStatusError     = "error"
	ProviderStatusDisabled  = "disabled"
)

// Helper 函数

// NewChatRequest 创建聊天请求
func NewChatRequest(model string, messages []Message) *ChatRequest {
	return &ChatRequest{
		Model:    model,
		Messages: messages,
		Stream:   false,
	}
}

// NewStreamChatRequest 创建流式聊天请求
func NewStreamChatRequest(model string, messages []Message) *ChatRequest {
	req := NewChatRequest(model, messages)
	req.Stream = true
	return req
}

// AddMessage 添加消息
func (r *ChatRequest) AddMessage(role, content string) {
	r.Messages = append(r.Messages, Message{
		Role:    role,
		Content: content,
	})
}

// SetParameters 设置参数
func (r *ChatRequest) SetParameters(temperature *float32, maxTokens *int) {
	r.Temperature = temperature
	r.MaxTokens = maxTokens
}

// ValidateMessages 验证消息
func (r *ChatRequest) ValidateMessages() error {
	if len(r.Messages) == 0 {
		return &APIError{
			Code:    ErrorCodeInvalidRequest,
			Message: "消息列表不能为空",
		}
	}

	for i, msg := range r.Messages {
		if msg.Role == "" {
			return &APIError{
				Code:    ErrorCodeInvalidRequest,
				Message: fmt.Sprintf("第%d条消息缺少角色", i+1),
			}
		}
		if msg.Content == "" {
			return &APIError{
				Code:    ErrorCodeInvalidRequest,
				Message: fmt.Sprintf("第%d条消息内容为空", i+1),
			}
		}
	}

	return nil
}

// GetLastAssistantMessage 获取最后一条助手消息
func (r *ChatResponse) GetLastAssistantMessage() string {
	if len(r.Choices) > 0 {
		return r.Choices[0].Message.Content
	}
	return ""
}

// IsSuccess 检查响应是否成功
func (r *ChatStreamResponse) IsSuccess() bool {
	return r.Error == nil
}

// GetContent 获取内容
func (r *ChatStreamResponse) GetContent() string {
	if len(r.Choices) > 0 {
		return r.Choices[0].Delta.Content
	}
	return ""
}

// StreamReader 流式读取器接口
type StreamReader interface {
	io.Reader
	io.Closer
}

// ProviderFactory 提供商工厂接口
type ProviderFactory interface {
	CreateProvider(config interface{}) (AIProvider, error)
	GetProviderType() string
}

// 工具函数

// FormatTokenUsage 格式化token使用量
func FormatTokenUsage(usage model.TokenUsage) string {
	return fmt.Sprintf("输入: %d tokens, 输出: %d tokens, 总计: %d tokens",
		usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens)
}

// FormatCost 格式化费用
func FormatCost(cost float64) string {
	return fmt.Sprintf("$%.6f", cost)
}
