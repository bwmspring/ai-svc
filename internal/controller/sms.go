package controller

import (
	"ai-svc/internal/middleware"
	"ai-svc/internal/model"
	"ai-svc/internal/service"
	"ai-svc/pkg/response"
	"ai-svc/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// SMSController 短信控制器.
type SMSController struct {
	smsService service.SMSService
	validator  *validator.Validate
}

// NewSMSController 创建短信控制器实例.
func NewSMSController(smsService service.SMSService) *SMSController {
	return &SMSController{
		smsService: smsService,
		validator:  validator.New(),
	}
}

// SendSMS 发送短信验证码.
func (ctrl *SMSController) SendSMS(c *gin.Context) {
	var req model.SendSMSRequest
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
	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// 检查是否需要用户登录
	var userID *uint
	if model.IsHighSecurityPurpose(req.Purpose) {
		// 高安全级别操作需要用户登录
		currentUserID := middleware.GetCurrentUserID(c)
		if currentUserID == 0 {
			response.Error(c, response.UNAUTHORIZED, "高安全级别操作需要用户登录")
			return
		}
		userID = &currentUserID
	}

	err := ctrl.smsService.SendVerificationCode(&req, clientIP, userAgent, userID)
	if err != nil {
		// 在错误响应中脱敏手机号
		maskedPhone := utils.MaskPhone(req.Phone)
		response.Error(c, response.ERROR, "发送验证码失败，手机号: "+maskedPhone)
		return
	}

	// 在成功响应中脱敏手机号
	maskedPhone := utils.MaskPhone(req.Phone)
	response.SuccessWithMessage(c, "短信验证码已发送到 "+maskedPhone, nil)
}

// ValidateSMS 验证短信验证码
func (ctrl *SMSController) ValidateSMS(c *gin.Context) {
	var req model.ValidateSMSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.INVALID_PARAMS, "参数错误: "+err.Error())
		return
	}

	// 参数验证
	if err := ctrl.validator.Struct(&req); err != nil {
		response.Error(c, response.INVALID_PARAMS, "参数验证失败: "+err.Error())
		return
	}

	// 检查是否需要用户登录
	if model.IsHighSecurityPurpose(req.Purpose) {
		userID := middleware.GetCurrentUserID(c)
		if userID == 0 {
			response.Error(c, response.UNAUTHORIZED, "高安全级别操作需要用户登录")
			return
		}

		err := ctrl.smsService.ValidateVerificationCodeWithUser(req.Phone, req.Code, req.Purpose, req.Token, userID)
		if err != nil {
			response.Error(c, response.ERROR, err.Error())
			return
		}
	} else {
		err := ctrl.smsService.ValidateVerificationCode(req.Phone, req.Code, req.Purpose, req.Token)
		if err != nil {
			response.Error(c, response.ERROR, err.Error())
			return
		}
	}

	response.SuccessWithMessage(c, "验证码验证成功", nil)
}
