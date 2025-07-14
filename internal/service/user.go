package service

import (
	"errors"
	"time"

	"ai-svc/internal/model"
	"ai-svc/internal/repository"
	"ai-svc/pkg/logger"

	"gorm.io/gorm"
)

// UserService 用户服务接口
type UserService interface {
	// 认证相关
	LoginWithSMS(req *model.LoginWithSMSRequest, ip, userAgent string) (*model.LoginResponse, bool, error)

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

// userService 用户服务实现
type userService struct {
	userRepo      repository.UserRepository
	smsService    SMSService
	deviceService DeviceService
	jwtService    JWTService
}

// NewUserService 创建用户服务实例
func NewUserService(userRepo repository.UserRepository, smsService SMSService, deviceService DeviceService) UserService {
	return &userService{
		userRepo:      userRepo,
		smsService:    smsService,
		deviceService: deviceService,
		jwtService:    NewJWTService(),
	}
}

// LoginWithSMS 手机号+验证码登录（同时完成注册）
func (s *userService) LoginWithSMS(req *model.LoginWithSMSRequest, ip, userAgent string) (*model.LoginResponse, bool, error) {
	// 验证验证码
	if err := s.smsService.ValidateVerificationCode(req.Phone, req.Code, "login"); err != nil {
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
				logger.Error("创建用户失败", map[string]interface{}{"error": err.Error()})
				return nil, false, errors.New("创建用户失败")
			}
			isNewUser = true
		} else {
			logger.Error("查询用户失败", map[string]interface{}{"error": err.Error()})
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
			logger.Error("更新用户登录信息失败", map[string]interface{}{"error": err.Error()})
			// 这里不返回错误，登录依然成功
		}
	}

	// 注册/更新设备
	device, err := s.deviceService.RegisterDevice(user.ID, req.DeviceInfo, ip, userAgent)
	if err != nil {
		logger.Error("注册设备失败", map[string]interface{}{"error": err.Error()})
		return nil, false, errors.New("设备注册失败")
	}

	// 创建会话记录
	session, err := s.deviceService.CreateSession(user.ID, device.DeviceID, "")
	if err != nil {
		logger.Error("创建会话失败", map[string]interface{}{"error": err.Error()})
		return nil, false, errors.New("创建会话失败")
	}

	// 生成包含会话信息的JWT Token
	jwtToken, err := s.jwtService.GenerateToken(user, device, session.SessionToken)
	if err != nil {
		logger.Error("生成JWT Token失败", map[string]interface{}{"error": err.Error()})
		return nil, false, errors.New("生成Token失败")
	}

	// 更新会话记录中的JWT Token
	session.JWTToken = jwtToken
	if err := s.deviceService.UpdateSession(session); err != nil {
		logger.Warn("更新会话JWT Token失败", map[string]interface{}{"error": err.Error()})
		// 不影响登录流程，只记录警告
	}

	logger.Info("用户登录成功", map[string]interface{}{
		"user_id":     user.ID,
		"phone":       user.Phone,
		"device_id":   device.DeviceID,
		"device_type": device.DeviceType,
		"is_new_user": isNewUser,
		"session_id":  session.ID,
	})

	return &model.LoginResponse{
		User:  s.convertToResponse(user),
		Token: jwtToken,
	}, isNewUser, nil
}

// GetUserByID 根据ID获取用户
func (s *userService) GetUserByID(id uint) (*model.UserResponse, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		logger.Error("获取用户失败", map[string]interface{}{"error": err.Error(), "id": id})
		return nil, errors.New("获取用户失败")
	}

	return s.convertToResponse(user), nil
}

// UpdateUser 更新用户
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
		logger.Error("更新用户失败", map[string]interface{}{"error": err.Error(), "id": id})
		return nil, errors.New("更新用户失败")
	}

	return s.convertToResponse(user), nil
}

// DeleteUser 删除用户
func (s *userService) DeleteUser(id uint) error {
	if _, err := s.userRepo.GetByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		return errors.New("获取用户失败")
	}

	if err := s.userRepo.Delete(id); err != nil {
		logger.Error("删除用户失败", map[string]interface{}{"error": err.Error(), "id": id})
		return errors.New("删除用户失败")
	}

	return nil
}

// GetUserList 获取用户列表
func (s *userService) GetUserList(page, size int) ([]*model.UserResponse, int64, error) {
	offset := (page - 1) * size
	users, total, err := s.userRepo.List(offset, size)
	if err != nil {
		logger.Error("获取用户列表失败", map[string]interface{}{"error": err.Error()})
		return nil, 0, errors.New("获取用户列表失败")
	}

	responses := make([]*model.UserResponse, len(users))
	for i, user := range users {
		responses[i] = s.convertToResponse(user)
	}

	return responses, total, nil
}

// SearchUsers 搜索用户
func (s *userService) SearchUsers(keyword string, page, size int) ([]*model.UserResponse, int64, error) {
	offset := (page - 1) * size
	users, total, err := s.userRepo.Search(keyword, offset, size)
	if err != nil {
		logger.Error("搜索用户失败", map[string]interface{}{"error": err.Error(), "keyword": keyword})
		return nil, 0, errors.New("搜索用户失败")
	}

	responses := make([]*model.UserResponse, len(users))
	for i, user := range users {
		responses[i] = s.convertToResponse(user)
	}

	return responses, total, nil
}

// GetUserDevices 获取用户设备列表
func (s *userService) GetUserDevices(userID uint) (*model.UserDevicesResponse, error) {
	return s.deviceService.GetUserDevices(userID)
}

// KickDevices 踢出设备
func (s *userService) KickDevices(userID uint, req *model.KickDeviceRequest) error {
	return s.deviceService.KickDevices(userID, req.DeviceIDs)
}

// convertToResponse 转换为响应结构
func (s *userService) convertToResponse(user *model.User) *model.UserResponse {
	return &model.UserResponse{
		ID:          user.ID,
		Phone:       user.Phone,
		Username:    user.Username,
		Email:       user.Email,
		Nickname:    user.Nickname,
		Avatar:      user.Avatar,
		RealName:    user.RealName,
		Gender:      user.Gender,
		VIPLevel:    user.VIPLevel,
		VIPExpireAt: user.VIPExpireAt,
		Status:      user.Status,
		LastLoginIP: user.LastLoginIP,
		LastLoginAt: user.LastLoginAt,
		LoginCount:  user.LoginCount,
		Birthday:    user.Birthday,
		Address:     user.Address,
		Bio:         user.Bio,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}
