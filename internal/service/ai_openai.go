package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"ai-svc/internal/config"
	"ai-svc/internal/model"
)

// OpenAIProvider OpenAI 提供商实现
type OpenAIProvider struct {
	config     config.ProviderConfig
	httpClient *http.Client
	baseURL    string
	apiKey     string
}

// NewOpenAIProvider 创建 OpenAI 提供商
func NewOpenAIProvider(cfg config.ProviderConfig) *OpenAIProvider {
	return &OpenAIProvider{
		config:  cfg,
		baseURL: cfg.BaseURL,
		apiKey:  cfg.APIKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GetName 获取提供商名称
func (p *OpenAIProvider) GetName() string {
	return "openai"
}

// Chat 发送聊天请求
func (p *OpenAIProvider) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	// 构建 OpenAI API 请求
	openaiReq := &OpenAIChatRequest{
		Model:       req.Model,
		Messages:    convertToOpenAIMessages(req.Messages),
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		TopP:        req.TopP,
		Stop:        req.Stop,
		User:        req.User,
		Stream:      false,
	}

	// 序列化请求
	reqBody, err := json.Marshal(openaiReq)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	// 创建 HTTP 请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	// 设置请求头
	p.setHeaders(httpReq)

	// 发送请求
	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return nil, p.handleErrorResponse(resp)
	}

	// 解析响应
	var openaiResp OpenAIChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&openaiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 转换为标准响应
	return p.convertToStandardResponse(&openaiResp), nil
}

// ChatStream 发送流式聊天请求
func (p *OpenAIProvider) ChatStream(ctx context.Context, req *ChatRequest) (<-chan *ChatStreamResponse, error) {
	// 构建流式请求
	openaiReq := &OpenAIChatRequest{
		Model:       req.Model,
		Messages:    convertToOpenAIMessages(req.Messages),
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		TopP:        req.TopP,
		Stop:        req.Stop,
		User:        req.User,
		Stream:      true,
	}

	// 序列化请求
	reqBody, err := json.Marshal(openaiReq)
	if err != nil {
		ch := make(chan *ChatStreamResponse, 1)
		ch <- &ChatStreamResponse{
			Error: &APIError{
				Code:    ErrorCodeInvalidRequest,
				Message: fmt.Sprintf("序列化请求失败: %v", err),
			},
		}
		close(ch)
		return ch, nil
	}

	// 创建 HTTP 请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		ch := make(chan *ChatStreamResponse, 1)
		ch <- &ChatStreamResponse{
			Error: &APIError{
				Code:    ErrorCodeNetworkError,
				Message: fmt.Sprintf("创建HTTP请求失败: %v", err),
			},
		}
		close(ch)
		return ch, nil
	}

	// 设置请求头
	p.setHeaders(httpReq)

	// 发送请求
	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		ch := make(chan *ChatStreamResponse, 1)
		ch <- &ChatStreamResponse{
			Error: &APIError{
				Code:    ErrorCodeNetworkError,
				Message: fmt.Sprintf("发送请求失败: %v", err),
			},
		}
		close(ch)
		return ch, nil
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		ch := make(chan *ChatStreamResponse, 1)
		ch <- &ChatStreamResponse{
			Error: &APIError{
				Code:    ErrorCodeProviderError,
				Message: fmt.Sprintf("HTTP错误: %d", resp.StatusCode),
			},
		}
		close(ch)
		return ch, nil
	}

	// 创建流式响应通道
	ch := make(chan *ChatStreamResponse, 10)

	// 启动goroutine处理流式响应
	go p.handleStreamResponse(ctx, resp.Body, ch)

	return ch, nil
}

// ListModels 列出可用模型
func (p *OpenAIProvider) ListModels(ctx context.Context) ([]ModelInfo, error) {
	// 从配置中获取模型信息
	var models []ModelInfo
	for _, modelCfg := range p.config.Models {
		models = append(models, ModelInfo{
			ID:          modelCfg.Name,
			Name:        modelCfg.Name,
			MaxTokens:   modelCfg.MaxTokens,
			InputPrice:  modelCfg.Pricing.Input,
			OutputPrice: modelCfg.Pricing.Output,
			Provider:    p.GetName(),
		})
	}
	return models, nil
}

