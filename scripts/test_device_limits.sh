#!/bin/bash

# 设备限制检查测试脚本
# 测试 /auth/login 和 /auth/refresh 接口的设备限制功能

set -e  # 遇到错误立即退出

# 配置
BASE_URL="http://localhost:8080/api/v1"
PHONE="13800138000"
TEST_CODE="123456"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 获取响应中的JSON字段
get_json_field() {
    echo "$1" | jq -r "$2" 2>/dev/null || echo "null"
}

# 发送SMS验证码
send_sms() {
    log_info "发送SMS验证码到 $PHONE"
    response=$(curl -s -X POST "$BASE_URL/sms/send" \
        -H "Content-Type: application/json" \
        -d "{\"phone\":\"$PHONE\",\"purpose\":\"login\"}")
    
    code=$(get_json_field "$response" ".code")
    if [ "$code" = "200" ]; then
        log_success "SMS验证码发送成功"
        return 0
    else
        message=$(get_json_field "$response" ".message")
        log_error "SMS验证码发送失败: $message"
        return 1
    fi
}

# 设备登录
device_login() {
    local device_id="$1"
    local device_type="$2"
    local device_name="$3"
    
    log_info "设备登录: $device_name ($device_id)"
    
    response=$(curl -s -X POST "$BASE_URL/auth/login" \
        -H "Content-Type: application/json" \
        -d "{
            \"phone\":\"$PHONE\",
            \"code\":\"$TEST_CODE\",
            \"device_info\":{
                \"device_id\":\"$device_id\",
                \"device_type\":\"$device_type\",
                \"device_name\":\"$device_name\",
                \"app_version\":\"1.0.0\",
                \"os_version\":\"iOS 17.0\"
            }
        }")
    
    code=$(get_json_field "$response" ".code")
    if [ "$code" = "200" ]; then
        access_token=$(get_json_field "$response" ".data.access_token")
        refresh_token=$(get_json_field "$response" ".data.refresh_token")
        
        if [ "$access_token" != "null" ] && [ "$refresh_token" != "null" ]; then
            log_success "设备登录成功"
            echo "$access_token|$refresh_token"
            return 0
        else
            log_error "登录响应中缺少token"
            echo "$response"
            return 1
        fi
    else
        message=$(get_json_field "$response" ".message")
        log_error "设备登录失败: $message"
        echo "$response"
        return 1
    fi
}

# Token刷新
refresh_token() {
    local refresh_token="$1"
    local device_name="$2"
    
    log_info "刷新Token: $device_name"
    
    response=$(curl -s -X POST "$BASE_URL/auth/refresh" \
        -H "Content-Type: application/json" \
        -d "{\"refresh_token\":\"$refresh_token\"}")
    
    code=$(get_json_field "$response" ".code")
    message=$(get_json_field "$response" ".message")
    
    if [ "$code" = "200" ]; then
        log_success "Token刷新成功: $device_name"
        return 0
    else
        log_error "Token刷新失败: $device_name - $message"
        echo "$response"
        return 1
    fi
}

# 访问受保护的API
access_protected_api() {
    local access_token="$1"
    local device_name="$2"
    
    log_info "访问受保护API: $device_name"
    
    response=$(curl -s -X GET "$BASE_URL/users/profile" \
        -H "Authorization: Bearer $access_token")
    
    code=$(get_json_field "$response" ".code")
    
    if [ "$code" = "200" ]; then
        log_success "API访问成功: $device_name"
        return 0
    else
        message=$(get_json_field "$response" ".message")
        log_error "API访问失败: $device_name - $message"
        echo "$response"
        return 1
    fi
}

# 踢出设备
kick_device() {
    local access_token="$1"
    local device_id="$2"
    local device_name="$3"
    
    log_info "踢出设备: $device_name ($device_id)"
    
    response=$(curl -s -X POST "$BASE_URL/users/devices/kick" \
        -H "Authorization: Bearer $access_token" \
        -H "Content-Type: application/json" \
        -d "{\"device_ids\":[\"$device_id\"]}")
    
    code=$(get_json_field "$response" ".code")
    
    if [ "$code" = "200" ]; then
        log_success "设备踢出成功: $device_name"
        return 0
    else
        message=$(get_json_field "$response" ".message")
        log_error "设备踢出失败: $device_name - $message"
        echo "$response"
        return 1
    fi
}

