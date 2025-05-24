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

// CreateDefaultConfig 创建默认配置
func CreateDefaultConfig() *Config {
	return &Config{
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
