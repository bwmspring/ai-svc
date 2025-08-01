-- 消息系统数据库迁移脚本
-- 创建时间: 2024-01-01
-- 描述: 创建站内信系统的核心表结构

-- 1. 创建消息定义表 (message_definitions)
-- 存储消息的通用信息，所有用户共享，避免重复存储消息内容
CREATE TABLE IF NOT EXISTS `message_definitions` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at` timestamp NULL DEFAULT NULL COMMENT '删除时间',
    
    -- 消息基本信息
    `title` varchar(200) NOT NULL COMMENT '消息标题',
    `content` text NOT NULL COMMENT '消息内容',
    `message_type` tinyint NOT NULL DEFAULT 1 COMMENT '消息类型 1:系统通知 2:用户消息 3:管理员消息 4:业务通知',
    `priority` tinyint NOT NULL DEFAULT 1 COMMENT '优先级 1:普通 2:重要 3:紧急',
    
    -- 发送者信息
    `sender_id` bigint unsigned NOT NULL COMMENT '发送者ID，0表示系统',
    `sender_type` tinyint NOT NULL DEFAULT 1 COMMENT '发送者类型 1:用户 2:系统 3:管理员 4:业务系统',
    
    -- 消息属性
    `is_broadcast` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否为广播消息',
    `target_users` json DEFAULT NULL COMMENT '目标用户ID列表，为空表示所有用户',
    
    -- 扩展字段
    `extra_data` json DEFAULT NULL COMMENT '扩展数据',
    `expire_at` timestamp NULL DEFAULT NULL COMMENT '过期时间',
    
    -- 统计信息
    `total_recipients` int NOT NULL DEFAULT 0 COMMENT '总接收者数量',
    `read_count` int NOT NULL DEFAULT 0 COMMENT '已读数量',
    
    PRIMARY KEY (`id`),
    KEY `idx_message_type` (`message_type`),
    KEY `idx_priority` (`priority`),
    KEY `idx_sender_id` (`sender_id`),
    KEY `idx_is_broadcast` (`is_broadcast`),
    KEY `idx_expire_at` (`expire_at`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='消息定义表';

-- 2. 创建用户消息表 (user_messages)
-- 存储用户个人的消息状态，只存储用户相关的状态信息，不重复存储消息内容
CREATE TABLE IF NOT EXISTS `user_messages` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at` timestamp NULL DEFAULT NULL COMMENT '删除时间',
    
    -- 关联消息定义
    `message_definition_id` bigint unsigned NOT NULL COMMENT '消息定义ID',
    
    -- 接收者信息
    `recipient_id` bigint unsigned NOT NULL COMMENT '接收者ID',
    
    -- 消息状态
    `is_read` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否已读',
    `is_deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否已删除',
    `read_at` timestamp NULL DEFAULT NULL COMMENT '阅读时间',
    
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_message` (`message_definition_id`, `recipient_id`),
    KEY `idx_recipient_id` (`recipient_id`),
    KEY `idx_message_definition_id` (`message_definition_id`),
    KEY `idx_is_read` (`is_read`),
    KEY `idx_is_deleted` (`is_deleted`),
    KEY `idx_read_at` (`read_at`),
    KEY `idx_created_at` (`created_at`),
    
    -- 外键约束
    CONSTRAINT `fk_user_messages_message_definition` FOREIGN KEY (`message_definition_id`) REFERENCES `message_definitions` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_user_messages_recipient` FOREIGN KEY (`recipient_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户消息表';

-- 3. 创建索引优化查询性能
-- 复合索引用于常见查询场景
CREATE INDEX `idx_user_messages_recipient_status` ON `user_messages` (`recipient_id`, `is_deleted`, `is_read`);
CREATE INDEX `idx_user_messages_recipient_created` ON `user_messages` (`recipient_id`, `created_at` DESC);
CREATE INDEX `idx_message_definitions_broadcast_created` ON `message_definitions` (`is_broadcast`, `created_at` DESC);

-- 4. 插入示例数据（可选）
-- 插入一条系统欢迎消息
INSERT INTO `message_definitions` (
    `title`, 
    `content`, 
    `message_type`, 
    `priority`, 
    `sender_id`, 
    `sender_type`, 
    `is_broadcast`, 
    `total_recipients`
) VALUES (
    '欢迎使用AI服务',
    '感谢您注册我们的AI服务平台！我们将为您提供优质的智能服务体验。',
    1,  -- 系统通知
    1,  -- 普通优先级
    0,  -- 系统发送
    2,  -- 系统类型
    1,  -- 广播消息
    0   -- 初始接收者数量为0
) ON DUPLICATE KEY UPDATE `updated_at` = CURRENT_TIMESTAMP;

-- 5. 创建视图用于简化查询（可选）
-- 用户消息详情视图
CREATE OR REPLACE VIEW `user_message_details` AS
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
FROM `user_messages` um
INNER JOIN `message_definitions` md ON um.message_definition_id = md.id
WHERE um.deleted_at IS NULL AND md.deleted_at IS NULL;

