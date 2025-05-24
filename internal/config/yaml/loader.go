package yaml

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadConfig 从YAML文件加载配置
func LoadConfig(filePath string) (*Config, error) {
	// 获取文件绝对路径
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("无法获取配置文件的绝对路径: %v", err)
	}

	// 检查文件是否存在
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("配置文件不存在: %s", absPath)
	}

	// 读取文件内容
	data, err := ioutil.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("无法读取配置文件: %v", err)
	}

	// 解析YAML
	config := &Config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("无法解析YAML配置: %v", err)
	}

	// 处理包含的配置文件
	if len(config.Includes) > 0 {
		// 获取主配置文件所在目录，用于解析相对路径
		baseDir := filepath.Dir(absPath)
		
		// 加载并合并所有包含的配置文件
		mergedConfig, err := loadAndMergeIncludes(config, baseDir)
		if err != nil {
			return nil, err
		}
		
		config = mergedConfig
	}

	// 设置默认值
	setDefaults(config)

	// 验证配置
	if err := validateConfig(config); err != nil {
		return nil, err
	}

	return config, nil
}

// setDefaults 设置配置的默认值
func setDefaults(config *Config) {
	// 设置默认超时时间
	if config.Timeout == 0 {
		config.Timeout = 30 // 默认30秒
	}

	// 设置默认输出目录
	if config.OutputDir == "" {
		config.OutputDir = "./test-reports"
	}

	// 确保请求头映射存在
	if config.Request.Headers == nil {
		config.Request.Headers = make(map[string]string)
	}

	// 确保路径参数映射存在
	if config.Request.PathParams == nil {
		config.Request.PathParams = make(map[string]string)
	}

	// 确保查询参数映射存在
	if config.Request.QueryParams == nil {
		config.Request.QueryParams = make(map[string]string)
	}

	// 确保请求体映射存在
	if config.Request.RequestBodies == nil {
		config.Request.RequestBodies = make(map[string]string)
	}

	// 设置Web UI默认端口
	if config.WebUI.Enabled && config.WebUI.Port == 0 {
		config.WebUI.Port = 8080 // 默认端口8080
	}
}

// validateConfig 验证配置是否有效
func validateConfig(config *Config) error {
	// 验证API规范文件
	// 如果 Spec 为空，检查 SpecFiles 是否有值
	if config.Spec == "" && len(config.SpecFiles) == 0 {
		return fmt.Errorf("API规范文件路径不能为空")
	}

	// 验证基础URL
	if config.BaseURL == "" {
		return fmt.Errorf("API基础URL不能为空")
	}

	return nil
}

// SaveConfig 将配置保存到YAML文件
func SaveConfig(config *Config, filePath string) error {
	// 将配置转换为YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("无法序列化配置: %v", err)
	}

	// 创建目录（如果不存在）
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("无法创建目录: %v", err)
	}

	// 写入文件
	if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("无法写入配置文件: %v", err)
	}

	return nil
}

// loadAndMergeIncludes 加载并合并包含的配置文件
func loadAndMergeIncludes(baseConfig *Config, baseDir string) (*Config, error) {
	// 创建一个新的配置对象作为合并结果
	result := &Config{}
	
	// 先复制基础配置
	*result = *baseConfig
	
	// 清除 includes 字段，避免循环引用
	includes := baseConfig.Includes
	result.Includes = nil
	
	// 处理每个包含的文件
	for _, includePath := range includes {
		// 如果是相对路径，则基于主配置文件目录解析
		if !filepath.IsAbs(includePath) {
			includePath = filepath.Join(baseDir, includePath)
		}
		
		// 检查文件是否存在
		if _, err := os.Stat(includePath); os.IsNotExist(err) {
			return nil, fmt.Errorf("包含的配置文件不存在: %s", includePath)
		}
		
		// 读取文件内容
		data, err := ioutil.ReadFile(includePath)
		if err != nil {
			return nil, fmt.Errorf("无法读取包含的配置文件: %v", err)
		}
		
		// 解析YAML
		includeConfig := &Config{}
		if err := yaml.Unmarshal(data, includeConfig); err != nil {
			return nil, fmt.Errorf("无法解析包含的YAML配置: %v", err)
		}
		
		// 如果包含的文件中也有 includes，递归处理
		if len(includeConfig.Includes) > 0 {
			// 获取包含文件所在目录
			includeDir := filepath.Dir(includePath)
			
			// 递归处理
			mergedInclude, err := loadAndMergeIncludes(includeConfig, includeDir)
			if err != nil {
				return nil, err
			}
			
			includeConfig = mergedInclude
		}
		
		// 合并配置
		result = mergeConfigs(result, includeConfig)
	}
	
	return result, nil
}

