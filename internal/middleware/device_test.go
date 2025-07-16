package middleware

import (
	"ai-svc/pkg/logger"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDeviceService 模拟设备服务.
type MockDeviceService struct {
	mock.Mock
}

func (m *MockDeviceService) ValidateDeviceSession(userID uint, deviceID, sessionID string) (bool, error) {
	args := m.Called(userID, deviceID, sessionID)
	return args.Bool(0), args.Error(1)
}

func (m *MockDeviceService) UpdateDeviceActivity(deviceID string) error {
	args := m.Called(deviceID)
	return args.Error(0)
}

// 测试辅助函数：创建测试上下文.
func createTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// 设置请求
	req := httptest.NewRequest("GET", "/test", nil)
	c.Request = req

	return c, w
}

// 设置设备信息到上下文.
func setDeviceContext(c *gin.Context, userID uint, deviceID, sessionID, requestID string) {
	c.Set("user_id", userID)
	c.Set("device_id", deviceID)
	c.Set("session_id", sessionID)
	c.Set("request_id", requestID)
}

func TestMain(m *testing.M) {
	// 初始化logger用于测试
	logger.Init("error", "text", "stdout")
	m.Run()
}

// 测试 DeviceInfo 结构体和方法.
func TestDeviceInfo(t *testing.T) {
	t.Run("extractDeviceInfo", func(t *testing.T) {
		c, _ := createTestContext()
		setDeviceContext(c, 123, "device_123", "session_456", "req_789")

		deviceInfo := extractDeviceInfo(c)

		assert.Equal(t, uint(123), deviceInfo.UserID)
		assert.Equal(t, "device_123", deviceInfo.DeviceID)
		assert.Equal(t, "session_456", deviceInfo.SessionID)
		assert.Equal(t, "req_789", deviceInfo.RequestID)
	})

	t.Run("validateBasicInfo", func(t *testing.T) {
		tests := []struct {
			name     string
			userID   uint
			deviceID string
			expected bool
		}{
			{"有效信息", 123, "device_123", true},
			{"用户ID为0", 0, "device_123", false},
			{"设备ID为空", 123, "", false},
			{"都无效", 0, "", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				deviceInfo := &DeviceInfo{
					UserID:   tt.userID,
					DeviceID: tt.deviceID,
				}
				assert.Equal(t, tt.expected, deviceInfo.validateBasicInfo())
			})
		}
	})

	t.Run("logFields", func(t *testing.T) {
		deviceInfo := &DeviceInfo{
			RequestID: "req_123",
			UserID:    456,
			DeviceID:  "device_789",
			SessionID: "session_abc",
		}

		fields := deviceInfo.logFields()

		assert.Equal(t, "req_123", fields["request_id"])
		assert.Equal(t, uint(456), fields["user_id"])
		assert.Equal(t, "device_789", fields["device_id"])
		assert.Equal(t, "session_abc", fields["session_id"])
	})

	t.Run("logFields_无SessionID", func(t *testing.T) {
		deviceInfo := &DeviceInfo{
			RequestID: "req_123",
			UserID:    456,
			DeviceID:  "device_789",
		}

		fields := deviceInfo.logFields()

		assert.Equal(t, "req_123", fields["request_id"])
		assert.Equal(t, uint(456), fields["user_id"])
		assert.Equal(t, "device_789", fields["device_id"])
		assert.NotContains(t, fields, "session_id")
	})
}

// 测试配置.
func TestDeviceValidationConfig(t *testing.T) {
	t.Run("DefaultDeviceValidationConfig", func(t *testing.T) {
		config := DefaultDeviceValidationConfig()

		assert.True(t, config.Enabled)
		assert.False(t, config.RequireSessionID)
		assert.Equal(t, DefaultDeviceValidationTimeout, config.Timeout)
		assert.True(t, config.UpdateActivity)
	})
}

