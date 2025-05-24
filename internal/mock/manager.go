package mock

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gaoyong06/api-tester/internal/config/yaml"
	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v3"
)

// TestDataManager 测试数据管理器
type TestDataManager struct {
	// 测试数据存储
	Data map[string]interface{}
	// 数据源配置
	Sources []yaml.DataSource
	// 配置
	Config *yaml.Config
	// OpenAPI 文档
	Doc *openapi3.T
	// Mock 生成器
	Generator *Generator
}

// NewTestDataManager 创建测试数据管理器
func NewTestDataManager(config *yaml.Config, doc *openapi3.T) *TestDataManager {
	return &TestDataManager{
		Data:      make(map[string]interface{}),
		Sources:   config.TestData.Sources,
		Config:    config,
		Doc:       doc,
		Generator: NewGenerator(doc),
	}
}

// LoadData 加载测试数据
func (m *TestDataManager) LoadData() error {
	// 加载所有数据源
	for _, source := range m.Sources {
		data, err := m.loadDataSource(source)
		if err != nil {
			return err
		}

		// 合并数据
		for k, v := range data {
			m.Data[k] = v
		}
	}

	return nil
}

// loadDataSource 加载单个数据源
func (m *TestDataManager) loadDataSource(source yaml.DataSource) (map[string]interface{}, error) {
	switch source.Type {
	case "file":
		return m.loadFromFile(source.Path)
	case "sql":
		return m.loadFromSQL(source.Path)
	case "api":
		return m.loadFromAPI(source.Path)
	default:
		return nil, fmt.Errorf("不支持的数据源类型: %s", source.Type)
	}
}

// loadFromFile 从文件加载数据
func (m *TestDataManager) loadFromFile(path string) (map[string]interface{}, error) {
	// 获取文件绝对路径
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("无法获取文件的绝对路径: %v", err)
	}

	// 检查文件是否存在
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("文件不存在: %s", absPath)
	}

	// 读取文件内容
	data, err := ioutil.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("无法读取文件: %v", err)
	}

	// 根据文件扩展名解析数据
	ext := filepath.Ext(absPath)
	var result map[string]interface{}

	switch strings.ToLower(ext) {
	case ".json":
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, fmt.Errorf("无法解析 JSON 文件: %v", err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &result); err != nil {
			return nil, fmt.Errorf("无法解析 YAML 文件: %v", err)
		}
	default:
		return nil, fmt.Errorf("不支持的文件格式: %s", ext)
	}

	return result, nil
}

// loadFromSQL 从 SQL 数据库加载数据
func (m *TestDataManager) loadFromSQL(connectionString string) (map[string]interface{}, error) {
	// 这里简化实现，实际项目中应该使用数据库驱动
	return nil, fmt.Errorf("SQL 数据源加载未实现")
}

// loadFromAPI 从 API 加载数据
func (m *TestDataManager) loadFromAPI(url string) (map[string]interface{}, error) {
	// 这里简化实现，实际项目中应该使用 HTTP 客户端
	return nil, fmt.Errorf("API 数据源加载未实现")
}

// InitData 初始化测试数据
func (m *TestDataManager) InitData() error {
	// 如果有初始化脚本，则执行
	if m.Config.TestData.InitScript != "" {
		cmd := exec.Command("sh", m.Config.TestData.InitScript)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("执行初始化脚本失败: %v", err)
		}
	}

	return nil
}

// CleanupData 清理测试数据
func (m *TestDataManager) CleanupData() error {
	// 如果有清理脚本，则执行
	if m.Config.TestData.CleanupScript != "" {
		cmd := exec.Command("sh", m.Config.TestData.CleanupScript)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("执行清理脚本失败: %v", err)
		}
	}

	return nil
}

// GenerateMockData 生成 Mock 数据
func (m *TestDataManager) GenerateMockData(strategy Strategy, count int) (map[string][]interface{}, error) {
	// 设置生成策略
	m.Generator.WithStrategy(strategy)

	// 生成数据
	return m.Generator.GenerateAll(count)
}

// SaveMockData 保存 Mock 数据到文件
func (m *TestDataManager) SaveMockData(data map[string][]interface{}, outputFile string) error {
	// 根据文件扩展名确定输出格式
	ext := filepath.Ext(outputFile)
	var bytes []byte
	var err error

	switch strings.ToLower(ext) {
	case ".json":
		bytes, err = json.MarshalIndent(data, "", "  ")
		if err != nil {
			return fmt.Errorf("无法序列化 JSON 数据: %v", err)
		}
	case ".yaml", ".yml":
		bytes, err = yaml.Marshal(data)
		if err != nil {
			return fmt.Errorf("无法序列化 YAML 数据: %v", err)
		}
	default:
		return fmt.Errorf("不支持的输出格式: %s", ext)
	}

	// 创建目录（如果不存在）
	dir := filepath.Dir(outputFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("无法创建目录: %v", err)
	}

	// 写入文件
	if err := ioutil.WriteFile(outputFile, bytes, 0644); err != nil {
		return fmt.Errorf("无法写入文件: %v", err)
	}

	return nil
}

// AddCustomRule 添加自定义生成规则
func (m *TestDataManager) AddCustomRule(pattern string, generator func() interface{}) error {
	return m.Generator.AddRule(pattern, generator)
}

// GetData 获取数据
func (m *TestDataManager) GetData(key string) (interface{}, bool) {
	value, exists := m.Data[key]
	return value, exists
}

// SetData 设置数据
func (m *TestDataManager) SetData(key string, value interface{}) {
	m.Data[key] = value
}
