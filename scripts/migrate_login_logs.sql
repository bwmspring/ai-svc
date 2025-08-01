-- 登录日志功能数据库迁移脚本
-- 为 user_behavior_logs 表添加登录相关字段

-- 1. 添加 Location 字段（地理位置信息）
ALTER TABLE user_behavior_logs 
ADD COLUMN location VARCHAR(200) COMMENT '地理位置信息';

-- 2. 添加 LoginTime 字段（登录时间）
ALTER TABLE user_behavior_logs 
ADD COLUMN login_time TIMESTAMP NULL COMMENT '登录时间（专门字段）';

-- 3. 为 LoginTime 字段添加索引，提高查询性能
CREATE INDEX idx_user_behavior_logs_login_time ON user_behavior_logs(login_time);

-- 4. 为 Location 字段添加索引，提高地理位置查询性能
CREATE INDEX idx_user_behavior_logs_location ON user_behavior_logs(location);

-- 5. 创建复合索引，优化登录历史查询
CREATE INDEX idx_user_behavior_logs_user_action_time ON user_behavior_logs(user_id, action, created_at);

-- 6. 创建复合索引，优化按用户和登录时间查询
CREATE INDEX idx_user_behavior_logs_user_login_time ON user_behavior_logs(user_id, login_time);

-- 7. 添加注释说明字段用途
COMMENT ON COLUMN user_behavior_logs.location IS '地理位置信息，格式：国家,省份,城市,区县,详细地址 或 纬度,经度';
COMMENT ON COLUMN user_behavior_logs.login_time IS '登录时间（专门字段），与created_at区分，created_at为数据库记录创建时间';

-- 8. 查看表结构确认
DESCRIBE user_behavior_logs; 