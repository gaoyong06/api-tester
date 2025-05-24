package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config 保存测试工具的配置
type Config struct {
	// OpenAPI/Swagger 规范文件路径（单个文件，向后兼容）
	SpecFile string
	// OpenAPI/Swagger 规范文件路径数组（多个文件）
	SpecFiles []string
	// API 基础 URL
	BaseURL string
	// 请求头 (JSON 格式)
	Headers map[string]string
	// 测试报告输出目录
	OutputDir string
	// 是否显示详细日志
	Verbose bool
	// 请求超时时间 (秒)
	Timeout int
	// 路径参数替换映射
	PathParams map[string]string
	// 请求体模板映射
	RequestBodies map[string]interface{}
}

// NewConfig 创建新的配置
func NewConfig(specFile, baseURL, headersJSON, outputDir string, verbose bool, timeout int, pathParamsFile, requestBodiesFile string) (*Config, error) {
	// 验证规范文件是否存在
	if _, err := os.Stat(specFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("规范文件不存在: %s", specFile)
	}

	// 创建单个规范文件的配置
	specFiles := []string{specFile}

	// 解析请求头
	headers := make(map[string]string)
	if headersJSON != "" {
		if err := json.Unmarshal([]byte(headersJSON), &headers); err != nil {
			return nil, fmt.Errorf("无法解析请求头 JSON: %v", err)
		}
	}

	// 创建输出目录（如果不存在）
	if outputDir != "" {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return nil, fmt.Errorf("无法创建输出目录: %v", err)
		}
	}

	// 解析路径参数
	pathParams := make(map[string]string)
	if pathParamsFile != "" {
		// 读取路径参数文件
		pathParamsData, err := os.ReadFile(pathParamsFile)
		if err != nil {
			return nil, fmt.Errorf("无法读取路径参数文件: %v", err)
		}

		// 解析 JSON
		if err := json.Unmarshal(pathParamsData, &pathParams); err != nil {
			return nil, fmt.Errorf("无法解析路径参数 JSON: %v", err)
		}
	}

	// 解析请求体模板
	requestBodies := make(map[string]interface{})
	if requestBodiesFile != "" {
		// 读取请求体模板文件
		requestBodiesData, err := os.ReadFile(requestBodiesFile)
		if err != nil {
			return nil, fmt.Errorf("无法读取请求体模板文件: %v", err)
		}

		// 解析 JSON
		if err := json.Unmarshal(requestBodiesData, &requestBodies); err != nil {
			return nil, fmt.Errorf("无法解析请求体模板 JSON: %v", err)
		}
	}

	// 返回配置
	return &Config{
		SpecFile:     filepath.Clean(specFile),
		SpecFiles:    specFiles,  // 添加对多个规范文件的支持
		BaseURL:      baseURL,
		Headers:      headers,
		OutputDir:    outputDir,
		Verbose:      verbose,
		Timeout:      timeout,
		PathParams:   pathParams,
		RequestBodies: requestBodies,
	}, nil
}
