# 站内信系统设计指南

## 概述

本文档详细说明了站内信系统的设计方案，重点解决大规模用户场景下的消息存储和性能问题。

## 存储方案设计

### 问题分析

在传统的消息存储方案中，如果每个用户都存储一条完整的消息记录，会导致以下问题：

1. **存储空间浪费**：1000万用户 × 每条消息 = 1000万条记录
2. **数据冗余**：相同消息内容被重复存储
3. **维护困难**：消息内容更新需要修改所有用户记录
4. **性能问题**：大量重复数据影响查询性能

### 解决方案：分离式存储

我们采用**消息定义表 + 用户消息表**的分离式存储方案：

#### 1. 消息定义表 (message_definitions)
存储消息的通用信息，所有用户共享：

```sql
CREATE TABLE message_definitions (
    id BIGINT UNSIGNED PRIMARY KEY,
    title VARCHAR(200) NOT NULL,           -- 消息标题
    content TEXT NOT NULL,                 -- 消息内容
    message_type TINYINT DEFAULT 1,        -- 消息类型
    priority TINYINT DEFAULT 1,            -- 优先级
    sender_id BIGINT UNSIGNED NOT NULL,    -- 发送者ID
    sender_type TINYINT DEFAULT 1,         -- 发送者类型
    is_broadcast BOOLEAN DEFAULT FALSE,    -- 是否为广播消息
    target_users JSON,                     -- 目标用户列表
    extra_data JSON,                       -- 扩展数据
    expire_at TIMESTAMP NULL,              -- 过期时间
    total_recipients INT DEFAULT 0,        -- 总接收者数量
    read_count INT DEFAULT 0,              -- 已读数量
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);
```

#### 2. 用户消息表 (user_messages)
存储用户个人的消息状态，只存储状态信息：

```sql
CREATE TABLE user_messages (
    id BIGINT UNSIGNED PRIMARY KEY,
    message_definition_id BIGINT UNSIGNED NOT NULL,  -- 关联消息定义
    recipient_id BIGINT UNSIGNED NOT NULL,           -- 接收者ID
    is_read BOOLEAN DEFAULT FALSE,                   -- 是否已读
    is_deleted BOOLEAN DEFAULT FALSE,                -- 是否已删除
    read_at TIMESTAMP NULL,                          -- 阅读时间
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    UNIQUE KEY uk_user_message (message_definition_id, recipient_id),
    FOREIGN KEY (message_definition_id) REFERENCES message_definitions(id) ON DELETE CASCADE,
    FOREIGN KEY (recipient_id) REFERENCES users(id) ON DELETE CASCADE
);
```

## 存储优势分析

### 1. 存储空间优化

**传统方案**：
- 每条消息存储：标题(200字节) + 内容(500字节) + 其他字段(200字节) = 900字节
- 1000万用户 × 900字节 = 9GB

**优化方案**：
- 消息定义：900字节（只存储一份）
- 用户消息：约50字节（只存储状态信息）
- 1000万用户 × 50字节 = 500MB
- **节省存储空间：约95%**

### 2. 查询性能优化

#### 索引策略

```sql
-- 消息定义表索引
CREATE INDEX idx_message_type ON message_definitions(message_type);
CREATE INDEX idx_priority ON message_definitions(priority);
CREATE INDEX idx_sender_id ON message_definitions(sender_id);
CREATE INDEX idx_is_broadcast ON message_definitions(is_broadcast);
CREATE INDEX idx_expire_at ON message_definitions(expire_at);
CREATE INDEX idx_created_at ON message_definitions(created_at);

-- 用户消息表索引
CREATE INDEX idx_recipient_id ON user_messages(recipient_id);
CREATE INDEX idx_message_definition_id ON user_messages(message_definition_id);
CREATE INDEX idx_is_read ON user_messages(is_read);
CREATE INDEX idx_is_deleted ON user_messages(is_deleted);
CREATE INDEX idx_read_at ON user_messages(read_at);
CREATE INDEX idx_created_at ON user_messages(created_at);

-- 复合索引优化常见查询
CREATE INDEX idx_user_messages_recipient_status ON user_messages(recipient_id, is_deleted, is_read);
CREATE INDEX idx_user_messages_recipient_created ON user_messages(recipient_id, created_at DESC);
```

