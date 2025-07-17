-- 设备ID唯一性迁移脚本
-- 执行前请务必备份数据库！

USE ai_svc;

-- 1. 检查现有重复数据
SELECT '=== 检查重复设备ID ===' as step;
SELECT device_id, COUNT(*) as duplicate_count, 
       GROUP_CONCAT(DISTINCT user_id) as user_ids,
       GROUP_CONCAT(id) as device_record_ids
FROM user_devices 
GROUP BY device_id 
HAVING duplicate_count > 1
ORDER BY duplicate_count DESC;

-- 2. 显示即将删除的重复记录（保留最新的）
SELECT '=== 即将删除的重复记录 ===' as step;
SELECT ud1.id, ud1.device_id, ud1.user_id, ud1.created_at, '将被删除' as action
FROM user_devices ud1
INNER JOIN user_devices ud2 
WHERE ud1.device_id = ud2.device_id 
  AND ud1.created_at < ud2.created_at
ORDER BY ud1.device_id, ud1.created_at;

-- 3. 备份即将删除的记录到临时表
SELECT '=== 创建备份表 ===' as step;
CREATE TABLE IF NOT EXISTS user_devices_backup_before_uniqueness AS
SELECT * FROM user_devices WHERE 1=0;

INSERT INTO user_devices_backup_before_uniqueness
SELECT ud1.*
FROM user_devices ud1
INNER JOIN user_devices ud2 
WHERE ud1.device_id = ud2.device_id 
  AND ud1.created_at < ud2.created_at;

SELECT COUNT(*) as backup_records_count FROM user_devices_backup_before_uniqueness;

-- 4. 删除重复数据（保留最新的记录）
SELECT '=== 删除重复记录 ===' as step;
DELETE ud1 FROM user_devices ud1
INNER JOIN user_devices ud2 
WHERE ud1.device_id = ud2.device_id 
  AND ud1.created_at < ud2.created_at;

-- 5. 验证删除结果
SELECT '=== 验证删除结果 ===' as step;
SELECT device_id, COUNT(*) as count 
FROM user_devices 
GROUP BY device_id 
HAVING count > 1;

-- 6. 添加唯一约束
SELECT '=== 添加唯一约束 ===' as step;
ALTER TABLE user_devices 
ADD UNIQUE INDEX idx_device_id_unique (device_id);

-- 7. 验证约束添加成功
SELECT '=== 验证约束 ===' as step;
SHOW INDEX FROM user_devices WHERE Key_name = 'idx_device_id_unique';

-- 8. 最终统计
SELECT '=== 迁移完成统计 ===' as step;
SELECT 
    COUNT(*) as total_devices,
    COUNT(DISTINCT device_id) as unique_device_ids,
    COUNT(DISTINCT user_id) as users_with_devices
FROM user_devices;

SELECT '迁移完成！备份数据保存在 user_devices_backup_before_uniqueness 表中' as final_message; 