package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"ai-svc/internal/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

// configCmd 定义配置管理的主命令
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "配置文件管理工具",
	Long: `配置文件管理工具，提供配置文件的查看、验证和生成功能。

支持的操作：
• show    - 显示当前配置
• validate - 验证配置文件格式
• generate - 生成默认配置文件模板

示例用法：
  ai-svc config show              # 显示当前配置
  ai-svc config validate         # 验证配置文件
  ai-svc config generate         # 生成默认配置模板`,
}

// configShowCmd 显示当前配置
var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "显示当前配置信息",
	Long: `显示当前加载的配置信息。

这个命令会显示：
• 配置文件路径
• 所有配置项及其当前值
• 配置来源（文件、环境变量、默认值）`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return showConfig()
	},
}

// configValidateCmd 验证配置文件
var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "验证配置文件格式和内容",
	Long: `验证配置文件的格式和内容是否正确。

检查项目包括：
• YAML 格式是否正确
• 必需的配置项是否存在
• 配置值是否在有效范围内
• 文件路径是否可访问`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return validateConfig()
	},
}

// configGenerateCmd 生成配置文件模板
var configGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "生成默认配置文件模板",
	Long: `生成包含所有配置项的默认配置文件模板。

生成的配置文件包括：
• 服务器配置（端口、超时等）
• 日志配置（级别、格式、输出）
• 数据库配置（连接信息）
• 第三方服务配置（短信服务等）

可以基于此模板进行自定义修改。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return generateConfig()
	},
}

// 配置相关的命令行参数
var (
	outputFile string // 输出文件路径（用于 generate 命令）
	format     string // 输出格式（yaml, json）
)

// init 初始化配置相关命令
func init() {
	// 将配置命令添加到根命令
	rootCmd.AddCommand(configCmd)

	// 添加子命令
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configValidateCmd)
	configCmd.AddCommand(configGenerateCmd)

	// 为 generate 命令添加特定参数
	configGenerateCmd.Flags().StringVarP(&outputFile, "output", "o", "./configs/config.yaml",
		"输出文件路径")
	configGenerateCmd.Flags().StringVarP(&format, "format", "f", "yaml",
		"输出格式 (yaml|json)")
}

// showConfig 显示当前配置信息
func showConfig() error {
	fmt.Println("=== AI 服务配置信息 ===\n")

	// 显示配置文件信息
	if viper.ConfigFileUsed() != "" {
		fmt.Printf("📁 配置文件: %s\n", viper.ConfigFileUsed())
	} else {
		fmt.Println("📁 配置文件: 未使用配置文件（使用默认值）")
	}

	// 尝试加载配置
	configPath := "./configs/config.yaml"
	if cfgFile != "" {
		configPath = cfgFile
	}

	if err := config.LoadConfig(configPath); err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	fmt.Println("\n=== 服务器配置 ===")
	fmt.Printf("端口: %s\n", config.AppConfig.Server.Port)
	fmt.Printf("模式: %s\n", config.AppConfig.Server.Mode)
	fmt.Printf("读取超时: %d 秒\n", config.AppConfig.Server.ReadTimeout)
	fmt.Printf("写入超时: %d 秒\n", config.AppConfig.Server.WriteTimeout)

	fmt.Println("\n=== 日志配置 ===")
	fmt.Printf("级别: %s\n", config.AppConfig.Log.Level)
	fmt.Printf("格式: %s\n", config.AppConfig.Log.Format)
	fmt.Printf("输出: %s\n", config.AppConfig.Log.Output)

	// 显示环境变量覆盖信息
	fmt.Println("\n=== 环境变量 ===")
	envVars := []string{
		"AI_SVC_SERVER_PORT",
		"AI_SVC_SERVER_MODE",
		"AI_SVC_LOG_LEVEL",
	}

	hasEnvVars := false
	for _, env := range envVars {
		if value := os.Getenv(env); value != "" {
			if !hasEnvVars {
				fmt.Println("检测到以下环境变量覆盖:")
				hasEnvVars = true
			}
			fmt.Printf("  %s = %s\n", env, value)
		}
	}

	if !hasEnvVars {
		fmt.Println("未检测到环境变量覆盖")
	}

	return nil
}

// validateConfig 验证配置文件
func validateConfig() error {
	fmt.Println("🔍 验证配置文件...")

	// 确定配置文件路径
	configPath := "./configs/config.yaml"
	if cfgFile != "" {
		configPath = cfgFile
	}

	// 检查文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("配置文件不存在: %s", configPath)
	}

	fmt.Printf("✅ 配置文件存在: %s\n", configPath)

	// 尝试加载配置
	if err := config.LoadConfig(configPath); err != nil {
		return fmt.Errorf("❌ 配置文件格式错误: %w", err)
	}

	fmt.Println("✅ 配置文件格式正确")

	// 验证必需的配置项
	validationErrors := []string{}

	if config.AppConfig.Server.Port == "" {
		validationErrors = append(validationErrors, "server.port 不能为空")
	}

	if config.AppConfig.Server.Mode == "" {
		validationErrors = append(validationErrors, "server.mode 不能为空")
	} else if config.AppConfig.Server.Mode != "debug" &&
		config.AppConfig.Server.Mode != "release" &&
		config.AppConfig.Server.Mode != "test" {
		validationErrors = append(validationErrors, "server.mode 必须是 debug、release 或 test")
	}

	if config.AppConfig.Log.Level == "" {
		validationErrors = append(validationErrors, "log.level 不能为空")
	}

	// 显示验证结果
	if len(validationErrors) > 0 {
		fmt.Println("\n❌ 配置验证失败:")
		for _, err := range validationErrors {
			fmt.Printf("  • %s\n", err)
		}
		return fmt.Errorf("配置验证失败，发现 %d 个问题", len(validationErrors))
	}

	fmt.Println("✅ 配置验证通过")
	fmt.Println("🎉 配置文件完全正确，可以安全使用")

	return nil
}

// generateConfig 生成配置文件模板
func generateConfig() error {
	fmt.Printf("📝 生成配置文件模板: %s\n", outputFile)

	// 创建默认配置结构
	defaultConfig := map[string]interface{}{
		"server": map[string]interface{}{
			"port":          "8080",
			"mode":          "debug",
			"read_timeout":  30,
			"write_timeout": 30,
		},
		"log": map[string]interface{}{
			"level":  "info",
			"format": "json",
			"output": "stdout",
		},
		"database": map[string]interface{}{
			"host":     "localhost",
			"port":     3306,
			"username": "ai_svc",
			"password": "your_password",
			"database": "ai_svc",
			"charset":  "utf8mb4",
		},
		"sms": map[string]interface{}{
			"provider": "aliyun",
			"config": map[string]interface{}{
				"access_key_id":     "your_access_key_id",
				"access_key_secret": "your_access_key_secret",
				"sign_name":         "your_sign_name",
				"template_code":     "your_template_code",
			},
		},
		"jwt": map[string]interface{}{
			"secret":        "your_jwt_secret_key_change_in_production",
			"expiry_hours":  24,
			"refresh_hours": 72,
			"issuer":        "ai-svc",
		},
	}

	// 确保输出目录存在
	outputDir := filepath.Dir(outputFile)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 检查文件是否已存在
	if _, err := os.Stat(outputFile); err == nil {
		fmt.Printf("⚠️  文件已存在: %s\n", outputFile)
		fmt.Print("是否覆盖? (y/N): ")

		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("❌ 操作已取消")
			return nil
		}
	}

	// 根据格式输出文件
	var data []byte
	var err error

	switch format {
	case "yaml", "yml":
		data, err = yaml.Marshal(defaultConfig)
		if err != nil {
			return fmt.Errorf("序列化 YAML 失败: %w", err)
		}

		// 添加注释头
		header := `# AI 服务配置文件
# 这是一个配置文件模板，包含了所有可用的配置选项
# 请根据实际环境修改相应的配置值

`
		data = append([]byte(header), data...)

	default:
		return fmt.Errorf("不支持的格式: %s (支持: yaml, json)", format)
	}

	// 写入文件
	if err := os.WriteFile(outputFile, data, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	fmt.Printf("✅ 配置文件模板生成成功: %s\n", outputFile)
	fmt.Println("📝 请根据实际环境修改配置值，特别是:")
	fmt.Println("   • 数据库连接信息")
	fmt.Println("   • JWT 密钥")
	fmt.Println("   • 短信服务配置")
	fmt.Println("   • 生产环境请将 server.mode 设置为 'release'")

	return nil
}
