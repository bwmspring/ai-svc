#!/bin/bash

# API测试脚本
BASE_URL="http://localhost:8080"

echo "=== AI-SVC API 测试脚本 ==="

# 健康检查
echo "1. 健康检查"
curl -s "$BASE_URL/health" | jq . || echo "请求失败"
echo ""

# 用户注册
echo "2. 用户注册"
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com", 
    "password": "123456",
    "nickname": "测试用户"
  }')
echo $REGISTER_RESPONSE | jq . || echo $REGISTER_RESPONSE
echo ""

# 用户登录
echo "3. 用户登录"
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "123456"
  }')
echo $LOGIN_RESPONSE | jq . || echo $LOGIN_RESPONSE

# 提取token
TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.data.token // empty')
echo ""

if [ ! -z "$TOKEN" ] && [ "$TOKEN" != "null" ]; then
    echo "4. 获取用户信息"
    curl -s -X GET "$BASE_URL/api/v1/users/profile" \
      -H "Authorization: Bearer $TOKEN" | jq . || echo "请求失败"
    echo ""

    echo "5. 更新用户信息"
    curl -s -X PUT "$BASE_URL/api/v1/users/profile" \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json" \
      -d '{
        "nickname": "更新后的昵称"
      }' | jq . || echo "请求失败"
    echo ""

    echo "6. 获取用户列表"
    curl -s -X GET "$BASE_URL/api/v1/users/list?page=1&size=10" \
      -H "Authorization: Bearer $TOKEN" | jq . || echo "请求失败"
    echo ""
else
    echo "登录失败，无法进行后续测试"
fi

echo "=== 测试完成 ==="
