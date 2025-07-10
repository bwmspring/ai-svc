package service

import (
	"errors"

	"ai-svc/internal/model"
	"ai-svc/internal/repository"
	"ai-svc/pkg/logger"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserService 用户服务接口
type UserService interface {
	CreateUser(req *model.CreateUserRequest) (*model.UserResponse, error)
	GetUserByID(id uint) (*model.UserResponse, error)
	UpdateUser(id uint, req *model.UpdateUserRequest) (*model.UserResponse, error)
	DeleteUser(id uint) error
	Login(req *model.LoginRequest, ip string) (*model.UserResponse, error)
	ChangePassword(id uint, req *model.ChangePasswordRequest) error
	GetUserList(page, size int) ([]*model.UserResponse, int64, error)
	SearchUsers(keyword string, page, size int) ([]*model.UserResponse, int64, error)
}

// userService 用户服务实现
type userService struct {
	userRepo repository.UserRepository
}

// NewUserService 创建用户服务实例
func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

// CreateUser 创建用户
func (s *userService) CreateUser(req *model.CreateUserRequest) (*model.UserResponse, error) {
	// 检查用户名是否存在
	if _, err := s.userRepo.GetByUsername(req.Username); err == nil {
		return nil, errors.New("用户名已存在")
	}

	// 检查邮箱是否存在
	if _, err := s.userRepo.GetByEmail(req.Email); err == nil {
		return nil, errors.New("邮箱已存在")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("密码加密失败", map[string]interface{}{"error": err.Error()})
		return nil, errors.New("密码加密失败")
	}

	// 创建用户
	user := &model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		Nickname: req.Nickname,
		Status:   1,
	}

	if err := s.userRepo.Create(user); err != nil {
		logger.Error("创建用户失败", map[string]interface{}{"error": err.Error()})
		return nil, errors.New("创建用户失败")
	}

	return s.convertToResponse(user), nil
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
	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
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

// Login 用户登录
func (s *userService) Login(req *model.LoginRequest, ip string) (*model.UserResponse, error) {
	// 获取用户
	user, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户名或密码错误")
		}
		return nil, errors.New("登录失败")
	}

	// 检查用户状态
	if user.Status == 0 {
		return nil, errors.New("账户已被禁用")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	// 更新最后登录信息
	if err := s.userRepo.UpdateLastLogin(user.ID, ip); err != nil {
		logger.Warn("更新最后登录信息失败", map[string]interface{}{"error": err.Error(), "user_id": user.ID})
	}

	// 重新获取用户信息（包含更新后的最后登录时间）
	user, _ = s.userRepo.GetByID(user.ID)

	return s.convertToResponse(user), nil
}

// ChangePassword 修改密码
func (s *userService) ChangePassword(id uint, req *model.ChangePasswordRequest) error {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		return errors.New("获取用户失败")
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		return errors.New("原密码错误")
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("密码加密失败", map[string]interface{}{"error": err.Error()})
		return errors.New("密码加密失败")
	}

	// 更新密码
	if err := s.userRepo.UpdatePassword(id, string(hashedPassword)); err != nil {
		logger.Error("更新密码失败", map[string]interface{}{"error": err.Error(), "id": id})
		return errors.New("更新密码失败")
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

// convertToResponse 转换为响应结构
func (s *userService) convertToResponse(user *model.User) *model.UserResponse {
	return &model.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Nickname:  user.Nickname,
		Avatar:    user.Avatar,
		Status:    user.Status,
		LastIP:    user.LastIP,
		LastTime:  user.LastTime,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
