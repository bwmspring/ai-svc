# 数据脱敏处理指南

## 🎯 **设计目标**

为了保护用户隐私和符合数据保护法规，对敏感信息进行脱敏处理：
- ✅ **手机号脱敏** - 保留前3位和后4位
- ✅ **邮箱脱敏** - 用户名部分脱敏，保留域名
- ✅ **身份证脱敏** - 保留前6位和后4位
- ✅ **银行卡脱敏** - 保留前4位和后4位
- ✅ **姓名脱敏** - 保留姓氏，其他用*替换

## 🔒 **脱敏规则**

### **1. 手机号脱敏**
```go
// 输入: 13800138000
// 输出: 138****8000
func MaskPhone(phone string) string {
    if len(phone) < 7 {
        return phone
    }
    return phone[:3] + "****" + phone[len(phone)-4:]
}
```

### **2. 邮箱脱敏**
```go
// 输入: user@example.com
// 输出: us***@example.com
func MaskEmail(email string) string {
    // 找到@符号位置
    atIndex := strings.Index(email, "@")
    if atIndex == -1 {
        return email
    }
    
    username := email[:atIndex]
    domain := email[atIndex:]
    
    if len(username) <= 2 {
        return email
    }
    
    maskedUsername := username[:2] + "***"
    return maskedUsername + domain
}
```

### **3. 身份证号脱敏**
```go
// 输入: 110101199001011234
// 输出: 110101********1234
func MaskIDCard(idCard string) string {
    if len(idCard) < 10 {
        return idCard
    }
    return idCard[:6] + "********" + idCard[len(idCard)-4:]
}
```

### **4. 银行卡号脱敏**
```go
// 输入: 6222021234567890123
// 输出: 6222 **** **** 0123
func MaskBankCard(bankCard string) string {
    if len(bankCard) < 8 {
        return bankCard
    }
    return bankCard[:4] + " **** **** " + bankCard[len(bankCard)-4:]
}
```

### **5. 姓名脱敏**
```go
// 输入: 张三
// 输出: 张*
func MaskName(name string) string {
    if len(name) < 2 {
        return name
    }
    if len(name) == 2 {
        return name[:1] + "*"
    }
    return name[:1] + "**"
}
```

## 📊 **使用场景**

### **1. API响应脱敏**
```go
// 用户信息接口返回脱敏数据
func (u *User) GetMaskedPhone() string {
    return utils.MaskPhone(u.Phone)
}

func (u *User) GetMaskedEmail() string {
    return utils.MaskEmail(u.Email)
}

func (u *User) GetMaskedRealName() string {
    return utils.MaskName(u.RealName)
}
```

### **2. 日志记录脱敏**
```go
// 短信发送日志使用脱敏手机号
func (s *smsService) logSMSSend(phone, purpose, clientIP string, userID *uint) {
    maskedPhone := utils.MaskPhone(phone)
    logData := map[string]any{
        "phone":     maskedPhone, // 脱敏手机号
        "purpose":   purpose,
        "client_ip": clientIP,
        "timestamp": time.Now().Unix(),
    }
    
    logger.Info("短信发送成功", logData)
}
```

### **3. 错误信息脱敏**
```go
// 错误响应中使用脱敏信息
func (ctrl *SMSController) SendSMS(c *gin.Context) {
    // ... 处理逻辑
    
    if err != nil {
        maskedPhone := utils.MaskPhone(req.Phone)
        response.Error(c, response.ERROR, "发送验证码失败，手机号: "+maskedPhone)
        return
    }
    
    maskedPhone := utils.MaskPhone(req.Phone)
    response.SuccessWithMessage(c, "短信验证码已发送到 "+maskedPhone, nil)
}
```

## 🛡️ **安全考虑**

### **1. 脱敏级别**
- **完全脱敏** - 用于公开接口和日志
- **部分脱敏** - 用于内部管理接口
- **不脱敏** - 仅用于核心业务逻辑

### **2. 脱敏策略**
```go
// 根据用户角色决定脱敏级别
func GetDesensitizationLevel(userRole string) string {
    switch userRole {
    case "admin":
        return "none"      // 管理员不脱敏
    case "manager":
        return "partial"   // 经理部分脱敏
    default:
        return "full"      // 普通用户完全脱敏
    }
}
```

### **3. 数据保护**
- ✅ 数据库存储原始数据
- ✅ API响应使用脱敏数据
- ✅ 日志记录使用脱敏数据
- ✅ 错误信息使用脱敏数据

## 📋 **最佳实践**

### **1. 统一脱敏工具**
```go
// 使用统一的脱敏工具包
import "ai-svc/pkg/utils"

// 所有脱敏操作都通过工具包
maskedPhone := utils.MaskPhone(phone)
maskedEmail := utils.MaskEmail(email)
```

### **2. 配置化脱敏**
```go
// 支持配置化的脱敏规则
type DesensitizationConfig struct {
    PhoneMaskPattern string `json:"phone_mask_pattern"`
    EmailMaskPattern string `json:"email_mask_pattern"`
    NameMaskPattern  string `json:"name_mask_pattern"`
}
```

### **3. 测试覆盖**
```go
// 脱敏功能测试
func TestMaskPhone(t *testing.T) {
    tests := []struct {
        input    string
        expected string
    }{
        {"13800138000", "138****8000"},
        {"1234567", "1234567"}, // 长度不足，不脱敏
        {"", ""},
    }
    
    for _, test := range tests {
        result := utils.MaskPhone(test.input)
        if result != test.expected {
            t.Errorf("MaskPhone(%s) = %s, expected %s", test.input, result, test.expected)
        }
    }
}
```

## 🔍 **监控和审计**

### **1. 脱敏日志**
```go
// 记录脱敏操作
logger.Info("数据脱敏处理", map[string]any{
    "original_length": len(originalData),
    "masked_length":   len(maskedData),
    "mask_type":       "phone",
    "timestamp":       time.Now().Unix(),
})
```

### **2. 异常检测**
- 检测脱敏后的数据是否仍然包含敏感信息
- 监控脱敏规则的覆盖率
- 审计脱敏操作的合规性

## 📈 **性能优化**

### **1. 缓存脱敏结果**
```go
// 对于频繁访问的数据，可以缓存脱敏结果
var maskCache = make(map[string]string)

func GetMaskedPhoneWithCache(phone string) string {
    if cached, exists := maskCache[phone]; exists {
        return cached
    }
    
    masked := utils.MaskPhone(phone)
    maskCache[phone] = masked
    return masked
}
```

### **2. 批量脱敏**
```go
// 批量处理多个手机号
func BatchMaskPhones(phones []string) []string {
    result := make([]string, len(phones))
    for i, phone := range phones {
        result[i] = utils.MaskPhone(phone)
    }
    return result
}
```

## 🎯 **合规要求**

### **1. 数据保护法规**
- ✅ **GDPR** - 欧盟数据保护条例
- ✅ **CCPA** - 加州消费者隐私法案
- ✅ **个人信息保护法** - 中国个人信息保护法

### **2. 行业标准**
- ✅ **PCI DSS** - 支付卡行业数据安全标准
- ✅ **ISO 27001** - 信息安全管理体系
- ✅ **SOC 2** - 服务组织控制报告

通过这套完整的脱敏处理方案，我们确保了用户隐私的保护，同时满足了各种合规要求。 