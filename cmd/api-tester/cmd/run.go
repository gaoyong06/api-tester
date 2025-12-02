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
	specFile      string
	baseURL       string
	headers       string
	timeout       int
	pathParams    string
	requestBodies string
	scenarioFile  string
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
		outputFlag := cmd.Flags().Lookup("output")
		outputFlagChanged := outputFlag != nil && outputFlag.Changed

		if cfgFile != "" {
			// 从配置文件加载配置
			yamlConfig, err := yaml.LoadConfig(cfgFile)
			if err != nil {
				log.Fatalf("无法加载配置文件: %v", err)
			}

			// 解析 spec 文件路径（相对于配置文件目录）
			cfgFileAbs, _ := filepath.Abs(cfgFile)
			cfgFileDir := filepath.Dir(cfgFileAbs)
			resolvedSpecFile := yamlConfig.Spec
			if !filepath.IsAbs(resolvedSpecFile) {
				// 如果是相对路径，则相对于配置文件目录解析
				resolvedSpecFile = filepath.Clean(filepath.Join(cfgFileDir, resolvedSpecFile))
			}
			resolvedSpecFiles := make([]string, len(yamlConfig.SpecFiles))
			for i, specFile := range yamlConfig.SpecFiles {
				if !filepath.IsAbs(specFile) {
					resolvedSpecFiles[i] = filepath.Clean(filepath.Join(cfgFileDir, specFile))
				} else {
					resolvedSpecFiles[i] = specFile
				}
			}

			// 转换为内部配置格式
			cfg = &config.Config{
				SpecFile:   resolvedSpecFile,
				SpecFiles:  resolvedSpecFiles, // 添加对多个规范文件的支持
				BaseURL:    yamlConfig.BaseURL,
				Headers:    yamlConfig.Request.Headers,
				OutputDir:  yamlConfig.OutputDir,
				Verbose:    verbose,
				Timeout:    yamlConfig.Timeout,
				PathParams: yamlConfig.Request.PathParams,
				// 将 map[string]string 转换为 map[string]interface{}
				RequestBodies: convertStringMapToInterfaceMap(yamlConfig.Request.RequestBodies),
				// 保存YAML配置对象，用于场景测试
				YamlConfig: yamlConfig,
			}

			// 如果命令行参数提供了值，覆盖配置文件中的值
			if specFile != "" {
				cfg.SpecFile = specFile
			}
			if baseURL != "" {
				cfg.BaseURL = baseURL
			}
			if headers != "" {
				// 将字符串转换为 map[string]string
				headerMap := make(map[string]string)
				// 这里应该实现一个解析字符串为 map 的函数
				// 为了简化，我们这里使用空的 map
				cfg.Headers = headerMap
			}
			if outputFlagChanged && outputDir != "" {
				cfg.OutputDir = outputDir
			}
			if timeout != 30 {
				cfg.Timeout = timeout
			}
			if pathParams != "" {
				// 将字符串转换为 map[string]string
				pathParamsMap := make(map[string]string)
				// 这里应该实现一个解析字符串为 map 的函数
				// 为了简化，我们这里使用空的 map
				cfg.PathParams = pathParamsMap
			}
			if requestBodies != "" {
				// 将字符串转换为 map[string]interface{}
				requestBodiesMap := make(map[string]interface{})
				// 这里应该实现一个解析字符串为 map 的函数
				// 为了简化，我们这里使用空的 map
				cfg.RequestBodies = requestBodiesMap
			}
		} else {
			effectiveOutputDir := outputDir
			if !outputFlagChanged || effectiveOutputDir == "" {
				effectiveOutputDir = "./reports"
			}
			// 验证必填参数
			if specFile == "" || baseURL == "" {
				cmd.Help()
				fmt.Println("\n错误: 必须提供 spec 和 url 参数，或者使用配置文件")
				os.Exit(1)
			}

			// 从命令行参数创建配置
			cfg, err = config.NewConfig(specFile, baseURL, headers, effectiveOutputDir, verbose, timeout, pathParams, requestBodies)
			if err != nil {
				log.Fatalf("配置错误: %v", err)
			}
		}

		if cfg.OutputDir == "" {
			cfg.OutputDir = "./reports"
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
			var apiDef *parser.APIDefinition

			// 如果指定了单个规范文件，使用它
			if cfg.SpecFile != "" {
				apiDef, err = parser.ParseSwaggerFile(cfg.SpecFile)
				if err != nil {
					log.Fatalf("无法解析API定义: %v", err)
				}
			} else if len(cfg.SpecFiles) > 0 {
				// 如果指定了多个规范文件，合并所有文件的信息
				// 创建一个空的 API 定义来存储合并的结果
				apiDef = &parser.APIDefinition{
					Title:     "合并的 API 定义",
					Version:   "v1",
					Endpoints: []*parser.Endpoint{},
				}

				// 遍历所有规范文件并合并端点
				for i, specFile := range cfg.SpecFiles {
					fmt.Printf("解析规范文件 [%d/%d]: %s\n", i+1, len(cfg.SpecFiles), specFile)
					tempDef, err := parser.ParseSwaggerFile(specFile)
					if err != nil {
						log.Printf("警告: 无法解析规范文件 %s: %v", specFile, err)
						continue
					}

					// 将当前文件的端点添加到合并的 API 定义中
					apiDef.Endpoints = append(apiDef.Endpoints, tempDef.Endpoints...)
					fmt.Printf("  找到 %d 个端点\n", len(tempDef.Endpoints))
				}

				fmt.Printf("总端点数量: %d\n\n", len(apiDef.Endpoints))

				// 检查是否成功解析了任何端点
				if len(apiDef.Endpoints) == 0 {
					log.Fatalf("无法从任何规范文件中解析出端点")
				}
			} else {
				log.Fatalf("未指定API规范文件")
			}

			// 确保输出目录存在
			outputDir := cfg.OutputDir
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				log.Fatalf("无法创建输出目录: %v", err)
			}

			// 将测试结果转换为端点测试结果数组，用于报告生成
			endpointResults := results.Results

			// 输出测试结果信息，用于调试
			fmt.Printf("测试结果数量: %d\n", len(endpointResults))

			// 创建一个字符串切片来存储所有生成的报告路径
			reportPaths := []string{}

			// 始终生成 JSON 报告，方便机器处理
			fmt.Println("正在生成 JSON 报告...")
			jsonReportPath, err := machine.GenerateReport(apiDef, endpointResults, outputDir, "json")
			if err != nil {
				log.Printf("警告: 无法生成 JSON 报告: %v", err)
			} else {
				reportPaths = append(reportPaths, jsonReportPath)
				fmt.Printf("JSON 报告已成功保存到: %s\n", jsonReportPath)
				// 检查文件是否存在
				if _, err := os.Stat(jsonReportPath); os.IsNotExist(err) {
					log.Printf("警告: JSON 报告文件不存在: %s", jsonReportPath)
				}
			}

			// 根据报告类型生成其他格式的报告
			var reportPath string

			switch reportType {
			case "json":
				// 已经生成了 JSON 报告，不需要重复生成
				reportPath = jsonReportPath
			case "xml":
				reportPath, err = machine.GenerateReport(apiDef, endpointResults, outputDir, "xml")
			case "junit":
				reportPath, err = machine.GenerateJUnitReport(apiDef, endpointResults, outputDir)
			default:
				// 默认生成 HTML 报告
				reportPath, err = machine.GenerateReport(apiDef, endpointResults, outputDir, "html")
			}

			if err != nil {
				log.Fatalf("无法生成机器可读报告: %v", err)
			}

			fmt.Printf("机器可读报告已保存到: %s\n", reportPath)

			// 如果有测试失败，返回非零退出码
			if results.Failed > 0 {
				os.Exit(1)
			}
		}
	},
}

// convertStringMapToInterfaceMap 将 map[string]string 转换为 map[string]interface{}
func convertStringMapToInterfaceMap(strMap map[string]string) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range strMap {
		result[k] = v
	}
	return result
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
