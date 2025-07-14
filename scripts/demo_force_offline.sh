#!/bin/bash

# 强制下线机制演示脚本

BASE_URL="http://localhost:8080/api/v1"
PHONE="13800138000"

echo "========================================"
echo "强制下线机制演示"
echo "========================================"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 辅助函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# 发送短信验证码
send_sms() {
    log_step "发送短信验证码..."
    curl -s -X POST "$BASE_URL/sms/send" \
        -H "Content-Type: application/json" \
        -d "{\"phone\": \"$PHONE\", \"purpose\": \"login\"}" > /dev/null
    log_info "验证码已发送"
}

# 登录设备
login_device() {
    local device_id=$1
    local device_type=$2
    local device_name=$3
    
    log_step "登录设备: $device_name ($device_type)"
    
    RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
        -H "Content-Type: application/json" \
        -d "{
            \"phone\": \"$PHONE\",
            \"code\": \"123456\",
            \"device_info\": {
                \"device_id\": \"$device_id\",
                \"device_type\": \"$device_type\",
                \"device_name\": \"$device_name\",
                \"app_version\": \"1.0.0\",
                \"os_version\": \"Test OS\"
            }
        }")
    
    TOKEN=$(echo "$RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
    
    if [ -n "$TOKEN" ]; then
        log_info "✓ 设备登录成功: $device_name"
        echo "$TOKEN"
    else
        log_error "✗ 设备登录失败: $device_name"
        echo "$RESPONSE"
        echo ""
    fi
}

# 获取设备列表
get_device_list() {
    local token=$1
    
    if [ -z "$token" ]; then
        log_error "Token为空，无法获取设备列表"
        return
    fi
    
    log_step "获取设备列表..."
    
    RESPONSE=$(curl -s -X GET "$BASE_URL/users/devices" \
        -H "Authorization: Bearer $token")
    
    echo "$RESPONSE" | jq -r '
        if .code == 200 then
            .data | 
            "设备统计: 总数=\(.total_count), 在线=\(.online_count), 限制=\(.max_devices)",
            "设备列表:",
            (.devices[] | "  - \(.device_name) (\(.device_type)) - \(if .is_online then "在线" else "离线" end) - 登录时间: \(.login_at)")
        else
            "错误: \(.message)"
        end
    '
}

# 验证设备Token是否有效
verify_token() {
    local token=$1
    local device_name=$2
    
    if [ -z "$token" ]; then
        return
    fi
    
    RESPONSE=$(curl -s -X GET "$BASE_URL/users/profile" \
        -H "Authorization: Bearer $token")
    
    if echo "$RESPONSE" | grep -q '"code":200'; then
        log_info "✓ $device_name Token有效"
    else
        log_warn "✗ $device_name Token无效或已过期"
    fi
}

