# 限流功能时间参数详细说明

## 核心设计原理

### 令牌桶算法
```go
type TokenBucket struct {
    tokens         int           // 当前令牌数量
    capacity       int           // 令牌桶最大容量
    refillRate     int           // 每个时间间隔补充的令牌数
    refillInterval time.Duration // 补充时间间隔
    lastRefill     time.Time     // 上次补充时间
    mu             sync.Mutex    // 互斥锁
}
```

### 时间参数计算逻辑
```go
func (tb *TokenBucket) Allow() bool {
    now := time.Now()
    elapsed := now.Sub(tb.lastRefill)
    intervals := int(elapsed / tb.refillInterval)
    
    if intervals > 0 {
        tokensToAdd := intervals * tb.refillRate
        tb.tokens += tokensToAdd
        if tb.tokens > tb.capacity {
            tb.tokens = tb.capacity
        }
        tb.lastRefill = tb.lastRefill.Add(time.Duration(intervals) * tb.refillInterval)
    }
    
    return tb.tokens > 0
}
```

## 预设配置详细说明

### 1. SMSRateLimitConfig - SMS短信限流
```go
SMSRateLimitConfig = RateLimitConfig{
    Capacity:       1,           // 令牌桶容量：1个
    RefillRate:     1,           // 每次补充：1个令牌
    RefillInterval: time.Minute, // 补充间隔：每分钟
    ErrorMsg:       "短信发送过于频繁，请稍后再试",
}
```

**实际效果**：
- 初始可发送1次短信
- 发送后需等待1分钟才能再次发送
- 最大突发量：1次
- 稳定发送频率：每分钟1次

**时间线示例**：
```
00:00 - 发送成功 (令牌: 1 -> 0)
00:30 - 发送失败 (令牌: 0, 未到补充时间)
01:00 - 补充令牌 (令牌: 0 -> 1)
01:01 - 发送成功 (令牌: 1 -> 0)
```

### 2. LoginRateLimitConfig - 登录限流
```go
LoginRateLimitConfig = RateLimitConfig{
    Capacity:       5,           // 令牌桶容量：5个
    RefillRate:     5,           // 每次补充：5个令牌
    RefillInterval: time.Minute, // 补充间隔：每分钟
    ErrorMsg:       "登录尝试过于频繁，请稍后再试",
}
```

**实际效果**：
- 初始可连续登录5次
- 用完后需等待1分钟补充5个令牌
- 最大突发量：5次
- 稳定登录频率：每分钟5次

**时间线示例**：
```
00:00 - 连续登录5次成功 (令牌: 5 -> 0)
00:30 - 登录失败 (令牌: 0, 未到补充时间)
01:00 - 补充令牌 (令牌: 0 -> 5)
01:01 - 可再次连续登录5次
```

### 3. APIRateLimitConfig - API限流
```go
APIRateLimitConfig = RateLimitConfig{
    Capacity:       10,          // 令牌桶容量：10个
    RefillRate:     10,          // 每次补充：10个令牌
    RefillInterval: time.Second, // 补充间隔：每秒
    ErrorMsg:       "请求过于频繁，请稍后再试",
}
```

**实际效果**：
- 初始可连续请求10次
- 每秒补充10个令牌
- 最大突发量：10次
- 稳定请求频率：每秒10次

**时间线示例**：
```
00:00.000 - 连续请求10次成功 (令牌: 10 -> 0)
00:00.500 - 请求失败 (令牌: 0, 未到补充时间)
00:01.000 - 补充令牌 (令牌: 0 -> 10)
00:01.100 - 可再次连续请求10次
```

### 4. StrictRateLimitConfig - 严格限流
```go
StrictRateLimitConfig = RateLimitConfig{
    Capacity:       3,           // 令牌桶容量：3个
    RefillRate:     3,           // 每次补充：3个令牌
    RefillInterval: time.Minute, // 补充间隔：每分钟
    ErrorMsg:       "操作过于频繁，请稍后再试",
}
```

**实际效果**：
- 初始可操作3次
- 每分钟补充3个令牌
- 最大突发量：3次
- 稳定操作频率：每分钟3次

### 5. LaxRateLimitConfig - 宽松限流
```go
LaxRateLimitConfig = RateLimitConfig{
    Capacity:       50,          // 令牌桶容量：50个
    RefillRate:     50,          // 每次补充：50个令牌
    RefillInterval: time.Second, // 补充间隔：每秒
    ErrorMsg:       "请求过于频繁，请稍后再试",
}
```

**实际效果**：
- 初始可连续请求50次
- 每秒补充50个令牌
- 最大突发量：50次
- 稳定请求频率：每秒50次

## 自定义配置示例

### 每小时1次的严格限制
```go
hourlyConfig := RateLimitConfig{
    Capacity:       1,
    RefillRate:     1,
    RefillInterval: time.Hour, // 每小时补充
    ErrorMsg:       "此操作每小时只能执行1次",
}
```

### 每10秒5次的中等限制
```go
mediumConfig := RateLimitConfig{
    Capacity:       5,
    RefillRate:     5,
    RefillInterval: 10 * time.Second, // 每10秒补充
    ErrorMsg:       "请求频率过高，请稍后再试",
}
```

### 每100毫秒1次的高频限制
```go
highFreqConfig := RateLimitConfig{
    Capacity:       1,
    RefillRate:     1,
    RefillInterval: 100 * time.Millisecond, // 每100毫秒补充
    ErrorMsg:       "请求间隔太短",
}
```

## 关键时间概念

### 1. 容量 (Capacity)
- 决定了**突发请求**的最大数量
- 也是令牌桶的上限
- 影响系统的瞬时处理能力

### 2. 补充速率 (RefillRate)
- 每个时间间隔补充的令牌数量
- 决定了**稳定状态**下的请求频率
- 与 RefillInterval 配合使用

### 3. 补充间隔 (RefillInterval)
- 令牌补充的时间间隔
- 支持任意 time.Duration：毫秒、秒、分钟、小时等
- 决定了限流的**时间粒度**

### 4. 实际限流频率计算
```
稳定频率 = RefillRate / RefillInterval
```

例如：
- RefillRate=5, RefillInterval=time.Minute → 每分钟5次 = 每12秒1次
- RefillRate=10, RefillInterval=time.Second → 每秒10次
- RefillRate=1, RefillInterval=time.Hour → 每小时1次

## 多接口隔离

每个IP的每个接口都有独立的令牌桶：
```go
endpointKey := c.Request.Method + ":" + c.FullPath()
// 例如: "POST:/api/v1/sms/send", "GET:/api/v1/users/profile"
```

这意味着：
- 同一IP访问不同接口有各自的限流计数
- 不同IP访问同一接口有各自的限流计数
- 提供了细粒度的限流控制

## 内存清理机制

访问者记录会在3小时无活动后自动清理：
```go
if time.Since(visitor.lastSeen) > 3*time.Hour {
    delete(rl.visitors, ip)
}
```

这确保了：
- 内存使用可控
- 长期运行的服务稳定性
- 自动清理不活跃的IP记录
