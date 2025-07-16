package middleware

import (
	"ai-svc/pkg/response"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimitConfig 限流配置.
type RateLimitConfig struct {
	Capacity       int           // 令牌桶容量
	RefillRate     int           // 补充的令牌数
	RefillInterval time.Duration // 补充间隔（例如：time.Minute 表示每分钟补充）
	ErrorMsg       string        // 自定义错误消息
}

// RateLimiter 频率限制器.
type RateLimiter struct {
	visitors map[string]*Visitor
	mu       sync.RWMutex
}

// Visitor 访问者信息.
type Visitor struct {
	limiters map[string]*TokenBucket // 支持多个限流器（用于不同接口）
	lastSeen time.Time
}

// TokenBucket 令牌桶.
type TokenBucket struct {
	tokens         int
	capacity       int
	refillRate     int
	refillInterval time.Duration
	lastRefill     time.Time
	mu             sync.Mutex
}

// NewRateLimiter 创建频率限制器.
func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*Visitor),
	}

	// 启动清理协程
	go rl.cleanupVisitors()

	return rl
}

// NewTokenBucket 创建令牌桶.
func NewTokenBucket(capacity, refillRate int, refillInterval time.Duration) *TokenBucket {
	return &TokenBucket{
		tokens:         capacity,
		capacity:       capacity,
		refillRate:     refillRate,
		refillInterval: refillInterval,
		lastRefill:     time.Now(),
	}
}

// Allow 检查是否允许请求.
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	// 计算经过的时间间隔数
	elapsed := now.Sub(tb.lastRefill)
	intervals := int(elapsed / tb.refillInterval)

	if intervals > 0 {
		tokensToAdd := intervals * tb.refillRate
		tb.tokens += tokensToAdd
		if tb.tokens > tb.capacity {
			tb.tokens = tb.capacity
		}
		// 更新最后补充时间，对齐到间隔边界
		tb.lastRefill = tb.lastRefill.Add(time.Duration(intervals) * tb.refillInterval)
	}

	if tb.tokens > 0 {
		tb.tokens--
		return true
	}
	return false
}

// GetVisitor 获取访问者.
func (rl *RateLimiter) GetVisitor(ip string) *Visitor {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	visitor, exists := rl.visitors[ip]
	if !exists {
		visitor = &Visitor{
			limiters: make(map[string]*TokenBucket),
			lastSeen: time.Now(),
		}
		rl.visitors[ip] = visitor
	}

	visitor.lastSeen = time.Now()
	return visitor
}

// GetLimiter 获取指定接口的限流器.
func (v *Visitor) GetLimiter(endpoint string, config RateLimitConfig) *TokenBucket {
	limiter, exists := v.limiters[endpoint]
	if !exists {
		limiter = NewTokenBucket(config.Capacity, config.RefillRate, config.RefillInterval)
		v.limiters[endpoint] = limiter
	}
	return limiter
}

// cleanupVisitors 清理过期的访问者.
func (rl *RateLimiter) cleanupVisitors() {
	for {
		time.Sleep(time.Minute)
		rl.mu.Lock()
		for ip, visitor := range rl.visitors {
			if time.Since(visitor.lastSeen) > 3*time.Hour {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// DefaultRateLimitConfig 默认限流配置.
var DefaultRateLimitConfig = RateLimitConfig{
	Capacity:       1,
	RefillRate:     1,
	RefillInterval: time.Minute, // 每分钟补充1个令牌
	ErrorMsg:       "请求过于频繁，请稍后再试",
}

// CustomRateLimit 自定义限流中间件.
func CustomRateLimit(limiter *RateLimiter, config ...RateLimitConfig) gin.HandlerFunc {
	// 如果没有传入配置，使用默认配置
	cfg := DefaultRateLimitConfig
	if len(config) > 0 {
		cfg = config[0]
		// 填充未设置的字段为默认值
		if cfg.Capacity <= 0 {
			cfg.Capacity = DefaultRateLimitConfig.Capacity
		}
		if cfg.RefillRate <= 0 {
			cfg.RefillRate = DefaultRateLimitConfig.RefillRate
		}
		if cfg.RefillInterval <= 0 {
			cfg.RefillInterval = DefaultRateLimitConfig.RefillInterval
		}
		if cfg.ErrorMsg == "" {
			cfg.ErrorMsg = DefaultRateLimitConfig.ErrorMsg
		}
	}

	return func(c *gin.Context) {
		ip := c.ClientIP()
		endpoint := c.Request.Method + ":" + c.FullPath()

		visitor := limiter.GetVisitor(ip)
		tokenBucket := visitor.GetLimiter(endpoint, cfg)

		if !tokenBucket.Allow() {
			response.Error(c, response.ERROR, cfg.ErrorMsg)
			c.Abort()
			return
		}

		c.Next()
	}
}

// SMSRateLimit SMS发送频率限制中间件（保持向后兼容）.
func SMSRateLimit(limiter *RateLimiter) gin.HandlerFunc {
	smsConfig := RateLimitConfig{
		Capacity:       1,
		RefillRate:     1,
		RefillInterval: time.Minute, // 每分钟补充1个令牌
		ErrorMsg:       "发送过于频繁，请稍后再试",
	}
	return CustomRateLimit(limiter, smsConfig)
}
