package machine

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/gaoyong06/api-tester/internal/parser"
	"github.com/gaoyong06/api-tester/internal/types"
)

// MachineReport 表示机器可读的测试报告
type MachineReport struct {
	// 基本信息
	Metadata struct {
		// API标题
		Title string `json:"title" xml:"title"`
		// API版本
		Version string `json:"version" xml:"version"`
		// 测试时间
		Timestamp string `json:"timestamp" xml:"timestamp"`
		// 测试环境
		Environment string `json:"environment" xml:"environment"`
		// 基础URL
		BaseURL string `json:"base_url" xml:"base_url"`
	} `json:"metadata" xml:"metadata"`

	// 统计数据
	Summary struct {
		// 总测试数
		Total int `json:"total" xml:"total"`
		// 通过测试数
		Passed int `json:"passed" xml:"passed"`
		// 失败测试数
		Failed int `json:"failed" xml:"failed"`
		// 通过率
		PassRate float64 `json:"pass_rate" xml:"pass_rate"`
		// 总响应时间
		TotalResponseTime int64 `json:"total_response_time" xml:"total_response_time"`
		// 平均响应时间
		AvgResponseTime float64 `json:"avg_response_time" xml:"avg_response_time"`
		// 最小响应时间
		MinResponseTime int64 `json:"min_response_time" xml:"min_response_time"`
		// 最大响应时间
		MaxResponseTime int64 `json:"max_response_time" xml:"max_response_time"`
		// 百分位响应时间
		Percentiles map[string]int64 `json:"percentiles" xml:"percentiles"`
	} `json:"summary" xml:"summary"`

	// 详细测试结果
	Results []TestResult `json:"results" xml:"results>result"`

	// 错误分析
	ErrorAnalysis struct {
		// 常见错误
		CommonErrors []string `json:"common_errors" xml:"common_errors>error"`
		// 错误分布
		ErrorDistribution map[string]int `json:"error_distribution" xml:"error_distribution"`
		// 状态码分布
		StatusCodeDistribution map[string]int `json:"status_code_distribution" xml:"status_code_distribution"`
	} `json:"error_analysis" xml:"error_analysis"`

	// 测试覆盖率
	Coverage struct {
		// 已测试端点数
		EndpointsTested int `json:"endpoints_tested" xml:"endpoints_tested"`
		// 总端点数
		TotalEndpoints int `json:"total_endpoints" xml:"total_endpoints"`
		// 覆盖率百分比
		CoveragePercent float64 `json:"coverage_percent" xml:"coverage_percent"`
		// 已测试路径
		TestedPaths []string `json:"tested_paths" xml:"tested_paths>path"`
		// 未测试路径
		UntestedPaths []string `json:"untested_paths" xml:"untested_paths>path"`
	} `json:"coverage" xml:"coverage"`
}

// TestResult 表示单个测试结果
type TestResult struct {
	// 端点信息
	Endpoint struct {
		Path        string   `json:"path" xml:"path"`
		Method      string   `json:"method" xml:"method"`
		OperationID string   `json:"operation_id" xml:"operation_id"`
		Description string   `json:"description" xml:"description"`
		Tags        []string `json:"tags" xml:"tags>tag"`
	} `json:"endpoint" xml:"endpoint"`

	// 验证结果
	Validation struct {
		Passed         bool   `json:"passed" xml:"passed"`
		FailureReason  string `json:"failure_reason,omitempty" xml:"failure_reason,omitempty"`
		ExpectedStatus string `json:"expected_status,omitempty" xml:"expected_status,omitempty"`
		ActualStatus   int    `json:"actual_status" xml:"actual_status"`
		ResponseTime   int64  `json:"response_time" xml:"response_time"`
		ResponseBody   string `json:"response_body,omitempty" xml:"response_body,omitempty"`
	} `json:"validation" xml:"validation"`

	// 测试时间
	Timestamp string `json:"timestamp" xml:"timestamp"`
}

// GenerateReport 生成机器可读的测试报告
func GenerateReport(apiDef *parser.APIDefinition, results []*types.EndpointTestResult, outputDir string, format string) (string, error) {
	// 创建输出目录
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("无法创建输出目录: %v", err)
	}

	// 准备报告数据
	report := prepareMachineReport(apiDef, results)

	// 生成报告文件名
	extension := ".json"
	if format == "xml" {
		extension = ".xml"
	}
	reportFileName := fmt.Sprintf("api-test-report-%s%s", time.Now().Format("20060102-150405"), extension)
	reportPath := filepath.Join(outputDir, reportFileName)

	// 创建报告文件
	reportFile, err := os.Create(reportPath)
	if err != nil {
		return "", fmt.Errorf("无法创建报告文件: %v", err)
	}
	defer reportFile.Close()

	// 根据格式生成报告
	var data []byte
	switch format {
	case "xml":
		data, err = xml.MarshalIndent(report, "", "  ")
	default: // 默认使用 JSON
		data, err = json.MarshalIndent(report, "", "  ")
	}

	if err != nil {
		return "", fmt.Errorf("无法序列化报告数据: %v", err)
	}

	// 写入文件
	if err := ioutil.WriteFile(reportPath, data, 0644); err != nil {
		return "", fmt.Errorf("无法写入报告文件: %v", err)
	}

	return reportPath, nil
}

