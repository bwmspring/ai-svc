package main

import (
	"ai-svc/internal/model"
	"ai-svc/internal/repository"
	"ai-svc/internal/service"
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

// BroadcastMessageExample 广播消息发送示例
func BroadcastMessageExample() {
	// 初始化消息服务
	messageRepo := repository.NewMessageRepository()
	userRepo := repository.NewUserRepository()
	messageService := service.NewMessageService(messageRepo, userRepo)

	// 示例1: 发送系统维护通知（所有用户）
	fmt.Println("=== 示例1: 发送系统维护通知 ===")

	maintenanceReq := &model.SendBroadcastMessageRequest{
		Title:       "系统维护通知",
		Content:     "系统将于今晚22:00-24:00进行维护升级，期间可能影响正常使用，请提前做好准备。",
		MessageType: 1, // 系统通知
		Priority:    2, // 重要
		SenderID:    0, // 系统发送
		SenderType:  2, // 系统类型
		ExtraData:   `{"type": "maintenance", "duration": "2小时"}`,
		ExpireAt:    timePtr(time.Now().Add(24 * time.Hour)), // 24小时后过期
	}

	start := time.Now()
	err := messageService.SendBroadcastMessageAsync(context.Background(), maintenanceReq)
	if err != nil {
		log.Printf("发送系统维护通知失败: %v", err)
		return
	}

	duration := time.Since(start)
	fmt.Printf("✅ 系统维护通知发送成功，耗时: %v\n", duration)

	// 示例2: 发送新功能通知（指定用户组）
	fmt.Println("\n=== 示例2: 发送新功能通知 ===")

	// 假设有1000万用户，我们选择前1000个用户进行测试
	targetUsers := make([]uint, 1000)
	for i := 0; i < 1000; i++ {
		targetUsers[i] = uint(i + 1)
	}

	featureReq := &model.SendBroadcastMessageRequest{
		Title:       "新功能上线通知",
		Content:     "我们推出了全新的消息系统功能，支持更丰富的消息类型和更好的用户体验！",
		MessageType: 1, // 系统通知
		Priority:    1, // 普通
		SenderID:    0, // 系统发送
		SenderType:  2, // 系统类型
		TargetUsers: targetUsers,
		ExtraData:   `{"type": "feature", "version": "2.0"}`,
		ExpireAt:    timePtr(time.Now().Add(7 * 24 * time.Hour)), // 7天后过期
	}

	start = time.Now()
	err = messageService.SendBroadcastMessageAsync(context.Background(), featureReq)
	if err != nil {
		log.Printf("发送新功能通知失败: %v", err)
		return
	}

	duration = time.Since(start)
	fmt.Printf("✅ 新功能通知发送成功，耗时: %v\n", duration)

	// 示例3: 批量发送个性化消息
	fmt.Println("\n=== 示例3: 批量发送个性化消息 ===")

	// 模拟为不同用户发送个性化消息
	personalizedMessages := []struct {
		UserID  uint
		Title   string
		Content string
	}{
		{1, "个性化推荐", "根据您的使用习惯，为您推荐了新的AI功能。"},
		{2, "账户安全提醒", "您的账户安全状态良好，继续保持！"},
		{3, "使用统计", "本月您已使用AI服务50次，感谢您的支持！"},
	}

	for _, msg := range personalizedMessages {
		personalReq := &model.SendMessageRequest{
			Title:       msg.Title,
			Content:     msg.Content,
			MessageType: 4, // 业务通知
			Priority:    1, // 普通
			RecipientID: msg.UserID,
			SenderID:    0, // 系统发送
			SenderType:  2, // 系统类型
			ExtraData:   `{"type": "personalized"}`,
		}

		err := messageService.SendMessageAsync(context.Background(), personalReq)
		if err != nil {
			log.Printf("发送个性化消息失败 (用户ID: %d): %v", msg.UserID, err)
			continue
		}
		fmt.Printf("✅ 个性化消息发送成功 (用户ID: %d)\n", msg.UserID)
	}
}

// 性能测试示例
func PerformanceTestExample() {
	fmt.Println("\n=== 性能测试示例 ===")

	messageRepo := repository.NewMessageRepository()
	userRepo := repository.NewUserRepository()
	messageService := service.NewMessageService(messageRepo, userRepo)

	// 测试大规模用户消息发送性能
	userCounts := []int{100, 1000, 10000, 100000}

	for _, count := range userCounts {
		fmt.Printf("\n测试 %d 个用户的消息发送性能:\n", count)

		// 生成测试用户ID
		targetUsers := make([]uint, count)
		for i := 0; i < count; i++ {
			targetUsers[i] = uint(i + 1)
		}

		// 发送测试消息
		testReq := &model.SendBroadcastMessageRequest{
			Title:       fmt.Sprintf("性能测试消息 (%d用户)", count),
			Content:     "这是一条性能测试消息，用于测试大规模用户消息发送的性能。",
			MessageType: 1,
			Priority:    1,
			SenderID:    0,
			SenderType:  2,
			TargetUsers: targetUsers,
		}

		start := time.Now()
		err := messageService.SendBroadcastMessageAsync(context.Background(), testReq)
		duration := time.Since(start)

		if err != nil {
			fmt.Printf("❌ 发送失败: %v\n", err)
		} else {
			fmt.Printf("✅ 发送成功，耗时: %v\n", duration)
			fmt.Printf("   平均每用户耗时: %v\n", duration/time.Duration(count))
		}
	}
}

// 消息查询性能测试
func QueryPerformanceTestExample() {
	fmt.Println("\n=== 消息查询性能测试 ===")

	messageRepo := repository.NewMessageRepository()
	userRepo := repository.NewUserRepository()
	messageService := service.NewMessageService(messageRepo, userRepo)

	// 测试不同用户的消息查询性能
	testUsers := []uint{1, 100, 1000, 10000}

	for _, userID := range testUsers {
		fmt.Printf("\n测试用户 %d 的消息查询性能:\n", userID)

		// 测试获取消息列表
		params := &model.MessageQueryParams{
			Page: 1,
			Size: 20,
		}

		start := time.Now()
		response, err := messageService.GetUserMessages(userID, params)
		duration := time.Since(start)

		if err != nil {
			fmt.Printf("❌ 查询失败: %v\n", err)
		} else {
			fmt.Printf("✅ 查询成功，耗时: %v\n", duration)
			fmt.Printf("   返回消息数量: %d\n", len(response.Messages))
			fmt.Printf("   总消息数量: %d\n", response.Pagination.Total)
		}

		// 测试获取未读消息数量
		start = time.Now()
		unreadResponse, err := messageService.GetUnreadCount(userID)
		duration = time.Since(start)

		if err != nil {
			fmt.Printf("❌ 未读数量查询失败: %v\n", err)
		} else {
			fmt.Printf("✅ 未读数量查询成功，耗时: %v\n", duration)
			fmt.Printf("   未读消息数量: %d\n", unreadResponse.UnreadCount)
		}
	}
}

// 消息状态管理示例
func MessageStatusManagementExample() {
	fmt.Println("\n=== 消息状态管理示例 ===")

	messageRepo := repository.NewMessageRepository()
	userRepo := repository.NewUserRepository()
	messageService := service.NewMessageService(messageRepo, userRepo)
	userID := uint(1)

	// 1. 获取用户消息列表
	fmt.Println("1. 获取用户消息列表:")
	params := &model.MessageQueryParams{
		Page: 1,
		Size: 10,
	}

	response, err := messageService.GetUserMessages(userID, params)
	if err != nil {
		log.Printf("获取消息列表失败: %v", err)
		return
	}

	fmt.Printf("   总消息数量: %d\n", response.Pagination.Total)

	// 2. 标记消息为已读
	if len(response.Messages) > 0 {
		fmt.Println("\n2. 标记消息为已读:")
		messageIDs := make([]uint, 0, len(response.Messages))
		for _, msg := range response.Messages {
			if !msg.IsRead {
				messageIDs = append(messageIDs, msg.ID)
			}
		}

		if len(messageIDs) > 0 {
			batchReq := &model.BatchReadRequest{
				MessageIDs: messageIDs,
			}
			err := messageService.BatchMarkAsRead(batchReq, userID)
			if err != nil {
				log.Printf("标记已读失败: %v", err)
			} else {
				fmt.Printf("   成功标记 %d 条消息为已读\n", len(messageIDs))
			}
		}
	}

	// 3. 删除消息
	if len(response.Messages) > 0 {
		fmt.Println("\n3. 删除消息:")
		messageID := response.Messages[0].ID

		err := messageService.DeleteMessage(messageID, userID)
		if err != nil {
			log.Printf("删除消息失败: %v", err)
		} else {
			fmt.Printf("   成功删除消息 (ID: %d)\n", messageID)
		}
	}

	// 4. 获取未读消息数量
	fmt.Println("\n4. 获取未读消息数量:")
	unreadResponse, err := messageService.GetUnreadCount(userID)
	if err != nil {
		log.Printf("获取未读数量失败: %v", err)
	} else {
		fmt.Printf("   未读消息数量: %d\n", unreadResponse.UnreadCount)
	}
}

// 数据清理示例
func DataCleanupExample() {
	fmt.Println("\n=== 数据清理示例 ===")

	// 启动清理任务（服务初始化时已启动）
	fmt.Println("1. 清理任务已启动:")
	fmt.Println("   ✅ 过期消息清理任务运行中")
	fmt.Println("   ✅ 数据归档任务运行中")

	// 2. 获取消息统计
	fmt.Println("\n2. 获取消息统计:")
	// 注意：这里需要实现GetMessageStats方法，暂时跳过
	fmt.Println("   📊 消息统计功能需要额外实现")
}

// 主函数
func main() {
	fmt.Println("🚀 站内信系统示例程序")
	fmt.Println(strings.Repeat("=", 50))

	// 运行各种示例
	BroadcastMessageExample()
	PerformanceTestExample()
	QueryPerformanceTestExample()
	MessageStatusManagementExample()
	DataCleanupExample()

	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("✅ 所有示例执行完成")
}

// 辅助函数
func timePtr(t time.Time) *time.Time {
	return &t
}
