# 登录日志功能设计文档

## 概述

登录日志功能是ai-svc系统的重要组成部分，用于完整记录用户的登录行为信息，包括登录人、登录时间、登录地点、登录系统、登录方式等，从而完整记录用户在什么时间、什么地点、采用哪种方式、登录了哪个系统。

## 功能特性

### 1. 完整的登录行为记录
- ✅ **登录人**：通过 `user_id` 字段记录
- ✅ **登录时间**：通过 `login_time` 字段记录（专门字段）
- ✅ **登录地点**：通过 `location` 字段记录地理位置信息
- ✅ **登录系统**：通过 `device_type` 和 `device_name` 记录
- ✅ **登录方式**：通过 `login_type` 记录（SMS、OAuth等）

### 2. 多种登录状态记录
- **登录尝试** (`action: login`)：记录用户开始登录
- **登录成功** (`action: login_success`)：记录登录成功
- **登录失败** (`action: login_failed`)：记录登录失败及原因
- **登出** (`action: logout`)：记录用户登出

### 3. 地理位置支持
- 支持IP地理位置解析
- 支持用户主动提供的地理位置信息
- 地理位置信息格式：`国家,省份,城市,区县,详细地址` 或 `纬度,经度`

### 4. 统计分析功能
- 用户登录历史查询
- 登录统计（总次数、成功次数、失败次数）
- 唯一设备统计
- 唯一地理位置统计

## 数据库设计

### UserBehaviorLog 表结构

```sql
CREATE TABLE user_behavior_logs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    action VARCHAR(50) NOT NULL COMMENT '行为类型：login, login_success, login_failed, logout',
    resource VARCHAR(100) COMMENT '资源信息，如登录方式、设备信息等',
    ip VARCHAR(45) COMMENT 'IP地址',
    user_agent VARCHAR(500) COMMENT '用户代理',
    location VARCHAR(200) COMMENT '地理位置信息',
    login_time TIMESTAMP NULL COMMENT '登录时间（专门字段）',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '数据库创建时间',
    
    INDEX idx_user_id (user_id),
    INDEX idx_action (action),
    INDEX idx_login_time (login_time),
    INDEX idx_location (location),
    INDEX idx_user_action_time (user_id, action, created_at),
    INDEX idx_user_login_time (user_id, login_time)
);
```

### 字段说明

| 字段名 | 类型 | 说明 |
|--------|------|------|
| `user_id` | BIGINT | 用户ID，标识登录人 |
| `action` | VARCHAR(50) | 行为类型：login/login_success/login_failed/logout |
| `resource` | VARCHAR(100) | 资源信息，存储登录方式、设备信息、失败原因等 |
| `ip` | VARCHAR(45) | IP地址，用于地理位置解析 |
| `user_agent` | VARCHAR(500) | 用户代理，记录浏览器/设备信息 |
| `location` | VARCHAR(200) | 地理位置信息 |
| `login_time` | TIMESTAMP | 登录时间（专门字段） |
| `created_at` | TIMESTAMP | 数据库记录创建时间 |

## 核心常量

### 登录行为常量
```go
const (
    ActionLogin        = "login"        // 登录
    ActionLoginSuccess = "login_success" // 登录成功
    ActionLoginFailed  = "login_failed"  // 登录失败
    ActionLogout       = "logout"        // 登出
    ActionRefreshToken = "refresh_token" // 刷新Token
)
```

### 登录方式常量
```go
const (
    LoginTypeSMS      = "sms"      // 短信登录
    LoginTypePassword = "password" // 密码登录
    LoginTypeOAuth    = "oauth"    // 第三方登录
    LoginTypeRefresh  = "refresh"  // Token刷新
)
```

## 服务架构

### 1. LoginLogService 接口
```go
type LoginLogService interface {
    // 记录登录尝试
    LogLoginAttempt(ctx context.Context, userID uint, phone, loginType, ip, userAgent string, location *model.LocationInfo) error
    
    // 记录登录成功
    LogLoginSuccess(ctx context.Context, userID uint, phone, loginType, deviceID, deviceType, ip, userAgent string, isNewUser bool, location *model.LocationInfo) error
    
    // 记录登录失败
    LogLoginFailed(ctx context.Context, userID uint, phone, loginType, reason, ip, userAgent string, location *model.LocationInfo) error
    
    // 记录登出
    LogLogout(ctx context.Context, userID uint, ip, userAgent string, location *model.LocationInfo) error
    
    // 获取用户登录历史
    GetUserLoginHistory(ctx context.Context, userID uint, page, size int) ([]*model.UserBehaviorLog, int64, error)
    
    // 获取登录统计
    GetLoginStats(ctx context.Context, userID uint, days int) (*LoginStats, error)
}
```

