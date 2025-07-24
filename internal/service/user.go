package service

import (
	"errors"
	"time"

	"ai-svc/internal/model"
	"ai-svc/internal/repository"
	"ai-svc/pkg/logger"

	"gorm.io/gorm"
)

// UserService 用户服务接口.
type UserService interface {
	// 认证相关
	LoginWithSMS(req *model.LoginWithSMSRequest, ip, userAgent string) (*model.LoginResponse, bool, error)
	RefreshToken(refreshToken string) (*model.TokenPair, error) // 新增Token刷新方法

	// 用户管理
	GetUserByID(id uint) (*model.UserResponse, error)
	UpdateUser(id uint, req *model.UpdateUserRequest) (*model.UserResponse, error)
	DeleteUser(id uint) error
	GetUserList(page, size int) ([]*model.UserResponse, int64, error)
	SearchUsers(keyword string, page, size int) ([]*model.UserResponse, int64, error)

	// 设备管理
	GetUserDevices(userID uint) (*model.UserDevicesResponse, error)
	KickDevices(userID uint, req *model.KickDeviceRequest) error
}

// userService 用户服务实现.
type userService struct {
	userRepo      repository.UserRepository
	smsService    SMSService
	deviceService DeviceService
	jwtService    JWTService
}

// NewUserService 创建用户服务实例.
func NewUserService(
	userRepo repository.UserRepository,
	smsService SMSService,
	deviceService DeviceService,
) UserService {
	return &userService{
		userRepo:      userRepo,
		smsService:    smsService,
		deviceService: deviceService,
		jwtService:    NewJWTServiceWithDeviceService(deviceService),
	}
}

// LoginWithSMS 手机号+验证码登录（同时完成注册)
func (s *userService) LoginWithSMS(
	req *model.LoginWithSMSRequest,
	ip, userAgent string,
) (*model.LoginResponse, bool, error) {
	// 验证验证码
	if err := s.smsService.ValidateVerificationCode(req.Phone, req.Code, "login", req.Token); err != nil {
		return nil, false, err
	}

	// 验证设备信息
	if req.DeviceInfo == nil {
		return nil, false, errors.New("设备信息不能为空")
	}

	// 查找用户
	user, err := s.userRepo.GetByPhone(req.Phone)
	isNewUser := false

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 用户不存在，创建新用户
			user = &model.User{
				Phone:       req.Phone,
				Status:      1,
				LastLoginIP: ip,
				LoginCount:  1,
			}
			now := time.Now()
			user.LastLoginAt = &now

			if err := s.userRepo.Create(user); err != nil {
				logger.Error("创建用户失败", map[string]any{"error": err.Error()})
				return nil, false, errors.New("创建用户失败")
			}
			isNewUser = true
		} else {
			logger.Error("查询用户失败", map[string]any{"error": err.Error()})
			return nil, false, errors.New("登录失败")
		}
	} else {
		// 用户存在，检查状态
		if user.Status == 0 {
			return nil, false, errors.New("账户已被禁用")
		}

		// 更新登录信息
		now := time.Now()
		user.LastLoginAt = &now
		user.LastLoginIP = ip
		user.LoginCount++

		if err := s.userRepo.Update(user); err != nil {
			logger.Error("更新用户登录信息失败", map[string]any{"error": err.Error()})
			// 这里不返回错误，登录依然成功
		}
	}

	// 处理设备登录
	device, err := s.deviceService.HandleDeviceLogin(user.ID, req.DeviceInfo, ip, userAgent)
	if err != nil {
		logger.Error("设备登录失败", map[string]any{"error": err.Error()})
		return nil, false, errors.New("设备登录失败")
	}

	// 生成包含设备信息的JWT Token
	accessToken, err := s.jwtService.GenerateToken(user, device, device.DeviceID) // 使用deviceID作为唯一标识
	if err != nil {
		logger.Error("生成Access Token失败", map[string]any{"error": err.Error()})
		return nil, false, errors.New("生成Access Token失败")
	}

	// 生成刷新令牌
	refreshToken, err := s.jwtService.GenerateRefreshToken(user, device)
	if err != nil {
		logger.Error("生成Refresh Token失败", map[string]any{"error": err.Error()})
		return nil, false, errors.New("生成Refresh Token失败")
	}

	logger.Info("用户登录成功", map[string]any{
		"user_id":     user.ID,
		"phone":       user.Phone,
		"device_id":   device.DeviceID,
		"device_type": device.DeviceType,
		"is_new_user": isNewUser,
	})

	return &model.LoginResponse{
		User:         s.convertToResponse(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    24 * 3600, // 24小时，单位秒
		TokenType:    "Bearer",
	}, isNewUser, nil
}

// RefreshToken 刷新Token（通过JWT服务）.
func (s *userService) RefreshToken(refreshToken string) (*model.TokenPair, error) {
	return s.jwtService.RefreshToken(refreshToken)
}

// GetUserByID 根据ID获取用户.
func (s *userService) GetUserByID(id uint) (*model.UserResponse, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		logger.Error("获取用户失败", map[string]any{"error": err.Error(), "id": id})
		return nil, errors.New("获取用户失败")
	}

	return s.convertToResponse(user), nil
}

