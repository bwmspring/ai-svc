package model

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// 登录相关的Action常量
const (
	ActionLogin        = "login"         // 登录
	ActionLoginSuccess = "login_success" // 登录成功
	ActionLoginFailed  = "login_failed"  // 登录失败
	ActionLogout       = "logout"        // 登出
	ActionRefreshToken = "refresh_token" // 刷新Token
)

// 登录方式常量
const (
	LoginTypeSMS      = "sms"      // 短信登录
	LoginTypePassword = "password" // 密码登录
	LoginTypeOAuth    = "oauth"    // 第三方登录
	LoginTypeRefresh  = "refresh"  // Token刷新
)

// 登录状态常量
const (
	LoginStatusSuccess = "success" // 登录成功
	LoginStatusFailed  = "failed"  // 登录失败
	LoginStatusBlocked = "blocked" // 登录被阻止
)

// UserBehaviorLog 用户行为日志（频繁变更的数据独立存储）.
type UserBehaviorLog struct {
	ID        uint       `gorm:"primarykey"                json:"id"`
	UserID    uint       `gorm:"index;not null"            json:"user_id"`
	Action    string     `gorm:"type:varchar(50);not null" json:"action"`   // login, logout, view, purchase等
	Resource  string     `gorm:"type:varchar(100)"         json:"resource"` // 操作的资源
	IP        string     `gorm:"type:varchar(45)"          json:"ip"`
	UserAgent string     `gorm:"type:varchar(500)"         json:"user_agent"`
	Location  string     `gorm:"type:varchar(200)"         json:"location"`   // 地理位置信息
	LoginTime *time.Time `gorm:"index"                     json:"login_time"` // 登录时间（专门字段）
	CreatedAt time.Time  `                                 json:"created_at"` // 数据库创建时间（保留）
}

// TableName 表名.
func (UserBehaviorLog) TableName() string {
	return "user_behavior_logs"
}

// LocationInfo 地理位置信息
type LocationInfo struct {
	Latitude  float64 `json:"latitude"`  // 纬度
	Longitude float64 `json:"longitude"` // 经度
	Address   string  `json:"address"`   // 详细地址
	City      string  `json:"city"`      // 城市
	Province  string  `json:"province"`  // 省份
	Country   string  `json:"country"`   // 国家
	District  string  `json:"district"`  // 区县
}

// ToString 将LocationInfo转换为字符串存储
func (l *LocationInfo) ToString() string {
	if l == nil {
		return ""
	}

	parts := []string{}
	if l.Country != "" {
		parts = append(parts, l.Country)
	}
	if l.Province != "" {
		parts = append(parts, l.Province)
	}
	if l.City != "" {
		parts = append(parts, l.City)
	}
	if l.District != "" {
		parts = append(parts, l.District)
	}
	if l.Address != "" {
		parts = append(parts, l.Address)
	}

	if len(parts) > 0 {
		return strings.Join(parts, ",")
	}

	// 如果没有详细地址，返回经纬度
	if l.Latitude != 0 && l.Longitude != 0 {
		return fmt.Sprintf("%.6f,%.6f", l.Latitude, l.Longitude)
	}

	return ""
}

// ParseLocationInfo 从字符串解析LocationInfo
func ParseLocationInfo(locationStr string) *LocationInfo {
	if locationStr == "" {
		return nil
	}

	// 尝试解析经纬度格式
	if strings.Contains(locationStr, ",") && !strings.Contains(locationStr, "省") &&
		!strings.Contains(locationStr, "市") {
		parts := strings.Split(locationStr, ",")
		if len(parts) == 2 {
			if lat, err := strconv.ParseFloat(parts[0], 64); err == nil {
				if lng, err := strconv.ParseFloat(parts[1], 64); err == nil {
					return &LocationInfo{
						Latitude:  lat,
						Longitude: lng,
					}
				}
			}
		}
	}

	// 其他格式可根据实际需要扩展
	return &LocationInfo{
		Address: locationStr,
	}
}

// GetLocationInfo 获取地理位置信息
func (log *UserBehaviorLog) GetLocationInfo() *LocationInfo {
	return ParseLocationInfo(log.Location)
}

// SetLocationInfo 设置地理位置信息
func (log *UserBehaviorLog) SetLocationInfo(location *LocationInfo) {
	if location != nil {
		log.Location = location.ToString()
	} else {
		log.Location = ""
	}
}