// ValidateConfig 验证配置是否有效
func (p *OpenAIProvider) ValidateConfig() error {
	if p.apiKey == "" {
		return fmt.Errorf("OpenAI API密钥不能为空")
	}
	if p.baseURL == "" {
		return fmt.Errorf("OpenAI API基础URL不能为空")
	}
	return nil
}

// Close 关闭连接
func (p *OpenAIProvider) Close() error {
	// OpenAI HTTP客户端不需要显式关闭
	return nil
}

// 私有方法

// setHeaders 设置请求头
func (p *OpenAIProvider) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	if p.config.Organization != "" {
		req.Header.Set("OpenAI-Organization", p.config.Organization)
	}

	req.Header.Set("User-Agent", "ai-svc/1.0")
}

// handleErrorResponse 处理错误响应
func (p *OpenAIProvider) handleErrorResponse(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取错误响应失败: %w", err)
	}

	var errorResp OpenAIErrorResponse
	if err := json.Unmarshal(body, &errorResp); err != nil {
		return fmt.Errorf("HTTP错误 %d: %s", resp.StatusCode, string(body))
	}

	return &APIError{
		Code:    mapOpenAIErrorCode(errorResp.Error.Type),
		Message: errorResp.Error.Message,
		Type:    errorResp.Error.Type,
	}
}

// handleStreamResponse 处理流式响应
func (p *OpenAIProvider) handleStreamResponse(ctx context.Context, body io.ReadCloser, ch chan<- *ChatStreamResponse) {
	defer close(ch)
	defer body.Close()

	scanner := bufio.NewScanner(body)
	var totalUsage *model.TokenUsage

	for scanner.Scan() {
		line := scanner.Text()

		// 跳过空行和非数据行
		if line == "" || !strings.HasPrefix(line, "data: ") {
			continue
		}

		// 移除 "data: " 前缀
		data := strings.TrimPrefix(line, "data: ")

		// 检查结束标志
		if data == "[DONE]" {
			if totalUsage != nil {
				ch <- &ChatStreamResponse{
					Usage:    totalUsage,
					Done:     true,
					Provider: p.GetName(),
				}
			}
			break
		}

		// 解析JSON数据
		var chunk OpenAIStreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			ch <- &ChatStreamResponse{
				Error: &APIError{
					Code:    ErrorCodeProviderError,
					Message: fmt.Sprintf("解析流式数据失败: %v", err),
				},
				Provider: p.GetName(),
			}
			continue
		}

		// 转换为标准响应
		streamResp := p.convertToStreamResponse(&chunk)

		// 保存usage信息
		if chunk.Usage != nil {
			totalUsage = &model.TokenUsage{
				PromptTokens:     chunk.Usage.PromptTokens,
				CompletionTokens: chunk.Usage.CompletionTokens,
				TotalTokens:      chunk.Usage.TotalTokens,
			}
		}

		ch <- streamResp
	}

	if err := scanner.Err(); err != nil {
		ch <- &ChatStreamResponse{
			Error: &APIError{
				Code:    ErrorCodeNetworkError,
				Message: fmt.Sprintf("读取流式数据失败: %v", err),
			},
			Provider: p.GetName(),
		}
	}
}

// convertToStandardResponse 转换为标准响应
func (p *OpenAIProvider) convertToStandardResponse(resp *OpenAIChatResponse) *ChatResponse {
	choices := make([]Choice, len(resp.Choices))
	for i, choice := range resp.Choices {
		choices[i] = Choice{
			Index: choice.Index,
			Message: Message{
				Role:    choice.Message.Role,
				Content: choice.Message.Content,
			},
			FinishReason: choice.FinishReason,
		}
	}

	return &ChatResponse{
		ID:      resp.ID,
		Object:  resp.Object,
		Created: resp.Created,
		Model:   resp.Model,
		Choices: choices,
		Usage: model.TokenUsage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
		Provider: p.GetName(),
	}
}

