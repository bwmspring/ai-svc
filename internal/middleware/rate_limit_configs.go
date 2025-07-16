package middleware

import "time"

// 预定义的限流配置
// RefillRate 表示每个 RefillInterval 补充的令牌数

// 容量：1个令牌，每分钟补充1个令牌 = 每分钟最多1次.
var SMSRateLimitConfig = RateLimitConfig{
	Capacity:       1,
	RefillRate:     1,
	RefillInterval: time.Minute, // 每分钟补充1个
	ErrorMsg:       "短信发送过于频繁，请稍后再试",
}

// 容量：5个令牌，每分钟补充5个令牌 = 每分钟最多5次.
var LoginRateLimitConfig = RateLimitConfig{
	Capacity:       5,
	RefillRate:     5,
	RefillInterval: time.Minute, // 每分钟补充5个
	ErrorMsg:       "登录尝试过于频繁，请稍后再试",
}

// 容量：10个令牌，每秒补充10个令牌 = 每秒最多10次.
var APIRateLimitConfig = RateLimitConfig{
	Capacity:       10,
	RefillRate:     10,
	RefillInterval: time.Second, // 每秒补充10个
	ErrorMsg:       "请求过于频繁，请稍后再试",
}

// 容量：3个令牌，每分钟补充3个令牌 = 每分钟最多3次.
var StrictRateLimitConfig = RateLimitConfig{
	Capacity:       3,
	RefillRate:     3,
	RefillInterval: time.Minute, // 每分钟补充3个
	ErrorMsg:       "操作过于频繁，请稍后再试",
}

// 容量：50个令牌，每秒补充50个令牌 = 每秒最多50次.
var LaxRateLimitConfig = RateLimitConfig{
	Capacity:       50,
	RefillRate:     50,
	RefillInterval: time.Second, // 每秒补充50个
	ErrorMsg:       "请求过于频繁，请稍后再试",
}
