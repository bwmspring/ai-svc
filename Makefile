# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=ai-svc
BINARY_UNIX=$(BINARY_NAME)_unix

# Main package path
MAIN_PATH=./cmd/server

.PHONY: all build clean test deps tidy run dev help

all: test build

## build: 编译应用程序
build:
	$(GOBUILD) -o $(BINARY_NAME) -v $(MAIN_PATH)

## clean: 清理编译产物
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)

## test: 运行测试
test:
	$(GOTEST) -v ./...

## test-coverage: 运行测试并生成覆盖率报告
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

## deps: 下载依赖
deps:
	$(GOGET) -d -v ./...

## tidy: 整理依赖
tidy:
	$(GOMOD) tidy

## run: 运行应用程序
run:
	$(GOBUILD) -o $(BINARY_NAME) -v $(MAIN_PATH) && ./$(BINARY_NAME)

## dev: 开发模式运行
dev:
	$(GOCMD) run $(MAIN_PATH)

## build-linux: 为Linux编译
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v $(MAIN_PATH)

## docker-build: 构建Docker镜像
docker-build:
	docker build -t $(BINARY_NAME):latest .

## docker-run: 运行Docker容器
docker-run:
	docker run -p 8080:8080 $(BINARY_NAME):latest

## help: 显示帮助信息
help: Makefile
	@echo
	@echo "选择一个命令运行:"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo
