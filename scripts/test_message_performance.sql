-- 消息系统性能测试脚本
-- 用于测试大规模用户场景下的消息系统性能

-- 1. 创建测试数据生成函数
DELIMITER //
CREATE PROCEDURE GenerateTestData(
    IN user_count INT,
    IN message_count INT
)
BEGIN
    DECLARE i INT DEFAULT 1;
    DECLARE j INT DEFAULT 1;
    DECLARE message_id BIGINT UNSIGNED;
    
    -- 生成测试消息定义
    WHILE i <= message_count DO
        INSERT INTO message_definitions (
            title, 
            content, 
            message_type, 
            priority, 
            sender_id, 
            sender_type, 
            is_broadcast,
            total_recipients
        ) VALUES (
            CONCAT('测试消息 ', i),
            CONCAT('这是第 ', i, ' 条测试消息的内容。用于测试大规模数据场景下的性能表现。'),
            (i % 4) + 1,  -- 消息类型 1-4
            (i % 3) + 1,  -- 优先级 1-3
            0,            -- 系统发送
            2,            -- 系统类型
            1,            -- 广播消息
            0             -- 初始接收者数量
        );
        
        SET message_id = LAST_INSERT_ID();
        
        -- 为每个用户创建消息记录
        SET j = 1;
        WHILE j <= user_count DO
            INSERT INTO user_messages (
                message_definition_id,
                recipient_id,
                is_read,
                is_deleted
            ) VALUES (
                message_id,
                j,
                (RAND() > 0.7),  -- 70%概率已读
                (RAND() > 0.9)   -- 10%概率已删除
            );
            SET j = j + 1;
        END WHILE;
        
        -- 更新消息定义中的接收者数量
        UPDATE message_definitions 
        SET total_recipients = user_count
        WHERE id = message_id;
        
        SET i = i + 1;
    END WHILE;
    
    SELECT CONCAT('生成了 ', message_count, ' 条消息，', user_count, ' 个用户，总计 ', message_count * user_count, ' 条用户消息记录') AS result;
END //
DELIMITER ;

-- 2. 性能测试查询

-- 测试1: 用户消息列表查询性能
SELECT '测试1: 用户消息列表查询' AS test_name;
SET @start_time = NOW(6);

SELECT 
    um.id,
    um.message_definition_id,
    um.recipient_id,
    um.is_read,
    um.is_deleted,
    um.read_at,
    um.created_at,
    md.title,
    md.content,
    md.message_type,
    md.priority,
    md.sender_id,
    md.sender_type
FROM user_messages um
INNER JOIN message_definitions md ON um.message_definition_id = md.id
WHERE um.recipient_id = 1 
  AND um.is_deleted = 0
  AND um.deleted_at IS NULL 
  AND md.deleted_at IS NULL
ORDER BY um.created_at DESC
LIMIT 20;

SET @end_time = NOW(6);
SELECT TIMESTAMPDIFF(MICROSECOND, @start_time, @end_time) / 1000 AS duration_ms;

-- 测试2: 未读消息数量查询性能
SELECT '测试2: 未读消息数量查询' AS test_name;
SET @start_time = NOW(6);

SELECT COUNT(*) as unread_count
FROM user_messages um
WHERE um.recipient_id = 1 
  AND um.is_read = 0 
  AND um.is_deleted = 0
  AND um.deleted_at IS NULL;

SET @end_time = NOW(6);
SELECT TIMESTAMPDIFF(MICROSECOND, @start_time, @end_time) / 1000 AS duration_ms;

-- 测试3: 消息类型筛选查询性能
SELECT '测试3: 消息类型筛选查询' AS test_name;
SET @start_time = NOW(6);

SELECT 
    um.id,
    md.title,
    md.content,
    md.message_type,
    um.is_read,
    um.created_at
FROM user_messages um
INNER JOIN message_definitions md ON um.message_definition_id = md.id
WHERE um.recipient_id = 1 
  AND md.message_type = 1
  AND um.is_deleted = 0
  AND um.deleted_at IS NULL 
  AND md.deleted_at IS NULL
ORDER BY um.created_at DESC
LIMIT 10;

SET @end_time = NOW(6);
SELECT TIMESTAMPDIFF(MICROSECOND, @start_time, @end_time) / 1000 AS duration_ms;

-- 测试4: 广播消息查询性能
SELECT '测试4: 广播消息查询' AS test_name;
SET @start_time = NOW(6);

SELECT 
    md.id,
    md.title,
    md.content,
    md.total_recipients,
    md.read_count,
    md.created_at
FROM message_definitions md
WHERE md.is_broadcast = 1
  AND md.deleted_at IS NULL
ORDER BY md.created_at DESC
LIMIT 10;

SET @end_time = NOW(6);
SELECT TIMESTAMPDIFF(MICROSECOND, @start_time, @end_time) / 1000 AS duration_ms;

