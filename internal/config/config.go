package config

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/viper"
)

// Config 应用配置结构
type Config struct {
	Server    ServerConfig          `mapstructure:"server"`
	Database  DatabaseConfig        `mapstructure:"database"`
	Redis     RedisConfig           `mapstructure:"redis"`
	JWT       JWTConfig             `mapstructure:"jwt"`
	SMS       SMSConfig             `mapstructure:"sms"`
	RateLimit GlobalRateLimitConfig `mapstructure:"rate_limit"`
	Logger    LoggerConfig          `mapstructure:"logger"`
	Device    DeviceConfig          `mapstructure:"device"`
	AI        AIConfig              `mapstructure:"ai"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host           string        `mapstructure:"host"`
	Port           int           `mapstructure:"port"`
	Mode           string        `mapstructure:"mode"`
	ReadTimeout    time.Duration `mapstructure:"read_timeout"`
	WriteTimeout   time.Duration `mapstructure:"write_timeout"`
	MaxHeaderBytes int           `mapstructure:"max_header_bytes"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	Database        string `mapstructure:"database"`
	Charset         string `mapstructure:"charset"`
	ParseTime       bool   `mapstructure:"parse_time"`
	Loc             string `mapstructure:"loc"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret             string `mapstructure:"secret"`
	ExpireHours        int    `mapstructure:"expire_hours"`
	RefreshExpireHours int    `mapstructure:"refresh_expire_hours"`
}

// GlobalRateLimitConfig 全局限流配置
type GlobalRateLimitConfig struct {
	SMS   RateLimitItemConfig `mapstructure:"sms"`
	API   RateLimitItemConfig `mapstructure:"api"`
	Login RateLimitItemConfig `mapstructure:"login"`
}

// RateLimitItemConfig 单个限流配置
type RateLimitItemConfig struct {
	Capacity       int           `mapstructure:"capacity"`
	RefillRate     int           `mapstructure:"refill_rate"`
	RefillInterval time.Duration `mapstructure:"refill_interval"`
	ErrorMessage   string        `mapstructure:"error_message"`
}

// LoggerConfig 日志配置
type LoggerConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

// DeviceConfig 设备管理配置
type DeviceConfig struct {
	Limits   DeviceLimits      `mapstructure:"limits"`
	Activity ActivityConfig    `mapstructure:"activity"`
	Kickout  KickoutConfig     `mapstructure:"kickout"`
	Cache    DeviceCacheConfig `mapstructure:"cache"`
}

// DeviceLimits 设备数量限制配置
type DeviceLimits struct {
	MobileDevices      int `mapstructure:"mobile_devices"`
	PCDevices          int `mapstructure:"pc_devices"`
	MiniprogramDevices int `mapstructure:"miniprogram_devices"`
	WebDevices         int `mapstructure:"web_devices"`
}

// ActivityConfig 设备活跃管理配置
type ActivityConfig struct {
	OnlineTimeoutMinutes int `mapstructure:"online_timeout_minutes"`
	CleanupIntervalHours int `mapstructure:"cleanup_interval_hours"`
}

// KickoutConfig 踢出策略配置
type KickoutConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Strategy string `mapstructure:"strategy"`
}

// DeviceCacheConfig 设备缓存配置
type DeviceCacheConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	KeyPrefix   string `mapstructure:"key_prefix"`
	ExpireHours int    `mapstructure:"expire_hours"`
}

var AppConfig *Config

// LoadConfig 加载配置文件
func LoadConfig(configPath string) error {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// 设置默认值
	setDefaults()

	// 读取环境变量
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("读取配置文件失败: %v", err)
		return err
	}

	AppConfig = &Config{}
	if err := viper.Unmarshal(AppConfig); err != nil {
		log.Printf("解析配置文件失败: %v", err)
		return err
	}

	return nil
}

// setDefaults 设置默认配置值
func setDefaults() {
	// 服务器默认配置
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("server.read_timeout", "60s")
	viper.SetDefault("server.write_timeout", "60s")
	viper.SetDefault("server.max_header_bytes", 1048576)

	// 数据库默认配置
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 3306)
	viper.SetDefault("database.charset", "utf8mb4")
	viper.SetDefault("database.parse_time", true)
	viper.SetDefault("database.loc", "Local")
	viper.SetDefault("database.max_idle_conns", 10)
	viper.SetDefault("database.max_open_conns", 100)
	viper.SetDefault("database.conn_max_lifetime", 3600)

	// Redis默认配置
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.pool_size", 10)
	viper.SetDefault("redis.min_idle_conns", 5)

	// JWT默认配置
	viper.SetDefault("jwt.secret", "your-secret-key")
	viper.SetDefault("jwt.expire_hours", 24)
	viper.SetDefault("jwt.refresh_expire_hours", 168)

	// 短信默认配置
	viper.SetDefault("sms.provider", "aliyun")

	// 限流默认配置
	viper.SetDefault("rate_limit.sms.capacity", 1)
	viper.SetDefault("rate_limit.sms.refill_rate", 1)
	viper.SetDefault("rate_limit.sms.refill_interval", "60s")
	viper.SetDefault("rate_limit.api.capacity", 100)
	viper.SetDefault("rate_limit.api.refill_rate", 10)
	viper.SetDefault("rate_limit.api.refill_interval", "1s")
	viper.SetDefault("rate_limit.login.capacity", 5)
	viper.SetDefault("rate_limit.login.refill_rate", 1)
	viper.SetDefault("rate_limit.login.refill_interval", "300s")

	// 日志默认配置
	viper.SetDefault("logger.level", "info")
	viper.SetDefault("logger.format", "json")
	viper.SetDefault("logger.output", "stdout")

	// 设备管理默认配置
	viper.SetDefault("device.max_devices", 5)
	viper.SetDefault("device.kick_strategy", "oldest")
	viper.SetDefault("device.session_timeout", "24h")
	viper.SetDefault("device.offline_timeout", "30m")
	viper.SetDefault("device.heartbeat_interval", "5m")
	viper.SetDefault("device.cleanup_interval", "1h")
	viper.SetDefault("device.expired_cleanup_enabled", true)
	viper.SetDefault("device.offline_cleanup_enabled", true)
	viper.SetDefault("device.device_fingerprint_enabled", true)
	viper.SetDefault("device.suspicious_login_detection", true)
	viper.SetDefault("device.max_login_attempts", 5)
	viper.SetDefault("device.login_lockout_duration", "15m")
	viper.SetDefault("device.kick_notification_enabled", true)
	viper.SetDefault("device.login_notification_enabled", true)
	viper.SetDefault("device.cache_enabled", true)
	viper.SetDefault("device.cache_ttl", "10m")
	viper.SetDefault("device.batch_update_enabled", true)
	viper.SetDefault("device.batch_size", 100)

	// AI服务默认配置
	viper.SetDefault("ai.default_provider", "openai")
	viper.SetDefault("ai.timeout", "30s")
	viper.SetDefault("ai.max_retries", 3)
	viper.SetDefault("ai.features.streaming", true)
	viper.SetDefault("ai.features.history.enabled", true)
	viper.SetDefault("ai.features.history.max_messages", 20)
	viper.SetDefault("ai.features.history.max_tokens", 8000)
	viper.SetDefault("ai.features.content_filter.enabled", true)
	viper.SetDefault("ai.features.usage_tracking.enabled", true)
	viper.SetDefault("ai.features.cache.enabled", true)
	viper.SetDefault("ai.features.cache.ttl", 3600)
}

// GetDSN 获取数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
		c.Username, c.Password, c.Host, c.Port, c.Database, c.Charset, c.ParseTime, c.Loc)
}

// GetRedisAddr 获取Redis地址
func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// GetServerAddr 获取服务器地址
func (c *ServerConfig) GetServerAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// GetJWTExpireDuration 获取JWT过期时间
func (c *JWTConfig) GetJWTExpireDuration() time.Duration {
	return time.Duration(c.ExpireHours) * time.Hour
}

// GetJWTRefreshExpireDuration 获取JWT刷新token过期时间
func (c *JWTConfig) GetJWTRefreshExpireDuration() time.Duration {
	return time.Duration(c.RefreshExpireHours) * time.Hour
}
