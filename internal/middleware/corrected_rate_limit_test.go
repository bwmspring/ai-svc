package middleware

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCorrectedTokenBucket(t *testing.T) {
	// 测试每分钟补充1个令牌的令牌桶
	tb := NewTokenBucket(1, 1, time.Minute)

	// 第一个请求应该成功
	assert.True(t, tb.Allow(), "第一个请求应该成功")

	// 第二个请求应该失败（没有令牌了）
	assert.False(t, tb.Allow(), "第二个请求应该失败")

	// 等待超过1分钟后，应该能再次请求成功
	tb.lastRefill = tb.lastRefill.Add(-65 * time.Second) // 模拟时间过去
	assert.True(t, tb.Allow(), "等待后的请求应该成功")
}

func TestTokenBucketWithSecondInterval(t *testing.T) {
	// 测试每秒补充2个令牌的令牌桶
	tb := NewTokenBucket(5, 2, time.Second)

	// 消耗所有5个令牌
	for i := 0; i < 5; i++ {
		assert.True(t, tb.Allow(), "初始令牌应该可用")
	}

	// 第6个应该失败
	assert.False(t, tb.Allow(), "超过容量的请求应该失败")

	// 模拟时间过去2.5秒，应该补充5个令牌（2.5秒 * 2个/秒 = 5个，但容量限制为5）
	tb.lastRefill = tb.lastRefill.Add(-2500 * time.Millisecond)

	// 现在应该有5个令牌可用
	for i := 0; i < 5; i++ {
		assert.True(t, tb.Allow(), "补充后的令牌应该可用")
	}

	// 第6个应该失败
	assert.False(t, tb.Allow(), "超过容量的请求应该失败")
}

func TestSMSRateLimitConfig(t *testing.T) {
	// 验证SMS配置的实际效果
	config := SMSRateLimitConfig

	assert.Equal(t, 1, config.Capacity, "SMS限流容量应该为1")
	assert.Equal(t, 1, config.RefillRate, "SMS限流每次补充1个令牌")
	assert.Equal(t, time.Minute, config.RefillInterval, "SMS限流每分钟补充一次")

	// 创建令牌桶测试
	tb := NewTokenBucket(config.Capacity, config.RefillRate, config.RefillInterval)

	// 第一个请求成功
	assert.True(t, tb.Allow(), "第一次SMS请求应该成功")

	// 第二个请求失败
	assert.False(t, tb.Allow(), "第二次SMS请求应该失败")

	// 模拟1分钟后
	tb.lastRefill = tb.lastRefill.Add(-61 * time.Second)

	// 现在应该可以再次请求
	assert.True(t, tb.Allow(), "1分钟后的SMS请求应该成功")
}

func TestLoginRateLimitConfig(t *testing.T) {
	// 验证登录配置的实际效果
	config := LoginRateLimitConfig

	assert.Equal(t, 5, config.Capacity, "登录限流容量应该为5")
	assert.Equal(t, 5, config.RefillRate, "登录限流每次补充5个令牌")
	assert.Equal(t, time.Minute, config.RefillInterval, "登录限流每分钟补充一次")

	tb := NewTokenBucket(config.Capacity, config.RefillRate, config.RefillInterval)

	// 连续5次请求应该都成功
	for i := 0; i < 5; i++ {
		assert.True(t, tb.Allow(), "前5次登录请求应该成功")
	}

	// 第6次应该失败
	assert.False(t, tb.Allow(), "第6次登录请求应该失败")

	// 模拟1分钟后，应该补充5个令牌
	tb.lastRefill = tb.lastRefill.Add(-61 * time.Second)

	// 现在又可以连续5次成功
	for i := 0; i < 5; i++ {
		assert.True(t, tb.Allow(), "1分钟后的登录请求应该成功")
	}
}

func TestAPIRateLimitConfig(t *testing.T) {
	// 验证API配置的实际效果
	config := APIRateLimitConfig

	assert.Equal(t, 10, config.Capacity, "API限流容量应该为10")
	assert.Equal(t, 10, config.RefillRate, "API限流每次补充10个令牌")
	assert.Equal(t, time.Second, config.RefillInterval, "API限流每秒补充一次")

	tb := NewTokenBucket(config.Capacity, config.RefillRate, config.RefillInterval)

	// 连续10次请求应该都成功
	for i := 0; i < 10; i++ {
		assert.True(t, tb.Allow(), "前10次API请求应该成功")
	}

	// 第11次应该失败
	assert.False(t, tb.Allow(), "第11次API请求应该失败")

	// 模拟1秒后，应该补充10个令牌
	tb.lastRefill = tb.lastRefill.Add(-1100 * time.Millisecond)

	// 现在又可以连续10次成功
	for i := 0; i < 10; i++ {
		assert.True(t, tb.Allow(), "1秒后的API请求应该成功")
	}
}
