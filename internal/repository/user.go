package repository

import (
	"ai-svc/internal/model"
	"ai-svc/pkg/database"

	"gorm.io/gorm"
)

// UserRepository 用户仓储接口.
type UserRepository interface {
	Create(user *model.User) error
	GetByID(id uint) (*model.User, error)
	GetByUsername(username string) (*model.User, error)
	GetByEmail(email string) (*model.User, error)
	GetByPhone(phone string) (*model.User, error)
	Update(user *model.User) error
	Delete(id uint) error
	List(offset, limit int) ([]*model.User, int64, error)
	Search(keyword string, offset, limit int) ([]*model.User, int64, error)

	// SMS 验证码相关方法
	CreateSMSCode(smsCode *model.SMSVerificationCode) error
	GetSMSCode(phone, purpose string) (*model.SMSVerificationCode, error)
	UpdateSMSCode(smsCode *model.SMSVerificationCode) error
}

// userRepository 用户仓储实现.
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓储实例.
func NewUserRepository() UserRepository {
	return &userRepository{
		db: database.GetDB(),
	}
}

// Create 创建用户.
func (r *userRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

// GetByID 根据ID获取用户.
func (r *userRepository) GetByID(id uint) (*model.User, error) {
	var user model.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByUsername 根据用户名获取用户.
func (r *userRepository) GetByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail 根据邮箱获取用户.
func (r *userRepository) GetByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByPhone 根据手机号获取用户.
func (r *userRepository) GetByPhone(phone string) (*model.User, error) {
	var user model.User
	err := r.db.Where("phone = ?", phone).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update 更新用户.
func (r *userRepository) Update(user *model.User) error {
	return r.db.Save(user).Error
}

// Delete 删除用户.
func (r *userRepository) Delete(id uint) error {
	return r.db.Delete(&model.User{}, id).Error
}

// List 获取用户列表.
func (r *userRepository) List(offset, limit int) ([]*model.User, int64, error) {
	var users []*model.User
	var total int64

	// 获取总数
	if err := r.db.Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err := r.db.Offset(offset).Limit(limit).Find(&users).Error
	return users, total, err
}

// Search 搜索用户.
func (r *userRepository) Search(keyword string, offset, limit int) ([]*model.User, int64, error) {
	var users []*model.User
	var total int64

	query := r.db.Model(&model.User{}).Where("username LIKE ? OR email LIKE ? OR nickname LIKE ? OR phone LIKE ?",
		"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err := query.Offset(offset).Limit(limit).Find(&users).Error
	return users, total, err
}

// CreateSMSCode 创建短信验证码.
func (r *userRepository) CreateSMSCode(smsCode *model.SMSVerificationCode) error {
	return r.db.Create(smsCode).Error
}

// GetSMSCode 根据手机号和用途获取最新的未使用验证码.
func (r *userRepository) GetSMSCode(phone, purpose string) (*model.SMSVerificationCode, error) {
	var smsCode model.SMSVerificationCode
	err := r.db.Where("phone = ? AND purpose = ? AND used_at IS NULL", phone, purpose).
		Order("created_at DESC").First(&smsCode).Error
	if err != nil {
		return nil, err
	}
	return &smsCode, nil
}

// UpdateSMSCode 更新短信验证码.
func (r *userRepository) UpdateSMSCode(smsCode *model.SMSVerificationCode) error {
	return r.db.Save(smsCode).Error
}
