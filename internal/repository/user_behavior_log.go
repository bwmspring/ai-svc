package repository

import (
	"ai-svc/internal/model"
	"ai-svc/pkg/database"
	"time"

	"gorm.io/gorm"
)

// UserBehaviorLogRepository 用户行为日志仓储接口
type UserBehaviorLogRepository interface {
	// 基础CRUD
	Create(log *model.UserBehaviorLog) error
	GetByID(id uint) (*model.UserBehaviorLog, error)
	Update(log *model.UserBehaviorLog) error
	Delete(id uint) error

	// 查询方法
	GetUserLogs(userID uint, action string, page, size int) ([]*model.UserBehaviorLog, int64, error)
	GetUserLoginHistory(userID uint, page, size int) ([]*model.UserBehaviorLog, int64, error)
	GetRecentLogins(userID uint, hours int) ([]*model.UserBehaviorLog, error)
	GetFailedLogins(userID uint, hours int) ([]*model.UserBehaviorLog, error)

	// 按登录时间查询
	GetLoginsByTimeRange(userID uint, startTime, endTime time.Time) ([]*model.UserBehaviorLog, error)
	GetLoginsByDate(userID uint, date time.Time) ([]*model.UserBehaviorLog, error)

	// 统计方法
	CountUserLogins(userID uint, days int) (int64, error)
	CountFailedLogins(userID uint, days int) (int64, error)
	CountLoginsByType(userID uint, loginType string, days int) (int64, error)
}

// userBehaviorLogRepository 用户行为日志仓储实现
type userBehaviorLogRepository struct {
	db *gorm.DB
}

// NewUserBehaviorLogRepository 创建用户行为日志仓储实例
func NewUserBehaviorLogRepository() UserBehaviorLogRepository {
	return &userBehaviorLogRepository{
		db: database.GetDB(),
	}
}

// Create 创建用户行为日志
func (r *userBehaviorLogRepository) Create(log *model.UserBehaviorLog) error {
	return r.db.Create(log).Error
}