#### 查询优化

**用户消息列表查询**：
```sql
-- 使用视图简化查询
CREATE VIEW user_message_details AS
SELECT 
    um.id,
    um.message_definition_id,
    um.recipient_id,
    um.is_read,
    um.is_deleted,
    um.read_at,
    um.created_at,
    um.updated_at,
    md.title,
    md.content,
    md.message_type,
    md.priority,
    md.sender_id,
    md.sender_type,
    md.extra_data,
    md.expire_at
FROM user_messages um
INNER JOIN message_definitions md ON um.message_definition_id = md.id
WHERE um.deleted_at IS NULL AND md.deleted_at IS NULL;

-- 用户消息列表查询
SELECT * FROM user_message_details 
WHERE recipient_id = ? AND is_deleted = 0
ORDER BY created_at DESC 
LIMIT ?, ?;
```

### 3. 功能实现

#### 消息发送流程

```go
// 1. 创建消息定义
func (s *MessageService) SendMessage(req *model.SendMessageRequest) error {
    // 创建消息定义
    messageDef := &model.MessageDefinition{
        Title:       req.Title,
        Content:     req.Content,
        MessageType: req.MessageType,
        Priority:    req.Priority,
        SenderID:    req.SenderID,
        SenderType:  req.SenderType,
        ExtraData:   req.ExtraData,
        ExpireAt:    req.ExpireAt,
    }
    
    if err := s.repo.CreateMessageDefinition(messageDef); err != nil {
        return err
    }
    
    // 创建用户消息记录
    userMessage := &model.UserMessage{
        MessageDefinitionID: messageDef.ID,
        RecipientID:        req.RecipientID,
    }
    
    return s.repo.CreateUserMessage(userMessage)
}
```

#### 广播消息发送

```go
// 2. 发送广播消息
func (s *MessageService) SendBroadcastMessage(req *model.SendBroadcastMessageRequest) error {
    // 创建消息定义
    messageDef := &model.MessageDefinition{
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
    
    if err := s.repo.CreateMessageDefinition(messageDef); err != nil {
        return err
    }
    
    // 批量创建用户消息记录
    if len(req.TargetUsers) > 0 {
        // 指定目标用户
        return s.repo.BatchCreateUserMessages(messageDef.ID, req.TargetUsers)
    } else {
        // 所有用户
        return s.repo.BatchCreateUserMessagesForAllUsers(messageDef.ID)
    }
}
```

#### 消息状态管理

```go
// 3. 标记消息已读
func (s *MessageService) MarkAsRead(userID uint, messageIDs []uint) error {
    now := time.Now()
    return s.repo.BatchUpdateUserMessages(userID, messageIDs, map[string]interface{}{
        "is_read": true,
        "read_at": &now,
    })
}

// 4. 删除消息
func (s *MessageService) DeleteMessages(userID uint, messageIDs []uint) error {
    return s.repo.BatchUpdateUserMessages(userID, messageIDs, map[string]interface{}{
        "is_deleted": true,
    })
}

// 5. 获取未读消息数量
func (s *MessageService) GetUnreadCount(userID uint) (int, error) {
    return s.repo.CountUnreadMessages(userID)
}
```

## 性能优化策略

### 1. 分页查询优化

```go
// 使用游标分页，避免深度分页问题
func (s *MessageService) GetUserMessages(userID uint, params *model.MessageQueryParams) (*model.MessageListResponse, error) {
    // 使用游标分页
    messages, err := s.repo.GetUserMessagesWithCursor(userID, params)
    if err != nil {
        return nil, err
    }
    
    // 获取总数（可选，用于显示分页信息）
    total, err := s.repo.CountUserMessages(userID, params)
    if err != nil {
        return nil, err
    }
    
    return &model.MessageListResponse{
        Messages: messages,
        Pagination: model.PaginationInfo{
            Page:       params.Page,
            Size:       params.Size,
            Total:      total,
            TotalPages: (total + params.Size - 1) / params.Size,
        },
    }, nil
}
```

### 2. 缓存策略