-- 测试5: 批量更新性能（标记已读）
SELECT '测试5: 批量标记已读性能' AS test_name;
SET @start_time = NOW(6);

UPDATE user_messages 
SET is_read = 1, read_at = NOW()
WHERE recipient_id = 1 
  AND is_read = 0 
  AND is_deleted = 0
  AND deleted_at IS NULL
LIMIT 100;

SET @end_time = NOW(6);
SELECT TIMESTAMPDIFF(MICROSECOND, @start_time, @end_time) / 1000 AS duration_ms;

-- 测试6: 统计查询性能
SELECT '测试6: 消息统计查询' AS test_name;
SET @start_time = NOW(6);

SELECT 
    COUNT(DISTINCT md.id) as total_messages,
    COUNT(um.id) as total_user_messages,
    SUM(CASE WHEN um.is_read = 0 AND um.is_deleted = 0 THEN 1 ELSE 0 END) as total_unread,
    SUM(CASE WHEN DATE(um.created_at) = CURDATE() THEN 1 ELSE 0 END) as today_messages
FROM message_definitions md
LEFT JOIN user_messages um ON md.id = um.message_definition_id
WHERE md.deleted_at IS NULL;

SET @end_time = NOW(6);
SELECT TIMESTAMPDIFF(MICROSECOND, @start_time, @end_time) / 1000 AS duration_ms;

-- 3. 索引使用情况分析
SELECT '索引使用情况分析' AS analysis;

-- 查看表的索引信息
SHOW INDEX FROM user_messages;
SHOW INDEX FROM message_definitions;

-- 分析查询执行计划
EXPLAIN SELECT 
    um.id,
    md.title,
    md.content,
    um.is_read,
    um.created_at
FROM user_messages um
INNER JOIN message_definitions md ON um.message_definition_id = md.id
WHERE um.recipient_id = 1 
  AND um.is_deleted = 0
  AND um.deleted_at IS NULL 
  AND md.deleted_at IS NULL
ORDER BY um.created_at DESC
LIMIT 20;

-- 4. 表大小和行数统计
SELECT '表大小和行数统计' AS statistics;

SELECT 
    table_name,
    table_rows,
    ROUND(((data_length + index_length) / 1024 / 1024), 2) AS 'Size (MB)',
    ROUND((data_length / 1024 / 1024), 2) AS 'Data (MB)',
    ROUND((index_length / 1024 / 1024), 2) AS 'Index (MB)'
FROM information_schema.tables 
WHERE table_schema = DATABASE() 
  AND table_name IN ('message_definitions', 'user_messages');

-- 5. 性能优化建议
SELECT '性能优化建议' AS recommendations;

-- 检查慢查询
SELECT 
    sql_text,
    exec_count,
    avg_timer_wait / 1000000000 as avg_time_sec,
    max_timer_wait / 1000000000 as max_time_sec
FROM performance_schema.events_statements_summary_by_digest
WHERE schema_name = DATABASE()
  AND digest_text LIKE '%user_messages%'
  OR digest_text LIKE '%message_definitions%'
ORDER BY avg_timer_wait DESC
LIMIT 10;

-- 6. 清理测试数据
DELIMITER //
CREATE PROCEDURE CleanupTestData()
BEGIN
    -- 删除测试数据
    DELETE FROM user_messages WHERE recipient_id <= 1000;
    DELETE FROM message_definitions WHERE title LIKE '测试消息%';
    
    -- 重置自增ID（可选）
    ALTER TABLE user_messages AUTO_INCREMENT = 1;
    ALTER TABLE message_definitions AUTO_INCREMENT = 1;
    
    SELECT '测试数据清理完成' AS result;
END //
DELIMITER ;

-- 7. 使用示例
-- 生成测试数据（1000个用户，100条消息）
-- CALL GenerateTestData(1000, 100);

-- 清理测试数据
-- CALL CleanupTestData();

-- 8. 性能基准测试
DELIMITER //
CREATE PROCEDURE PerformanceBenchmark()
BEGIN
    DECLARE i INT DEFAULT 1;
    DECLARE start_time TIMESTAMP(6);
    DECLARE end_time TIMESTAMP(6);
    DECLARE duration_ms DECIMAL(10,3);
    
    -- 测试不同用户ID的查询性能
    WHILE i <= 10 DO
        SET start_time = NOW(6);
        
        -- 执行查询
        SELECT COUNT(*) 
        FROM user_messages um
        WHERE um.recipient_id = i 
          AND um.is_read = 0 
          AND um.is_deleted = 0;
        
        SET end_time = NOW(6);
        SET duration_ms = TIMESTAMPDIFF(MICROSECOND, start_time, end_time) / 1000;
        
        SELECT CONCAT('用户 ', i, ' 查询耗时: ', duration_ms, ' ms') AS benchmark_result;
        
        SET i = i + 1;
    END WHILE;
END //
DELIMITER ;

-- 运行性能基准测试
-- CALL PerformanceBenchmark(); 