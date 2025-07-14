# AI 服务 Cobra CLI 使用指南

## 🎯 概述

本项目已成功集成 Cobra 框架，提供了现代化的命令行接口。主要特性：

- ✅ **简洁的 main.go**：只有 11 行代码，遵循开源项目最佳实践
- ✅ **丰富的子命令**：server、version、config 等多个实用命令
- ✅ **版本信息注入**：编译时自动注入 Git 信息和构建时间
- ✅ **详细的帮助系统**：每个命令都有完整的帮助文档
- ✅ **配置管理**：支持配置文件生成、验证和查看
- ✅ **优雅关闭**：支持信号处理和服务器优雅关闭

## 🏗️ 架构对比

### 重构前的 main.go (80+ 行)
```go
func main() {
    // 加载配置
    if err := config.LoadConfig("./configs/config.yaml"); err != nil {
        log.Fatalf("加载配置失败: %v", err)
    }
    
    // 初始化日志
    if err := logger.Init(...); err != nil {
        log.Fatalf("初始化日志失败: %v", err)
    }
    
    // ... 大量的启动逻辑
    // ... 服务器配置
    // ... 信号处理
    // ... 优雅关闭逻辑
}
```

### 重构后的 main.go (11 行)
```go
package main

import "ai-svc/cmd"

// main 函数是程序的入口点
// 它将控制权交给 Cobra 命令行框架来处理用户输入和命令执行
// 所有的复杂逻辑都被封装在 cmd 包中，保持 main.go 的简洁性
func main() {
    // 执行 Cobra 根命令
    // 这会解析命令行参数并路由到相应的子命令
    // 如果执行过程中出现错误，程序会自动退出并显示错误信息
    cmd.Execute()
}
```

## 📁 项目结构

```
ai-svc/
├── cmd/                    # Cobra 命令定义
│   ├── server/
│   │   └── main.go        # 简化的程序入口点 (11 行)
│   ├── root.go            # 根命令和全局配置
│   ├── server.go          # 服务器启动命令
│   ├── version.go         # 版本信息命令
│   └── config.go          # 配置管理命令
├── internal/              # 内部业务逻辑
├── pkg/                   # 可复用的包
├── configs/               # 配置文件
├── Makefile              # 构建脚本（支持版本注入）
└── README.md
```

## 🚀 使用方法

### 1. 基本命令

```bash
# 显示帮助信息
./ai-svc help

# 启动服务器（默认配置）
./ai-svc server

# 启动服务器（指定参数）
./ai-svc server --port 8080 --mode release

# 启动服务器（开发模式）
./ai-svc server --mode debug --verbose

# 显示版本信息
./ai-svc version

# 显示详细版本信息
./ai-svc version detail
```

### 2. 配置管理

```bash
# 生成默认配置文件
./ai-svc config generate

# 生成配置文件到指定位置
./ai-svc config generate --output ./my-config.yaml

# 验证配置文件
./ai-svc config validate

# 显示当前配置
./ai-svc config show

# 使用自定义配置文件
./ai-svc server --config ./my-config.yaml
```

### 3. Makefile 使用

```bash
# 构建应用程序（自动注入版本信息）
make build

# 开发模式运行
make dev

# 运行测试
make test

# 清理构建文件
make clean

# 更新依赖
make deps

# 显示帮助
make help
```

## 🎨 特色功能

### 1. 版本信息自动注入

编译时自动注入的信息：
- Git 提交哈希
- 构建时间
- Go 版本
- 平台信息

```bash
$ ./ai-svc version
AI 服务 (ai-svc)
版本: v1.2.3-5-g1234567-dirty
Git 提交: 1234567
构建时间: 2025-07-14 06:50:46 UTC
Go 版本: go1.21.6
平台: darwin/amd64
```

### 2. 智能配置管理

```bash
# 自动查找配置文件优先级：
# 1. --config 指定的文件
# 2. ./configs/config.yaml
# 3. ./config.yaml
# 4. $HOME/.ai-svc/config.yaml

# 环境变量覆盖（自动支持）
export AI_SVC_SERVER_PORT=9090
export AI_SVC_SERVER_MODE=release
./ai-svc server  # 会使用环境变量的值
```

