package service

import (
	"ai-svc/internal/model"
	"ai-svc/internal/repository"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// MessageService 消息服务接口
type MessageService interface {
	SendMessageAsync(ctx context.Context, req *model.SendMessageRequest) error
	SendBroadcastMessageAsync(ctx context.Context, req *model.SendBroadcastMessageRequest) error
	GetUserMessages(userID uint, params *model.MessageQueryParams) (*model.MessageListResponse, error)
	GetUnreadCount(userID uint) (*model.UnreadCountResponse, error)
	GetMessageByID(messageID uint, userID uint) (*model.MessageResponse, error)
	MarkAsRead(messageID uint, userID uint) error
	BatchMarkAsRead(req *model.BatchReadRequest, userID uint) error
	DeleteMessage(messageID uint, userID uint) error
	StartCleanupTask()
}

// messageService 消息服务实现
type messageService struct {
	messageRepo repository.MessageRepository
	userRepo    repository.UserRepository
	logger      *logrus.Logger
}

// NewMessageService 创建消息服务实例
func NewMessageService(messageRepo repository.MessageRepository, userRepo repository.UserRepository) MessageService {
	service := &messageService{
		messageRepo: messageRepo,
		userRepo:    userRepo,
		logger:      logrus.New(),
	}

	// 启动清理任务
	service.StartCleanupTask()

	return service
}

// SendMessageAsync 异步发送单条消息
func (s *messageService) SendMessageAsync(ctx context.Context, req *model.SendMessageRequest) error {
	// 启动goroutine异步处理
	go func() {
		defer func() {
			if r := recover(); r != nil {
				s.logger.Errorf("发送消息时发生panic: %v", r)
			}
		}()

		if err := s.processSingleMessage(ctx, req); err != nil {
			s.logger.Errorf("处理消息失败: %v", err)
		}
	}()

	return nil
}

// processSingleMessage 处理单条消息
func (s *messageService) processSingleMessage(ctx context.Context, req *model.SendMessageRequest) error {
	// 1. 验证接收者是否存在
	_, err := s.userRepo.GetByID(req.RecipientID)
	if err != nil {
		return fmt.Errorf("接收者不存在: %w", err)
	}

	// 2. 创建消息定义
	definition := &model.MessageDefinition{
		Title:           req.Title,
		Content:         req.Content,
		MessageType:     req.MessageType,
		Priority:        req.Priority,
		SenderID:        req.SenderID,
		SenderType:      req.SenderType,
		IsBroadcast:     false,
		TargetUsers:     []uint{req.RecipientID},
		ExtraData:       req.ExtraData,
		ExpireAt:        req.ExpireAt,
		TotalRecipients: 1,
	}

	if err := s.messageRepo.CreateMessageDefinition(definition); err != nil {
		return fmt.Errorf("创建消息定义失败: %w", err)
	}

	// 3. 创建用户消息记录
	userMessage := &model.UserMessage{
		MessageDefinitionID: definition.ID,
		RecipientID:         req.RecipientID,
		IsRead:              false,
		IsDeleted:           false,
	}

	if err := s.messageRepo.CreateUserMessage(userMessage); err != nil {
		return fmt.Errorf("创建用户消息失败: %w", err)
	}

	s.logger.Infof("单条消息发送成功: 定义ID=%d, 用户消息ID=%d, 接收者=%d",
		definition.ID, userMessage.ID, req.RecipientID)

	return nil
}

// SendBroadcastMessageAsync 异步发送广播消息
func (s *messageService) SendBroadcastMessageAsync(ctx context.Context, req *model.SendBroadcastMessageRequest) error {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				s.logger.Errorf("发送广播消息时发生panic: %v", r)
			}
		}()

		s.processBroadcastMessage(ctx, req)
	}()

	return nil
}

