package scenario

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gaoyong06/api-tester/internal/config/yaml"
	"github.com/gaoyong06/api-tester/internal/parser"
	"github.com/gaoyong06/api-tester/internal/types"
	"github.com/gaoyong06/api-tester/pkg/client"
	"github.com/tidwall/gjson"
)

// Manager 测试场景管理器
type Manager struct {
	// 场景列表
	Scenarios []*yaml.Scenario
	// API 定义
	APIDefinition *parser.APIDefinition
	// API 客户端
	Client *client.APIClient
	// 上下文数据
	Context *Context
}

// Context 测试上下文
type Context struct {
	// 变量存储
	Variables map[string]interface{}
	// 测试结果
	Results map[string]*types.EndpointTestResult
	// 步骤状态
	StepStatus map[string]bool
}

// NewManager 创建场景管理器
func NewManager(scenarios []*yaml.Scenario, apiDef *parser.APIDefinition, client *client.APIClient) *Manager {
	return &Manager{
		Scenarios:     scenarios,
		APIDefinition: apiDef,
		Client:        client,
		Context: &Context{
			Variables:  make(map[string]interface{}),
			Results:    make(map[string]*types.EndpointTestResult),
			StepStatus: make(map[string]bool),
		},
	}
}

// RunScenario 运行指定场景
func (m *Manager) RunScenario(scenarioName string) ([]*types.EndpointTestResult, error) {
	// 查找场景
	var scenario *yaml.Scenario
	for _, s := range m.Scenarios {
		if s.Name == scenarioName {
			scenario = s
			break
		}
	}

	if scenario == nil {
		return nil, fmt.Errorf("未找到场景: %s", scenarioName)
	}

	// 运行场景
	return m.runScenarioSteps(scenario)
}

// RunAllScenarios 运行所有场景
func (m *Manager) RunAllScenarios() ([]*types.EndpointTestResult, error) {
	var allResults []*types.EndpointTestResult

	for _, scenario := range m.Scenarios {
		results, err := m.runScenarioSteps(scenario)
		if err != nil {
			return nil, err
		}

		allResults = append(allResults, results...)
	}

	return allResults, nil
}

// runScenarioSteps 运行场景步骤
func (m *Manager) runScenarioSteps(scenario *yaml.Scenario) ([]*types.EndpointTestResult, error) {
	var results []*types.EndpointTestResult

	fmt.Printf("运行场景: %s\n", scenario.Name)
	if scenario.Description != "" {
		fmt.Printf("描述: %s\n", scenario.Description)
	}

	// 重置步骤状态
	m.Context.StepStatus = make(map[string]bool)

	// 运行所有步骤
	for _, step := range scenario.Steps {
		// 检查依赖是否已完成
		if !m.checkDependencies(step) {
			fmt.Printf("跳过步骤 %s，因为依赖未满足\n", step.Name)
			continue
		}

		fmt.Printf("执行步骤: %s (%s %s)\n", step.Name, step.Method, step.Endpoint)

		// 查找端点
		endpoint := m.findEndpoint(step.Endpoint, step.Method)
		if endpoint == nil {
			fmt.Printf("警告: 未在 API 定义中找到端点 %s %s\n", step.Method, step.Endpoint)
			// 创建一个临时端点
			endpoint = &parser.Endpoint{
				Path:        step.Endpoint,
				Method:      step.Method,
				OperationID: step.Name,
				Description: step.Name,
			}
		}

		// 处理变量替换
		pathParams, queryParams, requestBody := m.processVariables(step)

		// 发送请求
		response, err := m.Client.SendRequest(endpoint, pathParams, queryParams, []byte(requestBody))
		if err != nil {
			fmt.Printf("请求失败: %v\n", err)
			continue
		}

		// 提取变量
		if len(step.Extract) > 0 && len(response.Body) > 0 {
			m.extractVariables(step.Extract, response.Body)
		}

		// 创建测试结果
		result := &types.EndpointTestResult{
			Endpoint: endpoint,
			Validation: &types.ValidationResult{
				Passed:        response.StatusCode >= 200 && response.StatusCode < 300,
				ActualStatus:  response.StatusCode,
				ResponseTime:  response.ResponseTime,
				ResponseBody:  string(response.Body),
				FailureReason: "",
			},
			TestTime: time.Now(),
		}

		// 如果请求失败，设置失败原因
		if !result.Validation.Passed {
			result.Validation.FailureReason = fmt.Sprintf("状态码 %d 不在成功范围内", response.StatusCode)
		}

		// 保存结果
		m.Context.Results[step.Name] = result
		results = append(results, result)

		// 标记步骤已完成
		m.Context.StepStatus[step.Name] = true

		// 打印结果
		if result.Validation.Passed {
			fmt.Printf("步骤成功: %s (%d ms)\n", step.Name, result.Validation.ResponseTime)
		} else {
			fmt.Printf("步骤失败: %s - %s\n", step.Name, result.Validation.FailureReason)
		}
	}

	return results, nil
}

// checkDependencies 检查依赖是否已满足
func (m *Manager) checkDependencies(step *yaml.Step) bool {
	if len(step.Dependencies) == 0 {
		return true
	}

	for _, dep := range step.Dependencies {
		if !m.Context.StepStatus[dep] {
			return false
		}
	}

	return true
}

// findEndpoint 查找端点
func (m *Manager) findEndpoint(path, method string) *parser.Endpoint {
	for _, endpoint := range m.APIDefinition.Endpoints {
		if endpoint.Path == path && strings.EqualFold(endpoint.Method, method) {
			return endpoint
		}
	}

	return nil
}

// processVariables 处理变量替换
func (m *Manager) processVariables(step *yaml.Step) (map[string]string, map[string]string, string) {
	// 处理路径参数
	pathParams := make(map[string]string)
	// 处理查询参数
	queryParams := make(map[string]string)
	// 处理请求体
	requestBody := step.RequestBody

	// 替换请求体中的变量
	if requestBody != "" {
		requestBody = m.replaceVariables(requestBody)
	}

	return pathParams, queryParams, requestBody
}

// replaceVariables 替换字符串中的变量
func (m *Manager) replaceVariables(input string) string {
	// 匹配 {{.variable}} 格式的变量
	result := input
	for key, value := range m.Context.Variables {
		placeholder := "{{." + key + "}}"
		strValue := fmt.Sprintf("%v", value)
		result = strings.ReplaceAll(result, placeholder, strValue)
	}

	return result
}

// extractVariables 从响应中提取变量
func (m *Manager) extractVariables(extractors map[string]string, responseBody []byte) {
	// 检查响应是否是有效的 JSON
	if !json.Valid(responseBody) {
		fmt.Println("响应不是有效的 JSON，无法提取变量")
		return
	}

	// 使用 gjson 提取变量
	for name, path := range extractors {
		result := gjson.GetBytes(responseBody, path)
		if result.Exists() {
			m.Context.Variables[name] = result.Value()
			fmt.Printf("提取变量: %s = %v\n", name, result.Value())
		} else {
			fmt.Printf("警告: 无法从路径 %s 提取变量 %s\n", path, name)
		}
	}
}

// GetVariable 获取变量值
func (m *Manager) GetVariable(name string) (interface{}, bool) {
	value, exists := m.Context.Variables[name]
	return value, exists
}

// SetVariable 设置变量值
func (m *Manager) SetVariable(name string, value interface{}) {
	m.Context.Variables[name] = value
}

// GetStepResult 获取步骤结果
func (m *Manager) GetStepResult(stepName string) (*types.EndpointTestResult, bool) {
	result, exists := m.Context.Results[stepName]
	return result, exists
}
