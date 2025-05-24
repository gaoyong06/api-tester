package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gaoyong06/api-tester/internal/parser"
	"github.com/gaoyong06/api-tester/internal/reporter/machine"
	"github.com/gaoyong06/api-tester/internal/types"
	"github.com/spf13/cobra"
)

var (
	// report u547du4ee4u7684u6807u5fd7
	resultsFile string
	title       string
	description string
)

// reportCmd u8868u793a report u5b50u547du4ee4
var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "u751fu6210u6d4bu8bd5u62a5u544a",
	Long: `u751fu6210u6d4bu8bd5u62a5u544au547du4ee4u7528u4e8eu4eceu5df2u6709u7684u6d4bu8bd5u7ed3u679cu751fu6210u62a5u544au3002

u652fu6301u591au79cdu62a5u544au683cu5f0fuff0cu5305u62ec HTMLu3001JSONu3001XML u548c JUnitu3002
u53efu4ee5u4f7fu7528u5df2u4fddu5b58u7684u6d4bu8bd5u7ed3u679cu6587u4ef6u6216u8005u4eceu914du7f6eu6587u4ef6u4e2du52a0u8f7du6d4bu8bd5u7ed3u679cu3002`,
	Run: func(cmd *cobra.Command, args []string) {
		// u68c0u67e5u662fu5426u63d0u4f9bu4e86u6d4bu8bd5u7ed3u679cu6587u4ef6u6216u914du7f6eu6587u4ef6
		if resultsFile == "" && cfgFile == "" {
			cmd.Help()
			fmt.Println("\nu9519u8bef: u5fc5u987bu63d0u4f9bu6d4bu8bd5u7ed3u679cu6587u4ef6u6216u914du7f6eu6587u4ef6")
			os.Exit(1)
		}

		// u786eu4fddu8f93u51fau76eeu5f55u5b58u5728
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			log.Fatalf("u65e0u6cd5u521bu5efau8f93u51fau76eeu5f55: %v", err)
		}

		// u52a0u8f7du6d4bu8bd5u7ed3u679c
		var results []*types.EndpointTestResult
		var err error

		if resultsFile != "" {
			// u4eceu6587u4ef6u52a0u8f7du6d4bu8bd5u7ed3u679c			// 从文件加载测试结果
			testResult, err := types.LoadTestResultsFromFile(resultsFile)
			if err != nil {
				log.Fatalf("u65e0u6cd5u52a0u8f7du6d4bu8bd5u7ed3u679cu6587u4ef6: %v", err)
			}
			// 使用测试结果中的详细结果
			results = testResult.Results
		} else {
			// u4eceu914du7f6eu6587u4ef6u4e2du6307u5b9au7684u6700u65b0u6d4bu8bd5u7ed3u679cu52a0u8f7d
			// u8fd9u91ccu53efu4ee5u5b9eu73b0u4eceu914du7f6eu6587u4ef6u4e2du6307u5b9au7684u76eeu5f55u627eu5230u6700u65b0u7684u6d4bu8bd5u7ed3u679cu6587u4ef6
			// u8fd9u91ccu7b80u5316u5904u7406uff0cu5047u8bbeu6709u4e00u4e2au9ed8u8ba4u7684u7ed3u679cu6587u4ef6u4f4du7f6e			// 从配置文件中指定的最新测试结果加载
			// 这里可以实现从配置文件中指定的目录找到最新的测试结果文件
			// 这里简化处理，假设有一个默认的结果文件位置
			defaultResultsFile := filepath.Join(outputDir, "latest-results.json")
			testResult, err := types.LoadTestResultsFromFile(defaultResultsFile)
			if err != nil {
				log.Fatalf("u65e0u6cd5u52a0u8f7du9ed8u8ba4u6d4bu8bd5u7ed3u679cu6587u4ef6: %v", err)
			}
			// 使用测试结果中的详细结果
			results = testResult.Results
		}

		// u52a0u8f7d API u5b9au4e49
		var apiDef *parser.APIDefinition
		if specFile != "" {
			apiDef, err = parser.ParseSwaggerFile(specFile)
			if err != nil {
				log.Fatalf("u65e0u6cd5u89e3u6790 API u5b9au4e49: %v", err)
			}
		} else {
			// u5982u679cu6ca1u6709u63d0u4f9b spec u6587u4ef6uff0cu5c1du8bd5u4eceu6d4bu8bd5u7ed3u679cu4e2du63d0u53d6 API u5b9au4e49
			// u8fd9u91ccu7b80u5316u5904u7406uff0cu5047u8bbeu6d4bu8bd5u7ed3u679cu4e2du5305u542bu4e86 API u5b9au4e49u7684u5f15u7528
			// u5b9eu9645u5b9eu73b0u4e2du53efu80fdu9700u8981u66f4u590du6742u7684u903bu8f91
			log.Fatalf("u9519u8bef: u5fc5u987bu63d0u4f9b API u89c4u8303u6587u4ef6 (--spec)")
		}

		// u751fu6210u62a5u544a
		var reportPath string

		// u521bu5efau62a5u544au8f93u51fau76eeu5f55
		reportOutputDir := filepath.Join(outputDir, reportType)
		if err := os.MkdirAll(reportOutputDir, 0755); err != nil {
			log.Fatalf("u65e0u6cd5u521bu5efau62a5u544au8f93u51fau76eeu5f55: %v", err)
		}

		// u6839u636eu62a5u544au7c7bu578bu751fu6210u4e0du540cu683cu5f0fu7684u62a5u544a
		switch reportType {
		case "json":
			reportPath, err = machine.GenerateReport(apiDef, results, reportOutputDir, "json")
		case "xml":
			reportPath, err = machine.GenerateReport(apiDef, results, reportOutputDir, "xml")
		case "junit":
			reportPath, err = machine.GenerateJUnitReport(apiDef, results, reportOutputDir)
		case "html":
			// u8fd9u91ccu53efu4ee5u8c03u7528HTMLu62a5u544au751fu6210u5668
			// u5f53u524du7b80u5316u5904u7406uff0cu76f4u63a5u8f93u51fau9519u8befu4fe1u606f
			log.Fatalf("u5c1au672au5b9eu73b0HTMLu62a5u544au751fu6210u5668")
		default:
			log.Fatalf("u4e0du652fu6301u7684u62a5u544au7c7bu578b: %s", reportType)
		}

		if err != nil {
			log.Fatalf("u65e0u6cd5u751fu6210u62a5u544a: %v", err)
		}

		fmt.Printf("u62a5u544au751fu6210u6210u529f! u4fddu5b58u5230: %s\n", reportPath)
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)

	// u672cu5730u6807u5fd7
	reportCmd.Flags().StringVar(&resultsFile, "results", "", "u6d4bu8bd5u7ed3u679cu6587u4ef6u8defu5f84")
	reportCmd.Flags().StringVar(&title, "title", "", "u62a5u544au6807u9898")
	reportCmd.Flags().StringVar(&description, "description", "", "u62a5u544au63cfu8ff0")
}
