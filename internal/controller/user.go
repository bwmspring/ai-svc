package controller

import (
	"strconv"

	"ai-svc/internal/middleware"
	"ai-svc/internal/model"
	"ai-svc/internal/service"
	"ai-svc/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// UserController 用户控制器
type UserController struct {
	userService service.UserService
	validator   *validator.Validate
}

// NewUserController 创建用户控制器实例
func NewUserController(userService service.UserService) *UserController {
	return &UserController{
		userService: userService,
		validator:   validator.New(),
	}
}

// Register 用户注册
func (ctrl *UserController) Register(c *gin.Context) {
	var req model.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.INVALID_PARAMS, "参数错误: "+err.Error())
		return
	}

	// 参数验证
	if err := ctrl.validator.Struct(&req); err != nil {
		response.Error(c, response.INVALID_PARAMS, "参数验证失败: "+err.Error())
		return
	}

	user, err := ctrl.userService.CreateUser(&req)
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}

	response.SuccessWithMessage(c, "注册成功", user)
}

// Login 用户登录
func (ctrl *UserController) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.INVALID_PARAMS, "参数错误: "+err.Error())
		return
	}

	// 参数验证
	if err := ctrl.validator.Struct(&req); err != nil {
		response.Error(c, response.INVALID_PARAMS, "参数验证失败: "+err.Error())
		return
	}

	// 获取客户端IP
	ip := c.ClientIP()

	user, err := ctrl.userService.Login(&req, ip)
	if err != nil {
		response.Error(c, response.UNAUTHORIZED, err.Error())
		return
	}

	// 生成JWT令牌
	token, err := middleware.GenerateToken(user.ID, user.Username)
	if err != nil {
		response.Error(c, response.ERROR, "生成令牌失败")
		return
	}

	response.SuccessWithMessage(c, "登录成功", gin.H{
		"user":  user,
		"token": token,
	})
}

// GetProfile 获取用户信息
func (ctrl *UserController) GetProfile(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	if userID == 0 {
		response.Error(c, response.UNAUTHORIZED, "未授权")
		return
	}

	user, err := ctrl.userService.GetUserByID(userID)
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}

	response.Success(c, user)
}

// UpdateProfile 更新用户信息
func (ctrl *UserController) UpdateProfile(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	if userID == 0 {
		response.Error(c, response.UNAUTHORIZED, "未授权")
		return
	}

	var req model.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.INVALID_PARAMS, "参数错误: "+err.Error())
		return
	}

	// 参数验证
	if err := ctrl.validator.Struct(&req); err != nil {
		response.Error(c, response.INVALID_PARAMS, "参数验证失败: "+err.Error())
		return
	}

	user, err := ctrl.userService.UpdateUser(userID, &req)
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}

	response.SuccessWithMessage(c, "更新成功", user)
}

// ChangePassword 修改密码
func (ctrl *UserController) ChangePassword(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	if userID == 0 {
		response.Error(c, response.UNAUTHORIZED, "未授权")
		return
	}

	var req model.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.INVALID_PARAMS, "参数错误: "+err.Error())
		return
	}

	// 参数验证
	if err := ctrl.validator.Struct(&req); err != nil {
		response.Error(c, response.INVALID_PARAMS, "参数验证失败: "+err.Error())
		return
	}

	if err := ctrl.userService.ChangePassword(userID, &req); err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}

	response.SuccessWithMessage(c, "密码修改成功", nil)
}

// GetUserList 获取用户列表
func (ctrl *UserController) GetUserList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	users, total, err := ctrl.userService.GetUserList(page, size)
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}

	response.Page(c, users, total, page, size)
}

// SearchUsers 搜索用户
func (ctrl *UserController) SearchUsers(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		response.Error(c, response.INVALID_PARAMS, "搜索关键词不能为空")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	users, total, err := ctrl.userService.SearchUsers(keyword, page, size)
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}

	response.Page(c, users, total, page, size)
}

// GetUserByID 根据ID获取用户
func (ctrl *UserController) GetUserByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, response.INVALID_PARAMS, "无效的用户ID")
		return
	}

	user, err := ctrl.userService.GetUserByID(uint(id))
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}

	response.Success(c, user)
}

// DeleteUser 删除用户
func (ctrl *UserController) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, response.INVALID_PARAMS, "无效的用户ID")
		return
	}

	if err := ctrl.userService.DeleteUser(uint(id)); err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}

	response.SuccessWithMessage(c, "删除成功", nil)
}