// prepareMachineReport 准备机器可读的报告数据
func prepareMachineReport(apiDef *parser.APIDefinition, results []*types.EndpointTestResult) *MachineReport {
	// 创建报告
	report := &MachineReport{}

	// 设置元数据
	report.Metadata.Title = apiDef.Title
	report.Metadata.Version = apiDef.Version
	report.Metadata.Timestamp = time.Now().Format(time.RFC3339)
	report.Metadata.Environment = "development" // 可以从配置中获取

	// 计算统计数据
	total := len(results)
	passed := 0
	failed := 0
	totalResponseTime := int64(0)
	minResponseTime := int64(0)
	maxResponseTime := int64(0)
	statusCodeDistribution := make(map[string]int)
	errorDistribution := make(map[string]int)

	if total > 0 {
		minResponseTime = results[0].Validation.ResponseTime
	}

	// 处理测试结果
	report.Results = make([]TestResult, 0, total)
	for i, result := range results {
		// 更新统计数据
		if result.Validation.Passed {
			passed++
		} else {
			failed++
			// 记录错误分布
			errorDistribution[result.Validation.FailureReason]++
		}

		// 更新响应时间统计
		responseTime := result.Validation.ResponseTime
		totalResponseTime += responseTime

		if i == 0 || responseTime < minResponseTime {
			minResponseTime = responseTime
		}
		if responseTime > maxResponseTime {
			maxResponseTime = responseTime
		}

		// 更新状态码分布
		statusCode := fmt.Sprintf("%d", result.Validation.ActualStatus)
		statusCodeDistribution[statusCode]++

		// 添加详细测试结果
		testResult := TestResult{}
		
		// 设置端点信息
		endpoint := result.Endpoint.(*parser.Endpoint) // 类型断言
		testResult.Endpoint.Path = endpoint.Path
		testResult.Endpoint.Method = endpoint.Method
		testResult.Endpoint.OperationID = endpoint.OperationID
		testResult.Endpoint.Description = endpoint.Description
		testResult.Endpoint.Tags = endpoint.Tags

		// 设置验证结果
		testResult.Validation.Passed = result.Validation.Passed
		testResult.Validation.FailureReason = result.Validation.FailureReason
		testResult.Validation.ExpectedStatus = result.Validation.ExpectedStatus
		testResult.Validation.ActualStatus = result.Validation.ActualStatus
		testResult.Validation.ResponseTime = result.Validation.ResponseTime
		testResult.Validation.ResponseBody = result.Validation.ResponseBody

		// 设置测试时间
		testResult.Timestamp = result.TestTime.Format(time.RFC3339)

		report.Results = append(report.Results, testResult)
	}

	// 设置摘要信息
	report.Summary.Total = total
	report.Summary.Passed = passed
	report.Summary.Failed = failed
	report.Summary.TotalResponseTime = totalResponseTime
	report.Summary.MinResponseTime = minResponseTime
	report.Summary.MaxResponseTime = maxResponseTime

	// 计算通过率和平均响应时间
	if total > 0 {
		report.Summary.PassRate = float64(passed) / float64(total) * 100
		report.Summary.AvgResponseTime = float64(totalResponseTime) / float64(total)
	}

	// 设置错误分析
	report.ErrorAnalysis.ErrorDistribution = errorDistribution
	report.ErrorAnalysis.StatusCodeDistribution = statusCodeDistribution

	// 提取常见错误
	for err, count := range errorDistribution {
		if count > 1 { // 出现超过一次的错误被视为常见错误
			report.ErrorAnalysis.CommonErrors = append(report.ErrorAnalysis.CommonErrors, err)
		}
	}

	// 计算测试覆盖率
	allPaths := make(map[string]bool)
	testedPaths := make(map[string]bool)

	// 收集所有端点
	for _, endpoint := range apiDef.Endpoints {
		pathKey := fmt.Sprintf("%s %s", endpoint.Method, endpoint.Path)
		allPaths[pathKey] = true
	}

	// 收集已测试端点
	for _, result := range results {
		endpoint := result.Endpoint.(*parser.Endpoint) // 类型断言
		pathKey := fmt.Sprintf("%s %s", endpoint.Method, endpoint.Path)
		testedPaths[pathKey] = true
	}

	// 设置覆盖率信息
	report.Coverage.TotalEndpoints = len(allPaths)
	report.Coverage.EndpointsTested = len(testedPaths)

	if report.Coverage.TotalEndpoints > 0 {
		report.Coverage.CoveragePercent = float64(report.Coverage.EndpointsTested) / float64(report.Coverage.TotalEndpoints) * 100
	}

	// 收集已测试和未测试的路径
	for path := range testedPaths {
		report.Coverage.TestedPaths = append(report.Coverage.TestedPaths, path)
	}

	for path := range allPaths {
		if !testedPaths[path] {
			report.Coverage.UntestedPaths = append(report.Coverage.UntestedPaths, path)
		}
	}

	return report
}