```go
// 缓存用户未读消息数量
func (s *MessageService) GetUnreadCount(userID uint) (int, error) {
    cacheKey := fmt.Sprintf("user:unread_count:%d", userID)
    
    // 尝试从缓存获取
    if count, found := s.cache.Get(cacheKey); found {
        return count.(int), nil
    }
    
    // 从数据库查询
    count, err := s.repo.CountUnreadMessages(userID)
    if err != nil {
        return 0, err
    }
    
    // 缓存结果（5分钟）
    s.cache.Set(cacheKey, count, 5*time.Minute)
    
    return count, nil
}

// 更新缓存
func (s *MessageService) invalidateUnreadCountCache(userID uint) {
    cacheKey := fmt.Sprintf("user:unread_count:%d", userID)
    s.cache.Delete(cacheKey)
}
```

### 3. 批量操作优化

```go
// 批量插入用户消息记录
func (r *MessageRepository) BatchCreateUserMessages(messageDefID uint, userIDs []uint) error {
    // 分批处理，避免单次插入过多数据
    batchSize := 1000
    for i := 0; i < len(userIDs); i += batchSize {
        end := i + batchSize
        if end > len(userIDs) {
            end = len(userIDs)
        }
        
        batch := userIDs[i:end]
        userMessages := make([]*model.UserMessage, len(batch))
        
        for j, userID := range batch {
            userMessages[j] = &model.UserMessage{
                MessageDefinitionID: messageDefID,
                RecipientID:        userID,
            }
        }
        
        if err := r.db.CreateInBatches(userMessages, len(userMessages)).Error; err != nil {
            return err
        }
    }
    
    return nil
}
```

## 数据清理策略

### 1. 过期消息清理

```sql
-- 创建清理存储过程
DELIMITER //
CREATE PROCEDURE CleanupExpiredMessages()
BEGIN
    -- 删除过期的消息定义（级联删除用户消息记录）
    DELETE FROM message_definitions 
    WHERE expire_at IS NOT NULL 
    AND expire_at < NOW() 
    AND deleted_at IS NULL;
    
    -- 软删除用户已删除的消息记录（超过30天）
    UPDATE user_messages 
    SET deleted_at = NOW()
    WHERE is_deleted = 1 
    AND updated_at < DATE_SUB(NOW(), INTERVAL 30 DAY)
    AND deleted_at IS NULL;
END //
DELIMITER ;

-- 创建定时任务
CREATE EVENT cleanup_expired_messages_daily
ON SCHEDULE EVERY 1 DAY
STARTS CURRENT_TIMESTAMP
DO CALL CleanupExpiredMessages();
```

### 2. 归档策略

```go
// 归档旧消息
func (s *MessageService) ArchiveOldMessages(beforeDate time.Time) error {
    // 将旧消息移动到归档表
    return s.repo.ArchiveMessages(beforeDate)
}
```

## 监控和统计

### 1. 消息统计

```go
// 获取消息统计信息
func (s *MessageService) GetMessageStats() (*model.MessageStats, error) {
    stats := &model.MessageStats{}
    
    // 总消息数
    stats.TotalMessages, _ = s.repo.CountMessageDefinitions()
    
    // 总用户消息数
    stats.TotalUserMessages, _ = s.repo.CountUserMessages()
    
    // 未读消息数
    stats.TotalUnreadMessages, _ = s.repo.CountAllUnreadMessages()
    
    // 今日发送消息数
    stats.TodaySentMessages, _ = s.repo.CountTodaySentMessages()
    
    return stats, nil
}
```

### 2. 性能监控

```go
// 监控查询性能
func (s *MessageService) GetUserMessagesWithMetrics(userID uint, params *model.MessageQueryParams) (*model.MessageListResponse, error) {
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        s.metrics.RecordQueryDuration("get_user_messages", duration)
    }()
    
    return s.GetUserMessages(userID, params)
}
```

## 总结

通过采用**消息定义表 + 用户消息表**的分离式存储方案，我们实现了：

1. **存储空间优化**：节省约95%的存储空间
2. **查询性能提升**：通过合理的索引和查询优化
3. **功能完整性**：支持所有站内信功能
4. **可扩展性**：支持大规模用户场景
5. **维护性**：消息内容更新只需修改一条记录

这种方案特别适合大规模用户场景，能够有效解决您提到的1000万用户的消息存储问题。 