// UpdateUser 更新用户.
func (s *userService) UpdateUser(id uint, req *model.UpdateUserRequest) (*model.UserResponse, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, errors.New("获取用户失败")
	}

	// 更新字段
	if req.Username != "" {
		user.Username = req.Username
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}
	if req.RealName != "" {
		user.RealName = req.RealName
	}
	if req.Gender > 0 {
		user.Gender = req.Gender
	}
	if req.Birthday != nil {
		user.Birthday = req.Birthday
	}
	if req.Address != "" {
		user.Address = req.Address
	}
	if req.Bio != "" {
		user.Bio = req.Bio
	}

	if err := s.userRepo.Update(user); err != nil {
		logger.Error("更新用户失败", map[string]any{"error": err.Error(), "id": id})
		return nil, errors.New("更新用户失败")
	}

	return s.convertToResponse(user), nil
}

// DeleteUser 删除用户.
func (s *userService) DeleteUser(id uint) error {
	if _, err := s.userRepo.GetByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		return errors.New("获取用户失败")
	}

	if err := s.userRepo.Delete(id); err != nil {
		logger.Error("删除用户失败", map[string]any{"error": err.Error(), "id": id})
		return errors.New("删除用户失败")
	}

	return nil
}

// GetUserList 获取用户列表.
func (s *userService) GetUserList(page, size int) ([]*model.UserResponse, int64, error) {
	offset := (page - 1) * size
	users, total, err := s.userRepo.List(offset, size)
	if err != nil {
		logger.Error("获取用户列表失败", map[string]any{"error": err.Error()})
		return nil, 0, errors.New("获取用户列表失败")
	}

	responses := make([]*model.UserResponse, len(users))
	for i, user := range users {
		responses[i] = s.convertToResponse(user)
	}

	return responses, total, nil
}

// SearchUsers 搜索用户.
func (s *userService) SearchUsers(keyword string, page, size int) ([]*model.UserResponse, int64, error) {
	offset := (page - 1) * size
	users, total, err := s.userRepo.Search(keyword, offset, size)
	if err != nil {
		logger.Error("搜索用户失败", map[string]any{"error": err.Error(), "keyword": keyword})
		return nil, 0, errors.New("搜索用户失败")
	}

	responses := make([]*model.UserResponse, len(users))
	for i, user := range users {
		responses[i] = s.convertToResponse(user)
	}

	return responses, total, nil
}

// GetUserDevices 获取用户设备列表.
func (s *userService) GetUserDevices(userID uint) (*model.UserDevicesResponse, error) {
	deviceList, err := s.deviceService.GetUserDevices(userID, "")
	if err != nil {
		return nil, err
	}

	// 转换为旧的响应格式以保持兼容性
	deviceResponses := make([]*model.DeviceResponse, len(deviceList.Devices))
	for i, device := range deviceList.Devices {
		deviceResponses[i] = &model.DeviceResponse{
			ID:           device.ID,
			DeviceID:     device.DeviceID,
			DeviceType:   device.DeviceType,
			DeviceName:   device.DeviceName,
			AppVersion:   device.AppVersion,
			OSVersion:    device.OSVersion,
			ClientIP:     device.ClientIP,
			LoginAt:      device.LoginAt,
			LastActiveAt: device.LastActiveAt,
			IsOnline:     device.IsOnline,
			CreatedAt:    device.LoginAt, // 使用LoginAt作为CreatedAt的替代
		}
	}

	return &model.UserDevicesResponse{
		Devices:     deviceResponses,
		TotalCount:  deviceList.Summary.Total,
		OnlineCount: deviceList.Summary.Online,
		MaxDevices:  deviceList.Limits.Mobile + deviceList.Limits.PC + deviceList.Limits.Web + deviceList.Limits.Miniprogram,
	}, nil
}

// KickDevices 踢出设备.
func (s *userService) KickDevices(userID uint, req *model.KickDeviceRequest) error {
	return s.deviceService.KickDevices(userID, req.DeviceIDs)
}

// convertToResponse 转换为响应结构.
func (s *userService) convertToResponse(user *model.User) *model.UserResponse {
	// 辅助函数：将字符串转为字符串指针
	toStringPtr := func(s string) *string {
		if s == "" {
			return nil
		}
		return &s
	}

	return &model.UserResponse{
		ID:          user.ID,
		Phone:       user.GetMaskedPhone(), // 脱敏手机号
		Username:    toStringPtr(user.Username),
		Email:       toStringPtr(user.GetMaskedEmail()), // 脱敏邮箱
		Nickname:    toStringPtr(user.Nickname),
		Avatar:      toStringPtr(user.Avatar),
		RealName:    toStringPtr(user.GetMaskedRealName()), // 脱敏真实姓名
		Gender:      user.Gender,
		VIPLevel:    user.VIPLevel,
		VIPExpireAt: user.VIPExpireAt,
		Status:      user.Status,
		LastLoginIP: toStringPtr(user.LastLoginIP),
		LastLoginAt: user.LastLoginAt,
		LoginCount:  user.LoginCount,
		Birthday:    user.Birthday,
		Address:     toStringPtr(user.Address),
		Bio:         toStringPtr(user.Bio),
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}
