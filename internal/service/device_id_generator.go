package service

import (
	"ai-svc/internal/model"
	"ai-svc/internal/repository"
	"ai-svc/pkg/logger"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// DeviceIDGeneratorService 设备ID生成服务接口
type DeviceIDGeneratorService interface {
	// 生成设备ID
	GenerateDeviceID(deviceType string, userID uint, fingerprint string) (string, error)

	// 验证设备ID格式
	ValidateDeviceID(deviceID string) bool

	// 解析设备ID信息
	ParseDeviceID(deviceID string) (*DeviceIDInfo, error)

	// 生成设备指纹
	GenerateFingerprint(req *model.DeviceRegistrationRequest, clientIP string) string
}

// DeviceIDInfo 设备ID解析信息
type DeviceIDInfo struct {
	Prefix    string    `json:"prefix"`    // 前缀
	Type      string    `json:"type"`      // 设备类型
	Timestamp time.Time `json:"timestamp"` // 生成时间
	UserID    uint      `json:"user_id"`   // 用户ID
	Random    string    `json:"random"`    // 随机部分
	Checksum  string    `json:"checksum"`  // 校验和
}

// deviceIDGeneratorService 设备ID生成服务实现
type deviceIDGeneratorService struct {
	generators map[string]*model.DeviceIDGenerator // 不同设备类型的生成器配置
}

// NewDeviceIDGeneratorService 创建设备ID生成服务
func NewDeviceIDGeneratorService() DeviceIDGeneratorService {
	return &deviceIDGeneratorService{
		generators: map[string]*model.DeviceIDGenerator{
			model.DeviceTypeIOS: {
				Prefix:    "ios_",
				Length:    32,
				Timestamp: true,
				Random:    true,
				Checksum:  true,
			},
			model.DeviceTypeAndroid: {
				Prefix:    "and_",
				Length:    32,
				Timestamp: true,
				Random:    true,
				Checksum:  true,
			},
			model.DeviceTypePC: {
				Prefix:    "pc_",
				Length:    32,
				Timestamp: true,
				Random:    true,
				Checksum:  true,
			},
			model.DeviceTypeWeb: {
				Prefix:    "web_",
				Length:    32,
				Timestamp: true,
				Random:    true,
				Checksum:  true,
			},
			model.DeviceTypeMiniprogram: {
				Prefix:    "mp_",
				Length:    32,
				Timestamp: true,
				Random:    true,
				Checksum:  true,
			},
		},
	}
}

// GenerateDeviceID 生成设备ID
func (s *deviceIDGeneratorService) GenerateDeviceID(
	deviceType string,
	userID uint,
	fingerprint string,
) (string, error) {
	generator, exists := s.generators[deviceType]
	if !exists {
		return "", fmt.Errorf("不支持的设备类型: %s", deviceType)
	}

	var parts []string

	// 1. 添加前缀
	if generator.Prefix != "" {
		parts = append(parts, generator.Prefix)
	}

	// 2. 添加时间戳（精确到秒，10位）
	if generator.Timestamp {
		timestamp := time.Now().Unix()
		parts = append(parts, fmt.Sprintf("%x", timestamp)) // 十六进制时间戳
	}

	// 3. 添加用户ID（混淆）
	userIDHash := s.hashUserID(userID)
	parts = append(parts, userIDHash[:6]) // 取前6位

	// 4. 添加随机数
	if generator.Random {
		randomBytes := make([]byte, 4)
		if _, err := rand.Read(randomBytes); err != nil {
			return "", fmt.Errorf("生成随机数失败: %w", err)
		}
		parts = append(parts, hex.EncodeToString(randomBytes))
	}

	// 5. 添加指纹摘要
	if fingerprint != "" {
		fingerprintHash := sha256.Sum256([]byte(fingerprint))
		parts = append(parts, hex.EncodeToString(fingerprintHash[:4])) // 取前8字符
	}

	// 组合所有部分
	deviceID := strings.Join(parts, "")

	// 6. 添加校验和
	if generator.Checksum {
		checksum := s.calculateChecksum(deviceID)
		deviceID += checksum
	}

	// 7. 截断到指定长度
	if len(deviceID) > generator.Length {
		deviceID = deviceID[:generator.Length]
	}

	logger.Info("设备ID生成成功", map[string]any{
		"device_type": deviceType,
		"user_id":     userID,
		"device_id":   deviceID,
		"fingerprint": fingerprint[:16] + "...", // 只记录前16字符
	})

	return deviceID, nil
}

// ValidateDeviceID 验证设备ID格式
func (s *deviceIDGeneratorService) ValidateDeviceID(deviceID string) bool {
	if len(deviceID) < 16 || len(deviceID) > 64 {
		return false
	}

	// 检查是否包含有效前缀
	for _, generator := range s.generators {
		if strings.HasPrefix(deviceID, generator.Prefix) {
			return true
		}
	}

	return false
}

// ParseDeviceID 解析设备ID信息
func (s *deviceIDGeneratorService) ParseDeviceID(deviceID string) (*DeviceIDInfo, error) {
	if !s.ValidateDeviceID(deviceID) {
		return nil, fmt.Errorf("无效的设备ID格式")
	}

	info := &DeviceIDInfo{}

	// 识别设备类型和前缀
	for deviceType, generator := range s.generators {
		if strings.HasPrefix(deviceID, generator.Prefix) {
			info.Prefix = generator.Prefix
			info.Type = deviceType

			// 解析时间戳（如果存在）
			if generator.Timestamp {
				timestampStr := deviceID[len(generator.Prefix) : len(generator.Prefix)+8]
				if timestamp, err := strconv.ParseInt(timestampStr, 16, 64); err == nil {
					info.Timestamp = time.Unix(timestamp, 0)
				}
			}

			break
		}
	}

	if info.Type == "" {
		return nil, fmt.Errorf("无法识别设备类型")
	}

	return info, nil
}

// GenerateFingerprint 生成设备指纹
func (s *deviceIDGeneratorService) GenerateFingerprint(req *model.DeviceRegistrationRequest, clientIP string) string {
	// 组合设备特征信息
	features := []string{
		req.DeviceType,
		req.Platform,
		req.OSVersion,
		req.AppVersion,
		clientIP,
		req.ClientInfo,
	}

	// 添加时间因子（按天，减少指纹变化频率）
	dayFactor := time.Now().Format("2006-01-02")
	features = append(features, dayFactor)

	// 生成指纹
	combined := strings.Join(features, "|")
	hash := sha256.Sum256([]byte(combined))

	return hex.EncodeToString(hash[:])
}

// 私有方法

// hashUserID 对用户ID进行哈希混淆
func (s *deviceIDGeneratorService) hashUserID(userID uint) string {
	// 添加盐值防止彩虹表攻击
	salt := "device_id_salt_2024"
	data := fmt.Sprintf("%s_%d", salt, userID)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// calculateChecksum 计算校验和
func (s *deviceIDGeneratorService) calculateChecksum(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:2]) // 4字符校验和
}