// GetByID 根据ID获取用户行为日志
func (r *userBehaviorLogRepository) GetByID(id uint) (*model.UserBehaviorLog, error) {
	var log model.UserBehaviorLog
	err := r.db.First(&log, id).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

// Update 更新用户行为日志
func (r *userBehaviorLogRepository) Update(log *model.UserBehaviorLog) error {
	return r.db.Save(log).Error
}

// Delete 删除用户行为日志
func (r *userBehaviorLogRepository) Delete(id uint) error {
	return r.db.Delete(&model.UserBehaviorLog{}, id).Error
}

// GetUserLogs 获取用户日志
func (r *userBehaviorLogRepository) GetUserLogs(
	userID uint,
	action string,
	page, size int,
) ([]*model.UserBehaviorLog, int64, error) {
	var logs []*model.UserBehaviorLog
	var total int64

	// 参数校验
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10 // 默认每页10条
	}
	offset := (page - 1) * size

	// 构建查询
	query := r.db.Model(&model.UserBehaviorLog{}).Where("user_id = ?", userID)
	if action != "" {
		query = query.Where("action = ?", action)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 如果没有数据，直接返回空
	if total == 0 {
		return []*model.UserBehaviorLog{}, 0, nil
	}

	// 查询数据，按创建时间倒序排列
	err := query.Order("created_at DESC").
		Offset(offset).
		Limit(size).
		Find(&logs).Error
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetUserLoginHistory 获取用户登录历史（按登录时间排序）
func (r *userBehaviorLogRepository) GetUserLoginHistory(
	userID uint,
	page, size int,
) ([]*model.UserBehaviorLog, int64, error) {
	var logs []*model.UserBehaviorLog
	var total int64

	offset := (page - 1) * size

	// 查询登录相关的行为
	query := r.db.Where(
		"user_id = ? AND action IN (?)",
		userID,
		[]string{model.ActionLogin, model.ActionLoginSuccess, model.ActionLoginFailed},
	)

	// 获取总数
	if err := query.Model(&model.UserBehaviorLog{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 按登录时间倒序排列，如果没有登录时间则按创建时间
	err := query.Order("COALESCE(login_time, created_at) DESC").
		Offset(offset).
		Limit(size).
		Find(&logs).Error

	return logs, total, err
}

// GetRecentLogins 获取最近的登录记录
func (r *userBehaviorLogRepository) GetRecentLogins(userID uint, hours int) ([]*model.UserBehaviorLog, error) {
	var logs []*model.UserBehaviorLog

	since := time.Now().Add(-time.Duration(hours) * time.Hour)

	err := r.db.Where("user_id = ? AND action IN (?) AND created_at >= ?",
		userID,
		[]string{model.ActionLogin, model.ActionLoginSuccess, model.ActionLoginFailed},
		since,
	).Order("created_at DESC").Find(&logs).Error

	return logs, err
}

// GetFailedLogins 获取失败的登录记录
func (r *userBehaviorLogRepository) GetFailedLogins(userID uint, hours int) ([]*model.UserBehaviorLog, error) {
	var logs []*model.UserBehaviorLog

	since := time.Now().Add(-time.Duration(hours) * time.Hour)

	err := r.db.Where("user_id = ? AND action = ? AND created_at >= ?",
		userID,
		model.ActionLoginFailed,
		since,
	).Order("created_at DESC").Find(&logs).Error

	return logs, err
}

// GetLoginsByTimeRange 按时间范围查询登录记录
func (r *userBehaviorLogRepository) GetLoginsByTimeRange(
	userID uint,
	startTime, endTime time.Time,
) ([]*model.UserBehaviorLog, error) {
	var logs []*model.UserBehaviorLog

	err := r.db.Where("user_id = ? AND action IN (?) AND created_at BETWEEN ? AND ?",
		userID,
		[]string{model.ActionLogin, model.ActionLoginSuccess, model.ActionLoginFailed},
		startTime,
		endTime,
	).Order("created_at DESC").Find(&logs).Error

	return logs, err
}

// GetLoginsByDate 按日期查询登录记录
func (r *userBehaviorLogRepository) GetLoginsByDate(userID uint, date time.Time) ([]*model.UserBehaviorLog, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	return r.GetLoginsByTimeRange(userID, startOfDay, endOfDay)
}

// CountUserLogins 统计用户登录次数
func (r *userBehaviorLogRepository) CountUserLogins(userID uint, days int) (int64, error) {
	var count int64

	since := time.Now().Add(-time.Duration(days) * 24 * time.Hour)

	err := r.db.Model(&model.UserBehaviorLog{}).
		Where("user_id = ? AND action IN (?) AND created_at >= ?",
			userID,
			[]string{model.ActionLogin, model.ActionLoginSuccess, model.ActionLoginFailed},
			since,
		).Count(&count).Error

	return count, err
}

// CountFailedLogins 统计失败登录次数
func (r *userBehaviorLogRepository) CountFailedLogins(userID uint, days int) (int64, error) {
	var count int64

	since := time.Now().Add(-time.Duration(days) * 24 * time.Hour)

	err := r.db.Model(&model.UserBehaviorLog{}).
		Where("user_id = ? AND action = ? AND created_at >= ?",
			userID,
			model.ActionLoginFailed,
			since,
		).Count(&count).Error

	return count, err
}

// CountLoginsByType 按登录类型统计
func (r *userBehaviorLogRepository) CountLoginsByType(userID uint, loginType string, days int) (int64, error) {
	var count int64

	since := time.Now().Add(-time.Duration(days) * 24 * time.Hour)

	err := r.db.Model(&model.UserBehaviorLog{}).
		Where("user_id = ? AND action IN (?) AND resource LIKE ? AND created_at >= ?",
			userID,
			[]string{model.ActionLogin, model.ActionLoginSuccess, model.ActionLoginFailed},
			loginType+"%",
			since,
		).Count(&count).Error

	return count, err
}