// 测试基本的设备验证中间件.
func TestDeviceValidationMiddleware(t *testing.T) {
	t.Run("成功验证_有SessionID", func(t *testing.T) {
		mockService := new(MockDeviceService)
		mockService.On("ValidateDeviceSession", uint(123), "device_123", "session_456").Return(true, nil)
		mockService.On("UpdateDeviceActivity", "device_123").Return(nil)

		c, w := createTestContext()
		setDeviceContext(c, 123, "device_123", "session_456", "req_789")

		middleware := DeviceValidationMiddleware(mockService)
		middleware(c)

		assert.False(t, c.IsAborted())
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("成功验证_无SessionID", func(t *testing.T) {
		mockService := new(MockDeviceService)
		mockService.On("UpdateDeviceActivity", "device_123").Return(nil)

		c, w := createTestContext()
		setDeviceContext(c, 123, "device_123", "", "req_789")

		middleware := DeviceValidationMiddleware(mockService)
		middleware(c)

		assert.False(t, c.IsAborted())
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("基本信息验证失败_用户ID为0", func(t *testing.T) {
		mockService := new(MockDeviceService)

		c, w := createTestContext()
		setDeviceContext(c, 0, "device_123", "session_456", "req_789")

		middleware := DeviceValidationMiddleware(mockService)
		middleware(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), ErrMsgUserOrDeviceInfoMissing)
	})

	t.Run("基本信息验证失败_设备ID为空", func(t *testing.T) {
		mockService := new(MockDeviceService)

		c, w := createTestContext()
		setDeviceContext(c, 123, "", "session_456", "req_789")

		middleware := DeviceValidationMiddleware(mockService)
		middleware(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), ErrMsgUserOrDeviceInfoMissing)
	})

	t.Run("设备会话验证失败_服务错误", func(t *testing.T) {
		mockService := new(MockDeviceService)
		mockService.On("ValidateDeviceSession", uint(123), "device_123", "session_456").
			Return(false, errors.New("数据库连接失败"))

		c, w := createTestContext()
		setDeviceContext(c, 123, "device_123", "session_456", "req_789")

		middleware := DeviceValidationMiddleware(mockService)
		middleware(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), ErrMsgDeviceSessionValidationFailed)
		mockService.AssertExpectations(t)
	})

	t.Run("设备会话验证失败_会话无效", func(t *testing.T) {
		mockService := new(MockDeviceService)
		mockService.On("ValidateDeviceSession", uint(123), "device_123", "session_456").Return(false, nil)

		c, w := createTestContext()
		setDeviceContext(c, 123, "device_123", "session_456", "req_789")

		middleware := DeviceValidationMiddleware(mockService)
		middleware(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), ErrMsgDeviceSessionInvalid)
		mockService.AssertExpectations(t)
	})

	t.Run("更新设备活跃时间失败", func(t *testing.T) {
		mockService := new(MockDeviceService)
		mockService.On("ValidateDeviceSession", uint(123), "device_123", "session_456").Return(true, nil)
		mockService.On("UpdateDeviceActivity", "device_123").Return(errors.New("更新失败"))

		c, w := createTestContext()
		setDeviceContext(c, 123, "device_123", "session_456", "req_789")

		middleware := DeviceValidationMiddleware(mockService)
		middleware(c)

		// 更新活跃时间失败不应该中断请求
		assert.False(t, c.IsAborted())
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})
}

// 测试带配置的设备验证中间件.
func TestDeviceValidationMiddlewareWithConfig(t *testing.T) {
	t.Run("禁用设备验证", func(t *testing.T) {
		mockService := new(MockDeviceService)
		config := &DeviceValidationConfig{
			Enabled: false,
		}

		c, w := createTestContext()
		setDeviceContext(c, 0, "", "", "req_789") // 设置无效数据

		middleware := DeviceValidationMiddlewareWithConfig(mockService, config)
		middleware(c)

		assert.False(t, c.IsAborted())
		assert.Equal(t, http.StatusOK, w.Code)
		// 不应该调用任何mock方法
		mockService.AssertNotCalled(t, "ValidateDeviceSession")
		mockService.AssertNotCalled(t, "UpdateDeviceActivity")
	})

	t.Run("要求SessionID_但未提供", func(t *testing.T) {
		mockService := new(MockDeviceService)
		config := &DeviceValidationConfig{
			Enabled:          true,
			RequireSessionID: true,
		}

		c, w := createTestContext()
		setDeviceContext(c, 123, "device_123", "", "req_789")

		middleware := DeviceValidationMiddlewareWithConfig(mockService, config)
		middleware(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), ErrMsgDeviceSessionInvalid)
	})

	t.Run("禁用设备活跃时间更新", func(t *testing.T) {
		mockService := new(MockDeviceService)
		mockService.On("ValidateDeviceSession", uint(123), "device_123", "session_456").Return(true, nil)
		config := &DeviceValidationConfig{
			Enabled:        true,
			UpdateActivity: false,
			Timeout:        DefaultDeviceValidationTimeout, // 明确设置超时时间
		}

		c, w := createTestContext()
		setDeviceContext(c, 123, "device_123", "session_456", "req_789")

		middleware := DeviceValidationMiddlewareWithConfig(mockService, config)
		middleware(c)

		assert.False(t, c.IsAborted())
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertNotCalled(t, "UpdateDeviceActivity")
		mockService.AssertExpectations(t)
	})

	t.Run("超时测试", func(t *testing.T) {
		mockService := new(MockDeviceService)
		// 模拟慢响应的验证服务
		mockService.On("ValidateDeviceSession", uint(123), "device_123", "session_456").
			Return(true, nil).
			Run(func(args mock.Arguments) {
				time.Sleep(200 * time.Millisecond) // 延迟200ms
			})

		config := &DeviceValidationConfig{
			Enabled: true,
			Timeout: 100 * time.Millisecond, // 设置100ms超时
		}

		c, w := createTestContext()
		setDeviceContext(c, 123, "device_123", "session_456", "req_789")

		middleware := DeviceValidationMiddlewareWithConfig(mockService, config)
		middleware(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), ErrMsgDeviceSessionValidationFailed)
	})
}