// DeviceRegistrationService 设备注册服务接口
type DeviceRegistrationService interface {
	// 注册新设备
	RegisterDevice(
		userID uint,
		req *model.DeviceRegistrationRequest,
		clientIP, userAgent string,
	) (*model.DeviceRegistrationResponse, error)

	// 检查设备是否已注册
	CheckDeviceRegistration(fingerprint string) (*model.DeviceRegistrationResponse, error)

	// 更新设备指纹
	UpdateDeviceFingerprint(deviceID string, newFingerprint string) error

	// 获取设备注册信息
	GetDeviceRegistration(deviceID string) (*model.DeviceFingerprint, error)
}

// deviceRegistrationService 设备注册服务实现
type deviceRegistrationService struct {
	idGenerator DeviceIDGeneratorService
	deviceRepo  repository.DeviceRepository
}

// NewDeviceRegistrationService 创建设备注册服务
func NewDeviceRegistrationService(deviceRepo repository.DeviceRepository) DeviceRegistrationService {
	return &deviceRegistrationService{
		idGenerator: NewDeviceIDGeneratorService(),
		deviceRepo:  deviceRepo,
	}
}

// RegisterDevice 注册新设备
func (s *deviceRegistrationService) RegisterDevice(
	userID uint,
	req *model.DeviceRegistrationRequest,
	clientIP, userAgent string,
) (*model.DeviceRegistrationResponse, error) {
	// 1. 生成标准化设备指纹
	fingerprint := s.idGenerator.GenerateFingerprint(req, clientIP)

	// 2. 检查设备是否已注册
	if existing, err := s.CheckDeviceRegistration(fingerprint); err == nil && existing != nil {
		logger.Info("设备已注册，返回现有ID", map[string]any{
			"device_id":   existing.DeviceID,
			"fingerprint": fingerprint[:16] + "...",
			"user_id":     userID,
		})

		return &model.DeviceRegistrationResponse{
			DeviceID:  existing.DeviceID,
			IsNew:     false,
			ExpiresAt: existing.ExpiresAt,
		}, nil
	}

	// 3. 生成新的设备ID
	deviceID, err := s.idGenerator.GenerateDeviceID(req.DeviceType, userID, fingerprint)
	if err != nil {
		return nil, fmt.Errorf("生成设备ID失败: %w", err)
	}

	// 4. 创建设备记录
	device := &model.UserDevice{
		UserID:       userID,
		DeviceID:     deviceID,
		DeviceType:   req.DeviceType,
		DeviceName:   req.DeviceName,
		AppVersion:   req.AppVersion,
		OSVersion:    req.OSVersion,
		ClientIP:     clientIP,
		UserAgent:    userAgent,
		Status:       1,
		LoginAt:      time.Now(),
		LastActiveAt: time.Now(),
	}

	if err := s.deviceRepo.CreateDevice(device); err != nil {
		return nil, fmt.Errorf("创建设备记录失败: %w", err)
	}

	// TODO: 5. 创建指纹记录（需要实现指纹仓储）
	// 将来可以添加指纹记录功能来检测重复注册

	logger.Info("新设备注册成功", map[string]any{
		"user_id":     userID,
		"device_id":   deviceID,
		"device_type": req.DeviceType,
		"fingerprint": fingerprint[:16] + "...",
	})

	return &model.DeviceRegistrationResponse{
		DeviceID:  deviceID,
		IsNew:     true,
		ExpiresAt: time.Now().AddDate(1, 0, 0), // 1年有效期
	}, nil
}

// CheckDeviceRegistration 检查设备是否已注册
func (s *deviceRegistrationService) CheckDeviceRegistration(
	fingerprint string,
) (*model.DeviceRegistrationResponse, error) {
	// 这里需要查询指纹数据库
	// 暂时返回nil表示未注册
	return nil, fmt.Errorf("设备未注册")
}

// UpdateDeviceFingerprint 更新设备指纹
func (s *deviceRegistrationService) UpdateDeviceFingerprint(deviceID string, newFingerprint string) error {
	// 实现指纹更新逻辑
	return fmt.Errorf("功能待实现")
}

// GetDeviceRegistration 获取设备注册信息
func (s *deviceRegistrationService) GetDeviceRegistration(deviceID string) (*model.DeviceFingerprint, error) {
	// 实现获取注册信息逻辑
	return nil, fmt.Errorf("功能待实现")
}
