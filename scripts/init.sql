-- 创建数据库
CREATE DATABASE IF NOT EXISTS ai_svc CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 创建用户
CREATE USER IF NOT EXISTS 'ai_svc'@'%' IDENTIFIED BY 'ai_svc123';

-- 授权
GRANT ALL PRIVILEGES ON ai_svc.* TO 'ai_svc'@'%';

-- 刷新权限
FLUSH PRIVILEGES;

-- 使用数据库
USE ai_svc;

-- 用户表会由GORM自动创建，这里只是预留脚本空间
-- 如果需要额外的初始化数据，可以在这里添加
