package controller

import (
	"ai-svc/internal/middleware"
	"ai-svc/internal/model"
	"ai-svc/internal/service"
	"ai-svc/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// UserController 用户控制器.
type UserController struct {
	userService service.UserService
	smsService  service.SMSService
	validator   *validator.Validate
}

// NewUserController 创建用户控制器实例.
func NewUserController(userService service.UserService, smsService service.SMSService) *UserController {
	return &UserController{
		userService: userService,
		smsService:  smsService,
		validator:   validator.New(),
	}
}

// LoginWithSMS 手机号+验证码登录（同时完成注册）.
func (ctrl *UserController) LoginWithSMS(c *gin.Context) {
	var req model.LoginWithSMSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.INVALID_PARAMS, "参数错误: "+err.Error())
		return
	}

	// 参数验证
	if err := ctrl.validator.Struct(&req); err != nil {
		response.Error(c, response.INVALID_PARAMS, "参数验证失败: "+err.Error())
		return
	}

	// 获取客户端IP和User-Agent
	ip := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	loginResp, isNewUser, err := ctrl.userService.LoginWithSMS(&req, ip, userAgent)
	if err != nil {
		response.Error(c, response.UNAUTHORIZED, err.Error())
		return
	}

	message := "登录成功"
	if isNewUser {
		message = "注册并登录成功"
	}

	response.SuccessWithMessage(c, message, loginResp)
}

// GetProfile 获取用户信息.
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

// UpdateProfile 更新用户信息.
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

// GetUserList 获取用户列表.
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

// SearchUsers 搜索用户.
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

// GetUserByID 根据ID获取用户.
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

// DeleteUser 删除用户.
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

// GetUserDevices 获取用户设备列表.
func (ctrl *UserController) GetUserDevices(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	if userID == 0 {
		response.Error(c, response.UNAUTHORIZED, "未授权")
		return
	}

	devices, err := ctrl.userService.GetUserDevices(userID)
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}

	response.Success(c, devices)
}

// KickDevices 踢出设备.
func (ctrl *UserController) KickDevices(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	if userID == 0 {
		response.Error(c, response.UNAUTHORIZED, "未授权")
		return
	}

	var req model.KickDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.INVALID_PARAMS, "参数错误: "+err.Error())
		return
	}

	// 参数验证
	if err := ctrl.validator.Struct(&req); err != nil {
		response.Error(c, response.INVALID_PARAMS, "参数验证失败: "+err.Error())
		return
	}

	if err := ctrl.userService.KickDevices(userID, &req); err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}

	response.SuccessWithMessage(c, "设备已被踢出", nil)
}

// RefreshToken 刷新Token.
func (ctrl *UserController) RefreshToken(c *gin.Context) {
	var req model.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.INVALID_PARAMS, "参数错误: "+err.Error())
		return
	}

	// 参数验证
	if err := ctrl.validator.Struct(&req); err != nil {
		response.Error(c, response.INVALID_PARAMS, "参数验证失败: "+err.Error())
		return
	}

	// 通过用户服务刷新token（包含设备验证）
	tokenPair, err := ctrl.userService.RefreshToken(req.RefreshToken)
	if err != nil {
		response.Error(c, response.UNAUTHORIZED, err.Error())
		return
	}

	response.SuccessWithMessage(c, "Token刷新成功", tokenPair)
}