// 测试组合中间件.
func TestAuthWithDeviceValidation(t *testing.T) {
	t.Run("认证失败时不执行设备验证", func(t *testing.T) {
		mockService := new(MockDeviceService)

		c, _ := createTestContext()
		// 不设置任何认证信息，模拟认证失败

		// 模拟认证中间件中断了请求
		c.Abort()
		c.JSON(http.StatusUnauthorized, gin.H{"error": "认证失败"})

		middleware := AuthWithDeviceValidation(mockService)
		middleware(c)

		assert.True(t, c.IsAborted())
		// 设备服务不应该被调用
		mockService.AssertNotCalled(t, "ValidateDeviceSession")
		mockService.AssertNotCalled(t, "UpdateDeviceActivity")
	})
}

// 测试设备验证开关功能.
func TestDeviceValidationEnabled(t *testing.T) {
	t.Run("SetDeviceValidationEnabled", func(t *testing.T) {
		c, _ := createTestContext()

		middleware := SetDeviceValidationEnabled(false)
		middleware(c)

		enabled := IsDeviceValidationEnabled(c)
		assert.False(t, enabled)
	})

	t.Run("IsDeviceValidationEnabled_默认启用", func(t *testing.T) {
		c, _ := createTestContext()

		enabled := IsDeviceValidationEnabled(c)
		assert.True(t, enabled) // 默认启用
	})

	t.Run("IsDeviceValidationEnabled_已设置", func(t *testing.T) {
		c, _ := createTestContext()
		c.Set(ContextKeyDeviceValidation, false)

		enabled := IsDeviceValidationEnabled(c)
		assert.False(t, enabled)
	})
}

// 测试辅助函数.
func TestMergeMap(t *testing.T) {
	map1 := map[string]any{
		"key1": "value1",
		"key2": 123,
	}
	map2 := map[string]any{
		"key3": "value3",
		"key2": 456, // 覆盖map1中的key2
	}

	result := mergeMap(map1, map2)

	assert.Equal(t, "value1", result["key1"])
	assert.Equal(t, 456, result["key2"]) // 应该被map2覆盖
	assert.Equal(t, "value3", result["key3"])
	assert.Len(t, result, 3)
}

// 测试自定义错误类型.
func TestDeviceValidationError(t *testing.T) {
	err := &DeviceValidationError{Message: "测试错误"}
	assert.Equal(t, "测试错误", err.Error())
}

// 性能测试.
func BenchmarkDeviceValidationMiddleware(b *testing.B) {
	mockService := new(MockDeviceService)
	mockService.On("ValidateDeviceSession", mock.AnythingOfType("uint"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).
		Return(true, nil)
	mockService.On("UpdateDeviceActivity", mock.AnythingOfType("string")).Return(nil)

	middleware := DeviceValidationMiddleware(mockService)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c, _ := createTestContext()
		setDeviceContext(c, 123, "device_123", "session_456", "req_789")
		middleware(c)
	}
}

// 集成测试示例.
func TestDeviceValidationMiddleware_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockDeviceService)
	mockService.On("ValidateDeviceSession", uint(123), "device_123", "session_456").Return(true, nil)
	mockService.On("UpdateDeviceActivity", "device_123").Return(nil)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		// 模拟JWT中间件设置用户信息
		c.Set("request_id", "integration_test_req")
		c.Set("user_id", uint(123))
		c.Set("device_id", "device_123")
		c.Set("session_id", "session_456")
		c.Next()
	})
	router.Use(DeviceValidationMiddleware(mockService))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
	mockService.AssertExpectations(t)
}

// 测试错误场景的表格驱动测试.
func TestDeviceValidationMiddleware_ErrorScenarios(t *testing.T) {
	tests := []struct {
		name           string
		userID         uint
		deviceID       string
		sessionID      string
		setupMock      func(*MockDeviceService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "用户ID无效",
			userID:         0,
			deviceID:       "device_123",
			sessionID:      "session_456",
			setupMock:      func(m *MockDeviceService) {},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  ErrMsgUserOrDeviceInfoMissing,
		},
		{
			name:           "设备ID无效",
			userID:         123,
			deviceID:       "",
			sessionID:      "session_456",
			setupMock:      func(m *MockDeviceService) {},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  ErrMsgUserOrDeviceInfoMissing,
		},
		{
			name:      "验证服务返回错误",
			userID:    123,
			deviceID:  "device_123",
			sessionID: "session_456",
			setupMock: func(m *MockDeviceService) {
				m.On("ValidateDeviceSession", uint(123), "device_123", "session_456").
					Return(false, errors.New("service error"))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  ErrMsgDeviceSessionValidationFailed,
		},
		{
			name:      "会话无效",
			userID:    123,
			deviceID:  "device_123",
			sessionID: "session_456",
			setupMock: func(m *MockDeviceService) {
				m.On("ValidateDeviceSession", uint(123), "device_123", "session_456").Return(false, nil)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  ErrMsgDeviceSessionInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockDeviceService)
			tt.setupMock(mockService)

			c, w := createTestContext()
			setDeviceContext(c, tt.userID, tt.deviceID, tt.sessionID, "req_789")

			middleware := DeviceValidationMiddleware(mockService)
			middleware(c)

			assert.True(t, c.IsAborted())
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedError)
			mockService.AssertExpectations(t)
		})
	}
}
