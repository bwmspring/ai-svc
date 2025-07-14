# 强制下线机制详细实现说明

## 概述

当用户设备数量超过限制时，系统会自动踢出最旧的设备，确保新设备能够正常登录。强制下线是通过删除设备记录和相关会话来实现的。

## 实现机制

### 1. 触发条件

强制下线在以下情况下触发：
- 用户新设备登录时
- 当前设备数量 >= 最大允许设备数量（默认5台）
- 系统自动调用 `CheckDeviceLimit` 方法

### 2. 核心实现流程

```go
// 设备登录时的检查流程
func (s *deviceService) RegisterDevice(userID uint, deviceInfo *model.DeviceInfo, clientIP, userAgent string) (*model.UserDevice, error) {
    // 1. 检查设备是否已存在
    existingDevice, err := s.deviceRepo.GetDeviceByDeviceID(deviceInfo.DeviceID)
    if existingDevice != nil {
        // 更新现有设备信息（不占用新的设备位置）
        return s.updateExistingDevice(existingDevice, deviceInfo, clientIP, userAgent)
    }
    
    // 2. 新设备需要检查数量限制
    if err := s.CheckDeviceLimit(userID); err != nil {
        return nil, err
    }
    
    // 3. 创建新设备记录
    return s.createNewDevice(userID, deviceInfo, clientIP, userAgent)
}
```

### 3. 设备数量限制检查

```go
// CheckDeviceLimit 检查设备数量限制
func (s *deviceService) CheckDeviceLimit(userID uint) error {
    // 统计当前用户设备数量
    count, err := s.deviceRepo.CountUserDevices(userID)
    if err != nil {
        return err
    }
    
    // 如果设备数量达到或超过上限，踢出最旧设备
    if int(count) >= s.maxDevices {
        if err := s.KickOldestDevice(userID); err != nil {
            return errors.New("设备数量已达上限，且无法踢出旧设备")
        }
    }
    
    return nil
}
```

### 4. 踢出最旧设备的核心逻辑

```go
// KickOldestDevice 踢出最旧的设备
func (s *deviceService) KickOldestDevice(userID uint) error {
    // 1. 获取用户所有设备（按最后活跃时间降序排列）
    devices, err := s.deviceRepo.GetUserDevices(userID)
    if err != nil {
        return err
    }
    
    if len(devices) == 0 {
        return nil
    }
    
    // 2. 找到最旧的设备（数组最后一个元素）
    // 因为查询时使用了 ORDER BY last_active_at DESC
    // 所以最旧的设备在数组末尾
    oldestDevice := devices[len(devices)-1]
    
    // 3. 删除设备的所有会话（强制下线）
    if err := s.deviceRepo.DeleteDeviceSessions(oldestDevice.DeviceID); err != nil {
        logger.Error("删除最旧设备会话失败", map[string]interface{}{
            "error": err.Error(), 
            "device_id": oldestDevice.DeviceID
        })
    }
    
    // 4. 删除设备记录
    if err := s.deviceRepo.DeleteDevice(oldestDevice.ID); err != nil {
        return err
    }
    
    // 5. 记录操作日志
    logger.Info("踢出最旧设备", map[string]interface{}{
        "user_id":     userID,
        "device_id":   oldestDevice.DeviceID,
        "device_name": oldestDevice.DeviceName,
        "last_active": oldestDevice.LastActiveAt,
    })
    
    return nil
}
```

### 5. 数据库层面的实现

#### 设备查询排序
```go
// GetUserDevices 获取用户所有设备（按最后活跃时间降序）
func (r *deviceRepository) GetUserDevices(userID uint) ([]*model.UserDevice, error) {
    var devices []*model.UserDevice
    err := r.db.Where("user_id = ?", userID).
        Order("last_active_at DESC").  // 最活跃的在前，最旧的在后
        Find(&devices).Error
    return devices, err
}
```

#### 会话删除
```go
// DeleteDeviceSessions 删除设备的所有会话
func (r *deviceRepository) DeleteDeviceSessions(deviceID string) error {
    return r.db.Where("device_id = ?", deviceID).Delete(&model.UserSession{}).Error
}
```

#### 设备删除
```go
// DeleteDevice 删除设备
func (r *deviceRepository) DeleteDevice(id uint) error {
    return r.db.Delete(&model.UserDevice{}, id).Error
}
```

## 强制下线的具体效果

### 1. 立即效果
- **会话失效**: 被踢出设备的所有JWT Token立即失效
- **API拒绝**: 该设备后续的API请求会返回401未授权
- **设备记录删除**: 设备从用户设备列表中移除

