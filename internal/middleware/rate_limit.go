package middleware

import (
	"sync"
	"time"

	"ai-svc/pkg/response"

	"github.com/gin-gonic/gin"
)

// RateLimiter 频率限制器
type RateLimiter struct {
	visitors map[string]*Visitor
	mu       sync.RWMutex
}

// Visitor 访问者信息
type Visitor struct {
	limiter  *TokenBucket
	lastSeen time.Time
}

// TokenBucket 令牌桶
type TokenBucket struct {
	tokens     int
	capacity   int
	refillRate int
	lastRefill time.Time
	mu         sync.Mutex
}

// NewRateLimiter 创建频率限制器
func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*Visitor),
	}

	// 启动清理协程
	go rl.cleanupVisitors()

	return rl
}

// NewTokenBucket 创建令牌桶
func NewTokenBucket(capacity, refillRate int) *TokenBucket {
	return &TokenBucket{
		tokens:     capacity,
		capacity:   capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Allow 检查是否允许请求
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	// 计算应该添加的令牌数
	elapsed := now.Sub(tb.lastRefill)
	tokensToAdd := int(elapsed.Seconds()) * tb.refillRate

	if tokensToAdd > 0 {
		tb.tokens += tokensToAdd
		if tb.tokens > tb.capacity {
			tb.tokens = tb.capacity
		}
		tb.lastRefill = now
	}

	if tb.tokens > 0 {
		tb.tokens--
		return true
	}
	return false
}

// GetVisitor 获取访问者
func (rl *RateLimiter) GetVisitor(ip string) *Visitor {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	visitor, exists := rl.visitors[ip]
	if !exists {
		// 创建新的访问者，SMS接口限制：每分钟最多1次，每小时最多5次
		visitor = &Visitor{
			limiter:  NewTokenBucket(1, 1), // 容量1，每60秒补充1个令牌
			lastSeen: time.Now(),
		}
		rl.visitors[ip] = visitor
	}

	visitor.lastSeen = time.Now()
	return visitor
}

// cleanupVisitors 清理过期的访问者
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

// SMSRateLimit SMS发送频率限制中间件
func SMSRateLimit(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		visitor := limiter.GetVisitor(ip)

		if !visitor.limiter.Allow() {
			response.Error(c, response.ERROR, "发送过于频繁，请稍后再试")
			c.Abort()
			return
		}

		c.Next()
	}
}
