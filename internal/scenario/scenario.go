package scenario

import (
	"encoding/json"
	"fmt"
	"regexp"
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
func NewManager(scenarios []*yaml.Scenario, apiDef *parser.APIDefinition, client *client.APIClient, config *yaml.Config) *Manager {
	// 创建上下文变量存储
	variables := make(map[string]interface{})

	// 如果配置中有默认值，加载到上下文变量中
	if config != nil && len(config.DefaultValues) > 0 {
		// 将默认值添加到 default_values 变量中
		defaultValuesMap := make(map[string]interface{})
		for k, v := range config.DefaultValues {
			defaultValuesMap[k] = v
		}
		variables["default_values"] = defaultValuesMap
		fmt.Printf("从配置中加载了 %d 个默认值\n", len(config.DefaultValues))
	}

	return &Manager{
		Scenarios:     scenarios,
		APIDefinition: apiDef,
		Client:        client,
		Context: &Context{
			Variables:  variables,
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
		if !m.checkDependencies(&step) {
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
		pathParams, queryParams, requestBody := m.processVariables(&step)

		// 更新端点路径，使用处理后的路径
		endpoint.Path = step.Endpoint

		// 发送请求，包括请求体
		response, err := m.Client.SendRequest(endpoint, pathParams, queryParams, requestBody)
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
	var requestBodyStr string

	// 处理路径参数
	if step.PathParams != nil {
		for key, value := range step.PathParams {
			// 替换变量
			// 先处理Go模板语法
			value = m.replaceGoTemplateVars(value)
			// 再处理普通变量
			pathParams[key] = m.replaceVariables(value)
			fmt.Printf("路径参数: %s = %s\n", key, pathParams[key])
		}
	}

	// 处理查询参数
	if step.QueryParams != nil {
		for key, value := range step.QueryParams {
			// 替换变量
			// 先处理Go模板语法
			value = m.replaceGoTemplateVars(value)
			// 再处理普通变量
			queryParams[key] = m.replaceVariables(value)
			fmt.Printf("查询参数: %s = %s\n", key, queryParams[key])
		}
	}

	// 处理端点路径中的变量占位符
	if step.Endpoint != "" {
		// 匹配并替换路径中的参数占位符 {param_name}
		endpoint := step.Endpoint

		// 查找所有占位符 {param_name}
		re := regexp.MustCompile(`\{([^}]+)\}`)
		matches := re.FindAllStringSubmatch(endpoint, -1)

		// 记录占位符替换过程
		fmt.Printf("开始处理端点: %s\n", endpoint)
		fmt.Printf("变量替换优先级: 1.路径参数 > 2.上下文变量 > 3.相似名称变量 > 4.默认值\n")

		for _, match := range matches {
			if len(match) > 1 {
				paramName := match[1]
				placeholder := match[0] // 完整的占位符，如 {event_id}

				// 记录当前处理的占位符
				fmt.Printf("处理占位符: %s (参数名: %s)\n", placeholder, paramName)

				// 1. 首先检查 path_params 中是否有对应的值
				if value, exists := pathParams[paramName]; exists {
					endpoint = strings.ReplaceAll(endpoint, placeholder, value)
					fmt.Printf("  [优先级1] 替换占位符 %s 为路径参数值: %s\n", placeholder, value)
					continue
				} else {
					fmt.Printf("  [优先级1] 路径参数中未找到 %s 的值\n", paramName)
				}

				// 2. 检查上下文变量
				if value, exists := m.Context.Variables[paramName]; exists {
					strValue := fmt.Sprintf("%v", value)
					endpoint = strings.ReplaceAll(endpoint, placeholder, strValue)
					fmt.Printf("  [优先级2] 替换占位符 %s 为上下文变量值: %s\n", placeholder, strValue)

					// 同时添加到路径参数中，以便后续处理
					pathParams[paramName] = strValue
					continue
				} else {
					fmt.Printf("  [优先级2] 上下文变量中未找到 %s 的值\n", paramName)
				}

				// 3. 如果还没有替换，尝试使用相似名称的变量
				// 例如，将 event_id 转换为 eventId 或 eventID
				alternativeNames := []string{
					toCamelCase(paramName),
					toSnakeCase(paramName),
				}

				var altFound bool
				for _, altName := range alternativeNames {
					if altName == paramName {
						continue // 跳过相同的名称
					}

					fmt.Printf("  [优先级3] 尝试相似名称: %s\n", altName)

					if value, exists := m.Context.Variables[altName]; exists {
						strValue := fmt.Sprintf("%v", value)
						endpoint = strings.ReplaceAll(endpoint, placeholder, strValue)
						fmt.Printf("  [优先级3] 替换占位符 %s 为相似名称变量 %s 的值: %s\n", placeholder, altName, strValue)

						// 同时添加到路径参数中，以便后续处理
						pathParams[paramName] = strValue
						altFound = true
						break
					}
				}

				if !altFound {
					fmt.Printf("  [优先级3] 未找到相似名称的变量\n")

					// 4. 使用默认值替换
					// 从配置中获取默认值
					defaultValues := m.getDefaultValues()

					var defaultFound bool
					for defName, defValue := range defaultValues {
						if paramName == defName {
							endpoint = strings.ReplaceAll(endpoint, placeholder, defValue)
							fmt.Printf("  [优先级4] 替换占位符 %s 为默认值: %s\n", placeholder, defValue)

							// 同时添加到路径参数中，以便后续处理
							pathParams[paramName] = defValue
							defaultFound = true
							break
						}
					}

					if !defaultFound {
						fmt.Printf("  [优先级4] 未找到默认值，占位符 %s 将保持不变\n", placeholder)
					}
				}
			}
		}

		// 如果还有未替换的参数，输出警告
		if strings.Contains(endpoint, "{") && strings.Contains(endpoint, "}") {
			fmt.Printf("警告: 端点 %s 仍然包含未替换的参数占位符\n", endpoint)
		}

		// 更新步骤的端点路径
		step.Endpoint = endpoint
		fmt.Printf("处理后的端点: %s\n", endpoint)
	}

	// 处理不同类型的 RequestBody
	switch body := step.RequestBody.(type) {
	case string:
		// 如果是字符串，直接替换变量
		if body != "" {
			requestBodyStr = m.replaceVariables(body)
		}
	case map[string]interface{}, map[interface{}]interface{}:
		// 如果是对象，先转成 JSON 字符串
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			fmt.Printf("警告: 无法将请求体转换为 JSON: %v\n", err)
		} else {
			requestBodyStr = m.replaceVariables(string(jsonBytes))
		}
	case nil:
		// 如果为空，不需要处理
		requestBodyStr = ""
	default:
		// 其他类型，尝试转成 JSON
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			fmt.Printf("警告: 无法将请求体类型 %T 转换为 JSON: %v\n", body, err)
		} else {
			requestBodyStr = m.replaceVariables(string(jsonBytes))
		}
	}

	return pathParams, queryParams, requestBodyStr
}

// replaceVariables 替换字符串中的变量
func (m *Manager) replaceVariables(input string) string {
	// 先处理 Go 模板语法 {{.variable}}
	result := input

	// 1. 使用正则表达式匹配 Go 模板语法
	re := regexp.MustCompile(`{{\s*\.([a-zA-Z0-9_]+)\s*}}`)
	matches := re.FindAllStringSubmatch(result, -1)

	for _, match := range matches {
		if len(match) > 1 {
			placeholder := match[0] // {{.varName}}
			varName := match[1]     // varName

			// 检查上下文变量
			if value, exists := m.Context.Variables[varName]; exists {
				strValue := fmt.Sprintf("%v", value)
				result = strings.ReplaceAll(result, placeholder, strValue)
				fmt.Printf("  替换Go模板变量 %s 为上下文变量值: %s\n", placeholder, strValue)
				continue
			}

			// 尝试相似名称
			alternativeNames := []string{
				toCamelCase(varName),
				toSnakeCase(varName),
			}

			var replaced bool
			for _, altName := range alternativeNames {
				if altName == varName {
					continue // 跳过相同的名称
				}

				if value, exists := m.Context.Variables[altName]; exists {
					strValue := fmt.Sprintf("%v", value)
					result = strings.ReplaceAll(result, placeholder, strValue)
					fmt.Printf("  替换Go模板变量 %s 为相似名称变量 %s 的值: %s\n", placeholder, altName, strValue)
					replaced = true
					break
				}
			}

			// 使用默认值
			if !replaced {
				// 对于常见的参数名称，使用默认值
				defaultValues := map[string]string{
					"id":              "1",
					"event_id":        "87",
					"eventId":         "87",
					"table_id":        "1",
					"tableId":         "1",
					"guest_id":        "1",
					"guestId":         "1",
					"seat_id":         "1",
					"seatId":          "1",
					"task_id":         "1",
					"taskId":          "1",
					"group_id":        "1",
					"groupId":         "1",
					"relationship_id": "1",
					"relationshipId":  "1",
					"template_id":     "1",
					"templateId":      "1",
					"token":           "test-token",
				}

				for defName, defValue := range defaultValues {
					if varName == defName {
						result = strings.ReplaceAll(result, placeholder, defValue)
						fmt.Printf("  替换Go模板变量 %s 为默认值: %s\n", placeholder, defValue)
						break
					}
				}
			}
		}
	}

	// 2. 处理 {variable} 格式的变量
	re = regexp.MustCompile(`\{([^{}]+)\}`)
	matches = re.FindAllStringSubmatch(result, -1)

	for _, match := range matches {
		if len(match) > 1 {
			placeholder := match[0] // {varName}
			varName := match[1]     // varName

			// 检查上下文变量
			if value, exists := m.Context.Variables[varName]; exists {
				strValue := fmt.Sprintf("%v", value)
				result = strings.ReplaceAll(result, placeholder, strValue)
				fmt.Printf("  替换占位符 %s 为上下文变量值: %s\n", placeholder, strValue)
				continue
			}

			// 尝试相似名称
			alternativeNames := []string{
				toCamelCase(varName),
				toSnakeCase(varName),
			}

			var replaced bool
			for _, altName := range alternativeNames {
				if altName == varName {
					continue // 跳过相同的名称
				}

				if value, exists := m.Context.Variables[altName]; exists {
					strValue := fmt.Sprintf("%v", value)
					result = strings.ReplaceAll(result, placeholder, strValue)
					fmt.Printf("  替换占位符 %s 为相似名称变量 %s 的值: %s\n", placeholder, altName, strValue)
					replaced = true
					break
				}
			}

			// 使用默认值
			if !replaced {
				// 对于常见的参数名称，使用默认值
				defaultValues := map[string]string{
					"id":              "1",
					"event_id":        "87",
					"eventId":         "87",
					"table_id":        "1",
					"tableId":         "1",
					"guest_id":        "1",
					"guestId":         "1",
					"seat_id":         "1",
					"seatId":          "1",
					"task_id":         "1",
					"taskId":          "1",
					"group_id":        "1",
					"groupId":         "1",
					"relationship_id": "1",
					"relationshipId":  "1",
					"template_id":     "1",
					"templateId":      "1",
					"token":           "test-token",
				}

				for defName, defValue := range defaultValues {
					if varName == defName {
						result = strings.ReplaceAll(result, placeholder, defValue)
						fmt.Printf("  替换占位符 %s 为默认值: %s\n", placeholder, defValue)
						break
					}
				}
			}
		}
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

	// 打印响应体结构以便调试
	fmt.Println("响应体结构:", string(responseBody))

	// 使用 gjson 提取变量
	for name, path := range extractors {
		// 打印当前尝试提取的路径
		fmt.Printf("尝试从路径 %s 提取变量 %s\n", path, name)

		// 标准化路径格式（确保路径以 $ 开头）
		if !strings.HasPrefix(path, "$") {
			path = "$" + path
		}

		// 使用 gjson 提取变量
		result := gjson.GetBytes(responseBody, path)

		// 打印提取结果
		fmt.Printf("  路径 %s 的提取结果存在: %v\n", path, result.Exists())

		if result.Exists() {
			m.Context.Variables[name] = result.Value()
			fmt.Printf("  成功提取变量: %s = %v\n", name, result.Value())
		} else {
			// 如果路径不存在，尝试不同的路径格式
			// 尝试去除路径中的 $. 前缀
			cleanPath := strings.TrimPrefix(path, "$.")
			result = gjson.GetBytes(responseBody, cleanPath)

			fmt.Printf("  尝试去除 $. 前缀后的路径 %s 的提取结果存在: %v\n", cleanPath, result.Exists())

			if result.Exists() {
				m.Context.Variables[name] = result.Value()
				fmt.Printf("  成功提取变量: %s = %v (使用去除前缀的路径: %s)\n", name, result.Value(), cleanPath)
			} else {
				// 如果还是失败，尝试直接使用数组索引
				// 例如，如果路径是 $.tables[0].id，尝试 tables.0.id
				arrayPath := regexp.MustCompile(`\[(\d+)\]`).ReplaceAllString(cleanPath, ".$1")
				result = gjson.GetBytes(responseBody, arrayPath)

				fmt.Printf("  尝试使用点表示法的路径 %s 的提取结果存在: %v\n", arrayPath, result.Exists())

				if result.Exists() {
					m.Context.Variables[name] = result.Value()
					fmt.Printf("  成功提取变量: %s = %v (使用点表示法的路径: %s)\n", name, result.Value(), arrayPath)
				} else {
					// 如果还是失败，尝试直接获取第一个元素
					// 从路径中提取数组名称
					parts := strings.Split(cleanPath, ".")
					if len(parts) > 0 {
						arrayName := parts[0]
						// 尝试获取数组的第一个元素
						arrayFirstPath := arrayName + ".0.id"
						result = gjson.GetBytes(responseBody, arrayFirstPath)

						fmt.Printf("  尝试获取数组第一个元素的路径 %s 的提取结果存在: %v\n", arrayFirstPath, result.Exists())

						if result.Exists() {
							m.Context.Variables[name] = result.Value()
							fmt.Printf("  成功提取变量: %s = %v (使用数组第一个元素的路径: %s)\n", name, result.Value(), arrayFirstPath)
						} else {
							fmt.Printf("\u8b66\u544a: \u65e0\u6cd5\u4ece\u8def\u5f84 %s \u63d0\u53d6\u53d8\u91cf %s\n", path, name)
						}
					} else {
						fmt.Printf("\u8b66\u544a: \u65e0\u6cd5\u4ece\u8def\u5f84 %s \u63d0\u53d6\u53d8\u91cf %s\n", path, name)
					}
				}
			}
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

// toCamelCase 将蛛形命名法转换为驼峰命名法
func toCamelCase(s string) string {
	// 先将字符串分割为单词
	words := strings.Split(s, "_")
	result := words[0]

	// 将后续单词首字母大写
	for i := 1; i < len(words); i++ {
		if len(words[i]) > 0 {
			result += strings.ToUpper(words[i][:1]) + words[i][1:]
		}
	}

	return result
}

// toSnakeCase 将驼峰命名法转换为蛛形命名法
func toSnakeCase(s string) string {
	var result strings.Builder

	// 正则表达式匹配大写字母
	re := regexp.MustCompile(`[A-Z]`)

	// 遍历字符串
	for i, c := range s {
		if i > 0 && re.MatchString(string(c)) {
			result.WriteString("_")
		}
		result.WriteString(strings.ToLower(string(c)))
	}

	return result.String()
}

// getDefaultValues 从配置中获取默认值
func (m *Manager) getDefaultValues() map[string]string {
	// 首先检查上下文中是否已经有默认值配置
	if defaultConfig, exists := m.Context.Variables["default_values"]; exists {
		if defaultMap, ok := defaultConfig.(map[string]interface{}); ok {
			// 将 map[string]interface{} 转换为 map[string]string
			result := make(map[string]string)
			for k, v := range defaultMap {
				result[k] = fmt.Sprintf("%v", v)
			}
			return result
		}
	}

	// 如果没有配置，使用通用的默认值
	return map[string]string{
		"id":     "1",
		"page":   "1",
		"limit":  "10",
		"offset": "0",
		"token":  "test-token",
	}
}

// replaceGoTemplateVars 替换Go模板语法变量
func (m *Manager) replaceGoTemplateVars(input string) string {
	// 匹配 Go 模板语法 {{.varName}}
	re := regexp.MustCompile(`{{\s*\.([a-zA-Z0-9_]+)\s*}}`)
	matches := re.FindAllStringSubmatch(input, -1)

	result := input

	for _, match := range matches {
		if len(match) > 1 {
			placeholder := match[0] // {{.varName}}
			varName := match[1]     // varName

			// 检查上下文变量
			if value, exists := m.Context.Variables[varName]; exists {
				strValue := fmt.Sprintf("%v", value)
				result = strings.ReplaceAll(result, placeholder, strValue)
				fmt.Printf("  替换Go模板变量 %s 为上下文变量值: %s\n", placeholder, strValue)
				continue
			}

			// 尝试相似名称
			alternativeNames := []string{
				toCamelCase(varName),
				toSnakeCase(varName),
			}

			var replaced bool
			for _, altName := range alternativeNames {
				if altName == varName {
					continue // 跳过相同的名称
				}

				if value, exists := m.Context.Variables[altName]; exists {
					strValue := fmt.Sprintf("%v", value)
					result = strings.ReplaceAll(result, placeholder, strValue)
					fmt.Printf("  替换Go模板变量 %s 为相似名称变量 %s 的值: %s\n", placeholder, altName, strValue)
					replaced = true
					break
				}
			}

			// 使用默认值
			if !replaced {
				// 从配置中获取默认值
				defaultValues := m.getDefaultValues()

				// 使用配置中的默认值，不添加特定业务领域的硬编码映射

				for defName, defValue := range defaultValues {
					if varName == defName {
						result = strings.ReplaceAll(result, placeholder, defValue)
						fmt.Printf("  替换Go模板变量 %s 为默认值: %s\n", placeholder, defValue)
						break
					}
				}
			}
		}
	}

	return result
}
