package utils

import (
	"strings"
	"unicode/utf8"
)

// DesensitizeUtils 脱敏工具函数
type DesensitizeUtils struct{}

// MaskPhone 手机号脱敏处理
func MaskPhone(phone string) string {
	if len(phone) < 7 {
		return phone
	}
	// 保留前3位和后4位，中间用*替换
	return phone[:3] + "****" + phone[len(phone)-4:]
}

// MaskEmail 邮箱脱敏处理
func MaskEmail(email string) string {
	if len(email) < 5 {
		return email
	}

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

// MaskIDCard 身份证号脱敏处理
func MaskIDCard(idCard string) string {
	if len(idCard) < 10 {
		return idCard
	}
	// 保留前6位和后4位，中间用*替换
	return idCard[:6] + "********" + idCard[len(idCard)-4:]
}

// MaskBankCard 银行卡号脱敏处理
func MaskBankCard(bankCard string) string {
	if len(bankCard) < 8 {
		return bankCard
	}
	// 保留前4位和后4位，中间用*替换
	return bankCard[:4] + " **** **** " + bankCard[len(bankCard)-4:]
}

// MaskName 姓名脱敏处理（支持中文）
func MaskName(name string) string {
	if len(name) == 0 {
		return name
	}

	// 计算字符数量（不是字节数）
	charCount := utf8.RuneCountInString(name)
	if charCount < 2 {
		return name
	}

	// 转换为rune切片以便处理UTF-8字符
	runes := []rune(name)
	if charCount == 2 {
		return string(runes[:1]) + "*"
	}

	// 保留姓氏，其他用*替换
	return string(runes[:1]) + "**"
}

// MaskAddress 地址脱敏处理（支持中文）
func MaskAddress(address string) string {
	if len(address) == 0 {
		return address
	}

	// 计算字符数量
	charCount := utf8.RuneCountInString(address)
	if charCount < 10 {
		return address
	}

	// 转换为rune切片
	runes := []rune(address)
	// 保留前5个字符和后5个字符，中间用*替换
	return string(runes[:5]) + "*****" + string(runes[charCount-5:])
}
