#!/bin/bash

# 登录日志功能测试脚本
# 用于测试登录日志记录功能

echo "=== 登录日志功能测试 ==="

# 设置API基础URL
API_BASE="http://localhost:8080/api/v1"

# 测试数据
PHONE="13800138000"
CODE="123456"
TOKEN="test_token_$(date +%s)"

echo "1. 测试发送短信验证码..."
curl -X POST "$API_BASE/sms/send" \
  -H "Content-Type: application/json" \
  -d "{
    \"phone\": \"$PHONE\",
    \"type\": \"login\"
  }" | jq '.'

echo -e "\n2. 测试登录（应该记录登录尝试和成功）..."
LOGIN_RESPONSE=$(curl -s -X POST "$API_BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"phone\": \"$PHONE\",
    \"code\": \"$CODE\",
    \"token\": \"$TOKEN\",
    \"device_info\": {
      \"device_id\": \"test_device_$(date +%s)\",
      \"device_type\": \"android\",
      \"device_name\": \"测试设备\",
      \"os_version\": \"Android 12\",
      \"app_version\": \"1.0.0\"
    }
  }")

echo "$LOGIN_RESPONSE" | jq '.'

# 提取用户ID和token
USER_ID=$(echo "$LOGIN_RESPONSE" | jq -r '.data.user.id // empty')
ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.access_token // empty')

if [ "$USER_ID" != "null" ] && [ "$USER_ID" != "" ]; then
    echo -e "\n3. 测试获取用户登录历史..."
    curl -X GET "$API_BASE/users/login-history?page=1&size=10" \
      -H "Authorization: Bearer $ACCESS_TOKEN" \
      -H "Content-Type: application/json" | jq '.'
    
    echo -e "\n4. 测试获取登录统计..."
    curl -X GET "$API_BASE/users/login-stats?days=7" \
      -H "Authorization: Bearer $ACCESS_TOKEN" \
      -H "Content-Type: application/json" | jq '.'
    
    echo -e "\n5. 测试获取今日登录记录..."
    curl -X GET "$API_BASE/users/today-logins" \
      -H "Authorization: Bearer $ACCESS_TOKEN" \
      -H "Content-Type: application/json" | jq '.'
    
    echo -e "\n6. 测试获取最近登录记录..."
    curl -X GET "$API_BASE/users/recent-logins?hours=24" \
      -H "Authorization: Bearer $ACCESS_TOKEN" \
      -H "Content-Type: application/json" | jq '.'
else
    echo "登录失败，无法获取用户ID"
fi

echo -e "\n7. 测试登录失败记录（使用错误验证码）..."
curl -X POST "$API_BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"phone\": \"$PHONE\",
    \"code\": \"wrong_code\",
    \"token\": \"$TOKEN\",
    \"device_info\": {
      \"device_id\": \"test_device_fail_$(date +%s)\",
      \"device_type\": \"ios\",
      \"device_name\": \"测试失败设备\",
      \"os_version\": \"iOS 15\",
      \"app_version\": \"1.0.0\"
    }
  }" | jq '.'

echo -e "\n=== 测试完成 ==="
echo "请检查数据库中的 user_behavior_logs 表，应该包含以下记录："
echo "1. 登录尝试记录 (action: login)"
echo "2. 登录成功记录 (action: login_success)"
echo "3. 登录失败记录 (action: login_failed)"
echo "4. 包含地理位置信息 (location 字段)"
echo "5. 包含登录时间 (login_time 字段)" 