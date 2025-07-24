-- 短信验证码安全升级迁移脚本
-- 添加token和user_agent字段，增强安全性

-- 1. 添加token字段
ALTER TABLE sms_verification_codes 
ADD COLUMN token VARCHAR(64) NOT NULL DEFAULT '' COMMENT '唯一验证token';

-- 2. 添加user_agent字段
ALTER TABLE sms_verification_codes 
ADD COLUMN user_agent VARCHAR(500) DEFAULT NULL COMMENT '用户代理';

-- 3. 为token字段添加唯一索引
ALTER TABLE sms_verification_codes 
ADD UNIQUE INDEX idx_token (token);

-- 4. 为purpose字段添加索引，提高查询性能
ALTER TABLE sms_verification_codes 
ADD INDEX idx_purpose (purpose);

-- 5. 为phone和purpose组合添加索引
ALTER TABLE sms_verification_codes 
ADD INDEX idx_phone_purpose (phone, purpose);

-- 6. 更新现有记录的token（为现有记录生成随机token）
UPDATE sms_verification_codes 
SET token = CONCAT(
    SUBSTRING(MD5(RAND()), 1, 8),
    SUBSTRING(MD5(RAND()), 1, 8),
    SUBSTRING(MD5(RAND()), 1, 8),
    SUBSTRING(MD5(RAND()), 1, 8)
)
WHERE token = '';

-- 7. 添加验证码用途约束检查
-- 注意：MySQL不支持CHECK约束，需要在应用层验证
-- 建议的用途值：login, register, reset, change, payment, withdraw, security, device

-- 8. 创建验证码安全审计视图
CREATE VIEW sms_security_audit AS
SELECT 
    id,
    phone,
    purpose,
    client_ip,
    user_agent,
    created_at,
    expired_at,
    used_at,
    CASE 
        WHEN used_at IS NOT NULL THEN 'used'
        WHEN expired_at < NOW() THEN 'expired'
        ELSE 'valid'
    END as status
FROM sms_verification_codes
ORDER BY created_at DESC;

-- 9. 创建高安全级别验证码统计视图
CREATE VIEW high_security_sms_stats AS
SELECT 
    purpose,
    COUNT(*) as total_count,
    COUNT(CASE WHEN used_at IS NOT NULL THEN 1 END) as used_count,
    COUNT(CASE WHEN expired_at < NOW() AND used_at IS NULL THEN 1 END) as expired_count,
    COUNT(CASE WHEN created_at > DATE_SUB(NOW(), INTERVAL 1 HOUR) THEN 1 END) as last_hour_count
FROM sms_verification_codes
WHERE purpose IN ('change', 'payment', 'withdraw', 'security')
GROUP BY purpose;

-- 10. 添加注释说明
-- 验证码用途说明：
-- login: 登录验证
-- register: 注册验证  
-- reset: 重置密码
-- change: 变更个人信息（需要登录）
-- payment: 支付验证（需要登录）
-- withdraw: 提现验证（需要登录）
-- security: 安全设置变更（需要登录）
-- device: 设备绑定（需要登录） 