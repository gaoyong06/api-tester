package types

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"
)

// ValidationResult 表示API验证结果
type ValidationResult struct {
	// 是否通过验证
	Passed bool
	// 失败原因
	FailureReason string
	// 预期状态码
	ExpectedStatus string
	// 实际状态码
	ActualStatus int
	// 响应时间（毫秒）
	ResponseTime int64
	// 响应体
	ResponseBody string
}

// EndpointTestResult 表示单个端点的测试结果
type EndpointTestResult struct {
	// 端点信息（使用指针避免循环导入）
	Endpoint interface{}
	// 验证结果
	Validation *ValidationResult
	// 测试时间
	TestTime time.Time
}

// TestResult 表示测试结果摘要
type TestResult struct {
	// 总测试数
	Total int
	// 通过测试数
	Passed int
	// 失败测试数
	Failed int
	// 测试报告路径
	ReportPath string
	// 测试结果详情
	Results []*EndpointTestResult `json:"results"`
}

// LoadTestResultsFromFile 从文件加载测试结果
func LoadTestResultsFromFile(filePath string) (*TestResult, error) {
	// 读取文件内容
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("无法读取测试结果文件: %v", err)
	}

	// 解析 JSON
	result := &TestResult{}
	if err := json.Unmarshal(data, result); err != nil {
		return nil, fmt.Errorf("无法解析测试结果: %v", err)
	}

	return result, nil
}
