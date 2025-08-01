package service

import (
	"ai-svc/internal/model"
	"ai-svc/internal/repository"
	"ai-svc/pkg/logger"
	"context"
	"fmt"
	"strings"
	"time"
)

// LoginLogService 登录日志服务接口
type LoginLogService interface {
	// 记录登录尝试
	LogLoginAttempt(
		ctx context.Context,
		userID uint,
		phone, loginType, ip, userAgent string,
		location *model.LocationInfo,
	) error

	// 记录登录成功
	LogLoginSuccess(
		ctx context.Context,
		userID uint,
		phone, loginType, deviceID, deviceType, ip, userAgent string,
		isNewUser bool,
		location *model.LocationInfo,
	) error

	// 记录登录失败
	LogLoginFailed(
		ctx context.Context,
		userID uint,
		phone, loginType, reason, ip, userAgent string,
		location *model.LocationInfo,
	) error

	// 记录登出
	LogLogout(ctx context.Context, userID uint, ip, userAgent string, location *model.LocationInfo) error

	// 获取用户登录历史
	GetUserLoginHistory(ctx context.Context, userID uint, page, size int) ([]*model.UserBehaviorLog, int64, error)

	// 获取登录统计
	GetLoginStats(ctx context.Context, userID uint, days int) (*LoginStats, error)

	// 获取用户今日登录记录
	GetUserTodayLogins(ctx context.Context, userID uint) ([]*model.UserBehaviorLog, error)

	// 获取用户最近登录记录
	GetUserRecentLogins(ctx context.Context, userID uint, hours int) ([]*model.UserBehaviorLog, error)
}

// LoginStats 登录统计
type LoginStats struct {
	TotalLogins     int64 `json:"total_logins"`
	SuccessLogins   int64 `json:"success_logins"`
	FailedLogins    int64 `json:"failed_logins"`
	UniqueDevices   int64 `json:"unique_devices"`
	UniqueLocations int64 `json:"unique_locations"`
}

// loginLogService 登录日志服务实现
type loginLogService struct {
	behaviorLogRepo repository.UserBehaviorLogRepository
	userRepo        repository.UserRepository
	locationService LocationService
}

// NewLoginLogService 创建登录日志服务实例
func NewLoginLogService(
	behaviorLogRepo repository.UserBehaviorLogRepository,
	userRepo repository.UserRepository,
	locationService LocationService,
) LoginLogService {
	return &loginLogService{
		behaviorLogRepo: behaviorLogRepo,
		userRepo:        userRepo,
		locationService: locationService,
	}
}

// LogLoginAttempt 记录登录尝试
func (s *loginLogService) LogLoginAttempt(
	ctx context.Context,
	userID uint,
	phone, loginType, ip, userAgent string,
	location *model.LocationInfo,
) error {
	now := time.Now()
	log := &model.UserBehaviorLog{
		UserID:    userID,
		Action:    model.ActionLogin,
		Resource:  loginType, // 使用Resource字段存储登录方式
		IP:        ip,
		UserAgent: userAgent,
		LoginTime: &now, // 设置登录时间
		CreatedAt: now,  // 数据库创建时间
	}

	// 设置地理位置信息
	log.SetLocationInfo(location)

	err := s.behaviorLogRepo.Create(log)
	if err != nil {
		logger.Error("记录登录尝试失败", map[string]any{
			"user_id": userID,
			"phone":   phone,
			"error":   err.Error(),
		})
		return err
	}

	logger.Info("记录登录尝试成功", map[string]any{
		"user_id":    userID,
		"phone":      phone,
		"login_type": loginType,
		"ip":         ip,
	})

	return nil
}

// LogLoginSuccess 记录登录成功
func (s *loginLogService) LogLoginSuccess(
	ctx context.Context,
	userID uint,
	phone, loginType, deviceID, deviceType, ip, userAgent string,
	isNewUser bool,
	location *model.LocationInfo,
) error {
	now := time.Now()
	// 构建资源信息（包含设备信息）
	resource := fmt.Sprintf("%s|device:%s|type:%s|new_user:%t", loginType, deviceID, deviceType, isNewUser)

	log := &model.UserBehaviorLog{
		UserID:    userID,
		Action:    model.ActionLoginSuccess,
		Resource:  resource,
		IP:        ip,
		UserAgent: userAgent,
		LoginTime: &now, // 设置登录时间
		CreatedAt: now,  // 数据库创建时间
	}

	// 设置地理位置信息
	log.SetLocationInfo(location)

	err := s.behaviorLogRepo.Create(log)
	if err != nil {
		logger.Error("记录登录成功失败", map[string]any{
			"user_id": userID,
			"phone":   phone,
			"error":   err.Error(),
		})
		return err
	}

	logger.Info("记录登录成功", map[string]any{
		"user_id":     userID,
		"phone":       phone,
		"login_type":  loginType,
		"device_id":   deviceID,
		"device_type": deviceType,
		"is_new_user": isNewUser,
		"ip":          ip,
	})

	return nil
}

