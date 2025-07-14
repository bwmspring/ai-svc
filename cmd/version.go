package cmd

import (
	"fmt"
	"runtime"
	"time"

	"github.com/spf13/cobra"
)

// 版本信息常量
// 这些变量通常在编译时通过 -ldflags 参数注入
var (
	// Version 应用程序版本号
	// 在生产环境中，这个值会在编译时通过 go build -ldflags 注入
	Version = "dev"

	// GitCommit Git 提交哈希
	GitCommit = "unknown"

	// BuildTime 编译时间
	BuildTime = "unknown"

	// GoVersion Go 语言版本
	GoVersion = runtime.Version()

	// Platform 编译平台信息
	Platform = fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
)

// versionCmd 定义版本信息显示命令
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "显示版本信息",
	Long: `显示 AI 服务的详细版本信息。

包含以下信息：
• 应用程序版本号
• Git 提交哈希值
• 编译时间
• Go 语言版本
• 编译平台信息

这些信息对于问题排查和部署管理非常有用。`,

	// Run 函数执行版本信息显示逻辑
	Run: func(cmd *cobra.Command, args []string) {
		showVersion()
	},
}

// versionDetailCmd 定义详细版本信息命令
var versionDetailCmd = &cobra.Command{
	Use:   "detail",
	Short: "显示详细版本信息",
	Long:  `显示更详细的版本和构建信息，包括运行时环境信息。`,
	Run: func(cmd *cobra.Command, args []string) {
		showDetailedVersion()
	},
}

// init 初始化版本相关命令
func init() {
	// 将版本命令添加到根命令
	rootCmd.AddCommand(versionCmd)

	// 将详细版本命令添加到版本命令下
	versionCmd.AddCommand(versionDetailCmd)
}

// showVersion 显示基本版本信息
func showVersion() {
	fmt.Printf("AI 服务 (ai-svc)\n")
	fmt.Printf("版本: %s\n", Version)

	if GitCommit != "unknown" {
		fmt.Printf("Git 提交: %s\n", GitCommit)
	}

	if BuildTime != "unknown" {
		fmt.Printf("构建时间: %s\n", BuildTime)
	}

	fmt.Printf("Go 版本: %s\n", GoVersion)
	fmt.Printf("平台: %s\n", Platform)
}

// showDetailedVersion 显示详细版本信息
func showDetailedVersion() {
	// 基本版本信息
	showVersion()

	fmt.Println("\n=== 详细信息 ===")

	// 运行时信息
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	fmt.Printf("运行时信息:\n")
	fmt.Printf("  Go 版本: %s\n", runtime.Version())
	fmt.Printf("  编译器: %s\n", runtime.Compiler)
	fmt.Printf("  CPU 核心数: %d\n", runtime.NumCPU())
	fmt.Printf("  Goroutine 数量: %d\n", runtime.NumGoroutine())

	// 内存信息（单位：KB）
	fmt.Printf("内存信息:\n")
	fmt.Printf("  已分配内存: %d KB\n", bToKb(m.Alloc))
	fmt.Printf("  总分配内存: %d KB\n", bToKb(m.TotalAlloc))
	fmt.Printf("  系统内存: %d KB\n", bToKb(m.Sys))
	fmt.Printf("  GC 次数: %d\n", m.NumGC)

	// 编译信息
	fmt.Printf("编译信息:\n")
	fmt.Printf("  操作系统: %s\n", runtime.GOOS)
	fmt.Printf("  架构: %s\n", runtime.GOARCH)

	// 当前时间
	fmt.Printf("当前时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))

	// 如果是开发版本，显示提示
	if Version == "dev" {
		fmt.Println("\n⚠️  这是开发版本，不建议在生产环境使用")
	}
}

// bToKb 将字节转换为千字节
func bToKb(b uint64) uint64 {
	return b / 1024
}

// GetVersionInfo 返回版本信息的结构化数据
// 这个函数可以被其他包调用，获取版本信息用于日志记录或 API 响应
func GetVersionInfo() map[string]string {
	return map[string]string{
		"version":    Version,
		"git_commit": GitCommit,
		"build_time": BuildTime,
		"go_version": GoVersion,
		"platform":   Platform,
	}
}