// processBroadcastMessage 处理广播消息
func (s *messageService) processBroadcastMessage(ctx context.Context, req *model.SendBroadcastMessageRequest) {
	// 1. 创建消息定义
	definition := &model.MessageDefinition{
		Title:       req.Title,
		Content:     req.Content,
		MessageType: req.MessageType,
		Priority:    req.Priority,
		SenderID:    req.SenderID,
		SenderType:  req.SenderType,
		IsBroadcast: true,
		TargetUsers: req.TargetUsers,
		ExtraData:   req.ExtraData,
		ExpireAt:    req.ExpireAt,
	}

	if err := s.messageRepo.CreateMessageDefinition(definition); err != nil {
		s.logger.Errorf("创建广播消息定义失败: %v", err)
		return
	}

	// 2. 获取目标用户列表
	var targetUserIDs []uint
	var err error

	if len(req.TargetUsers) == 0 {
		// 发送给所有用户
		targetUserIDs, err = s.messageRepo.GetAllUserIDs()
		if err != nil {
			s.logger.Errorf("获取所有用户ID失败: %v", err)
			return
		}
	} else {
		targetUserIDs = req.TargetUsers
	}

	// 3. 批量创建用户消息记录
	s.batchCreateUserMessages(definition.ID, targetUserIDs)

	// 4. 更新消息定义统计信息
	if err := s.messageRepo.UpdateMessageDefinitionStats(definition.ID, len(targetUserIDs), 0); err != nil {
		s.logger.Errorf("更新消息定义统计信息失败: %v", err)
	}

	s.logger.Infof("广播消息发送完成: 定义ID=%d, 总接收者=%d", definition.ID, len(targetUserIDs))
}

// batchCreateUserMessages 批量创建用户消息记录
func (s *messageService) batchCreateUserMessages(definitionID uint, userIDs []uint) {
	// 使用工作池限制并发数
	workerCount := 10
	semaphore := make(chan struct{}, workerCount)
	var wg sync.WaitGroup

	// 分批处理，每批1000个用户
	batchSize := 1000
	totalBatches := (len(userIDs) + batchSize - 1) / batchSize

	for i := 0; i < totalBatches; i++ {
		wg.Add(1)
		go func(batchIndex int) {
			defer wg.Done()
			semaphore <- struct{}{}        // 获取信号量
			defer func() { <-semaphore }() // 释放信号量

			start := batchIndex * batchSize
			end := start + batchSize
			if end > len(userIDs) {
				end = len(userIDs)
			}

			batchUserIDs := userIDs[start:end]
			s.processUserMessageBatch(definitionID, batchUserIDs)
		}(i)
	}

	wg.Wait()
}

// processUserMessageBatch 处理一批用户消息
func (s *messageService) processUserMessageBatch(definitionID uint, userIDs []uint) {
	var userMessages []model.UserMessage

	for _, userID := range userIDs {
		userMessage := model.UserMessage{
			MessageDefinitionID: definitionID,
			RecipientID:         userID,
			IsRead:              false,
			IsDeleted:           false,
		}
		userMessages = append(userMessages, userMessage)
	}

	if err := s.messageRepo.BatchCreateUserMessages(userMessages); err != nil {
		s.logger.Errorf("批量创建用户消息失败，定义ID=%d, 用户数=%d: %v", definitionID, len(userIDs), err)
	} else {
		s.logger.Infof("批量创建用户消息成功，定义ID=%d, 用户数=%d", definitionID, len(userIDs))
	}
}