// LogLoginFailed 记录登录失败
func (s *loginLogService) LogLoginFailed(
	ctx context.Context,
	userID uint,
	phone, loginType, reason, ip, userAgent string,
	location *model.LocationInfo,
) error {
	now := time.Now()
	// 构建资源信息（包含失败原因）
	resource := fmt.Sprintf("%s|reason:%s", loginType, reason)

	log := &model.UserBehaviorLog{
		UserID:    userID,
		Action:    model.ActionLoginFailed,
		Resource:  resource,
		IP:        ip,
		UserAgent: userAgent,
		LoginTime: &now, // 设置登录时间
		CreatedAt: now,  // 数据库创建时间
	}

	// 设置地理位置信息
	log.SetLocationInfo(location)

	err := s.behaviorLogRepo.Create(log)
	if err != nil {
		logger.Error("记录登录失败失败", map[string]any{
			"user_id": userID,
			"phone":   phone,
			"error":   err.Error(),
		})
		return err
	}

	logger.Info("记录登录失败", map[string]any{
		"user_id":    userID,
		"phone":      phone,
		"login_type": loginType,
		"reason":     reason,
		"ip":         ip,
	})

	return nil
}

// LogLogout 记录登出
func (s *loginLogService) LogLogout(
	ctx context.Context,
	userID uint,
	ip, userAgent string,
	location *model.LocationInfo,
) error {
	now := time.Now()
	log := &model.UserBehaviorLog{
		UserID:    userID,
		Action:    model.ActionLogout,
		Resource:  "logout",
		IP:        ip,
		UserAgent: userAgent,
		LoginTime: &now, // 设置登出时间
		CreatedAt: now,  // 数据库创建时间
	}

	// 设置地理位置信息
	log.SetLocationInfo(location)

	err := s.behaviorLogRepo.Create(log)
	if err != nil {
		logger.Error("记录登出失败", map[string]any{
			"user_id": userID,
			"error":   err.Error(),
		})
		return err
	}

	logger.Info("记录登出成功", map[string]any{
		"user_id": userID,
		"ip":      ip,
	})

	return nil
}

// GetUserLoginHistory 获取用户登录历史
func (s *loginLogService) GetUserLoginHistory(
	ctx context.Context,
	userID uint,
	page, size int,
) ([]*model.UserBehaviorLog, int64, error) {
	logs, total, err := s.behaviorLogRepo.GetUserLoginHistory(userID, page, size)
	if err != nil {
		logger.Error("获取用户登录历史失败", map[string]any{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, 0, err
	}

	return logs, total, nil
}

// GetLoginStats 获取登录统计
func (s *loginLogService) GetLoginStats(ctx context.Context, userID uint, days int) (*LoginStats, error) {
	// 获取总登录次数
	totalLogins, err := s.behaviorLogRepo.CountUserLogins(userID, days)
	if err != nil {
		return nil, err
	}

	// 获取失败登录次数
	failedLogins, err := s.behaviorLogRepo.CountFailedLogins(userID, days)
	if err != nil {
		return nil, err
	}

	// 计算成功登录次数
	successLogins := totalLogins - failedLogins

	// 获取最近的登录记录来计算唯一设备和位置
	recentLogs, err := s.behaviorLogRepo.GetRecentLogins(userID, days*24)
	if err != nil {
		return nil, err
	}

	// 统计唯一设备
	uniqueDevices := make(map[string]bool)
	uniqueLocations := make(map[string]bool)

	for _, log := range recentLogs {
		if log.Action == model.ActionLoginSuccess {
			// 解析设备信息
			if deviceInfo := s.parseDeviceInfo(log.Resource); deviceInfo != "" {
				uniqueDevices[deviceInfo] = true
			}

			// 统计地理位置
			if log.Location != "" {
				uniqueLocations[log.Location] = true
			}
		}
	}

	return &LoginStats{
		TotalLogins:     totalLogins,
		SuccessLogins:   successLogins,
		FailedLogins:    failedLogins,
		UniqueDevices:   int64(len(uniqueDevices)),
		UniqueLocations: int64(len(uniqueLocations)),
	}, nil
}

// GetUserTodayLogins 获取用户今日登录记录
func (s *loginLogService) GetUserTodayLogins(ctx context.Context, userID uint) ([]*model.UserBehaviorLog, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	return s.behaviorLogRepo.GetLoginsByTimeRange(userID, startOfDay, endOfDay)
}

// GetUserRecentLogins 获取用户最近登录记录
func (s *loginLogService) GetUserRecentLogins(
	ctx context.Context,
	userID uint,
	hours int,
) ([]*model.UserBehaviorLog, error) {
	return s.behaviorLogRepo.GetRecentLogins(userID, hours)
}

// parseDeviceInfo 解析设备信息
func (s *loginLogService) parseDeviceInfo(resource string) string {
	if resource == "" {
		return ""
	}

	parts := strings.Split(resource, "|")
	for _, part := range parts {
		if strings.HasPrefix(part, "device:") {
			return strings.TrimPrefix(part, "device:")
		}
	}

	return ""
}

// LocationService 地理位置服务接口（可选实现）
type LocationService interface {
	// 根据IP获取地理位置
	GetLocationByIP(ip string) (*model.LocationInfo, error)

	// 批量获取地理位置
	BatchGetLocation(ips []string) (map[string]*model.LocationInfo, error)

	// 缓存地理位置信息
	CacheLocation(ip string, location *model.LocationInfo) error
}

// 默认地理位置服务实现（空实现，可根据需要扩展）
type defaultLocationService struct{}

func NewDefaultLocationService() LocationService {
	return &defaultLocationService{}
}

func (s *defaultLocationService) GetLocationByIP(ip string) (*model.LocationInfo, error) {
	// 这里可以实现IP地理位置解析
	// 目前返回空，表示不解析地理位置
	return nil, nil
}

func (s *defaultLocationService) BatchGetLocation(ips []string) (map[string]*model.LocationInfo, error) {
	return make(map[string]*model.LocationInfo), nil
}

func (s *defaultLocationService) CacheLocation(ip string, location *model.LocationInfo) error {
	return nil
}
