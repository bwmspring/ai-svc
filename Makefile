# AI 服务 Makefile
# 提供常用的开发和部署命令

# 应用程序名称
APP_NAME := ai-svc

# 版本信息（可以通过环境变量覆盖）
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u '+%Y-%m-%d %H:%M:%S UTC')
GO_VERSION := $(shell go version | awk '{print $$3}')

# 构建目录
BUILD_DIR := build
BINARY := $(BUILD_DIR)/$(APP_NAME)

# Go 相关变量
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOMOD := $(GOCMD) mod

# ldflags 用于在编译时注入版本信息
LDFLAGS := -ldflags "-X 'ai-svc/cmd.Version=$(VERSION)' \
                     -X 'ai-svc/cmd.GitCommit=$(GIT_COMMIT)' \
                     -X 'ai-svc/cmd.BuildTime=$(BUILD_TIME)'"

# 默认目标
.PHONY: all
all: clean deps test build

# 构建应用程序
.PHONY: build
build:
	@echo "🔨 构建应用程序..."
	@echo "   版本: $(VERSION)"
	@echo "   提交: $(GIT_COMMIT)"
	@echo "   时间: $(BUILD_TIME)"
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BINARY) ./cmd/server

# 运行应用程序
.PHONY: run
run: build
	@echo "🚀 启动应用程序..."
	./$(BINARY) server

# 运行开发模式
.PHONY: dev
dev:
	@echo "🔥 开发模式启动..."
	$(GOCMD) run ./cmd/server server --mode debug --verbose

# 运行测试
.PHONY: test
test:
	@echo "🧪 运行测试..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...

# 依赖管理
.PHONY: deps
deps:
	@echo "📦 更新依赖..."
	$(GOMOD) tidy
	$(GOMOD) download

# 清理构建文件
.PHONY: clean
clean:
	@echo "🧹 清理构建文件..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# 显示帮助信息
.PHONY: help
help:
	@echo "AI 服务 Makefile 命令："
	@echo ""
	@echo "构建相关："
	@echo "  build          构建应用程序"
	@echo "  clean          清理构建文件"
	@echo ""
	@echo "运行相关："
	@echo "  run            构建并运行应用程序"
	@echo "  dev            开发模式运行"
	@echo ""
	@echo "测试相关："
	@echo "  test           运行测试"
	@echo ""
	@echo "依赖管理："
	@echo "  deps           更新依赖"
	@echo ""
	@echo "其他："
	@echo "  help           显示此帮助信息"
