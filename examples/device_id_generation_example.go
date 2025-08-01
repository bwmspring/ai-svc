package main

import (
	"ai-svc/internal/model"
	"ai-svc/internal/service"
	"ai-svc/pkg/logger"
	"fmt"
	"log"
	"time"
)

// 演示后端统一生成设备ID的流程
func main() {
	// 初始化logger
	if err := logger.Init("info", "text", "stdout"); err != nil {
		log.Fatalf("初始化logger失败: %v", err)
	}

	fmt.Println("🎯 后端统一生成设备ID示例")
	fmt.Println(repeat("=", 50))

	// 1. 创建设备ID生成服务
	idGenerator := service.NewDeviceIDGeneratorService()

	// 2. 模拟不同类型的设备注册请求
	devices := []struct {
		name    string
		request *model.DeviceRegistrationRequest
		userID  uint
	}{
		{
			name: "iPhone 15 Pro",
			request: &model.DeviceRegistrationRequest{
				DeviceFingerprint: "a1b2c3d4e5f6789abcdef123456789abcdef123456789abcdef123456789abcdef",
				DeviceType:        model.DeviceTypeIOS,
				DeviceName:        "iPhone 15 Pro",
				AppVersion:        "1.0.0",
				OSVersion:         "iOS 17.0",
				Platform:          "mobile",
				ClientInfo:        "Apple A17 Pro",
			},
			userID: 1001,
		},
		{
			name: "Samsung Galaxy S24",
			request: &model.DeviceRegistrationRequest{
				DeviceFingerprint: "b2c3d4e5f6789abcdef123456789abcdef123456789abcdef123456789abcdef1",
				DeviceType:        model.DeviceTypeAndroid,
				DeviceName:        "Samsung Galaxy S24",
				AppVersion:        "1.0.0",
				OSVersion:         "Android 14",
				Platform:          "mobile",
				ClientInfo:        "Snapdragon 8 Gen 3",
			},
			userID: 1002,
		},
		{
			name: "MacBook Pro M3",
			request: &model.DeviceRegistrationRequest{
				DeviceFingerprint: "c3d4e5f6789abcdef123456789abcdef123456789abcdef123456789abcdef12a",
				DeviceType:        model.DeviceTypePC,
				DeviceName:        "MacBook Pro M3",
				AppVersion:        "1.0.0",
				OSVersion:         "macOS 14.0",
				Platform:          "desktop",
				ClientInfo:        "Apple M3 Pro",
			},
			userID: 1003,
		},
		{
			name: "Chrome Browser",
			request: &model.DeviceRegistrationRequest{
				DeviceFingerprint: "d4e5f6789abcdef123456789abcdef123456789abcdef123456789abcdef12ab3",
				DeviceType:        model.DeviceTypeWeb,
				DeviceName:        "Chrome Browser",
				AppVersion:        "1.0.0",
				OSVersion:         "Windows 11",
				Platform:          "web",
				ClientInfo:        "Chrome 120.0",
			},
			userID: 1004,
		},
		{
			name: "微信小程序",
			request: &model.DeviceRegistrationRequest{
				DeviceFingerprint: "e5f6789abcdef123456789abcdef123456789abcdef123456789abcdef12ab34c",
				DeviceType:        model.DeviceTypeMiniprogram,
				DeviceName:        "微信小程序",
				AppVersion:        "1.0.0",
				OSVersion:         "微信 8.0.30",
				Platform:          "miniprogram",
				ClientInfo:        "WeChat MP",
			},
			userID: 1005,
		},
	}

	fmt.Println("\n📱 设备ID生成示例：")
	fmt.Println(repeat("-", 80))

	for i, device := range devices {
		fmt.Printf("\n%d. %s (用户ID: %d)\n", i+1, device.name, device.userID)

		// 生成设备指纹
		clientIP := fmt.Sprintf("192.168.1.%d", 100+i)
		fingerprint := idGenerator.GenerateFingerprint(device.request, clientIP)
		fmt.Printf("   📝 设备指纹: %s...\n", fingerprint[:32])

		// 生成设备ID
		deviceID, err := idGenerator.GenerateDeviceID(
			device.request.DeviceType,
			device.userID,
			fingerprint,
		)
		if err != nil {
			log.Printf("❌ 生成设备ID失败: %v", err)
			continue
		}

		fmt.Printf("   🆔 设备ID: %s\n", deviceID)

		// 验证设备ID格式
		isValid := idGenerator.ValidateDeviceID(deviceID)
		fmt.Printf("   ✅ 格式验证: %v\n", isValid)

		// 解析设备ID信息
		if info, err := idGenerator.ParseDeviceID(deviceID); err == nil {
			fmt.Printf("   📊 解析信息:\n")
			fmt.Printf("      - 前缀: %s\n", info.Prefix)
			fmt.Printf("      - 类型: %s\n", info.Type)
			fmt.Printf("      - 生成时间: %s\n", info.Timestamp.Format("2006-01-02 15:04:05"))
		}
	}

	fmt.Println("\n🔒 安全特性演示：")
	fmt.Println(repeat("-", 50))

	// 展示同一用户多次生成的唯一性
	userID := uint(2001)
	deviceType := model.DeviceTypeIOS
	baseFingerprint := "test_fingerprint_12345678"

	fmt.Printf("同一用户(%d)同一设备类型(%s)的多次生成：\n", userID, deviceType)
	for i := 0; i < 3; i++ {
		deviceID, _ := idGenerator.GenerateDeviceID(deviceType, userID, baseFingerprint)
		fmt.Printf("  第%d次: %s\n", i+1, deviceID)
		time.Sleep(time.Second) // 确保时间戳不同
	}

	fmt.Println("\n🎯 唯一性保证:")
	fmt.Println("  ✅ 时间戳确保时序唯一性")
	fmt.Println("  ✅ 随机数确保并发唯一性")
	fmt.Println("  ✅ 用户ID哈希确保用户隔离")
	fmt.Println("  ✅ 指纹摘要确保设备绑定")
	fmt.Println("  ✅ 校验和确保数据完整性")

	fmt.Println("\n🛡️ 安全优势:")
	fmt.Println("  🔐 服务端生成，无法伪造")
	fmt.Println("  🎲 包含随机数，无法预测")
	fmt.Println("  🔍 支持解析和验证")
	fmt.Println("  📝 完整的审计追踪")
	fmt.Println("  🚫 杜绝设备ID冲突")

	fmt.Println("\n✨ 实施建议:")
	fmt.Println("  1. 客户端生成设备指纹")
	fmt.Println("  2. 登录时发送指纹到服务端")
	fmt.Println("  3. 服务端生成唯一设备ID")
	fmt.Println("  4. 客户端保存设备ID")
	fmt.Println("  5. 后续请求携带设备ID")

	fmt.Println("\n🎉 后端统一生成设备ID方案演示完成！")
}

func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