// GetUserMessages 获取用户消息列表
func (s *messageService) GetUserMessages(
	userID uint,
	params *model.MessageQueryParams,
) (*model.MessageListResponse, error) {
	// 设置默认值
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Size <= 0 {
		params.Size = 20
	}

	// 查询消息
	userMessages, total, err := s.messageRepo.GetUserMessages(
		userID,
		params.Page,
		params.Size,
		params.MessageType,
		params.IsRead,
		params.Priority,
	)
	if err != nil {
		return nil, fmt.Errorf("获取消息列表失败: %w", err)
	}

	// 转换为响应格式
	messageResponses := make([]model.MessageResponse, len(userMessages))
	for i, userMsg := range userMessages {
		messageResponses[i] = s.convertToMessageResponse(&userMsg)
	}

	// 计算分页信息
	totalPages := (total + params.Size - 1) / params.Size

	return &model.MessageListResponse{
		Messages: messageResponses,
		Pagination: model.PaginationInfo{
			Page:       params.Page,
			Size:       params.Size,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

// GetUnreadCount 获取未读消息数量
func (s *messageService) GetUnreadCount(userID uint) (*model.UnreadCountResponse, error) {
	count, err := s.messageRepo.GetUnreadCount(userID)
	if err != nil {
		return nil, fmt.Errorf("获取未读消息数量失败: %w", err)
	}

	return &model.UnreadCountResponse{
		UnreadCount: count,
	}, nil
}

// GetMessageByID 根据ID获取消息详情
func (s *messageService) GetMessageByID(messageID uint, userID uint) (*model.MessageResponse, error) {
	// 获取用户消息
	userMessages, err := s.messageRepo.GetUserMessagesByIDs([]uint{messageID}, userID)
	if err != nil {
		return nil, fmt.Errorf("获取消息失败: %w", err)
	}

	if len(userMessages) == 0 {
		return nil, fmt.Errorf("消息不存在或无权限查看")
	}

	// 转换为响应格式
	response := s.convertToMessageResponse(&userMessages[0])
	return &response, nil
}

// MarkAsRead 标记消息为已读
func (s *messageService) MarkAsRead(messageID uint, userID uint) error {
	// 验证消息是否存在且属于该用户
	userMessages, err := s.messageRepo.GetUserMessagesByIDs([]uint{messageID}, userID)
	if err != nil {
		return fmt.Errorf("获取消息失败: %w", err)
	}

	if len(userMessages) == 0 {
		return fmt.Errorf("消息不存在或无权限操作")
	}

	if err := s.messageRepo.MarkAsRead(messageID, userID); err != nil {
		return fmt.Errorf("标记已读失败: %w", err)
	}

	return nil
}

// BatchMarkAsRead 批量标记已读
func (s *messageService) BatchMarkAsRead(req *model.BatchReadRequest, userID uint) error {
	// 验证消息是否都属于该用户
	userMessages, err := s.messageRepo.GetUserMessagesByIDs(req.MessageIDs, userID)
	if err != nil {
		return fmt.Errorf("获取消息失败: %w", err)
	}

	if len(userMessages) != len(req.MessageIDs) {
		return fmt.Errorf("部分消息不存在或无权限操作")
	}

	if err := s.messageRepo.BatchMarkAsRead(req.MessageIDs, userID); err != nil {
		return fmt.Errorf("批量标记已读失败: %w", err)
	}

	return nil
}

// DeleteMessage 删除消息
func (s *messageService) DeleteMessage(messageID uint, userID uint) error {
	// 验证消息是否存在且属于该用户
	userMessages, err := s.messageRepo.GetUserMessagesByIDs([]uint{messageID}, userID)
	if err != nil {
		return fmt.Errorf("获取消息失败: %w", err)
	}

	if len(userMessages) == 0 {
		return fmt.Errorf("消息不存在或无权限操作")
	}

	if err := s.messageRepo.DeleteUserMessage(messageID, userID); err != nil {
		return fmt.Errorf("删除消息失败: %w", err)
	}

	return nil
}

// StartCleanupTask 启动清理任务
func (s *messageService) StartCleanupTask() {
	ticker := time.NewTicker(24 * time.Hour) // 每天执行一次
	go func() {
		for range ticker.C {
			s.cleanupExpiredMessages()
		}
	}()
}

// cleanupExpiredMessages 清理过期消息
func (s *messageService) cleanupExpiredMessages() {
	count, err := s.messageRepo.DeleteExpiredMessages()
	if err != nil {
		s.logger.Errorf("清理过期消息失败: %v", err)
		return
	}

	if count > 0 {
		s.logger.Infof("清理过期消息完成，共清理 %d 条消息", count)
	}
}

// convertToMessageResponse 转换为消息响应格式
func (s *messageService) convertToMessageResponse(userMessage *model.UserMessage) model.MessageResponse {
	definition := userMessage.MessageDefinition
	return model.MessageResponse{
		ID:                  userMessage.ID,
		MessageDefinitionID: userMessage.MessageDefinitionID,
		Title:               definition.Title,
		Content:             definition.Content,
		MessageType:         definition.MessageType,
		Priority:            definition.Priority,
		SenderID:            definition.SenderID,
		SenderType:          definition.SenderType,
		RecipientID:         userMessage.RecipientID,
		IsRead:              userMessage.IsRead,
		IsDeleted:           userMessage.IsDeleted,
		ReadAt:              userMessage.ReadAt,
		ExtraData:           definition.ExtraData,
		ExpireAt:            definition.ExpireAt,
		CreatedAt:           userMessage.CreatedAt,
		UpdatedAt:           userMessage.UpdatedAt,
	}
}
