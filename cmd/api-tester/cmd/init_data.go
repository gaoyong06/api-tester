package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/gaoyong06/api-tester/internal/config/yaml"
	"github.com/gaoyong06/api-tester/internal/mock"
	"github.com/spf13/cobra"
)

var (
	// init-data 命令的标志
	dataSourceFile string
	cleanupAfter   bool
)

// initDataCmd 表示 init-data 子命令
var initDataCmd = &cobra.Command{
	Use:   "init-data",
	Short: "初始化测试数据",
	Long: `初始化测试数据命令用于准备测试环境和数据。

可以从配置文件中加载数据源定义，或者通过命令行参数指定数据源文件。
支持多种数据源类型，包括文件、SQL和API。`,
	Run: func(cmd *cobra.Command, args []string) {
		// 检查是否提供了配置文件或数据源文件
		if cfgFile == "" && dataSourceFile == "" {
			cmd.Help()
			fmt.Println("\n错误: 必须提供配置文件或数据源文件")
			os.Exit(1)
		}

		var dataSources []yaml.DataSource

		if cfgFile != "" {
			// 从配置文件加载数据源
			yamlConfig, err := yaml.LoadConfig(cfgFile)
			if err != nil {
				log.Fatalf("无法加载配置文件: %v", err)
			}

			dataSources = yamlConfig.TestData.DataSources
		} else {
			// 从数据源文件加载
			dataSourceConfig, err := yaml.LoadDataSourcesFromFile(dataSourceFile)
			if err != nil {
				log.Fatalf("无法加载数据源文件: %v", err)
			}

			dataSources = dataSourceConfig
		}

		// 创建数据管理器
		dataManager := mock.NewTestDataManager()

		// 初始化数据
		fmt.Println("正在初始化测试数据...")
		for _, source := range dataSources {
			fmt.Printf("处理数据源: %s (%s)\n", source.Name, source.Type)
			err := dataManager.LoadDataSource(&source)
			if err != nil {
				log.Fatalf("无法加载数据源 %s: %v", source.Name, err)
			}
		}

		// 执行初始化脚本
		if cfgFile != "" {
			yamlConfig, _ := yaml.LoadConfig(cfgFile)
			if len(yamlConfig.TestData.InitScripts) > 0 {
				fmt.Println("执行初始化脚本...")
				for _, script := range yamlConfig.TestData.InitScripts {
					fmt.Printf("执行脚本: %s\n", script.Name)
					err := dataManager.ExecuteScript(script.Type, script.Content)
					if err != nil {
						log.Fatalf("无法执行初始化脚本 %s: %v", script.Name, err)
					}
				}
			}
		}

		// 如果设置了清理标志，注册清理函数
		if cleanupAfter {
			fmt.Println("注册数据清理函数...")
			// 这里可以注册一个在测试完成后执行的清理函数
			// 例如，可以将清理脚本保存到临时文件，供后续使用
		}

		fmt.Println("测试数据初始化完成!")
	},
}

func init() {
	rootCmd.AddCommand(initDataCmd)

	// 本地标志
	initDataCmd.Flags().StringVar(&dataSourceFile, "data-source", "", "数据源定义文件 (YAML 格式)")
	initDataCmd.Flags().BoolVar(&cleanupAfter, "cleanup", false, "测试完成后清理数据")
}
