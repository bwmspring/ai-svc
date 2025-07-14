package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCustomRateLimit(t *testing.T) {
	// 设置 Gin 为测试模式
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		config       RateLimitConfig
		requestCount int
		expectedPass int
		expectedFail int
		waitBetween  time.Duration
		description  string
	}{
		{
			name: "默认配置测试",
			config: RateLimitConfig{
				Capacity:   1,
				RefillRate: 1,
				ErrorMsg:   "默认限流触发",
			},
			requestCount: 3,
			expectedPass: 1,
			expectedFail: 2,
			waitBetween:  0,
			description:  "容量1，立即发送3个请求，应该只有1个通过",
		},
		{
			name:         "SMS限流配置测试",
			config:       SMSRateLimitConfig,
			requestCount: 2,
			expectedPass: 1,
			expectedFail: 1,
			waitBetween:  0,
			description:  "SMS配置，立即发送2个请求，应该只有1个通过",
		},
		{
			name:         "API限流配置测试",
			config:       APIRateLimitConfig,
			requestCount: 15,
			expectedPass: 10,
			expectedFail: 5,
			waitBetween:  0,
			description:  "API配置容量10，发送15个请求，应该有10个通过",
		},
		{
			name: "带等待时间的测试",
			config: RateLimitConfig{
				Capacity:   1,
				RefillRate: 2, // 每秒补充2个令牌
				ErrorMsg:   "等待测试限流",
			},
			requestCount: 3,
			expectedPass: 2,
			expectedFail: 1,
			waitBetween:  600 * time.Millisecond, // 等待0.6秒，应该能补充1个令牌
			description:  "测试令牌补充机制",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建限流器
			rateLimiter := NewRateLimiter()

			// 创建路由
			router := gin.New()
			router.POST("/test", CustomRateLimit(rateLimiter, tt.config), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			passCount := 0
			failCount := 0

			// 发送请求
			for i := 0; i < tt.requestCount; i++ {
				if i > 0 && tt.waitBetween > 0 {
					time.Sleep(tt.waitBetween)
				}

				w := httptest.NewRecorder()
				req, _ := http.NewRequest("POST", "/test", nil)
				req.Header.Set("X-Real-IP", "192.168.1.100") // 模拟同一IP

				router.ServeHTTP(w, req)

				if w.Code == http.StatusOK {
					passCount++
				} else {
					failCount++
				}
			}

			assert.Equal(t, tt.expectedPass, passCount, "通过的请求数不符合预期: %s", tt.description)
			assert.Equal(t, tt.expectedFail, failCount, "失败的请求数不符合预期: %s", tt.description)
		})
	}
}

func TestMultipleEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建限流器
	rateLimiter := NewRateLimiter()

	// 创建路由，不同接口使用不同配置
	router := gin.New()

	smsConfig := RateLimitConfig{Capacity: 1, RefillRate: 1, ErrorMsg: "SMS限流"}
	loginConfig := RateLimitConfig{Capacity: 3, RefillRate: 3, ErrorMsg: "登录限流"}

	router.POST("/sms", CustomRateLimit(rateLimiter, smsConfig), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"endpoint": "sms"})
	})

	router.POST("/login", CustomRateLimit(rateLimiter, loginConfig), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"endpoint": "login"})
	})

	// 测试SMS接口
	t.Run("SMS接口限流", func(t *testing.T) {
		// 第一个请求应该成功
		w1 := httptest.NewRecorder()
		req1, _ := http.NewRequest("POST", "/sms", nil)
		req1.Header.Set("X-Real-IP", "192.168.1.101")
		router.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusOK, w1.Code)

		// 第二个请求应该被限流
		w2 := httptest.NewRecorder()
		req2, _ := http.NewRequest("POST", "/sms", nil)
		req2.Header.Set("X-Real-IP", "192.168.1.101")
		router.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusBadRequest, w2.Code)
	})

	// 测试登录接口
	t.Run("登录接口限流", func(t *testing.T) {
		successCount := 0

		// 发送5个请求
		for i := 0; i < 5; i++ {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/login", nil)
			req.Header.Set("X-Real-IP", "192.168.1.102")
			router.ServeHTTP(w, req)

			if w.Code == http.StatusOK {
				successCount++
			}
		}

		// 应该有3个成功（容量为3）
		assert.Equal(t, 3, successCount)
	})

	// 测试不同IP不互相影响
	t.Run("不同IP独立限流", func(t *testing.T) {
		// IP1的请求
		w1 := httptest.NewRecorder()
		req1, _ := http.NewRequest("POST", "/sms", nil)
		req1.Header.Set("X-Real-IP", "192.168.1.103")
		router.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusOK, w1.Code)

		// IP2的请求，应该不受IP1影响
		w2 := httptest.NewRecorder()
		req2, _ := http.NewRequest("POST", "/sms", nil)
		req2.Header.Set("X-Real-IP", "192.168.1.104")
		router.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusOK, w2.Code)
	})
}