// convertToStreamResponse 转换为流式响应
func (p *OpenAIProvider) convertToStreamResponse(chunk *OpenAIStreamChunk) *ChatStreamResponse {
	choices := make([]StreamChoice, len(chunk.Choices))
	for i, choice := range chunk.Choices {
		choices[i] = StreamChoice{
			Index: choice.Index,
			Delta: struct {
				Role    string `json:"role,omitempty"`
				Content string `json:"content,omitempty"`
			}{
				Role:    choice.Delta.Role,
				Content: choice.Delta.Content,
			},
			FinishReason: choice.FinishReason,
		}
	}

	return &ChatStreamResponse{
		ID:       chunk.ID,
		Object:   chunk.Object,
		Created:  chunk.Created,
		Model:    chunk.Model,
		Choices:  choices,
		Provider: p.GetName(),
	}
}

// OpenAI API 结构体定义

// OpenAIChatRequest OpenAI 聊天请求
type OpenAIChatRequest struct {
	Model            string          `json:"model"`
	Messages         []OpenAIMessage `json:"messages"`
	Temperature      *float32        `json:"temperature,omitempty"`
	MaxTokens        *int            `json:"max_tokens,omitempty"`
	TopP             *float32        `json:"top_p,omitempty"`
	FrequencyPenalty *float32        `json:"frequency_penalty,omitempty"`
	PresencePenalty  *float32        `json:"presence_penalty,omitempty"`
	Stop             []string        `json:"stop,omitempty"`
	User             string          `json:"user,omitempty"`
	Stream           bool            `json:"stream"`
}

// OpenAIMessage OpenAI 消息
type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
}

// OpenAIChatResponse OpenAI 聊天响应
type OpenAIChatResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []OpenAIChoice `json:"choices"`
	Usage   OpenAIUsage    `json:"usage"`
}

// OpenAIChoice OpenAI 选择
type OpenAIChoice struct {
	Index        int           `json:"index"`
	Message      OpenAIMessage `json:"message"`
	FinishReason string        `json:"finish_reason"`
}

// OpenAIUsage OpenAI 使用量
type OpenAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// OpenAIStreamChunk OpenAI 流式块
type OpenAIStreamChunk struct {
	ID      string               `json:"id"`
	Object  string               `json:"object"`
	Created int64                `json:"created"`
	Model   string               `json:"model"`
	Choices []OpenAIStreamChoice `json:"choices"`
	Usage   *OpenAIUsage         `json:"usage,omitempty"`
}

// OpenAIStreamChoice OpenAI 流式选择
type OpenAIStreamChoice struct {
	Index int `json:"index"`
	Delta struct {
		Role    string `json:"role,omitempty"`
		Content string `json:"content,omitempty"`
	} `json:"delta"`
	FinishReason *string `json:"finish_reason"`
}

// OpenAIErrorResponse OpenAI 错误响应
type OpenAIErrorResponse struct {
	Error OpenAIError `json:"error"`
}

// OpenAIError OpenAI 错误
type OpenAIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// 辅助函数

// convertToOpenAIMessages 转换为 OpenAI 消息格式
func convertToOpenAIMessages(messages []Message) []OpenAIMessage {
	openaiMessages := make([]OpenAIMessage, len(messages))
	for i, msg := range messages {
		openaiMessages[i] = OpenAIMessage{
			Role:    msg.Role,
			Content: msg.Content,
			Name:    msg.Name,
		}
	}
	return openaiMessages
}

// mapOpenAIErrorCode 映射 OpenAI 错误代码
func mapOpenAIErrorCode(errorType string) string {
	switch errorType {
	case "invalid_request_error":
		return ErrorCodeInvalidRequest
	case "authentication_error":
		return ErrorCodeAuthenticationError
	case "rate_limit_exceeded":
		return ErrorCodeRateLimitExceeded
	case "quota_exceeded":
		return ErrorCodeQuotaExceeded
	default:
		return ErrorCodeProviderError
	}
}
