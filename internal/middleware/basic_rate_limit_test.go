package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestBasicTokenBucket(t *testing.T) {
	// 测试令牌桶基本功能
	tb := NewTokenBucket(2, 1, time.Second)

	// 应该有2个初始令牌
	assert.True(t, tb.Allow(), "第一个令牌应该可用")
	assert.True(t, tb.Allow(), "第二个令牌应该可用")
	assert.False(t, tb.Allow(), "第三个令牌应该不可用")
}

func TestBasicRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rateLimiter := NewRateLimiter()
	router := gin.New()

	config := RateLimitConfig{
		Capacity:   1,
		RefillRate: 1,
		ErrorMsg:   "限流测试",
	}

	router.POST("/test", CustomRateLimit(rateLimiter, config), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// 第一个请求应该成功
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/test", nil)
	req1.Header.Set("X-Real-IP", "192.168.1.100")
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// 第二个请求应该被限流
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/test", nil)
	req2.Header.Set("X-Real-IP", "192.168.1.100")
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusBadRequest, w2.Code)
}

func TestMultipleIPs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rateLimiter := NewRateLimiter()
	router := gin.New()

	config := RateLimitConfig{
		Capacity:   1,
		RefillRate: 1,
		ErrorMsg:   "限流测试",
	}

	router.GET("/test", CustomRateLimit(rateLimiter, config), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// 不同IP的请求应该互不影响
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/test", nil)
	req1.Header.Set("X-Real-IP", "192.168.1.101")
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/test", nil)
	req2.Header.Set("X-Real-IP", "192.168.1.102")
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
}

func TestSMSRateLimitBackwardCompatibility(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rateLimiter := NewRateLimiter()
	router := gin.New()

	// 使用原有的SMSRateLimit函数
	router.POST("/sms", SMSRateLimit(rateLimiter), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "sms"})
	})

	// 第一个请求应该成功
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/sms", nil)
	req1.Header.Set("X-Real-IP", "192.168.1.103")
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// 第二个请求应该被限流
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/sms", nil)
	req2.Header.Set("X-Real-IP", "192.168.1.103")
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusBadRequest, w2.Code)
}
