package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config 保存测试工具的配置
type Config struct {
	// OpenAPI/Swagger 规范文件路径
	SpecFile string
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
}

// NewConfig 创建新的配置
func NewConfig(specFile, baseURL, headersJSON, outputDir string, verbose bool, timeout int) (*Config, error) {
	// 验证规范文件是否存在
	if _, err := os.Stat(specFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("规范文件不存在: %s", specFile)
	}

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

	// 返回配置
	return &Config{
		SpecFile:  filepath.Clean(specFile),
		BaseURL:   baseURL,
		Headers:   headers,
		OutputDir: outputDir,
		Verbose:   verbose,
		Timeout:   timeout,
	}, nil
}
