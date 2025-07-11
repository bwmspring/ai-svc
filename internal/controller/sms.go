package controller

import (
	"ai-svc/internal/model"
	"ai-svc/internal/service"
	"ai-svc/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// SMSController 短信控制器
type SMSController struct {
	smsService service.SMSService
	validator  *validator.Validate
}

// NewSMSController 创建短信控制器实例
func NewSMSController(smsService service.SMSService) *SMSController {
	return &SMSController{
		smsService: smsService,
		validator:  validator.New(),
	}
}

// SendSMS 发送短信验证码
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

	// 获取客户端IP
	clientIP := c.ClientIP()

	err := ctrl.smsService.SendVerificationCode(&req, clientIP)
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}

	response.SuccessWithMessage(c, "短信验证码发送成功", nil)
}
