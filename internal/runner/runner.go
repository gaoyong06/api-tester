package runner

import (
	"fmt"
	"time"

	"github.com/gaoyong06/api-tester/internal/config"
	"github.com/gaoyong06/api-tester/internal/parser"
	"github.com/gaoyong06/api-tester/internal/reporter"
	"github.com/gaoyong06/api-tester/internal/types"
	"github.com/gaoyong06/api-tester/internal/validator"
	"github.com/gaoyong06/api-tester/pkg/client"
)

// Runner 是API测试运行器
type Runner struct {
	// 配置
	config *config.Config
	// API客户端
	client *client.APIClient
	// 测试结果
	results []*types.EndpointTestResult
}

// NewRunner 创建一个新的测试运行器
func NewRunner(cfg *config.Config) *Runner {
	return &Runner{
		config:  cfg,
		client:  client.NewAPIClient(cfg.BaseURL, cfg.Headers, cfg.Timeout, cfg.Verbose, cfg.RequestBodies),
		results: make([]*types.EndpointTestResult, 0),
	}
}

// Run 运行API测试
func (r *Runner) Run() (*types.TestResult, error) {
	// 解析OpenAPI规范
	apiDef, err := parser.ParseOpenAPI(r.config.SpecFile)
	if err != nil {
		return nil, fmt.Errorf("解析OpenAPI规范失败: %v", err)
	}

	// 打印测试信息
	fmt.Printf("开始测试 API: %s (版本 %s)\n", apiDef.Title, apiDef.Version)
	fmt.Printf("基础 URL: %s\n", r.config.BaseURL)
	fmt.Printf("端点数量: %d\n\n", len(apiDef.Endpoints))

	// 运行所有端点测试
	for i, endpoint := range apiDef.Endpoints {
		fmt.Printf("[%d/%d] 测试 %s %s... ", i+1, len(apiDef.Endpoints), endpoint.Method, endpoint.Path)

		// 提取路径参数和查询参数
		pathParams := client.ExtractPathParams(endpoint)
		// 使用配置中的路径参数覆盖默认值
		for name, value := range r.config.PathParams {
			pathParams[name] = value
		}
		queryParams := client.ExtractQueryParams(endpoint)

		// 发送请求
		response, err := r.client.SendRequest(endpoint, pathParams, queryParams)
		if err != nil {
			fmt.Printf("发送请求出错: %v\n", err)
			continue
		}

		// 验证响应
		validationResults := validator.ValidateResponse(endpoint, response)

		// 保存测试结果
		r.results = append(r.results, &types.EndpointTestResult{
			Endpoint:   endpoint,
			Validation: validationResults,
			TestTime:   time.Now(),
		})

		// 打印测试结果
		if validationResults.Passed {
			fmt.Printf("通过 (%d ms)\n", validationResults.ResponseTime)
		} else {
			fmt.Printf("失败: %s\n", validationResults.FailureReason)
		}
	}

	// 生成测试报告
	reportPath, err := reporter.GenerateReport(apiDef, r.results, r.config.OutputDir)
	if err != nil {
		return nil, fmt.Errorf("生成测试报告失败: %v", err)
	}

	// 统计测试结果
	total := len(r.results)
	passed := 0
	failed := 0

	for _, result := range r.results {
		if result.Validation.Passed {
			passed++
		} else {
			failed++
		}
	}

	return &types.TestResult{
		Total:      total,
		Passed:     passed,
		Failed:     failed,
		ReportPath: reportPath,
	}, nil
}
