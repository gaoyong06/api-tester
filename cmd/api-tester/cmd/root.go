package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// 全局标志
	cfgFile    string
	verbose    bool
	outputDir  string
	reportType string
)

// rootCmd 表示基础命令
var rootCmd = &cobra.Command{
	Use:   "api-tester",
	Short: "API测试工具 - 自动化API测试和验证",
	Long: `API测试工具是一个强大的命令行工具，用于自动化API测试和验证。
基于OpenAPI/Swagger规范，它可以自动生成测试、验证响应，并生成详细报告。

支持的功能包括：
- 运行API测试
- 初始化测试数据
- 生成测试报告
- 模拟数据生成
- 转换Swagger/OpenAPI规范`,
}

// Execute 添加所有子命令到根命令并设置标志
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// 全局标志
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "配置文件路径 (默认为 ./config.yaml)")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "启用详细输出")
	rootCmd.PersistentFlags().StringVar(&outputDir, "output", "", "输出目录路径（默认 ./reports，或配置文件中的 output_dir）")
	rootCmd.PersistentFlags().StringVar(&reportType, "report-type", "html", "报告类型 (html, json, xml, junit)")
}
