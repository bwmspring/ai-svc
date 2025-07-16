// Package main 是 AI 服务的入口点
// 使用 Cobra 框架提供清晰的命令行接口和丰富的功能
package main

import "ai-svc/cmd"

// 所有的复杂逻辑都被封装在 cmd 包中，保持 main.go 的简洁性.
func main() {
	// 执行 Cobra 根命令
	// 这会解析命令行参数并路由到相应的子命令
	// 如果执行过程中出现错误，程序会自动退出并显示错误信息
	cmd.Execute()
}
