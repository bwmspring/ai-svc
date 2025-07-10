# 项目架构设计文档

## 概述

本项目是一个基于Go语言的企业级后端服务架构，采用了经典的分层架构设计，集成了主流的Go生态工具和最佳实践。

## 技术选型

### 核心框架
- **Gin**: 高性能的HTTP Web框架
- **GORM**: 功能丰富的ORM库
- **Viper**: 灵活的配置管理库
- **Logrus**: 结构化日志库

### 数据库
- **MySQL 8.0**: 主数据库
- **Redis**: 缓存和会话存储

### 安全
- **JWT**: 无状态身份认证
- **bcrypt**: 密码加密

### 容器化
- **Docker**: 容器化部署
- **Docker Compose**: 本地开发环境

## 架构设计

### 目录结构说明

```
ai-svc/
├── cmd/                    # 应用程序入口点
│   └── server/            # 服务器主程序
├── internal/              # 内部应用代码（不对外暴露）
│   ├── config/           # 配置管理
│   ├── controller/       # HTTP控制器（控制器层）
│   ├── service/          # 业务逻辑层
│   ├── repository/       # 数据访问层
│   ├── routes/           # 路由配置
│   ├── model/            # 数据模型和DTO
│   └── middleware/       # 中间件
├── pkg/                   # 可重用的公共库
│   ├── database/         # 数据库连接管理
│   ├── logger/           # 日志工具
│   └── response/         # 统一响应格式
├── configs/              # 配置文件
└── docs/                 # 文档
```

### 分层架构

```
┌─────────────────┐
│   HTTP Client   │
└─────────────────┘
         │
┌─────────────────┐
│   Middleware    │  ← CORS, JWT, 日志, 恢复
└─────────────────┘
         │
┌─────────────────┐
│   Controller    │  ← HTTP请求处理
└─────────────────┘
         │
┌─────────────────┐
│    Service      │  ← 业务逻辑
└─────────────────┘
         │
┌─────────────────┐
│   Repository    │  ← 数据访问
└─────────────────┘
         │
┌─────────────────┐
│    Database     │  ← MySQL/Redis
└─────────────────┘
```

## 设计原则

### 1. 单一职责原则
- 每个层次都有明确的职责
- Controller负责HTTP请求处理和参数验证
- Service负责业务逻辑
- Repository负责数据访问

### 2. 依赖倒置原则
- Service层依赖Repository接口，而不是具体实现
- 通过接口定义契约，便于单元测试和扩展

### 3. 开闭原则
- 通过接口设计，便于扩展新功能
- 中间件机制支持功能插拔

### 4. 配置外部化
- 所有配置项都可通过配置文件或环境变量设置
- 支持不同环境的配置切换

## 核心功能

### 1. 用户管理
- 用户注册/登录
- 用户信息管理
- 密码管理
- 用户列表和搜索

### 2. 身份认证
- JWT令牌认证
- 中间件自动验证
- 令牌过期处理

### 3. 安全特性
- 密码bcrypt加密
- CORS跨域支持
- 输入参数验证
- SQL注入防护（通过GORM）

### 4. 运维支持
- 结构化日志
- 健康检查接口
- 优雅关闭
- Docker容器化

## 数据模型

### 用户表 (users)
```sql
CREATE TABLE users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    nickname VARCHAR(50),
    avatar VARCHAR(255),
    status TINYINT DEFAULT 1,
    last_ip VARCHAR(45),
    last_time DATETIME,
    created_at DATETIME,
    updated_at DATETIME,
    deleted_at DATETIME
);
```

### 用户资料表 (user_profiles)
```sql
CREATE TABLE user_profiles (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    real_name VARCHAR(20),
    phone VARCHAR(20),
    gender TINYINT DEFAULT 0,
    birthday DATETIME,
    address VARCHAR(255),
    bio TEXT,
    created_at DATETIME,
    updated_at DATETIME,
    deleted_at DATETIME
);
```

## API设计

### 响应格式标准
```json
{
    "code": 200,
    "message": "操作成功",
    "data": {}
}
```

### 分页响应格式
```json
{
    "code": 200,
    "message": "查询成功",
    "data": [],
    "total": 100,
    "page": 1,
    "size": 10
}
```

### 错误码规范
- 200: 成功
- 400: 参数错误
- 401: 未授权
- 403: 禁止访问
- 404: 资源不存在
- 500: 系统错误

## 部署方案

### 1. 本地开发
```bash
# 安装依赖
go mod tidy

# 运行服务
go run ./cmd/server
```

### 2. Docker部署
```bash
# 构建镜像
docker build -t ai-svc .

# 运行容器
docker run -p 8080:8080 ai-svc
```

### 3. Docker Compose
```bash
# 一键启动全套环境
docker-compose up -d
```

## 监控和日志

### 日志格式
项目使用结构化日志，便于日志分析和监控：

```json
{
    "level": "info",
    "time": "2023-12-07T10:00:00Z",
    "message": "HTTP请求",
    "method": "POST",
    "path": "/api/v1/login",
    "status": 200,
    "latency": "10ms",
    "ip": "127.0.0.1"
}
```

### 健康检查
- 接口: `GET /health`
- 返回服务状态和基本信息

## 性能优化

### 1. 数据库优化
- 连接池配置
- 索引优化
- 查询优化

### 2. 缓存策略
- Redis缓存热点数据
- JWT令牌缓存
- 用户会话管理

### 3. 并发处理
- Goroutine处理请求
- 数据库连接池
- 合理的超时设置

## 扩展建议

### 1. 功能扩展
- 角色权限系统
- 文件上传服务
- 消息推送系统
- 数据统计分析

### 2. 技术扩展
- gRPC支持
- 微服务拆分
- 消息队列集成
- 分布式缓存

### 3. 运维扩展
- Prometheus监控
- 链路追踪
- 自动化部署
- 负载均衡

### 4. 架构扩展
- 独立的路由层设计
- 控制器层的职责分离
- 更灵活的依赖注入
- 插件化架构支持

## 最佳实践

### 1. 代码规范
- 使用go fmt格式化代码
- 遵循Go命名规范
- 编写单元测试
- 使用golangci-lint检查代码

### 2. 安全实践
- 定期更新依赖
- 使用HTTPS
- 实施API限流
- 输入验证和输出编码

### 3. 部署实践
- 使用容器化部署
- 配置健康检查
- 实施滚动更新
- 备份重要数据

这个架构为企业级应用提供了坚实的基础，可以根据具体需求进行扩展和定制。