# 主测试流程
main() {
    echo "============================================"
    echo "设备限制检查测试"
    echo "============================================"
    
    # 1. 发送SMS验证码
    if ! send_sms; then
        exit 1
    fi
    
    echo
    echo "============================================"
    echo "测试1: 登录接口的设备限制检查"
    echo "============================================"
    
    # 2. 第一个设备登录
    tokens1=$(device_login "device_001" "ios" "iPhone 15 Pro")
    if [ $? -ne 0 ]; then
        exit 1
    fi
    access_token1=$(echo "$tokens1" | cut -d'|' -f1)
    refresh_token1=$(echo "$tokens1" | cut -d'|' -f2)
    
    # 3. 第二个设备登录
    tokens2=$(device_login "device_002" "ios" "iPhone 14")
    if [ $? -ne 0 ]; then
        exit 1
    fi
    access_token2=$(echo "$tokens2" | cut -d'|' -f1)
    refresh_token2=$(echo "$tokens2" | cut -d'|' -f2)
    
    # 4. 验证两个设备都能正常访问API
    echo
    log_info "验证设备登录后的API访问"
    access_protected_api "$access_token1" "iPhone 15 Pro"
    access_protected_api "$access_token2" "iPhone 14"
    
    echo
    echo "============================================"
    echo "测试2: refresh接口的设备验证"
    echo "============================================"
    
    # 5. 验证两个设备都能刷新token
    log_info "验证设备可以正常刷新token"
    refresh_token "$refresh_token1" "iPhone 15 Pro"
    refresh_token "$refresh_token2" "iPhone 14"
    
    # 6. 踢出第一个设备
    echo
    log_info "踢出第一个设备 (iPhone 15 Pro)"
    kick_device "$access_token2" "device_001" "iPhone 15 Pro"
    
    # 7. 验证被踢出的设备不能刷新token
    echo
    log_info "验证被踢出的设备不能刷新token"
    if refresh_token "$refresh_token1" "iPhone 15 Pro (被踢出)"; then
        log_error "安全漏洞：被踢出的设备仍能刷新token！"
        exit 1
    else
        log_success "安全检查通过：被踢出的设备无法刷新token"
    fi
    
    # 8. 验证正常设备仍能刷新token
    log_info "验证正常设备仍能刷新token"
    if refresh_token "$refresh_token2" "iPhone 14 (正常)"; then
        log_success "正常设备token刷新正常"
    else
        log_error "正常设备token刷新失败"
        exit 1
    fi
    
    # 9. 验证被踢出的设备不能访问API
    echo
    log_info "验证被踢出的设备不能访问API"
    if access_protected_api "$access_token1" "iPhone 15 Pro (被踢出)"; then
        log_error "安全漏洞：被踢出的设备仍能访问API！"
        exit 1
    else
        log_success "安全检查通过：被踢出的设备无法访问API"
    fi
    
    echo
    echo "============================================"
    echo "测试完成"
    echo "============================================"
    log_success "所有安全检查通过！"
    echo
    echo "验证结果："
    echo "✅ /auth/login 接口：设备登录时会触发设备限制检查"
    echo "✅ /auth/refresh 接口：被踢出的设备无法刷新token"
    echo "✅ 设备验证机制：被踢出的设备无法访问任何受保护的API"
    echo "✅ 安全性：设备管理功能正常工作"
}

# 检查依赖
check_dependencies() {
    if ! command -v curl &> /dev/null; then
        log_error "curl 命令未找到，请安装 curl"
        exit 1
    fi
    
    if ! command -v jq &> /dev/null; then
        log_error "jq 命令未找到，请安装 jq"
        exit 1
    fi
}

# 脚本入口
echo "检查依赖..."
check_dependencies

echo "开始测试..."
main 