### 2. 用户体验
- 被踢出的设备会在下次API调用时收到"会话已过期"错误
- 用户需要重新登录才能继续使用
- 新设备可以正常登录和使用

## 优化策略

### 1. 智能选择踢出设备
目前的实现是踢出最旧设备，也可以考虑其他策略：

```go
// 可以根据不同策略选择要踢出的设备
func (s *deviceService) selectDeviceToKick(devices []*model.UserDevice) *model.UserDevice {
    // 策略1: 踢出最旧设备（当前实现）
    return devices[len(devices)-1]
    
    // 策略2: 踢出离线设备
    // for _, device := range devices {
    //     if !device.IsOnline() {
    //         return device
    //     }
    // }
    
    // 策略3: 踢出特定类型设备（优先级低的）
    // devicePriority := map[string]int{
    //     "web": 1,
    //     "pc": 2,
    //     "android": 3,
    //     "ios": 4,
    //     "miniprogram": 5,
    // }
}
```

### 2. 通知机制
```go
// 可以添加通知机制，告知用户设备被踢出
func (s *deviceService) notifyDeviceKicked(device *model.UserDevice) {
    // 发送推送通知
    // 发送邮件通知
    // 记录操作日志
    logger.Info("设备被强制下线", map[string]interface{}{
        "user_id":     device.UserID,
        "device_id":   device.DeviceID,
        "device_name": device.DeviceName,
        "device_type": device.DeviceType,
        "kick_time":   time.Now(),
    })
}
```

### 3. 软下线 vs 硬下线
```go
// 软下线：只标记设备为离线，不删除记录
func (s *deviceService) softKickDevice(deviceID string) error {
    return s.deviceRepo.MarkDeviceOffline(deviceID)
}

// 硬下线：删除设备记录和会话（当前实现）
func (s *deviceService) hardKickDevice(deviceID string) error {
    // 删除会话
    s.deviceRepo.DeleteDeviceSessions(deviceID)
    // 删除设备记录
    return s.deviceRepo.DeleteDeviceByDeviceID(deviceID)
}
```

## 测试场景

### 1. 基本测试
```bash
# 1. 登录5台设备
curl -X POST "/api/v1/auth/login" -d '{"phone":"13800138000","code":"123456","device_info":{"device_id":"device_1","device_type":"ios"}}'
curl -X POST "/api/v1/auth/login" -d '{"phone":"13800138000","code":"123456","device_info":{"device_id":"device_2","device_type":"android"}}'
# ... 继续登录到5台设备

# 2. 登录第6台设备（应该踢出最旧的）
curl -X POST "/api/v1/auth/login" -d '{"phone":"13800138000","code":"123456","device_info":{"device_id":"device_6","device_type":"web"}}'

# 3. 查看设备列表（应该只有5台设备）
curl -X GET "/api/v1/users/devices" -H "Authorization: Bearer <token>"
```

### 2. 并发测试
```go
// 测试多个设备同时登录的情况
func TestConcurrentLogin(t *testing.T) {
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(deviceNum int) {
            defer wg.Done()
            // 模拟设备登录
            loginWithDevice(fmt.Sprintf("device_%d", deviceNum))
        }(i)
    }
    wg.Wait()
    
    // 验证最终只有5台设备
    devices := getUserDevices(userID)
    assert.Equal(t, 5, len(devices))
}
```

## 配置和监控

### 1. 配置项
```yaml
# config.yaml
device:
  max_devices: 5          # 最大设备数量
  kick_strategy: "oldest" # 踢出策略: oldest, offline, priority
  notify_kicked: true     # 是否通知被踢出的设备
  cleanup_interval: 1h    # 清理任务间隔
```

### 2. 监控指标
- 设备强制下线频率
- 各设备类型分布
- 用户平均设备数量
- 异常登录检测

## 总结

强制下线机制通过以下步骤实现：

1. **检查触发**: 新设备登录时检查设备数量限制
2. **选择目标**: 按最后活跃时间排序，选择最旧设备
3. **删除会话**: 删除目标设备的所有会话记录
4. **删除设备**: 删除目标设备记录
5. **记录日志**: 记录强制下线操作

这个机制确保了：
- 用户设备数量不超过限制
- 新设备可以正常登录
- 被踢出的设备立即失效
- 操作过程可追溯和监控

通过这种方式，系统能够有效管理多端设备登录，保证系统资源的合理使用和用户体验的平衡。
