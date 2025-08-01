package controller

import (
	"ai-svc/internal/middleware"
	"ai-svc/internal/model"
	"ai-svc/internal/service"
	"ai-svc/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// MessageController 消息控制器
type MessageController struct {
	messageService service.MessageService
}

// NewMessageController 创建消息控制器实例
func NewMessageController(messageService service.MessageService) *MessageController {
	return &MessageController{
		messageService: messageService,
	}
}

// SendMessage 发送单条消息
func (c *MessageController) SendMessage(ctx *gin.Context) {
	var req model.SendMessageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, response.INVALID_PARAMS, "请求参数错误")
		return
	}

	// 从JWT中获取发送者信息
	userID := middleware.GetCurrentUserID(ctx)
	req.SenderID = userID
	req.SenderType = 1 // 用户类型

	// 异步发送消息
	if err := c.messageService.SendMessageAsync(ctx, &req); err != nil {
		response.Error(ctx, response.ERROR, "消息发送失败")
		return
	}

	response.Success(ctx, gin.H{
		"status":  "processing",
		"message": "消息发送任务已提交",
	})
}

// SendBroadcastMessage 发送广播消息
func (c *MessageController) SendBroadcastMessage(ctx *gin.Context) {
	var req model.SendBroadcastMessageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, response.INVALID_PARAMS, "请求参数错误")
		return
	}

	// 从JWT中获取发送者信息
	userID := middleware.GetCurrentUserID(ctx)
	req.SenderID = userID
	req.SenderType = 1 // 用户类型

	// 异步发送广播消息
	if err := c.messageService.SendBroadcastMessageAsync(ctx, &req); err != nil {
		response.Error(ctx, response.ERROR, "广播消息发送失败")
		return
	}

	// 估算处理时间
	estimatedTime := "1-5分钟"
	if len(req.TargetUsers) > 1000000 {
		estimatedTime = "10-30分钟"
	} else if len(req.TargetUsers) > 100000 {
		estimatedTime = "5-15分钟"
	}

	response.Success(ctx, gin.H{
		"status":             "processing",
		"message":            "广播消息发送任务已提交",
		"estimated_time":     estimatedTime,
		"target_users_count": len(req.TargetUsers),
	})
}

// GetMessages 获取消息列表
func (c *MessageController) GetMessages(ctx *gin.Context) {
	userID := middleware.GetCurrentUserID(ctx)

	// 解析查询参数
	var params model.MessageQueryParams
	if err := ctx.ShouldBindQuery(&params); err != nil {
		response.Error(ctx, response.INVALID_PARAMS, "查询参数错误")
		return
	}

	// 获取消息列表
	result, err := c.messageService.GetUserMessages(userID, &params)
	if err != nil {
		response.Error(ctx, response.ERROR, "获取消息列表失败")
		return
	}

	response.Success(ctx, result)
}

// GetUnreadCount 获取未读消息数量
func (c *MessageController) GetUnreadCount(ctx *gin.Context) {
	userID := middleware.GetCurrentUserID(ctx)

	result, err := c.messageService.GetUnreadCount(userID)
	if err != nil {
		response.Error(ctx, response.ERROR, "获取未读消息数量失败")
		return
	}

	response.Success(ctx, result)
}

// MarkAsRead 标记消息为已读
func (c *MessageController) MarkAsRead(ctx *gin.Context) {
	userID := middleware.GetCurrentUserID(ctx)

	// 获取消息ID
	messageIDStr := ctx.Param("id")
	messageID, err := strconv.ParseUint(messageIDStr, 10, 32)
	if err != nil {
		response.Error(ctx, response.INVALID_PARAMS, "消息ID格式错误")
		return
	}

	if err := c.messageService.MarkAsRead(uint(messageID), userID); err != nil {
		response.Error(ctx, response.ERROR, "标记已读失败")
		return
	}

	response.Success(ctx, gin.H{
		"message": "消息已标记为已读",
	})
}

// BatchMarkAsRead 批量标记已读
func (c *MessageController) BatchMarkAsRead(ctx *gin.Context) {
	userID := middleware.GetCurrentUserID(ctx)

	var req model.BatchReadRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, response.INVALID_PARAMS, "请求参数错误")
		return
	}

	if err := c.messageService.BatchMarkAsRead(&req, userID); err != nil {
		response.Error(ctx, response.ERROR, "批量标记已读失败")
		return
	}

	response.Success(ctx, gin.H{
		"message": "批量标记成功",
		"count":   len(req.MessageIDs),
	})
}

// DeleteMessage 删除消息
func (c *MessageController) DeleteMessage(ctx *gin.Context) {
	userID := middleware.GetCurrentUserID(ctx)

	// 获取消息ID
	messageIDStr := ctx.Param("id")
	messageID, err := strconv.ParseUint(messageIDStr, 10, 32)
	if err != nil {
		response.Error(ctx, response.INVALID_PARAMS, "消息ID格式错误")
		return
	}

	if err := c.messageService.DeleteMessage(uint(messageID), userID); err != nil {
		response.Error(ctx, response.ERROR, "删除消息失败")
		return
	}

	response.Success(ctx, gin.H{
		"message": "消息已删除",
	})
}

// GetMessageDetail 获取消息详情
func (c *MessageController) GetMessageDetail(ctx *gin.Context) {
	userID := middleware.GetCurrentUserID(ctx)

	// 获取消息ID
	messageIDStr := ctx.Param("id")
	messageID, err := strconv.ParseUint(messageIDStr, 10, 32)
	if err != nil {
		response.Error(ctx, response.INVALID_PARAMS, "消息ID格式错误")
		return
	}

	// 获取消息详情
	message, err := c.messageService.GetMessageByID(uint(messageID), userID)
	if err != nil {
		response.Error(ctx, response.ERROR, "获取消息详情失败")
		return
	}

	response.Success(ctx, message)
}
