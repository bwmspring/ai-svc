# 多端设备管理系统设计文档

## 功能概述

本系统实现了多端设备管理功能，支持用户在不同设备上登录，并对设备数量进行限制。当用户设备数量达到上限时，系统会自动踢出最旧的设备。

## 设备类型支持

- **PC端** (`pc`): 桌面应用
- **iOS端** (`ios`): iOS移动应用
- **Android端** (`android`): Android移动应用
- **小程序** (`miniprogram`): 微信小程序等
- **Web端** (`web`): 网页应用

## 核心功能

### 1. 设备标识

每个设备通过以下信息进行标识：
- `device_id`: 设备唯一标识符
- `device_type`: 设备类型（pc/ios/android/miniprogram/web）
- `device_name`: 设备名称（用户可读）
- `app_version`: 应用版本号
- `os_version`: 操作系统版本

### 2. 设备数量限制

- 默认最多允许5台设备同时登录
- 超过限制时，自动踢出最旧的设备（按最后活跃时间排序）
- 支持动态配置设备数量限制

### 3. 强制下线机制

- 设备超限时自动踢出最旧设备
- 管理员可手动踢出指定设备
- 被踢出的设备会话立即失效

## 接口设计

### 1. 登录接口

```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "phone": "13800138000",
  "code": "123456",
  "device_info": {
    "device_id": "unique_device_id",
    "device_type": "ios",
    "device_name": "iPhone 13 Pro",
    "app_version": "1.0.0",
    "os_version": "iOS 15.0"
  }
}
```

**响应示例：**
```json
{
  "code": 200,
  "message": "登录成功",
  "data": {
    "user": {
      "id": 1,
      "phone": "13800138000",
      "nickname": "用户昵称",
      "avatar": "头像URL",
      "vip_level": 1,
      "status": 1,
      "created_at": "2024-01-01T00:00:00Z"
    },
    "token": "jwt_token_here"
  }
}
```

### 2. 获取用户设备列表

```http
GET /api/v1/users/devices
Authorization: Bearer <token>
```

**响应示例：**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "devices": [
      {
        "id": 1,
        "device_id": "device_1",
        "device_type": "ios",
        "device_name": "iPhone 13 Pro",
        "app_version": "1.0.0",
        "os_version": "iOS 15.0",
        "client_ip": "192.168.1.100",
        "status": 1,
        "login_at": "2024-01-01T10:00:00Z",
        "last_active_at": "2024-01-01T12:00:00Z",
        "is_online": true,
        "created_at": "2024-01-01T10:00:00Z"
      }
    ],
    "total_count": 3,
    "online_count": 2,
    "max_devices": 5
  }
}
```

### 3. 踢出设备

```http
POST /api/v1/users/devices/kick
Authorization: Bearer <token>
Content-Type: application/json

{
  "device_ids": ["device_1", "device_2"]
}
```

**响应示例：**
```json
{
  "code": 200,
  "message": "设备已被踢出",
  "data": null
}
```

## 数据库设计

### 1. 用户设备表 (user_devices)

```sql
CREATE TABLE user_devices (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    device_id VARCHAR(100) NOT NULL,
    device_type VARCHAR(20) NOT NULL,
    device_name VARCHAR(100),
    app_version VARCHAR(20),
    os_version VARCHAR(50),
    client_ip VARCHAR(45),
    user_agent VARCHAR(500),
    status TINYINT DEFAULT 1 COMMENT '状态 1:在线 0:离线',
    login_at TIMESTAMP NOT NULL,
    last_active_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_user_id (user_id),
    INDEX idx_device_id (device_id),
    INDEX idx_last_active (last_active_at)
);
```

### 2. 用户会话表 (user_sessions)

```sql
CREATE TABLE user_sessions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    device_id VARCHAR(100) NOT NULL,
    session_token VARCHAR(255) NOT NULL,
    jwt_token TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_user_id (user_id),
    INDEX idx_device_id (device_id),
    UNIQUE KEY uk_session_token (session_token)
);
```

## 客户端接入指南

### 1. 设备ID生成规则

不同平台生成设备ID的推荐方案：

- **iOS**: 使用 `UIDevice.current.identifierForVendor` 或 `KeyChain` 存储的UUID
- **Android**: 使用 `Settings.Secure.ANDROID_ID` 或生成UUID存储到SharedPreferences
- **Web**: 使用 `localStorage` 存储生成的UUID
- **PC**: 使用硬件信息生成唯一标识或UUID存储到本地文件
- **小程序**: 使用 `wx.getStorageSync` 存储生成的UUID

### 2. 设备信息获取

```javascript
// Web端示例
function getDeviceInfo() {
  return {
    device_id: localStorage.getItem('device_id') || generateUUID(),
    device_type: 'web',
    device_name: navigator.userAgent,
    app_version: '1.0.0',
    os_version: navigator.platform
  };
}

