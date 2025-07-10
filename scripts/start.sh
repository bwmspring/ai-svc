#!/bin/bash

# 启动脚本
set -e

echo "=== AI-SVC 服务启动脚本 ==="

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "错误: 未找到Go环境，请先安装Go"
    exit 1
fi

# 检查配置文件
if [ ! -f "./configs/config.yaml" ]; then
    echo "错误: 配置文件不存在，请检查 ./configs/config.yaml"
    exit 1
fi

# 整理依赖
echo "正在整理依赖..."
go mod tidy

# 编译项目
echo "正在编译项目..."
go build -o ai-svc ./cmd/server

# 运行项目
echo "正在启动服务..."
echo "服务将在 http://localhost:8080 启动"
echo "健康检查: http://localhost:8080/health"
echo "按 Ctrl+C 停止服务"
echo "========================="

./ai-svc
