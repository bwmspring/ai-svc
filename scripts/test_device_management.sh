#!/bin/bash

# 多端设备管理系统API测试脚本

BASE_URL="http://localhost:8080/api/v1"
PHONE="13800138000"

echo "========================================"
echo "多端设备管理系统API测试"
echo "========================================"

# 1. 发送短信验证码
echo "1. 发送短信验证码..."
curl -X POST "$BASE_URL/sms/send" \
  -H "Content-Type: application/json" \
  -d "{
    \"phone\": \"$PHONE\",
    \"purpose\": \"login\"
  }"
echo -e "\n"

# 2. 设备1登录 (iOS)
echo "2. 设备1登录 (iOS)..."
DEVICE1_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"phone\": \"$PHONE\",
    \"code\": \"123456\",
    \"device_info\": {
      \"device_id\": \"ios_device_001\",
      \"device_type\": \"ios\",
      \"device_name\": \"iPhone 13 Pro\",
      \"app_version\": \"1.0.0\",
      \"os_version\": \"iOS 15.0\"
    }
  }")
echo "$DEVICE1_RESPONSE"
TOKEN1=$(echo "$DEVICE1_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
echo -e "\n"

# 3. 设备2登录 (Android)  
echo "3. 设备2登录 (Android)..."
DEVICE2_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"phone\": \"$PHONE\",
    \"code\": \"123456\",
    \"device_info\": {
      \"device_id\": \"android_device_001\",
      \"device_type\": \"android\",
      \"device_name\": \"Samsung Galaxy S21\",
      \"app_version\": \"1.0.0\",
      \"os_version\": \"Android 11\"
    }
  }")
echo "$DEVICE2_RESPONSE"
TOKEN2=$(echo "$DEVICE2_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
echo -e "\n"

# 4. 设备3登录 (PC)
echo "4. 设备3登录 (PC)..."
DEVICE3_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"phone\": \"$PHONE\",
    \"code\": \"123456\",
    \"device_info\": {
      \"device_id\": \"pc_device_001\",
      \"device_type\": \"pc\",
      \"device_name\": \"MacBook Pro\",
      \"app_version\": \"1.0.0\",
      \"os_version\": \"macOS 12.0\"
    }
  }")
echo "$DEVICE3_RESPONSE"
TOKEN3=$(echo "$DEVICE3_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
echo -e "\n"

# 5. 设备4登录 (Web)
echo "5. 设备4登录 (Web)..."
DEVICE4_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"phone\": \"$PHONE\",
    \"code\": \"123456\",
    \"device_info\": {
      \"device_id\": \"web_device_001\",
      \"device_type\": \"web\",
      \"device_name\": \"Chrome Browser\",
      \"app_version\": \"1.0.0\",
      \"os_version\": \"Chrome 96.0\"
    }
  }")
echo "$DEVICE4_RESPONSE"
TOKEN4=$(echo "$DEVICE4_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
echo -e "\n"

# 6. 设备5登录 (小程序)
echo "6. 设备5登录 (小程序)..."
DEVICE5_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"phone\": \"$PHONE\",
    \"code\": \"123456\",
    \"device_info\": {
      \"device_id\": \"miniprogram_device_001\",
      \"device_type\": \"miniprogram\",
      \"device_name\": \"微信小程序\",
      \"app_version\": \"1.0.0\",
      \"os_version\": \"WeChat 8.0\"
    }
  }")
echo "$DEVICE5_RESPONSE"
TOKEN5=$(echo "$DEVICE5_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
echo -e "\n"

# 7. 获取用户设备列表
echo "7. 获取用户设备列表..."
curl -X GET "$BASE_URL/users/devices" \
  -H "Authorization: Bearer $TOKEN1"
echo -e "\n"

# 8. 设备6登录 (超过限制，应该踢出最旧设备)
echo "8. 设备6登录 (超过限制，应该踢出最旧设备)..."
DEVICE6_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"phone\": \"$PHONE\",
    \"code\": \"123456\",
    \"device_info\": {
      \"device_id\": \"ios_device_002\",
      \"device_type\": \"ios\",
      \"device_name\": \"iPhone 14\",
      \"app_version\": \"1.0.0\",
      \"os_version\": \"iOS 16.0\"
    }
  }")
echo "$DEVICE6_RESPONSE"
TOKEN6=$(echo "$DEVICE6_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
echo -e "\n"

# 9. 再次获取用户设备列表（应该只有5台设备）
echo "9. 再次获取用户设备列表（应该只有5台设备）..."
curl -X GET "$BASE_URL/users/devices" \
  -H "Authorization: Bearer $TOKEN6"
echo -e "\n"

# 10. 手动踢出指定设备
echo "10. 手动踢出指定设备..."
curl -X POST "$BASE_URL/users/devices/kick" \
  -H "Authorization: Bearer $TOKEN6" \
  -H "Content-Type: application/json" \
  -d "{
    \"device_ids\": [\"android_device_001\", \"pc_device_001\"]
  }"
echo -e "\n"

# 11. 最终获取用户设备列表
echo "11. 最终获取用户设备列表..."
curl -X GET "$BASE_URL/users/devices" \
  -H "Authorization: Bearer $TOKEN6"
echo -e "\n"

# 12. 获取用户信息
echo "12. 获取用户信息..."
curl -X GET "$BASE_URL/users/profile" \
  -H "Authorization: Bearer $TOKEN6"
echo -e "\n"

echo "========================================"
echo "测试完成"
echo "========================================"
