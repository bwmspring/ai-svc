# 用户系统重构总结

## 重构目标
将原有的用户名+密码注册登录系统，重构为手机号+验证码的免密登录系统。

## 主要变更

### 1. 数据模型变更 (`internal/model/user.go`)

#### 用户模型 (User)
- **移除**：`Password` 字段
- **调整**：`Phone` 字段设为唯一索引和必填项
- **调整**：`Username` 和 `Email` 字段改为可选
- **新增**：VIP 等级相关字段
  - `VIPLevel`: VIP等级 (0:普通用户 1:VIP1 2:VIP2 3:VIP3)
  - `VIPExpireAt`: VIP过期时间
- **优化**：登录行为字段
  - `LastLoginIP`: 最后登录IP
  - `LastLoginAt`: 最后登录时间
  - `LoginCount`: 登录次数

#### 新增模型
- **SMSVerificationCode**: 短信验证码模型
  - `Phone`: 手机号
  - `Code`: 验证码
  - `Purpose`: 用途 (login, register, reset等)
  - `UsedAt`: 使用时间
  - `ExpiredAt`: 过期时间

#### 请求/响应模型
- **新增**：`SendSMSRequest` - 发送短信验证码请求
- **新增**：`LoginWithSMSRequest` - 手机号+验证码登录请求
- **新增**：`LoginResponse` - 登录响应（包含用户信息和token）
- **移除**：`CreateUserRequest`、`LoginRequest`、`ChangePasswordRequest`
- **更新**：`UpdateUserRequest` - 支持更多用户资料字段

### 2. 服务层变更 (`internal/service/user.go`)

#### 新增方法
- `SendSMS()`: 发送短信验证码
- `LoginWithSMS()`: 手机号+验证码登录（同时完成注册）
- `validateSMSCode()`: 验证短信验证码
- `generateVerificationCode()`: 生成6位随机验证码

#### 移除方法
- `CreateUser()`: 创建用户
- `Login()`: 用户名+密码登录
- `ChangePassword()`: 修改密码

#### 核心逻辑
- **一键登录注册**：用户首次使用手机号+验证码登录时自动创建账户
- **验证码管理**：验证码5分钟过期，使用后自动失效
- **登录信息跟踪**：记录登录IP、时间、次数等信息

### 3. 数据访问层变更 (`internal/repository/user.go`)

#### 新增方法
- `GetByPhone()`: 根据手机号获取用户
- `CreateSMSCode()`: 创建短信验证码记录
- `GetSMSCode()`: 获取短信验证码
- `UpdateSMSCode()`: 更新短信验证码状态

#### 移除方法
- `UpdatePassword()`: 更新密码
- `UpdateLastLogin()`: 更新最后登录信息（集成到用户更新中）

### 4. 控制器变更 (`internal/controller/user.go`)

#### 新增方法
- `SendSMS()`: 发送短信验证码接口
- `LoginWithSMS()`: 手机号+验证码登录接口

#### 移除方法
- `Register()`: 用户注册接口
- `Login()`: 用户名+密码登录接口
- `ChangePassword()`: 修改密码接口

### 5. 路由变更 (`internal/routes/routes.go`)

#### 新增路由
- `POST /api/v1/sms/send`: 发送短信验证码
- `POST /api/v1/auth/login`: 手机号+验证码登录

#### 移除路由
- `POST /api/v1/register`: 用户注册
- `POST /api/v1/login`: 用户名+密码登录
- `POST /api/v1/users/change-password`: 修改密码

## 设计优势

### 1. 用户体验提升
- **无密码困扰**：用户无需记住密码，降低使用门槛
- **快速注册**：首次登录即完成注册，流程简化
- **安全性提升**：验证码时效性短，安全性更高

### 2. 数据库设计优化
- **避免过度拆分**：将常用用户信息集中存储，减少JOIN查询
- **合理冗余**：在规范化和查询性能之间找到平衡
- **行为日志独立**：高频变更的行为数据独立存储

### 3. 系统架构清晰
- **单一职责**：每个模块职责明确
- **易于扩展**：支持后续添加更多登录方式
- **兼容性好**：保留用户名和邮箱字段，支持未来功能扩展

## 数据库表结构

### users 表
```sql
CREATE TABLE users (
    id bigint PRIMARY KEY AUTO_INCREMENT,
    phone varchar(20) UNIQUE NOT NULL,
    username varchar(50) UNIQUE,
    email varchar(100) UNIQUE,
    nickname varchar(50),
    avatar varchar(255),
    real_name varchar(50),
    gender tinyint DEFAULT 0,
    vip_level tinyint DEFAULT 0,
    vip_expire_at datetime,
    status tinyint DEFAULT 1,
    last_login_ip varchar(45),
    last_login_at datetime,
    login_count int DEFAULT 0,
    birthday datetime,
    address varchar(255),
    bio text,
    created_at datetime,
    updated_at datetime,
    deleted_at datetime
);
```

### sms_verification_codes 表
```sql
CREATE TABLE sms_verification_codes (
    id bigint PRIMARY KEY AUTO_INCREMENT,
    phone varchar(20) NOT NULL,
    code varchar(10) NOT NULL,
    purpose varchar(20) NOT NULL,
    used_at datetime,
    expired_at datetime NOT NULL,
    created_at datetime,
    INDEX idx_phone_purpose (phone, purpose)
);
```

## API 接口

### 1. 发送短信验证码
```
POST /api/v1/sms/send
{
    "phone": "13800138000",
    "purpose": "login"
}
```

### 2. 手机号+验证码登录
```
POST /api/v1/auth/login
{
    "phone": "13800138000",
    "code": "123456"
}
```

### 3. 获取用户信息
```
GET /api/v1/users/profile
Authorization: Bearer <token>
```

### 4. 更新用户信息
```
PUT /api/v1/users/profile
Authorization: Bearer <token>
{
    "nickname": "新昵称",
    "avatar": "头像URL",
    "real_name": "真实姓名",
    "gender": 1
}
```

## 注意事项

1. **短信服务集成**：当前使用日志记录验证码，生产环境需要集成真实的短信服务
2. **验证码安全**：建议添加发送频率限制和IP限制
3. **数据迁移**：如果有旧数据，需要编写迁移脚本
4. **JWT密钥管理**：确保JWT密钥的安全存储和定期更换

这次重构实现了用户系统的现代化改造，提升了用户体验和系统安全性，同时保持了良好的代码架构和扩展性。
