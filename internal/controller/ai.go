package controller

import (
	"ai-svc/internal/config"
	"ai-svc/internal/service"
	"ai-svc/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// AIController AI 控制器
type AIController struct {
	aiService service.AIService
	config    *config.AIConfig
}

// NewAIController 创建 AI 控制器
func NewAIController(aiService service.AIService, config *config.AIConfig) *AIController {
	return &AIController{
		aiService: aiService,
		config:    config,
	}
}

// ChatRequest 聊天请求
type ChatRequest struct {
	Message   string                 `json:"message"              binding:"required"`
	SessionID string                 `json:"session_id,omitempty"`
	Provider  string                 `json:"provider,omitempty"`
	Model     string                 `json:"model,omitempty"`
	Stream    bool                   `json:"stream,omitempty"`
	Options   map[string]interface{} `json:"options,omitempty"`
}

// CreateConversationRequest 创建对话请求
type CreateConversationRequest struct {
	Title    string `json:"title"              binding:"required"`
	Provider string `json:"provider,omitempty"`
	Model    string `json:"model,omitempty"`
}

// Chat 发送聊天消息
func (c *AIController) Chat(ctx *gin.Context) {
	var req ChatRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, response.INVALID_PARAMS, "请求参数错误: "+err.Error())
		return
	}

	// 获取用户ID
	userID := getUserID(ctx)
	if userID == 0 {
		response.Error(ctx, response.UNAUTHORIZED, "用户未登录")
		return
	}

	// 设置默认提供商和模型
	if req.Provider == "" {
		req.Provider = c.config.DefaultProvider
	}

	// 构建聊天选项
	options := &service.ChatOptions{
		Provider: req.Provider,
		Model:    req.Model,
		Stream:   req.Stream,
	}

	// 处理流式响应
	if req.Stream {
		c.handleStreamChat(ctx, userID, req.SessionID, req.Message, options)
		return
	}

	// 发送消息
	message, err := c.aiService.SendMessage(ctx, userID, req.SessionID, req.Message, options)
	if err != nil {
		response.Error(ctx, response.ERROR, "发送消息失败: "+err.Error())
		return
	}

	response.Success(ctx, gin.H{
		"message": message,
	})
}

// handleStreamChat 处理流式聊天
func (c *AIController) handleStreamChat(
	ctx *gin.Context,
	userID uint,
	sessionID, content string,
	options *service.ChatOptions,
) {
	// 设置SSE响应头
	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("Access-Control-Allow-Origin", "*")

	// 获取流式响应
	stream, err := c.aiService.SendMessageStream(ctx, userID, sessionID, content, options)
	if err != nil {
		ctx.SSEvent("error", gin.H{"error": err.Error()})
		return
	}

	// 处理流式数据
	for chunk := range stream {
		if chunk.Error != nil {
			ctx.SSEvent("error", gin.H{"error": chunk.Error.Message})
			break
		}

		ctx.SSEvent("data", chunk)
		ctx.Writer.Flush()

		if chunk.Done {
			break
		}
	}

	ctx.SSEvent("done", gin.H{"message": "流式响应完成"})
}

// CreateConversation 创建对话
func (c *AIController) CreateConversation(ctx *gin.Context) {
	var req CreateConversationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, response.INVALID_PARAMS, "请求参数错误: "+err.Error())
		return
	}

	userID := getUserID(ctx)
	if userID == 0 {
		response.Error(ctx, response.UNAUTHORIZED, "用户未登录")
		return
	}

	// 设置默认值
	if req.Provider == "" {
		req.Provider = c.config.DefaultProvider
	}

	conversation, err := c.aiService.CreateConversation(ctx, userID, req.Title, req.Provider, req.Model)
	if err != nil {
		response.Error(ctx, response.ERROR, "创建对话失败: "+err.Error())
		return
	}

	response.Success(ctx, conversation)
}

// GetConversations 获取对话列表
func (c *AIController) GetConversations(ctx *gin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		response.Error(ctx, response.UNAUTHORIZED, "用户未登录")
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(ctx.DefaultQuery("size", "20"))

	conversations, total, err := c.aiService.ListConversations(ctx, userID, page, size)
	if err != nil {
		response.Error(ctx, response.ERROR, "获取对话列表失败: "+err.Error())
		return
	}

	response.Page(ctx, conversations, total, page, size)
}

// GetConversation 获取对话详情
func (c *AIController) GetConversation(ctx *gin.Context) {
	sessionID := ctx.Param("session_id")
	if sessionID == "" {
		response.Error(ctx, response.INVALID_PARAMS, "会话ID不能为空")
		return
	}

	userID := getUserID(ctx)
	if userID == 0 {
		response.Error(ctx, response.UNAUTHORIZED, "用户未登录")
		return
	}

	conversation, err := c.aiService.GetConversation(ctx, userID, sessionID)
	if err != nil {
		response.Error(ctx, response.ERROR, "获取对话失败: "+err.Error())
		return
	}

	response.Success(ctx, conversation)
}

// GetMessages 获取消息列表
func (c *AIController) GetMessages(ctx *gin.Context) {
	sessionID := ctx.Param("session_id")
	if sessionID == "" {
		response.Error(ctx, response.INVALID_PARAMS, "会话ID不能为空")
		return
	}

	userID := getUserID(ctx)
	if userID == 0 {
		response.Error(ctx, response.UNAUTHORIZED, "用户未登录")
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(ctx.DefaultQuery("size", "50"))

	messages, err := c.aiService.GetMessages(ctx, userID, sessionID, page, size)
	if err != nil {
		response.Error(ctx, response.ERROR, "获取消息列表失败: "+err.Error())
		return
	}

	response.Success(ctx, messages)
}

// ListProviders 获取提供商列表
func (c *AIController) ListProviders(ctx *gin.Context) {
	providers, err := c.aiService.ListProviders(ctx)
	if err != nil {
		response.Error(ctx, response.ERROR, "获取提供商列表失败: "+err.Error())
		return
	}

	response.Success(ctx, providers)
}

// GetUsageStats 获取使用统计
func (c *AIController) GetUsageStats(ctx *gin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		response.Error(ctx, response.UNAUTHORIZED, "用户未登录")
		return
	}

	startDate := ctx.Query("start_date")
	endDate := ctx.Query("end_date")

	stats, err := c.aiService.GetUsageStats(ctx, userID, startDate, endDate)
	if err != nil {
		response.Error(ctx, response.ERROR, "获取使用统计失败: "+err.Error())
		return
	}

	response.Success(ctx, stats)
}

// DeleteConversation 删除对话
func (c *AIController) DeleteConversation(ctx *gin.Context) {
	sessionID := ctx.Param("session_id")
	if sessionID == "" {
		response.Error(ctx, response.INVALID_PARAMS, "会话ID不能为空")
		return
	}

	userID := getUserID(ctx)
	if userID == 0 {
		response.Error(ctx, response.UNAUTHORIZED, "用户未登录")
		return
	}

	err := c.aiService.DeleteConversation(ctx, userID, sessionID)
	if err != nil {
		response.Error(ctx, response.ERROR, "删除对话失败: "+err.Error())
		return
	}

	response.Success(ctx, gin.H{"message": "删除成功"})
}

// 辅助函数

// getUserID 从上下文获取用户ID
func getUserID(ctx *gin.Context) uint {
	if userID, exists := ctx.Get("user_id"); exists {
		if id, ok := userID.(uint); ok {
			return id
		}
	}
	return 0
}
