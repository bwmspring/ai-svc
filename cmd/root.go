// Package cmd 提供命令行接口定义
// 使用 Cobra 框架构建清晰的命令行结构，支持多个子命令和丰富的参数配置
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// cfgFile 配置文件路径，可通过命令行参数 --config 指定.
	cfgFile string

	// verbose 详细输出模式，可通过 --verbose 或 -v 启用.
	verbose bool
)

// 当执行二进制文件但没有指定任何子命令时，会执行此命令.
var rootCmd = &cobra.Command{
	Use:   "ai-svc",
	Short: "AI服务 - 智能化的微服务应用",
	Long:  ``,

	// 当根命令执行时的回调函数
	// 如果用户只运行 ./ai-svc 而不带任何子命令，则显示帮助信息
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("欢迎使用 AI服务！")
		fmt.Println("使用 'ai-svc help' 查看可用命令")
		fmt.Println("使用 'ai-svc server' 启动服务器")
	},
}

// 如果执行过程中出现错误，程序将退出并返回错误代码.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "执行命令时发生错误: %v\n", err)
		os.Exit(1)
	}
}

// 用于初始化命令行参数和配置.
func init() {
	// 在 Cobra 初始化时调用配置初始化函数
	cobra.OnInitialize(initConfig)

	// 定义全局标志（flags），这些标志可以在所有子命令中使用

	// --config 标志：指定配置文件路径
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"配置文件路径 (默认查找 ./configs/config.yaml)")

	// --verbose 标志：启用详细输出模式
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false,
		"启用详细输出模式")

	// 将命令行标志绑定到 Viper 配置管理器
	// 这样可以通过 viper.GetBool("verbose") 来访问标志值
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
}

// 该函数会在每个命令执行前被调用，用于加载配置文件.
func initConfig() {
	if cfgFile != "" {
		// 如果用户指定了配置文件路径，则使用指定的文件
		viper.SetConfigFile(cfgFile)
	} else {
		// 如果没有指定配置文件，则按以下顺序查找：
		// 1. 当前目录下的 configs 文件夹
		// 2. 当前目录
		// 3. $HOME/.ai-svc 目录
		viper.AddConfigPath("./configs")
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME/.ai-svc")

		// 设置配置文件名（不包含扩展名）
		viper.SetConfigName("config")

		// 支持的配置文件类型
		viper.SetConfigType("yaml")
	}

	// 启用环境变量读取
	// 环境变量会覆盖配置文件中的同名配置项
	// 例如：AI_SVC_SERVER_PORT 环境变量会覆盖 server.port 配置
	viper.SetEnvPrefix("AI_SVC")
	viper.AutomaticEnv()

	// 尝试读取配置文件
	if err := viper.ReadInConfig(); err == nil {
		// 如果成功读取配置文件，且启用了详细模式，则输出配置文件路径
		if verbose {
			fmt.Printf("使用配置文件: %s\n", viper.ConfigFileUsed())
		}
	}
}
