package controller

import (
	"ai-svc/internal/service"
	"ai-svc/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// DeviceController 设备管理控制器
type DeviceController struct {
	deviceService service.DeviceService
	validator     *validator.Validate
}

// NewDeviceController 创建设备管理控制器实例
func NewDeviceController(deviceService service.DeviceService) *DeviceController {
	return &DeviceController{
		deviceService: deviceService,
		validator:     validator.New(),
	}
}

// GetMyDevices 获取我的设备列表
func (ctrl *DeviceController) GetMyDevices(c *gin.Context) {
	// 从JWT中获取用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, response.UNAUTHORIZED, "用户未登录")
		return
	}

	// 获取当前设备ID（如果有的话）
	currentDeviceID := c.GetHeader("X-Device-ID")

	// 获取设备列表
	deviceList, err := ctrl.deviceService.GetUserDevices(userID.(uint), currentDeviceID)
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}

	response.Success(c, deviceList)
}

// KickDevicesRequest 踢出设备请求
type KickDevicesRequest struct {
	DeviceIDs []string `json:"device_ids" validate:"required,min=1"`
}

// KickDevices 踢出指定设备
func (ctrl *DeviceController) KickDevices(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, response.UNAUTHORIZED, "用户未登录")
		return
	}

	var req KickDevicesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.INVALID_PARAMS, "请求参数错误")
		return
	}

	if err := ctrl.validator.Struct(&req); err != nil {
		response.Error(c, response.INVALID_PARAMS, "参数验证失败")
		return
	}

	// 踢出设备
	if err := ctrl.deviceService.KickDevices(userID.(uint), req.DeviceIDs); err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}

	response.Success(c, gin.H{
		"message": "设备已被踢出",
		"count":   len(req.DeviceIDs),
	})
}

// KickOtherDevices 踢出其他所有设备（保留当前设备）
func (ctrl *DeviceController) KickOtherDevices(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, response.UNAUTHORIZED, "用户未登录")
		return
	}

	// 获取当前设备ID
	currentDeviceID := c.GetHeader("X-Device-ID")
	if currentDeviceID == "" {
		response.Error(c, response.INVALID_PARAMS, "缺少设备ID")
		return
	}

	// 踢出其他设备
	if err := ctrl.deviceService.KickOtherDevices(userID.(uint), currentDeviceID); err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}

	response.Success(c, gin.H{
		"message": "其他设备已被踢出",
	})
}

// UpdateHeartbeat 更新设备活跃状态
func (ctrl *DeviceController) UpdateHeartbeat(c *gin.Context) {
	// 获取设备ID
	deviceID := c.GetHeader("X-Device-ID")
	if deviceID == "" {
		response.Error(c, response.INVALID_PARAMS, "缺少设备ID")
		return
	}

	// 更新设备活跃时间
	if err := ctrl.deviceService.UpdateDeviceActivity(deviceID); err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}

	response.Success(c, gin.H{
		"message": "设备活跃状态已更新",
	})
}

// GetDeviceStats 获取设备统计信息
func (ctrl *DeviceController) GetDeviceStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, response.UNAUTHORIZED, "用户未登录")
		return
	}

	currentDeviceID := c.GetHeader("X-Device-ID")

	// 获取设备列表
	deviceList, err := ctrl.deviceService.GetUserDevices(userID.(uint), currentDeviceID)
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}

	// 返回统计信息
	response.Success(c, gin.H{
		"summary": deviceList.Summary,
		"limits":  deviceList.Limits,
	})
}
