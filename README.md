# AI-SVC 企业级Go后端服务

一个基于 Gin + GORM + MySQL + Viper 构建的企业级Go后端服务架构。

## 技术栈

- **Web框架**: Gin
- **ORM**: GORM
- **数据库**: MySQL 8.0
- **配置管理**: Viper
- **日志**: Logrus
- **身份认证**: JWT
- **容器化**: Docker & Docker Compose
- **缓存**: Redis

## 项目结构

```
ai-svc/
├── cmd/                    # 应用程序入口
│   └── server/
│       └── main.go        # 主程序
├── internal/              # 内部应用代码
│   ├── config/           # 配置管理
│   ├── controller/       # HTTP控制器
│   ├── service/          # 业务逻辑层
│   ├── repository/       # 数据访问层
│   ├── model/            # 数据模型
│   └── middleware/       # 中间件
├── pkg/                   # 可重用的库代码
│   ├── database/         # 数据库连接
│   ├── logger/           # 日志工具
│   └── response/         # 统一响应格式
├── routes/               # 路由配置
├── configs/              # 配置文件
├── docs/                 # 文档
├── Dockerfile            # Docker镜像构建文件
├── docker-compose.yml    # Docker Compose配置
├── Makefile             # 构建脚本
└── README.md            # 项目说明
```

## 功能特性

- ✅ 用户注册/登录
- ✅ JWT身份认证
- ✅ 用户信息管理
- ✅ 密码修改
- ✅ 用户列表查询
- ✅ 用户搜索
- ✅ 统一错误处理
- ✅ 日志记录
- ✅ 跨域支持
- ✅ 优雅关闭
- ✅ Docker容器化

## 快速开始

### 环境要求

- Go 1.21+
- MySQL 8.0+
- Redis (可选)

### 1. 克隆项目

```bash
git clone <repository-url>
cd ai-svc
```

### 2. 安装依赖

```bash
go mod tidy
```

### 3. 配置数据库

修改 `configs/config.yaml` 中的数据库配置：

```yaml
database:
  host: "localhost"
  port: 3306
  username: "root"
  password: "your_password"
  database: "ai_svc"
```

### 4. 运行项目

```bash
# 开发模式运行
make dev

# 或者编译后运行
make build
./ai-svc
```

### 5. 使用Docker运行

```bash
# 使用Docker Compose一键启动（包含MySQL和Redis）
docker-compose up -d

# 仅启动应用
docker build -t ai-svc .
docker run -p 8080:8080 ai-svc
```

## API文档

### 公开接口

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | `/api/v1/register` | 用户注册 |
| POST | `/api/v1/login` | 用户登录 |
| GET | `/health` | 健康检查 |

### 认证接口 (需要JWT Token)

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/api/v1/users/profile` | 获取当前用户信息 |
| PUT | `/api/v1/users/profile` | 更新用户信息 |
| POST | `/api/v1/users/change-password` | 修改密码 |
| GET | `/api/v1/users/list` | 获取用户列表 |
| GET | `/api/v1/users/search` | 搜索用户 |
| GET | `/api/v1/users/:id` | 获取指定用户信息 |
| DELETE | `/api/v1/users/:id` | 删除用户 |

### 请求示例

#### 用户注册
```bash
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "123456",
    "nickname": "测试用户"
  }'
```

#### 用户登录
```bash
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "123456"
  }'
```

#### 获取用户信息
```bash
curl -X GET http://localhost:8080/api/v1/users/profile \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## 配置说明

### 服务器配置
```yaml
server:
  port: "8080"          # 服务端口
  mode: "debug"         # 运行模式: debug, release, test
  read_timeout: 60      # 读取超时时间(秒)
  write_timeout: 60     # 写入超时时间(秒)
```

### 数据库配置
```yaml
database:
  host: "localhost"     # 数据库主机
  port: 3306           # 数据库端口
  username: "root"     # 用户名
  password: "password" # 密码
  database: "ai_svc"   # 数据库名
  charset: "utf8mb4"   # 字符集
  max_idle_conns: 10   # 最大空闲连接数
  max_open_conns: 100  # 最大打开连接数
```

### 日志配置
```yaml
log:
  level: "info"        # 日志级别: debug, info, warn, error
  format: "json"       # 日志格式: json, text
  output: "stdout"     # 输出方式: stdout, file
```

### JWT配置
```yaml
jwt:
  secret: "your-secret-key"  # JWT密钥
  expire_time: 3600         # 过期时间(秒)
```

## 开发指南

### 添加新的API

1. 在 `internal/model/` 中定义数据模型
2. 在 `internal/repository/` 中实现数据访问层
3. 在 `internal/service/` 中实现业务逻辑层
4. 在 `internal/controller/` 中实现HTTP控制器
5. 在 `routes/routes.go` 中添加路由

### 数据库迁移

项目启动时会自动进行数据库迁移，如需手动迁移：

```go
// 在 main.go 中的 migrateDatabase 函数中添加新的模型
if err := db.AutoMigrate(
    &model.User{},
    &model.UserProfile{},
    // 添加新的模型
); err != nil {
    return fmt.Errorf("数据库迁移失败: %w", err)
}
```

### 中间件

项目包含以下中间件：

- **JWT认证**: 验证用户身份
- **CORS**: 跨域资源共享
- **日志**: 记录HTTP请求
- **恢复**: 从panic中恢复
- **限流**: 请求频率限制（待实现）

## 构建和部署

### 本地构建

```bash
# 编译
make build

# 运行测试
make test

# 生成测试覆盖率报告
make test-coverage

# 清理
make clean
```

### Docker部署

```bash
# 构建镜像
make docker-build

# 运行容器
make docker-run

# 使用Docker Compose
docker-compose up -d
```

## 监控和日志

### 健康检查

访问 `http://localhost:8080/health` 检查服务状态。

### 日志查看

```bash
# 查看应用日志
docker-compose logs -f app

# 查看数据库日志
docker-compose logs -f mysql
```

## 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 联系方式

如有问题或建议，请提交 Issue 或联系维护者。