-- 6. 创建存储过程用于批量发送广播消息（可选）
DELIMITER //
CREATE PROCEDURE `SendBroadcastMessage`(
    IN p_title VARCHAR(200),
    IN p_content TEXT,
    IN p_message_type TINYINT,
    IN p_priority TINYINT,
    IN p_sender_id BIGINT UNSIGNED,
    IN p_sender_type TINYINT,
    IN p_extra_data JSON,
    IN p_expire_at TIMESTAMP,
    IN p_target_users JSON
)
BEGIN
    DECLARE message_id BIGINT UNSIGNED;
    DECLARE done INT DEFAULT FALSE;
    DECLARE user_id BIGINT UNSIGNED;
    DECLARE user_cursor CURSOR FOR 
        SELECT id FROM users WHERE deleted_at IS NULL;
    DECLARE CONTINUE HANDLER FOR NOT FOUND SET done = TRUE;
    
    -- 插入消息定义
    INSERT INTO `message_definitions` (
        `title`, `content`, `message_type`, `priority`, 
        `sender_id`, `sender_type`, `is_broadcast`, 
        `target_users`, `extra_data`, `expire_at`
    ) VALUES (
        p_title, p_content, p_message_type, p_priority,
        p_sender_id, p_sender_type, 1,
        p_target_users, p_extra_data, p_expire_at
    );
    
    SET message_id = LAST_INSERT_ID();
    
    -- 如果指定了目标用户，则只为这些用户创建消息记录
    IF p_target_users IS NOT NULL AND JSON_LENGTH(p_target_users) > 0 THEN
        INSERT INTO `user_messages` (`message_definition_id`, `recipient_id`)
        SELECT message_id, JSON_UNQUOTE(JSON_EXTRACT(p_target_users, CONCAT('$[', numbers.n, ']')))
        FROM (
            SELECT 0 AS n UNION SELECT 1 UNION SELECT 2 UNION SELECT 3 UNION SELECT 4
            UNION SELECT 5 UNION SELECT 6 UNION SELECT 7 UNION SELECT 8 UNION SELECT 9
        ) numbers
        WHERE JSON_EXTRACT(p_target_users, CONCAT('$[', numbers.n, ']')) IS NOT NULL;
    ELSE
        -- 为所有用户创建消息记录
        OPEN user_cursor;
        read_loop: LOOP
            FETCH user_cursor INTO user_id;
            IF done THEN
                LEAVE read_loop;
            END IF;
            
            INSERT INTO `user_messages` (`message_definition_id`, `recipient_id`)
            VALUES (message_id, user_id);
        END LOOP;
        CLOSE user_cursor;
    END IF;
    
    -- 更新消息定义中的接收者数量
    UPDATE `message_definitions` 
    SET `total_recipients` = (
        SELECT COUNT(*) FROM `user_messages` 
        WHERE `message_definition_id` = message_id
    )
    WHERE `id` = message_id;
    
    SELECT message_id AS message_definition_id;
END //
DELIMITER ;

-- 7. 创建触发器用于自动更新统计信息（可选）
DELIMITER //
CREATE TRIGGER `update_message_read_count` 
AFTER UPDATE ON `user_messages`
FOR EACH ROW
BEGIN
    IF OLD.is_read != NEW.is_read THEN
        IF NEW.is_read = 1 THEN
            UPDATE `message_definitions` 
            SET `read_count` = `read_count` + 1
            WHERE `id` = NEW.message_definition_id;
        ELSE
            UPDATE `message_definitions` 
            SET `read_count` = `read_count` - 1
            WHERE `id` = NEW.message_definition_id;
        END IF;
    END IF;
END //
DELIMITER ;

-- 8. 创建清理过期消息的存储过程（可选）
DELIMITER //
CREATE PROCEDURE `CleanupExpiredMessages`()
BEGIN
    -- 删除过期的消息定义（同时会级联删除用户消息记录）
    DELETE FROM `message_definitions` 
    WHERE `expire_at` IS NOT NULL 
    AND `expire_at` < NOW() 
    AND `deleted_at` IS NULL;
    
    -- 软删除用户已删除的消息记录（超过30天）
    UPDATE `user_messages` 
    SET `deleted_at` = NOW()
    WHERE `is_deleted` = 1 
    AND `updated_at` < DATE_SUB(NOW(), INTERVAL 30 DAY)
    AND `deleted_at` IS NULL;
END //
DELIMITER ;

-- 9. 创建定时任务（需要在MySQL事件调度器中启用）
-- SET GLOBAL event_scheduler = ON;
-- 
-- CREATE EVENT `cleanup_expired_messages_daily`
-- ON SCHEDULE EVERY 1 DAY
-- STARTS CURRENT_TIMESTAMP
-- DO CALL CleanupExpiredMessages();

-- 10. 添加表注释和字段说明
ALTER TABLE `message_definitions` COMMENT = '消息定义表 - 存储消息的通用信息，所有用户共享';
ALTER TABLE `user_messages` COMMENT = '用户消息表 - 存储用户个人的消息状态';

-- 显示创建结果
SELECT '消息系统表创建完成' AS status; 