# 主要演示流程
main() {
    # 1. 发送验证码
    send_sms
    echo ""
    
    # 2. 登录5台设备（达到上限）
    log_step "=== 阶段1: 登录5台设备（达到上限）==="
    
    TOKEN1=$(login_device "ios_device_001" "ios" "iPhone 13 Pro")
    sleep 1
    TOKEN2=$(login_device "android_device_001" "android" "Samsung Galaxy S21")
    sleep 1
    TOKEN3=$(login_device "pc_device_001" "pc" "MacBook Pro")
    sleep 1
    TOKEN4=$(login_device "web_device_001" "web" "Chrome Browser")
    sleep 1
    TOKEN5=$(login_device "miniprogram_device_001" "miniprogram" "微信小程序")
    
    echo ""
    log_info "=== 当前设备状态 ==="
    get_device_list "$TOKEN5"
    echo ""
    
    # 3. 验证所有设备Token都有效
    log_step "=== 阶段2: 验证所有设备Token状态 ==="
    verify_token "$TOKEN1" "iPhone 13 Pro"
    verify_token "$TOKEN2" "Samsung Galaxy S21"
    verify_token "$TOKEN3" "MacBook Pro"
    verify_token "$TOKEN4" "Chrome Browser"
    verify_token "$TOKEN5" "微信小程序"
    echo ""
    
    # 4. 登录第6台设备（触发强制下线）
    log_step "=== 阶段3: 登录第6台设备（触发强制下线）==="
    log_warn "即将登录第6台设备，这将触发强制下线机制"
    
    TOKEN6=$(login_device "ios_device_002" "ios" "iPhone 14")
    
    echo ""
    log_info "=== 强制下线后的设备状态 ==="
    get_device_list "$TOKEN6"
    echo ""
    
    # 5. 验证被踢出设备的Token状态
    log_step "=== 阶段4: 验证被踢出设备的Token状态 ==="
    log_info "检查各设备Token是否仍然有效..."
    
    verify_token "$TOKEN1" "iPhone 13 Pro (最旧设备)"
    verify_token "$TOKEN2" "Samsung Galaxy S21"
    verify_token "$TOKEN3" "MacBook Pro"
    verify_token "$TOKEN4" "Chrome Browser"
    verify_token "$TOKEN5" "微信小程序"
    verify_token "$TOKEN6" "iPhone 14 (新设备)"
    echo ""
    
    # 6. 继续登录第7台设备
    log_step "=== 阶段5: 继续登录第7台设备 ==="
    log_warn "登录第7台设备，再次触发强制下线机制"
    
    TOKEN7=$(login_device "web_device_002" "web" "Firefox Browser")
    
    echo ""
    log_info "=== 第二次强制下线后的设备状态 ==="
    get_device_list "$TOKEN7"
    echo ""
    
    # 7. 最终验证
    log_step "=== 阶段6: 最终Token状态验证 ==="
    verify_token "$TOKEN1" "iPhone 13 Pro"
    verify_token "$TOKEN2" "Samsung Galaxy S21 (可能被踢出)"
    verify_token "$TOKEN3" "MacBook Pro"
    verify_token "$TOKEN4" "Chrome Browser"
    verify_token "$TOKEN5" "微信小程序"
    verify_token "$TOKEN6" "iPhone 14"
    verify_token "$TOKEN7" "Firefox Browser (最新设备)"
    echo ""
    
    # 8. 手动踢出设备测试
    log_step "=== 阶段7: 手动踢出设备测试 ==="
    log_info "手动踢出 MacBook Pro 设备"
    
    KICK_RESPONSE=$(curl -s -X POST "$BASE_URL/users/devices/kick" \
        -H "Authorization: Bearer $TOKEN7" \
        -H "Content-Type: application/json" \
        -d '{"device_ids": ["pc_device_001"]}')
    
    if echo "$KICK_RESPONSE" | grep -q '"code":200'; then
        log_info "✓ 手动踢出设备成功"
    else
        log_error "✗ 手动踢出设备失败"
        echo "$KICK_RESPONSE"
    fi
    
    echo ""
    log_info "=== 手动踢出后的设备状态 ==="
    get_device_list "$TOKEN7"
    echo ""
    
    # 9. 验证被手动踢出的设备
    log_step "=== 阶段8: 验证被手动踢出的设备 ==="
    verify_token "$TOKEN3" "MacBook Pro (被手动踢出)"
    echo ""
    
    # 10. 总结
    log_step "=== 演示总结 ==="
    log_info "强制下线机制演示完成！"
    log_info "观察要点："
    log_info "1. 设备数量始终保持在限制范围内（5台）"
    log_info "2. 超限时自动踢出最旧设备（按最后活跃时间）"
    log_info "3. 被踢出设备的Token立即失效"
    log_info "4. 新设备可以正常登录和使用"
    log_info "5. 支持手动踢出指定设备"
}

# 检查依赖
check_dependencies() {
    if ! command -v jq &> /dev/null; then
        log_error "jq 命令未找到，请安装 jq 用于JSON格式化"
        log_info "macOS: brew install jq"
        log_info "Ubuntu: sudo apt-get install jq"
        exit 1
    fi
}

# 执行演示
check_dependencies
main

echo "========================================"
echo "演示完成"
echo "========================================"