func TestTokenBucketRefill(t *testing.T) {
	// 测试令牌桶的补充机制
	tb := NewTokenBucket(2, 1, time.Second) // 容量2，每秒补充1个

	// 消耗所有令牌
	assert.True(t, tb.Allow())  // 第1个
	assert.True(t, tb.Allow())  // 第2个
	assert.False(t, tb.Allow()) // 第3个应该失败

	// 等待1.1秒，应该补充1个令牌
	time.Sleep(1100 * time.Millisecond)
	assert.True(t, tb.Allow())  // 现在应该有1个令牌
	assert.False(t, tb.Allow()) // 再次消耗后应该没有
}

func TestRateLimitConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		config   RateLimitConfig
		expected RateLimitConfig
	}{
		{
			name: "完整配置",
			config: RateLimitConfig{
				Capacity:   5,
				RefillRate: 10,
				ErrorMsg:   "自定义错误",
			},
			expected: RateLimitConfig{
				Capacity:   5,
				RefillRate: 10,
				ErrorMsg:   "自定义错误",
			},
		},
		{
			name: "部分配置自动补全",
			config: RateLimitConfig{
				Capacity: 3,
				// RefillRate 和 ErrorMsg 未设置
			},
			expected: RateLimitConfig{
				Capacity:   3,
				RefillRate: DefaultRateLimitConfig.RefillRate,
				ErrorMsg:   DefaultRateLimitConfig.ErrorMsg,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rateLimiter := NewRateLimiter()
			router := gin.New()

			// 应用限流中间件
			router.GET("/test", CustomRateLimit(rateLimiter, tt.config), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "ok"})
			})

			// 验证配置是否正确应用（通过行为验证）
			// 这里我们测试容量限制
			successCount := 0
			for i := 0; i < tt.expected.Capacity+2; i++ {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/test", nil)
				req.Header.Set("X-Real-IP", "192.168.1.200")
				router.ServeHTTP(w, req)

				if w.Code == http.StatusOK {
					successCount++
				}
			}

			assert.Equal(t, tt.expected.Capacity, successCount)
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rateLimiter := NewRateLimiter()
	router := gin.New()

	// 不传配置，使用默认配置
	router.GET("/default", CustomRateLimit(rateLimiter), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "default"})
	})

	// 第一个请求应该成功
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/default", nil)
	req1.Header.Set("X-Real-IP", "192.168.1.201")
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// 第二个请求应该被限流（默认容量为1）
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/default", nil)
	req2.Header.Set("X-Real-IP", "192.168.1.201")
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusBadRequest, w2.Code)
}

func TestBackwardCompatibility(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rateLimiter := NewRateLimiter()
	router := gin.New()

	// 使用原有的SMSRateLimit函数
	router.POST("/sms-old", SMSRateLimit(rateLimiter), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "sms-old"})
	})

	// 第一个请求应该成功
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/sms-old", nil)
	req1.Header.Set("X-Real-IP", "192.168.1.202")
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// 第二个请求应该被限流
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/sms-old", nil)
	req2.Header.Set("X-Real-IP", "192.168.1.202")
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusBadRequest, w2.Code)
}

// BenchmarkRateLimit 性能测试
func BenchmarkRateLimit(b *testing.B) {
	gin.SetMode(gin.TestMode)

	rateLimiter := NewRateLimiter()
	router := gin.New()

	config := RateLimitConfig{
		Capacity:   1000,
		RefillRate: 1000,
		ErrorMsg:   "限流",
	}

	router.GET("/bench", CustomRateLimit(rateLimiter, config), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "bench"})
	})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/bench", nil)
		req.Header.Set("X-Real-IP", "192.168.1.100")
		router.ServeHTTP(w, req)
	}
}