### 2. UserBehaviorLogRepository 接口
```go
type UserBehaviorLogRepository interface {
    // 基础CRUD
    Create(log *model.UserBehaviorLog) error
    GetByID(id uint) (*model.UserBehaviorLog, error)
    Update(log *model.UserBehaviorLog) error
    Delete(id uint) error
    
    // 查询方法
    GetUserLogs(userID uint, action string, page, size int) ([]*model.UserBehaviorLog, int64, error)
    GetUserLoginHistory(userID uint, page, size int) ([]*model.UserBehaviorLog, int64, error)
    GetRecentLogins(userID uint, hours int) ([]*model.UserBehaviorLog, error)
    GetFailedLogins(userID uint, hours int) ([]*model.UserBehaviorLog, error)
    
    // 统计方法
    CountUserLogins(userID uint, days int) (int64, error)
    CountFailedLogins(userID uint, days int) (int64, error)
    CountLoginsByType(userID uint, loginType string, days int) (int64, error)
}
```

## 地理位置信息

### LocationInfo 结构
```go
type LocationInfo struct {
    Latitude  float64 `json:"latitude"`  // 纬度
    Longitude float64 `json:"longitude"` // 经度
    Address   string  `json:"address"`   // 详细地址
    City      string  `json:"city"`      // 城市
    Province  string  `json:"province"`  // 省份
    Country   string  `json:"country"`   // 国家
    District  string  `json:"district"`  // 区县
}
```

### 地理位置服务接口
```go
type LocationService interface {
    // 根据IP获取地理位置
    GetLocationByIP(ip string) (*model.LocationInfo, error)
    
    // 批量获取地理位置
    BatchGetLocation(ips []string) (map[string]*model.LocationInfo, error)
    
    // 缓存地理位置信息
    CacheLocation(ip string, location *model.LocationInfo) error
}
```

## 使用示例

### 1. 记录登录尝试
```go
loginLogService.LogLoginAttempt(ctx, 0, "13800138000", "sms", "192.168.1.1", "Mozilla/5.0...", nil)
```

### 2. 记录登录成功
```go
loginLogService.LogLoginSuccess(ctx, 123, "13800138000", "sms", "device_001", "android", "192.168.1.1", "Mozilla/5.0...", false, locationInfo)
```

### 3. 记录登录失败
```go
loginLogService.LogLoginFailed(ctx, 123, "13800138000", "sms", "验证码错误", "192.168.1.1", "Mozilla/5.0...", nil)
```

### 4. 获取登录历史
```go
logs, total, err := loginLogService.GetUserLoginHistory(ctx, 123, 1, 10)
```

### 5. 获取登录统计
```go
stats, err := loginLogService.GetLoginStats(ctx, 123, 7) // 最近7天
```

## 数据库迁移

执行以下SQL脚本添加新字段：

```sql
-- 添加 Location 字段
ALTER TABLE user_behavior_logs 
ADD COLUMN location VARCHAR(200) COMMENT '地理位置信息';

-- 添加 LoginTime 字段
ALTER TABLE user_behavior_logs 
ADD COLUMN login_time TIMESTAMP NULL COMMENT '登录时间（专门字段）';

-- 添加索引
CREATE INDEX idx_user_behavior_logs_login_time ON user_behavior_logs(login_time);
CREATE INDEX idx_user_behavior_logs_location ON user_behavior_logs(location);
CREATE INDEX idx_user_behavior_logs_user_action_time ON user_behavior_logs(user_id, action, created_at);
CREATE INDEX idx_user_behavior_logs_user_login_time ON user_behavior_logs(user_id, login_time);
```

## 测试

使用提供的测试脚本进行功能验证：

```bash
# 运行测试脚本
./scripts/test_login_logs.sh
```

测试脚本会验证：
1. 登录尝试记录
2. 登录成功记录
3. 登录失败记录
4. 地理位置信息记录
5. 登录时间记录

## 性能优化

### 1. 索引优化
- 为 `login_time` 字段添加索引
- 为 `location` 字段添加索引
- 创建复合索引优化查询性能

### 2. 数据分区
- 按时间分区存储历史数据
- 定期归档旧数据

### 3. 缓存策略
- 缓存地理位置信息
- 缓存用户登录统计

## 安全考虑

### 1. 数据脱敏
- IP地址脱敏处理
- 地理位置信息脱敏
- 用户代理信息脱敏

### 2. 访问控制
- 登录日志查询需要认证
- 敏感操作需要权限验证

### 3. 数据保护
- 定期备份登录日志数据
- 数据加密存储

## 扩展功能

### 1. 地理位置解析
- 集成第三方IP地理位置服务
- 支持高精度地理位置定位

### 2. 安全分析
- 异常登录行为检测
- 地理位置异常检测
- 设备指纹分析

### 3. 报表功能
- 登录行为统计报表
- 地理位置分布报表
- 设备使用情况报表

## 总结

登录日志功能完整实现了用户登录行为的记录需求，包括：

1. ✅ **登录人**：通过 `user_id` 字段记录
2. ✅ **登录时间**：通过 `login_time` 专门字段记录
3. ✅ **登录地点**：通过 `location` 字段记录地理位置
4. ✅ **登录系统**：通过 `device_type` 和 `device_name` 记录
5. ✅ **登录方式**：通过 `login_type` 记录

该设计简洁实用，基于现有的 `UserBehaviorLog` 模型扩展，避免了过度复杂化，同时满足了所有功能需求。 