### 3. 详细的帮助系统

每个命令都有完整的帮助文档：

```bash
$ ./ai-svc server --help
启动 AI 服务的 HTTP 服务器。

服务器提供以下功能：
• RESTful API 接口服务
• 用户认证与授权
• 短信验证码发送
• 设备管理与安全控制
• 智能限流保护
• 健康检查接口

示例用法：
  ai-svc server                          # 使用默认配置启动
  ai-svc server --port 8080              # 指定端口启动
  ai-svc server --mode release           # 生产模式启动
  ai-svc server --config custom.yaml    # 使用自定义配置文件
  ai-svc server --verbose               # 启用详细日志输出

Usage:
  ai-svc server [flags]

Flags:
  -h, --help           help for server
  -m, --mode string    运行模式 (debug|release|test, 默认: debug)
  -p, --port string    服务器监听端口 (默认: 8080)
      --profile        启用 pprof 性能分析接口

Global Flags:
      --config string   配置文件路径 (默认查找 ./configs/config.yaml)
  -v, --verbose         启用详细输出模式
```

### 4. 优雅的启动和关闭

```bash
$ ./ai-svc server
🚀 正在启动 AI 服务...
✅ 配置文件加载成功: ./configs/config.yaml
🌟 服务器启动成功，监听端口: 8080
📊 运行模式: debug
🔗 访问地址: http://localhost:8080
💚 健康检查: http://localhost:8080/health

# 按 Ctrl+C 优雅关闭
^C
🛑 收到关闭信号: interrupt
⏳ 正在等待现有连接完成...
✅ 服务器已优雅关闭
```

## 🛠️ 开发者指南

### 添加新命令

1. 在 `cmd/` 目录下创建新文件，例如 `cmd/migrate.go`：

```go
package cmd

import (
    "github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
    Use:   "migrate",
    Short: "数据库迁移工具",
    Long:  `执行数据库迁移操作...`,
    RunE: func(cmd *cobra.Command, args []string) error {
        // 迁移逻辑
        return nil
    },
}

func init() {
    rootCmd.AddCommand(migrateCmd)
}
```

2. 在 `cmd/root.go` 的 `init()` 函数中会自动加载新命令

### 添加命令行参数

```go
var migrateCmd = &cobra.Command{
    // ...
}

func init() {
    rootCmd.AddCommand(migrateCmd)
    
    // 添加命令特定的标志
    migrateCmd.Flags().StringP("direction", "d", "up", "迁移方向 (up|down)")
    migrateCmd.Flags().IntP("steps", "s", 0, "迁移步数")
    
    // 绑定到 Viper（支持配置文件和环境变量）
    viper.BindPFlag("migrate.direction", migrateCmd.Flags().Lookup("direction"))
}
```

## 📦 部署建议

### 1. 生产环境构建

```bash
# 设置版本标签
export VERSION=v1.0.0

# 构建生产版本
make build

# 或者使用 Git 标签
git tag v1.0.0
make build  # 自动使用 Git 标签作为版本
```

### 2. Docker 化

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN make build

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/build/ai-svc .
COPY --from=builder /app/configs ./configs
ENTRYPOINT ["./ai-svc"]
CMD ["server"]
```

### 3. 系统服务

创建 systemd 服务文件：

```ini
[Unit]
Description=AI Service
After=network.target

[Service]
Type=simple
User=ai-svc
WorkingDirectory=/opt/ai-svc
ExecStart=/opt/ai-svc/ai-svc server --config /etc/ai-svc/config.yaml
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

## 🔍 最佳实践

### 1. 配置管理
- 使用环境变量覆盖敏感配置
- 为不同环境准备不同的配置文件
- 使用 `--verbose` 模式进行调试

### 2. 日志记录
- 生产环境使用 `--mode release`
- 开发环境使用 `--verbose` 获得详细日志
- 使用结构化日志格式

### 3. 监控和健康检查
- 使用 `/health` 端点进行健康检查
- 监控版本信息和构建信息
- 使用 `version detail` 查看运行时信息

这个重构为项目带来了现代化的命令行体验，遵循了 Go 社区的最佳实践，使得项目更易于维护和扩展。
