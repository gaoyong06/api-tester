package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gaoyong06/api-tester/internal/config"
	"github.com/gaoyong06/api-tester/internal/config/yaml"
	"github.com/gaoyong06/api-tester/internal/parser"
	"github.com/gaoyong06/api-tester/internal/reporter/machine"
	"github.com/gaoyong06/api-tester/internal/runner"
	"github.com/spf13/cobra"
)

var (
	// run 命令的标志
	specFile     string
	baseURL      string
	headers      string
	timeout      int
	pathParams   string
	requestBodies string
	scenarioFile string
)

// runCmd 表示 run 子命令
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "运行API测试",
	Long: `运行API测试命令用于执行API测试并生成报告。

可以通过命令行参数指定测试配置，也可以使用配置文件。
如果同时提供了命令行参数和配置文件，命令行参数将优先使用。`,
	Run: func(cmd *cobra.Command, args []string) {
		// 检查是否提供了配置文件
		var cfg *config.Config
		var err error

		if cfgFile != "" {
			// 从配置文件加载配置
			yamlConfig, err := yaml.LoadConfig(cfgFile)
			if err != nil {
				log.Fatalf("无法加载配置文件: %v", err)
			}

			// 转换为内部配置格式
			cfg = &config.Config{
				SpecFile:      yamlConfig.API.SpecFile,
				BaseURL:       yamlConfig.API.BaseURL,
				Headers:       yamlConfig.API.Headers,
				OutputDir:     yamlConfig.Output.Directory,
				Verbose:       verbose,
				Timeout:       yamlConfig.API.Timeout,
				PathParams:    yamlConfig.API.PathParams,
				RequestBodies: yamlConfig.API.RequestBodies,
			}

			// 如果命令行参数提供了值，覆盖配置文件中的值
			if specFile != "" {
				cfg.SpecFile = specFile
			}
			if baseURL != "" {
				cfg.BaseURL = baseURL
			}
			if headers != "" {
				cfg.Headers = headers
			}
			if outputDir != "" {
				cfg.OutputDir = outputDir
			}
			if timeout != 30 {
				cfg.Timeout = timeout
			}
			if pathParams != "" {
				cfg.PathParams = pathParams
			}
			if requestBodies != "" {
				cfg.RequestBodies = requestBodies
			}
		} else {
			// 验证必填参数
			if specFile == "" || baseURL == "" {
				cmd.Help()
				fmt.Println("\n错误: 必须提供 spec 和 url 参数，或者使用配置文件")
				os.Exit(1)
			}

			// 从命令行参数创建配置
			cfg, err = config.NewConfig(specFile, baseURL, headers, outputDir, verbose, timeout, pathParams, requestBodies)
			if err != nil {
				log.Fatalf("配置错误: %v", err)
			}
		}

		// 创建并运行测试
		r := runner.NewRunner(cfg)
		results, err := r.Run()
		if err != nil {
			log.Fatalf("测试运行失败: %v", err)
		}

		// 输出测试结果摘要
		fmt.Printf("\n测试完成! 总计: %d, 通过: %d, 失败: %d\n", 
			results.Total, results.Passed, results.Failed)
		fmt.Printf("详细报告已保存到: %s\n", results.ReportPath)

		// 如果需要生成机器可读报告
		if reportType == "json" || reportType == "xml" || reportType == "junit" {
			// 解析API定义
			apiDef, err := parser.ParseSwaggerFile(cfg.SpecFile)
			if err != nil {
				log.Fatalf("无法解析API定义: %v", err)
			}

			// 确保输出目录存在
			outputDir := filepath.Join(cfg.OutputDir, "machine")
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				log.Fatalf("无法创建输出目录: %v", err)
			}

			var reportPath string
			// 根据报告类型生成不同格式的报告
			switch reportType {
			case "json":
				reportPath, err = machine.GenerateReport(apiDef, results.Results, outputDir, "json")
			case "xml":
				reportPath, err = machine.GenerateReport(apiDef, results.Results, outputDir, "xml")
			case "junit":
				reportPath, err = machine.GenerateJUnitReport(apiDef, results.Results, outputDir)
			}

			if err != nil {
				log.Fatalf("无法生成机器可读报告: %v", err)
			}

			fmt.Printf("机器可读报告已保存到: %s\n", reportPath)
		}

		// 如果有测试失败，返回非零退出码
		if results.Failed > 0 {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	// 本地标志
	runCmd.Flags().StringVar(&specFile, "spec", "", "OpenAPI/Swagger 规范文件路径")
	runCmd.Flags().StringVar(&baseURL, "url", "", "API 基础 URL")
	runCmd.Flags().StringVar(&headers, "headers", "", "请求头 (JSON 格式)")
	runCmd.Flags().IntVar(&timeout, "timeout", 30, "请求超时时间 (秒)")
	runCmd.Flags().StringVar(&pathParams, "path-params", "", "路径参数文件 (JSON 格式)")
	runCmd.Flags().StringVar(&requestBodies, "request-bodies", "", "请求体模板文件 (JSON 格式)")
	runCmd.Flags().StringVar(&scenarioFile, "scenario", "", "测试场景文件 (YAML 格式)")
}
