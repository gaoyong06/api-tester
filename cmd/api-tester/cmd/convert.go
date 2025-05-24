package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	// convert 命令的标志
	inputFile  string
	outputFile string
	format     string
	version    string
)

// convertCmd 表示 convert 子命令
var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "转换 API 规范格式",
	Long: `转换 API 规范格式命令用于将不同格式的 API 规范转换为标准的 OpenAPI 格式。

支持的转换包括：
- Swagger 1.2 到 OpenAPI 3.0
- Swagger/OpenAPI 2.0 到 OpenAPI 3.0
- JSON 格式到 YAML 格式（反之亦然）`,
	Run: func(cmd *cobra.Command, args []string) {
		// 检查必要参数
		if inputFile == "" {
			cmd.Help()
			fmt.Println("\n错误: 必须提供输入文件 (--input)")
			os.Exit(1)
		}

		// 如果没有指定输出文件，则根据输入文件生成默认输出文件名
		if outputFile == "" {
			ext := filepath.Ext(inputFile)
			base := strings.TrimSuffix(inputFile, ext)

			// 根据指定的格式设置扩展名
			var newExt string
			if format == "json" {
				newExt = ".json"
			} else {
				newExt = ".yaml"
			}

			// 添加版本信息到文件名
			outputFile = fmt.Sprintf("%s.openapi%s%s", base, version, newExt)
		}

		// 读取输入文件
		data, err := ioutil.ReadFile(inputFile)
		if err != nil {
			log.Fatalf("无法读取输入文件: %v", err)
		}

		// 确定输入文件格式（JSON 或 YAML）
		var inputFormat string
		if strings.HasSuffix(strings.ToLower(inputFile), ".json") {
			inputFormat = "json"
		} else if strings.HasSuffix(strings.ToLower(inputFile), ".yaml") || strings.HasSuffix(strings.ToLower(inputFile), ".yml") {
			inputFormat = "yaml"
		} else {
			// 尝试解析为 JSON，如果失败则假设为 YAML
			var js json.RawMessage
			if err := json.Unmarshal(data, &js); err == nil {
				inputFormat = "json"
			} else {
				inputFormat = "yaml"
			}
		}

		// 根据版本执行不同的转换
		switch version {
		case "3":
			// 转换为 OpenAPI 3.0
			err = convertToOpenAPI3(data, inputFormat, outputFile, format)
		case "2":
			// 转换为 OpenAPI/Swagger 2.0
			err = convertToOpenAPI2(data, inputFormat, outputFile, format)
		default:
			log.Fatalf("不支持的 OpenAPI 版本: %s", version)
		}

		if err != nil {
			log.Fatalf("转换失败: %v", err)
		}

		fmt.Printf("转换成功! 输出文件: %s\n", outputFile)
	},
}

// convertToOpenAPI3 将 API 规范转换为 OpenAPI 3.0 格式
func convertToOpenAPI3(data []byte, inputFormat, outputFile, outputFormat string) error {
	// 解析输入文件
	var swagger2 openapi2.T
	var err error

	if inputFormat == "json" {
		err = json.Unmarshal(data, &swagger2)
	} else {
		err = yaml.Unmarshal(data, &swagger2)
	}

	if err != nil {
		return fmt.Errorf("无法解析输入文件: %v", err)
	}

	// 转换为 OpenAPI 3.0
	openapi3, err := openapi2conv.ToV3(&swagger2)
	if err != nil {
		return fmt.Errorf("无法转换为 OpenAPI 3.0: %v", err)
	}

	// 序列化为指定格式
	var outputData []byte
	if outputFormat == "json" {
		outputData, err = json.MarshalIndent(openapi3, "", "  ")
	} else {
		outputData, err = yaml.Marshal(openapi3)
	}

	if err != nil {
		return fmt.Errorf("无法序列化输出: %v", err)
	}

	// 写入输出文件
	if err := ioutil.WriteFile(outputFile, outputData, 0644); err != nil {
		return fmt.Errorf("无法写入输出文件: %v", err)
	}

	return nil
}

// convertToOpenAPI2 将 API 规范转换为 OpenAPI/Swagger 2.0 格式
func convertToOpenAPI2(data []byte, inputFormat, outputFile, outputFormat string) error {
	// 解析输入文件为 OpenAPI 3.0
	var openapi3 openapi3.T
	var err error

	loader := openapi3.NewLoader()

	if inputFormat == "json" {
		err = json.Unmarshal(data, &openapi3)
	} else {
		err = yaml.Unmarshal(data, &openapi3)
	}

	if err != nil {
		return fmt.Errorf("无法解析输入文件: %v", err)
	}

	// 转换为 OpenAPI/Swagger 2.0
	swagger2, err := openapi2conv.FromV3(&openapi3)
	if err != nil {
		return fmt.Errorf("无法转换为 OpenAPI 2.0: %v", err)
	}

	// 序列化为指定格式
	var outputData []byte
	if outputFormat == "json" {
		outputData, err = json.MarshalIndent(swagger2, "", "  ")
	} else {
		outputData, err = yaml.Marshal(swagger2)
	}

	if err != nil {
		return fmt.Errorf("无法序列化输出: %v", err)
	}

	// 写入输出文件
	if err := ioutil.WriteFile(outputFile, outputData, 0644); err != nil {
		return fmt.Errorf("无法写入输出文件: %v", err)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(convertCmd)

	// 本地标志
	convertCmd.Flags().StringVar(&inputFile, "input", "", "输入文件路径")
	convertCmd.Flags().StringVar(&outputFile, "output", "", "输出文件路径 (默认根据输入文件名生成)")
	convertCmd.Flags().StringVar(&format, "format", "yaml", "输出格式 (json 或 yaml)")
	convertCmd.Flags().StringVar(&version, "version", "3", "OpenAPI 版本 (2 或 3)")
}