// JUnitTestSuite JUnit测试套件结构
type JUnitTestSuite struct {
	XMLName    xml.Name        `xml:"testsuite"`
	Name       string          `xml:"name,attr"`
	Tests      int             `xml:"tests,attr"`
	Failures   int             `xml:"failures,attr"`
	Errors     int             `xml:"errors,attr"`
	Skipped    int             `xml:"skipped,attr"`
	Time       float64         `xml:"time,attr"`
	Timestamp  string          `xml:"timestamp,attr"`
	Properties []JUnitProperty `xml:"properties>property,omitempty"`
	TestCases  []JUnitTestCase `xml:"testcase"`
}

// JUnitProperty JUnit属性结构
type JUnitProperty struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

// JUnitTestCase JUnit测试用例结构
type JUnitTestCase struct {
	Name      string       `xml:"name,attr"`
	Classname string       `xml:"classname,attr"`
	Time      float64      `xml:"time,attr"`
	Failure   *JUnitFailure `xml:"failure,omitempty"`
}

// JUnitFailure JUnit失败信息结构
type JUnitFailure struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
	Content string `xml:",chardata"`
}

// GenerateJUnitReport 生成 JUnit 格式的测试报告
func GenerateJUnitReport(apiDef *parser.APIDefinition, results []*types.EndpointTestResult, outputDir string) (string, error) {

	// 创建输出目录
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("无法创建输出目录: %v", err)
	}

	// 准备 JUnit 报告数据
	testSuite := JUnitTestSuite{
		Name:      apiDef.Title,
		Tests:     len(results),
		Failures:  0,
		Errors:    0,
		Skipped:   0,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// 添加属性
	testSuite.Properties = append(testSuite.Properties, JUnitProperty{
		Name:  "version",
		Value: apiDef.Version,
	})

	// 计算总时间和失败数
	totalTime := 0.0
	for _, result := range results {
		endpoint := result.Endpoint.(*parser.Endpoint) // 类型断言
		testCase := JUnitTestCase{
			Name:      fmt.Sprintf("%s %s", endpoint.Method, endpoint.Path),
			Classname: endpoint.OperationID,
			Time:      float64(result.Validation.ResponseTime) / 1000.0, // 转换为秒
		}

		totalTime += testCase.Time

		// 如果测试失败，添加失败信息
		if !result.Validation.Passed {
			testSuite.Failures++
			testCase.Failure = &JUnitFailure{
				Message: result.Validation.FailureReason,
				Type:    "AssertionError",
				Content: fmt.Sprintf("Expected status: %s, Actual status: %d", 
					result.Validation.ExpectedStatus, result.Validation.ActualStatus),
			}
		}

		testSuite.TestCases = append(testSuite.TestCases, testCase)
	}

	testSuite.Time = totalTime

	// 生成报告文件名
	reportFileName := fmt.Sprintf("junit-report-%s.xml", time.Now().Format("20060102-150405"))
	reportPath := filepath.Join(outputDir, reportFileName)

	// 创建报告文件
	reportFile, err := os.Create(reportPath)
	if err != nil {
		return "", fmt.Errorf("无法创建报告文件: %v", err)
	}
	defer reportFile.Close()

	// 序列化为 XML
	data, err := xml.MarshalIndent(testSuite, "", "  ")
	if err != nil {
		return "", fmt.Errorf("无法序列化报告数据: %v", err)
	}

	// 添加 XML 头
	data = append([]byte(xml.Header), data...)

	// 写入文件
	if err := ioutil.WriteFile(reportPath, data, 0644); err != nil {
		return "", fmt.Errorf("无法写入报告文件: %v", err)
	}

	return reportPath, nil
}
