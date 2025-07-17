package service

import (
	"testing"

	"ai-svc/internal/config"
	"ai-svc/internal/model"
	"ai-svc/pkg/logger"
)

// TestJWTService 测试JWT服务.
func TestJWTService(t *testing.T) {
	// 加载配置
	if err := config.LoadConfig("../../configs/config.yaml"); err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 初始化logger
	if err := logger.Init("info", "text", "stdout"); err != nil {
		t.Fatalf("初始化logger失败: %v", err)
	}

	// 创建JWT服务
	jwtService := NewJWTService()

	// 模拟用户和设备数据
	user := &model.User{
		BaseModel: model.BaseModel{ID: 1},
		Phone:     "13800138000",
	}

	device := &model.UserDevice{
		DeviceID:   "test_device_123",
		DeviceType: "ios",
	}

	// 测试生成Token
	t.Run("GenerateToken", func(t *testing.T) {
		token, err := jwtService.GenerateToken(user, device, "session_123")
		if err != nil {
			t.Errorf("生成Token失败: %v", err)
			return
		}
		if token == "" {
			t.Error("生成的Token为空")
		}

		// 验证Token
		claims, err := jwtService.ValidateToken(token)
		if err != nil {
			t.Errorf("验证Token失败: %v", err)
			return
		}

		// 检查Claims内容
		if claims.UserID != user.ID {
			t.Errorf("用户ID不匹配: 期望%d, 实际%d", user.ID, claims.UserID)
		}
		if claims.Phone != user.Phone {
			t.Errorf("手机号不匹配: 期望%s, 实际%s", user.Phone, claims.Phone)
		}
		if claims.DeviceID != device.DeviceID {
			t.Errorf("设备ID不匹配: 期望%s, 实际%s", device.DeviceID, claims.DeviceID)
		}
		if claims.SessionID != "session_123" {
			t.Errorf("会话ID不匹配: 期望%s, 实际%s", "session_123", claims.SessionID)
		}
	})

	// 测试边界情况
	t.Run("EdgeCases", func(t *testing.T) {
		// 测试空用户
		_, err := jwtService.GenerateToken(nil, device, "session")
		if err == nil {
			t.Error("应该返回错误但没有")
		}

		// 测试空设备
		_, err = jwtService.GenerateToken(user, nil, "session")
		if err == nil {
			t.Error("应该返回错误但没有")
		}

		// 测试无效Token
		_, err = jwtService.ValidateToken("invalid_token")
		if err == nil {
			t.Error("应该验证失败但没有返回错误")
		}
	})
}
