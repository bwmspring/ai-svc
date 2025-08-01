package repository

import (
	"ai-svc/internal/model"
	"ai-svc/pkg/database"
	"time"

	"gorm.io/gorm"
)

// MessageRepository 消息仓储接口
type MessageRepository interface {
	// 消息定义相关
	CreateMessageDefinition(definition *model.MessageDefinition) error
	GetMessageDefinitionByID(id uint) (*model.MessageDefinition, error)
	UpdateMessageDefinitionStats(id uint, totalRecipients, readCount int) error

	// 用户消息相关
	CreateUserMessage(userMessage *model.UserMessage) error
	GetUserMessages(
		userID uint,
		page, size int,
		messageType *int,
		isRead *bool,
		priority *int,
	) ([]model.UserMessage, int, error)
	GetUnreadCount(userID uint) (int, error)
	MarkAsRead(id uint, userID uint) error
	BatchMarkAsRead(messageIDs []uint, userID uint) error
	DeleteUserMessage(id uint, userID uint) error
	DeleteExpiredMessages() (int, error)
	GetUserMessagesByIDs(messageIDs []uint, userID uint) ([]model.UserMessage, error)

	// 批量操作
	BatchCreateUserMessages(userMessages []model.UserMessage) error
	GetAllUserIDs() ([]uint, error)
	GetUserIDsByCondition(condition string, args ...interface{}) ([]uint, error)
}

// messageRepository 消息仓储实现
type messageRepository struct {
	db *gorm.DB
}

// NewMessageRepository 创建消息仓储实例
func NewMessageRepository() MessageRepository {
	return &messageRepository{
		db: database.GetDB(),
	}
}

// CreateMessageDefinition 创建消息定义
func (r *messageRepository) CreateMessageDefinition(definition *model.MessageDefinition) error {
	return r.db.Create(definition).Error
}

// GetMessageDefinitionByID 根据ID获取消息定义
func (r *messageRepository) GetMessageDefinitionByID(id uint) (*model.MessageDefinition, error) {
	var definition model.MessageDefinition
	err := r.db.First(&definition, id).Error
	if err != nil {
		return nil, err
	}
	return &definition, nil
}

// UpdateMessageDefinitionStats 更新消息定义统计信息
func (r *messageRepository) UpdateMessageDefinitionStats(id uint, totalRecipients, readCount int) error {
	return r.db.Model(&model.MessageDefinition{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"total_recipients": totalRecipients,
			"read_count":       readCount,
		}).Error
}

// CreateUserMessage 创建用户消息
func (r *messageRepository) CreateUserMessage(userMessage *model.UserMessage) error {
	return r.db.Create(userMessage).Error
}

// BatchCreateUserMessages 批量创建用户消息
func (r *messageRepository) BatchCreateUserMessages(userMessages []model.UserMessage) error {
	return r.db.CreateInBatches(userMessages, 1000).Error
}

// GetUserMessages 获取用户消息列表
func (r *messageRepository) GetUserMessages(
	userID uint,
	page, size int,
	messageType *int,
	isRead *bool,
	priority *int,
) ([]model.UserMessage, int, error) {
	var userMessages []model.UserMessage
	var total int64

	query := r.db.Model(&model.UserMessage{}).
		Joins("JOIN message_definitions ON user_messages.message_definition_id = message_definitions.id").
		Where("user_messages.recipient_id = ? AND user_messages.is_deleted = ?", userID, false)

	// 添加过滤条件
	if messageType != nil {
		query = query.Where("message_definitions.message_type = ?", *messageType)
	}
	if isRead != nil {
		query = query.Where("user_messages.is_read = ?", *isRead)
	}
	if priority != nil {
		query = query.Where("message_definitions.priority = ?", *priority)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * size
	err := query.Preload("MessageDefinition").
		Order("user_messages.created_at DESC").
		Offset(offset).
		Limit(size).
		Find(&userMessages).Error

	return userMessages, int(total), err
}

// GetUnreadCount 获取用户未读消息数量
func (r *messageRepository) GetUnreadCount(userID uint) (int, error) {
	var count int64
	err := r.db.Model(&model.UserMessage{}).
		Where("recipient_id = ? AND is_read = ? AND is_deleted = ?", userID, false, false).
		Count(&count).Error

	return int(count), err
}

// MarkAsRead 标记消息为已读
func (r *messageRepository) MarkAsRead(id uint, userID uint) error {
	now := time.Now()
	return r.db.Model(&model.UserMessage{}).
		Where("id = ? AND recipient_id = ?", id, userID).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": &now,
		}).Error
}

// BatchMarkAsRead 批量标记已读
func (r *messageRepository) BatchMarkAsRead(messageIDs []uint, userID uint) error {
	now := time.Now()
	return r.db.Model(&model.UserMessage{}).
		Where("id IN ? AND recipient_id = ?", messageIDs, userID).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": &now,
		}).Error
}

// DeleteUserMessage 删除用户消息（软删除）
func (r *messageRepository) DeleteUserMessage(id uint, userID uint) error {
	return r.db.Model(&model.UserMessage{}).
		Where("id = ? AND recipient_id = ?", id, userID).
		Update("is_deleted", true).Error
}

// DeleteExpiredMessages 删除过期消息
func (r *messageRepository) DeleteExpiredMessages() (int, error) {
	now := time.Now()
	result := r.db.Model(&model.UserMessage{}).
		Joins("JOIN message_definitions ON user_messages.message_definition_id = message_definitions.id").
		Where("message_definitions.expire_at IS NOT NULL AND message_definitions.expire_at < ?", now).
		Update("is_deleted", true)

	return int(result.RowsAffected), result.Error
}

// GetUserMessagesByIDs 根据ID列表获取用户消息
func (r *messageRepository) GetUserMessagesByIDs(messageIDs []uint, userID uint) ([]model.UserMessage, error) {
	var userMessages []model.UserMessage
	err := r.db.Where("id IN ? AND recipient_id = ? AND is_deleted = ?", messageIDs, userID, false).
		Find(&userMessages).Error
	return userMessages, err
}

// GetAllUserIDs 获取所有用户ID
func (r *messageRepository) GetAllUserIDs() ([]uint, error) {
	var userIDs []uint
	err := r.db.Model(&model.User{}).Pluck("id", &userIDs).Error
	return userIDs, err
}

// GetUserIDsByCondition 根据条件获取用户ID列表
func (r *messageRepository) GetUserIDsByCondition(condition string, args ...interface{}) ([]uint, error) {
	var userIDs []uint
	err := r.db.Model(&model.User{}).Where(condition, args...).Pluck("id", &userIDs).Error
	return userIDs, err
}