// 生成UUID
function generateUUID() {
  const uuid = 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
    const r = Math.random() * 16 | 0;
    const v = c === 'x' ? r : (r & 0x3 | 0x8);
    return v.toString(16);
  });
  localStorage.setItem('device_id', uuid);
  return uuid;
}
```

### 3. 登录流程

1. 客户端获取设备信息
2. 发送登录请求（包含设备信息）
3. 服务器验证验证码
4. 注册/更新设备信息
5. 检查设备数量限制
6. 超限时踢出最旧设备
7. 创建会话并返回Token

## 安全考虑

### 1. 设备验证

- 设备ID应该是唯一且不易伪造的
- 定期更新设备活跃时间
- 检测异常登录行为

### 2. 会话管理

- JWT Token包含设备ID信息
- 定期清理过期会话
- 支持会话刷新机制

### 3. 强制下线

- 被踢出的设备Token立即失效
- 支持批量踢出设备
- 记录设备操作日志

## 运维监控

### 1. 关键指标

- 用户平均设备数量
- 设备类型分布
- 强制下线频率
- 异常登录检测

### 2. 定时任务

- 清理过期会话
- 清理离线设备
- 设备活跃度统计

## 配置参数

```yaml
# config.yaml
device:
  max_devices: 5          # 最大设备数量
  session_timeout: 24h    # 会话超时时间
  offline_timeout: 30m    # 离线超时时间
  cleanup_interval: 1h    # 清理任务间隔
```

## 使用示例

### 前端登录实现

```javascript
// 登录函数
async function login(phone, code) {
  const deviceInfo = getDeviceInfo();
  
  const response = await fetch('/api/v1/auth/login', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      phone,
      code,
      device_info: deviceInfo
    })
  });
  
  const result = await response.json();
  
  if (result.code === 200) {
    // 保存Token
    localStorage.setItem('token', result.data.token);
    // 登录成功处理
    console.log('登录成功', result.data.user);
  } else {
    // 错误处理
    console.error('登录失败', result.message);
  }
}

// 获取设备列表
async function getDevices() {
  const token = localStorage.getItem('token');
  
  const response = await fetch('/api/v1/users/devices', {
    headers: {
      'Authorization': `Bearer ${token}`
    }
  });
  
  const result = await response.json();
  
  if (result.code === 200) {
    console.log('设备列表', result.data);
    return result.data;
  }
}

// 踢出设备
async function kickDevices(deviceIds) {
  const token = localStorage.getItem('token');
  
  const response = await fetch('/api/v1/users/devices/kick', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      device_ids: deviceIds
    })
  });
  
  const result = await response.json();
  
  if (result.code === 200) {
    console.log('设备踢出成功');
  }
}
```

这个多端设备管理系统提供了完整的设备标识、数量限制、强制下线等功能，满足了用户在不同设备上登录的需求，同时保证了系统的安全性和可管理性。
