package runner

import (
	"fmt"
	"time"

	"github.com/gaoyong06/api-tester/internal/config"
	"github.com/gaoyong06/api-tester/internal/config/yaml"
	"github.com/gaoyong06/api-tester/internal/parser"
	"github.com/gaoyong06/api-tester/internal/reporter"
	"github.com/gaoyong06/api-tester/internal/scenario"
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
	// 存储所有端点
	allEndpoints := []*parser.Endpoint{}
	
	// 判断是否有多个规范文件
	if len(r.config.SpecFiles) > 0 {
		// 处理多个规范文件
		fmt.Printf("检测到 %d 个 API 规范文件\n", len(r.config.SpecFiles))
		
		for i, specFile := range r.config.SpecFiles {
			fmt.Printf("解析规范文件 [%d/%d]: %s\n", i+1, len(r.config.SpecFiles), specFile)
			
			// 尝试解析 OpenAPI 3.0 格式
			apiDef, err := parser.ParseOpenAPI(specFile)
			if err != nil {
				// 如果失败，尝试解析 Swagger 2.0 格式
				apiDef, err = parser.ParseSwaggerFile(specFile)
				if err != nil {
					return nil, fmt.Errorf("解析规范文件 %s 失败: %v", specFile, err)
				}
			}
			
			// 将端点添加到总列表中
			allEndpoints = append(allEndpoints, apiDef.Endpoints...)
			
			fmt.Printf("  找到 %d 个端点\n", len(apiDef.Endpoints))
		}
	} else if r.config.SpecFile != "" {
		// 处理单个规范文件（向后兼容）
		fmt.Printf("解析规范文件: %s\n", r.config.SpecFile)
		
		// 尝试解析 OpenAPI 3.0 格式
		apiDef, err := parser.ParseOpenAPI(r.config.SpecFile)
		if err != nil {
			// 如果失败，尝试解析 Swagger 2.0 格式
			apiDef, err = parser.ParseSwaggerFile(r.config.SpecFile)
			if err != nil {
				return nil, fmt.Errorf("解析规范文件失败: %v", err)
			}
		}
		
		// 将端点添加到总列表中
		allEndpoints = append(allEndpoints, apiDef.Endpoints...)
		
		fmt.Printf("基础 URL: %s\n", r.config.BaseURL)
		fmt.Printf("端点数量: %d\n\n", len(apiDef.Endpoints))
	} else {
		return nil, fmt.Errorf("未指定规范文件")
	}
	
	// 打印总端点数量
	fmt.Printf("总端点数量: %d\n\n", len(allEndpoints))

	// 创建一个合并的 API 定义用于场景管理器
	mergedApiDef := &parser.APIDefinition{
		Title:     "合并的 API 定义",
		Version:   "1.0",
		Endpoints: allEndpoints,
	}

	// 检查是否有测试场景定义
	if r.config.YamlConfig != nil && len(r.config.YamlConfig.Scenarios) > 0 {
		// 使用场景管理器运行测试场景
		fmt.Printf("检测到 %d 个测试场景，使用场景模式运行测试\n\n", len(r.config.YamlConfig.Scenarios))
		
		// 将 []yaml.Scenario 转换为 []*yaml.Scenario
		scenarios := make([]*yaml.Scenario, 0, len(r.config.YamlConfig.Scenarios))
		for i := range r.config.YamlConfig.Scenarios {
			scenarios = append(scenarios, &r.config.YamlConfig.Scenarios[i])
		}
		
		// 创建场景管理器，传递配置对象
		scenarioManager := scenario.NewManager(scenarios, mergedApiDef, r.client, r.config.YamlConfig)
		
		// 运行所有场景
		scenarioResults, err := scenarioManager.RunAllScenarios()
		if err != nil {
			return nil, fmt.Errorf("运行测试场景失败: %v", err)
		}
		
		// 保存测试结果
		r.results = append(r.results, scenarioResults...)
	} else {
		// 如果没有定义测试场景，则运行所有端点测试（兼容旧版本）
		fmt.Printf("未检测到测试场景，使用端点模式运行测试\n\n")
		
		// 运行所有端点测试
		for i, endpoint := range allEndpoints {
			fmt.Printf("[%d/%d] 测试 %s %s... ", i+1, len(allEndpoints), endpoint.Method, endpoint.Path)

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
	}

	// 生成测试报告
	reportPath, err := reporter.GenerateReport(mergedApiDef, r.results, r.config.OutputDir)
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
		Results:    r.results, // 添加测试结果详情
	}, nil
}
