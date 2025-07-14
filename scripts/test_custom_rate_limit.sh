#!/bin/bash

# 定制化限流功能测试脚本

echo "=== 定制化限流功能测试 ==="
echo ""

# 服务器地址
SERVER="http://localhost:8080"

echo "1. 测试SMS接口限流（每分钟1次）"
echo "发送第一次请求..."
curl -X POST "$SERVER/api/v1/sms/send" \
  -H "Content-Type: application/json" \
  -d '{"phone": "13800138000"}' \
  -w "\nHTTP状态码: %{http_code}\n"

echo ""
echo "立即发送第二次请求（应该被限流）..."
curl -X POST "$SERVER/api/v1/sms/send" \
  -H "Content-Type: application/json" \
  -d '{"phone": "13800138000"}' \
  -w "\nHTTP状态码: %{http_code}\n"

echo ""
echo "=== 测试完成 ==="
echo ""

echo "2. 测试登录接口限流（每分钟5次）"
echo "连续发送6次登录请求..."
for i in {1..6}; do
  echo "第 $i 次登录请求:"
  curl -X POST "$SERVER/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"phone": "13800138000", "code": "123456"}' \
    -w "\nHTTP状态码: %{http_code}\n"
  echo ""
done

echo ""
echo "3. 测试自定义限流接口（每分钟1次）"
echo "发送第一次请求..."
curl -X POST "$SERVER/api/v1/dangerous-operation" \
  -H "Content-Type: application/json" \
  -w "\nHTTP状态码: %{http_code}\n"

echo ""
echo "立即发送第二次请求（应该被限流）..."
curl -X POST "$SERVER/api/v1/dangerous-operation" \
  -H "Content-Type: application/json" \
  -w "\nHTTP状态码: %{http_code}\n"

echo ""
echo "=== 所有测试完成 ==="
