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
GOFMT := $(GOCMD) fmt

# 工具变量
GOIMPORTS := goimports
GOLINES := golines
GOLANGCI_LINT := golangci-lint

# ldflags 用于在编译时注入版本信息
LDFLAGS := -ldflags "-X 'ai-svc/cmd.Version=$(VERSION)' \
                     -X 'ai-svc/cmd.GitCommit=$(GIT_COMMIT)' \
                     -X 'ai-svc/cmd.BuildTime=$(BUILD_TIME)'"

# 默认目标
.PHONY: all
all: clean deps format test build

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

# 生成测试覆盖率报告
.PHONY: coverage
coverage: test
	@echo "📊 生成覆盖率报告..."
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告已生成：coverage.html"

# 代码格式化
.PHONY: fmt
fmt:
	@echo "🎨 格式化代码..."
	$(GOFMT) ./...

# Import 排序和格式化
.PHONY: imports
imports:
	@echo "📦 排序和格式化 imports..."
	@command -v $(GOIMPORTS) >/dev/null 2>&1 || { \
		echo "安装 goimports..."; \
		$(GOCMD) install golang.org/x/tools/cmd/goimports@latest; \
	}
	$(GOIMPORTS) -w -local ai-svc .

# 控制代码行长度
.PHONY: lines
lines:
	@echo "📏 控制代码行长度..."
	@command -v $(GOLINES) >/dev/null 2>&1 || { \
		echo "安装 golines..."; \
		$(GOCMD) install github.com/segmentio/golines@latest; \
	}
	$(GOLINES) -w -m 120 --base-formatter=gofumpt .

# 代码检查和 lint
.PHONY: lint
lint:
	@echo "🔍 运行代码检查..."
	@command -v $(GOLANGCI_LINT) >/dev/null 2>&1 || { \
		echo "安装 golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin latest; \
	}
	$(GOLANGCI_LINT) run ./...

# 修复可自动修复的 lint 问题
.PHONY: lint-fix
lint-fix:
	@echo "🔧 修复 lint 问题..."
	@command -v $(GOLANGCI_LINT) >/dev/null 2>&1 || { \
		echo "安装 golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin latest; \
	}
	$(GOLANGCI_LINT) run --fix ./...

# 全面格式化（包含所有格式化操作）
.PHONY: format
format: fmt imports lines lint-fix
	@echo "🎯 代码格式化完成！"

# 代码质量检查（格式化 + 测试 + lint）
.PHONY: check
check: format test lint
	@echo "✅ 代码质量检查完成！"

# 安装开发工具
.PHONY: install-tools
install-tools:
	@echo "🛠️  安装开发工具..."
	$(GOCMD) install golang.org/x/tools/cmd/goimports@latest
	$(GOCMD) install github.com/segmentio/golines@latest
	@command -v $(GOLANGCI_LINT) >/dev/null 2>&1 || { \
		echo "安装 golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin latest; \
	}
	@echo "所有开发工具安装完成！"

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
	@echo "  coverage       生成测试覆盖率报告"
	@echo ""
	@echo "代码质量："
	@echo "  fmt            格式化代码"
	@echo "  imports        排序和格式化 imports"
	@echo "  lines          控制代码行长度（最大120字符）"
	@echo "  lint           运行代码检查"
	@echo "  lint-fix       修复可自动修复的 lint 问题"
	@echo "  format         全面格式化（fmt + imports + lines + lint-fix）"
	@echo "  check          代码质量检查（format + test + lint）"
	@echo ""
	@echo "工具管理："
	@echo "  install-tools  安装开发工具"
	@echo "  deps           更新依赖"
	@echo ""
	@echo "其他："
	@echo "  help           显示此帮助信息"
