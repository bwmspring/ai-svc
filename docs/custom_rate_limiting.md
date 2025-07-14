# 定制化限流功能使用指南

## 概述

重构后的限流功能支持针对不同接口设置不同的限流参数，提供了更加灵活和强大的限流能力。

## 核心特性

1. **定制化配置**: 每个接口可以有独立的限流配置
2. **预设配置**: 提供常用的限流配置模板
3. **向后兼容**: 保持原有 `SMSRateLimit` 函数的兼容性
4. **多接口支持**: 同一IP的不同接口有独立的令牌桶
5. **自动清理**: 自动清理过期的访问者记录

## 配置结构

```go
type RateLimitConfig struct {
    Capacity   int    // 令牌桶容量
    RefillRate int    // 每秒补充的令牌数
    ErrorMsg   string // 自定义错误消息
}
```

## 预设配置

### SMSRateLimitConfig
- **用途**: SMS短信发送
- **限制**: 每分钟1次
- **容量**: 1
- **补充速率**: 每秒1个令牌

### LoginRateLimitConfig
- **用途**: 登录接口
- **限制**: 每分钟5次
- **容量**: 5
- **补充速率**: 每秒5个令牌

### APIRateLimitConfig
- **用途**: 一般API接口
- **限制**: 每秒10次
- **容量**: 10
- **补充速率**: 每秒10个令牌

### StrictRateLimitConfig
- **用途**: 敏感操作
- **限制**: 每分钟3次
- **容量**: 3
- **补充速率**: 每秒3个令牌

### LaxRateLimitConfig
- **用途**: 查询类操作
- **限制**: 每秒50次
- **容量**: 50
- **补充速率**: 每秒50个令牌

## 使用方法

### 1. 使用预设配置

```go
// 在路由设置中
rateLimiter := middleware.NewRateLimiter()

// 使用SMS限流配置
api.POST("/sms/send", middleware.CustomRateLimit(rateLimiter, middleware.SMSRateLimitConfig), controller.SendSMS)

// 使用登录限流配置
api.POST("/auth/login", middleware.CustomRateLimit(rateLimiter, middleware.LoginRateLimitConfig), controller.Login)

// 使用一般API限流配置
api.GET("/profile", middleware.CustomRateLimit(rateLimiter, middleware.APIRateLimitConfig), controller.GetProfile)
```

### 2. 使用自定义配置

```go
// 创建自定义配置
customConfig := middleware.RateLimitConfig{
    Capacity:   2,           // 容量2
    RefillRate: 1,           // 每秒补充1个令牌
    ErrorMsg:   "操作过于频繁，请2秒后再试",
}

// 应用自定义配置
api.POST("/special-operation", middleware.CustomRateLimit(rateLimiter, customConfig), controller.SpecialOperation)
```

### 3. 使用默认配置

```go
// 不传配置参数，使用默认配置
api.GET("/default", middleware.CustomRateLimit(rateLimiter), controller.Default)
```

### 4. 向后兼容用法

```go
// 原有的SMS限流中间件仍然可用
api.POST("/sms/send", middleware.SMSRateLimit(rateLimiter), controller.SendSMS)
```

## 限流策略建议

### 短信发送接口
- **配置**: `SMSRateLimitConfig`
- **原因**: 防止短信轰炸，保护费用

### 登录接口
- **配置**: `LoginRateLimitConfig`
- **原因**: 防止暴力破解，允许合理重试

### 查询接口
- **配置**: `LaxRateLimitConfig`
- **原因**: 查询操作相对安全，可以较为宽松

### 写入操作
- **配置**: `APIRateLimitConfig`
- **原因**: 平衡性能和安全

### 敏感操作（删除、踢设备等）
- **配置**: `StrictRateLimitConfig`
- **原因**: 防止误操作和恶意操作

## 实现原理

1. **多接口隔离**: 每个IP访问者维护多个令牌桶，按 `方法:路径` 区分
2. **令牌桶算法**: 支持突发流量，平滑限流
3. **自动清理**: 3小时未活跃的访问者会被自动清理
4. **线程安全**: 使用互斥锁保证并发安全

## 配置参数说明

### Capacity (容量)
- 令牌桶的最大容量
- 决定允许的突发请求数量
- 建议值: 1-100

### RefillRate (补充速率)
- 每秒补充的令牌数
- 决定稳定状态下的请求频率
- 建议值: 1-100

### ErrorMsg (错误消息)
- 限流触发时返回的错误消息
- 建议包含具体的等待时间提示

## 监控和调试

可以通过日志观察限流效果：
- 查看请求ID和IP
- 观察限流触发频率
- 分析不同接口的访问模式

## 性能考虑

1. **内存使用**: 每个IP和接口组合占用少量内存
2. **清理策略**: 自动清理机制防止内存泄漏
3. **锁竞争**: 使用读写锁优化并发性能