// mergeConfigs 合并两个配置对象
func mergeConfigs(base, override *Config) *Config {
	// 创建一个新的配置对象作为合并结果
	result := &Config{}
	*result = *base
	
	// 如果覆盖配置中有值，则使用覆盖配置的值
	if override.Spec != "" {
		result.Spec = override.Spec
	}
	
	// 合并 SpecFiles
	if len(override.SpecFiles) > 0 {
		// 如果基础配置中没有 SpecFiles，直接使用覆盖配置的
		if len(result.SpecFiles) == 0 {
			result.SpecFiles = override.SpecFiles
		} else {
			// 否则合并两个列表，避免重复
			specFilesMap := make(map[string]bool)
			for _, file := range result.SpecFiles {
				specFilesMap[file] = true
			}
			
			for _, file := range override.SpecFiles {
				if !specFilesMap[file] {
					result.SpecFiles = append(result.SpecFiles, file)
				}
			}
		}
	}
	
	// 合并其他基本字段
	if override.BaseURL != "" {
		result.BaseURL = override.BaseURL
	}
	
	if override.OutputDir != "" {
		result.OutputDir = override.OutputDir
	}
	
	if override.Timeout != 0 {
		result.Timeout = override.Timeout
	}
	
	// 合并 Request 结构
	// 合并 Headers
	for k, v := range override.Request.Headers {
		result.Request.Headers[k] = v
	}
	
	// 合并 PathParams
	for k, v := range override.Request.PathParams {
		result.Request.PathParams[k] = v
	}
	
	// 合并 QueryParams
	for k, v := range override.Request.QueryParams {
		result.Request.QueryParams[k] = v
	}
	
	// 合并 RequestBodies
	for k, v := range override.Request.RequestBodies {
		result.Request.RequestBodies[k] = v
	}
	
	// 合并 TestData
	if override.TestData.InitScript != "" {
		result.TestData.InitScript = override.TestData.InitScript
	}
	
	if override.TestData.CleanupScript != "" {
		result.TestData.CleanupScript = override.TestData.CleanupScript
	}
	
	// 合并 TestData.Sources
	result.TestData.Sources = append(result.TestData.Sources, override.TestData.Sources...)
	
	// 合并 Scenarios
	result.Scenarios = append(result.Scenarios, override.Scenarios...)
	
	// 合并 CI 配置
	if override.CI.OutputFormat != "" {
		result.CI.OutputFormat = override.CI.OutputFormat
	}
	
	if override.CI.FailThreshold != 0 {
		result.CI.FailThreshold = override.CI.FailThreshold
	}
	
	if override.CI.Notifications.Slack != "" {
		result.CI.Notifications.Slack = override.CI.Notifications.Slack
	}
	
	if override.CI.Notifications.Email != "" {
		result.CI.Notifications.Email = override.CI.Notifications.Email
	}
	
	// 合并 WebUI 配置
	if override.WebUI.Enabled {
		result.WebUI.Enabled = true
	}
	
	if override.WebUI.Port != 0 {
		result.WebUI.Port = override.WebUI.Port
	}
	
	return result
}

// CreateDefaultConfig 创建默认配置
func CreateDefaultConfig() *Config {
	return &Config{
		Includes:  []string{},
		Spec:      "./api/swagger.json",
		BaseURL:   "http://localhost:8080",
		OutputDir: "./test-reports",
		Timeout:   30,
		Verbose:   false,
		Request: struct {
			Headers       map[string]string `yaml:"headers"`
			PathParams    map[string]string `yaml:"path_params"`
			QueryParams   map[string]string `yaml:"query_params"`
			RequestBodies map[string]string `yaml:"request_bodies"`
		}{
			Headers:       make(map[string]string),
			PathParams:    make(map[string]string),
			QueryParams:   make(map[string]string),
			RequestBodies: make(map[string]string),
		},
		TestData: struct {
			InitScript    string       `yaml:"init_script"`
			CleanupScript string       `yaml:"cleanup_script"`
			Sources       []DataSource `yaml:"sources"`
		}{
			Sources: []DataSource{},
		},
		Scenarios: []Scenario{},
		CI: struct {
			OutputFormat  string `yaml:"output_format"`
			FailThreshold float64 `yaml:"fail_threshold"`
			Notifications struct {
				Slack string `yaml:"slack"`
				Email string `yaml:"email"`
			} `yaml:"notifications"`
		}{
			OutputFormat:  "html",
			FailThreshold: 0,
		},
		WebUI: struct {
			Enabled bool `yaml:"enabled"`
			Port    int  `yaml:"port"`
		}{
			Enabled: false,
			Port:    8080,
		},
	}
}
