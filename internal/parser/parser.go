package parser

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v3"
)

// APIDefinition 表示从OpenAPI规范解析出的API定义
type APIDefinition struct {
	// API标题
	Title string
	// API版本
	Version string
	// API描述
	Description string
	// API路径和操作
	Endpoints []*Endpoint
	// 模式定义，用于生成模拟数据
	Schemas map[string]interface{}
}

// Endpoint 表示API端点
type Endpoint struct {
	// 路径 (例如 /users/{id})
	Path string
	// HTTP方法 (例如 GET, POST)
	Method string
	// 操作ID
	OperationID string
	// 操作描述
	Description string
	// 标签
	Tags []string
	// 请求体示例
	RequestBody string
	// 请求参数
	Parameters []*Parameter
	// 响应示例
	Responses map[string]string
}

// Parameter 表示API参数
type Parameter struct {
	// 参数名
	Name string
	// 参数位置 (path, query, header, cookie)
	In string
	// 是否必需
	Required bool
	// 参数描述
	Description string
	// 参数类型
	Type string
	// 示例值
	Example string
}

// ParseSwaggerFile 解析 Swagger 2.0 文件并转换为 APIDefinition
func ParseSwaggerFile(filePath string) (*APIDefinition, error) {
	// 获取文件绝对路径
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("无法获取规范文件的绝对路径: %v", err)
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

	// 尝试作为 JSON 解析
	swagger := &openapi2.T{}
	if err := json.Unmarshal(data, swagger); err != nil {
		// 尝试作为 YAML 解析
		if err := yaml.Unmarshal(data, swagger); err != nil {
			return nil, fmt.Errorf("无法解析 Swagger 规范: %v", err)
		}
	}

	// 将 Swagger 定义转换为 APIDefinition
	apiDef := &APIDefinition{
		Title:       swagger.Info.Title,
		Version:     swagger.Info.Version,
		Description: swagger.Info.Description,
		Endpoints:   make([]*Endpoint, 0),
		Schemas:     make(map[string]interface{}),
	}

	// 处理模式定义
	for name, schema := range swagger.Definitions {
		apiDef.Schemas[name] = schema
	}

	// 处理路径和操作
	for path, pathItem := range swagger.Paths {
		// 手动处理每个 HTTP 方法
		processOperation := func(method string, op *openapi2.Operation) {
			if op == nil || op.OperationID == "" {
				return
			}

			// 创建端点
			endpoint := &Endpoint{
				Path:        path,
				Method:      method,
				OperationID: op.OperationID,
				Description: op.Description,
				Tags:        op.Tags,
				Parameters:  make([]*Parameter, 0),
				Responses:   make(map[string]string),
			}

			// 处理参数
			for _, param := range op.Parameters {
				parameter := &Parameter{
					Name:        param.Name,
					In:          param.In,
					Required:    param.Required,
					Description: param.Description,
					Type:        "string", // 默认使用字符串类型
				}

				endpoint.Parameters = append(endpoint.Parameters, parameter)
			}

			apiDef.Endpoints = append(apiDef.Endpoints, endpoint)
		}

		// 调用每个 HTTP 方法的处理函数
		processOperation("GET", pathItem.Get)
		processOperation("POST", pathItem.Post)
		processOperation("PUT", pathItem.Put)
		processOperation("DELETE", pathItem.Delete)
		processOperation("PATCH", pathItem.Patch)
		processOperation("HEAD", pathItem.Head)
		processOperation("OPTIONS", pathItem.Options)
	}

	return apiDef, nil
}

// ParseOpenAPI 解析OpenAPI/Swagger规范文件
func ParseOpenAPI(filePath string) (*APIDefinition, error) {
	// 加载OpenAPI规范文件
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	// 获取文件绝对路径
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("无法获取规范文件的绝对路径: %v", err)
	}

	// 解析OpenAPI文档
	doc, err := loader.LoadFromFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("无法解析OpenAPI规范: %v", err)
	}

	// 验证文档
	if err := doc.Validate(loader.Context); err != nil {
		return nil, fmt.Errorf("OpenAPI规范验证失败: %v", err)
	}

	// 创建API定义
	apiDef := &APIDefinition{
		Title:       doc.Info.Title,
		Version:     doc.Info.Version,
		Description: doc.Info.Description,
		Endpoints:   make([]*Endpoint, 0),
	}

	// 解析路径和操作
	for path, pathItem := range doc.Paths.Map() {
		// 遍历每个HTTP方法
		for method, operation := range map[string]*openapi3.Operation{
			"GET":     pathItem.Get,
			"POST":    pathItem.Post,
			"PUT":     pathItem.Put,
			"DELETE":  pathItem.Delete,
			"PATCH":   pathItem.Patch,
			"HEAD":    pathItem.Head,
			"OPTIONS": pathItem.Options,
		} {
			if operation == nil {
				continue
			}

			// 创建端点
			endpoint := &Endpoint{
				Path:        path,
				Method:      method,
				OperationID: operation.OperationID,
				Description: operation.Description,
				Tags:        operation.Tags,
				Parameters:  make([]*Parameter, 0),
				Responses:   make(map[string]string),
			}

			// 解析请求体
			if operation.RequestBody != nil && operation.RequestBody.Value != nil {
				for contentType, mediaType := range operation.RequestBody.Value.Content {
					if strings.Contains(contentType, "json") && mediaType.Example != nil {
						endpoint.RequestBody = fmt.Sprintf("%v", mediaType.Example)
						break
					}
				}
			}

			// 解析参数
			for _, paramRef := range operation.Parameters {
				if paramRef.Value == nil {
					continue
				}

				param := &Parameter{
					Name:        paramRef.Value.Name,
					In:          paramRef.Value.In,
					Required:    paramRef.Value.Required,
					Description: paramRef.Value.Description,
				}

				// 获取参数类型和示例
				if paramRef.Value.Schema != nil && paramRef.Value.Schema.Value != nil {
					schema := paramRef.Value.Schema.Value
					// schema.Type 是 *openapi3.Types 类型（字符串切片的指针），需要转换为单一字符串
					if schema.Type != nil && len(schema.Type.Slice()) > 0 {
						param.Type = schema.Type.Slice()[0]
					}
					if schema.Example != nil {
						param.Example = fmt.Sprintf("%v", schema.Example)
					}
				}

				endpoint.Parameters = append(endpoint.Parameters, param)
			}

			// 解析响应
			// operation.Responses 是 *openapi3.Responses 类型，需要使用 Map() 方法获取可遍历的映射
			for statusCode, responseRef := range operation.Responses.Map() {
				if responseRef.Value == nil {
					continue
				}

				for contentType, mediaType := range responseRef.Value.Content {
					if strings.Contains(contentType, "json") && mediaType.Example != nil {
						endpoint.Responses[statusCode] = fmt.Sprintf("%v", mediaType.Example)
						break
					}
				}
			}

			apiDef.Endpoints = append(apiDef.Endpoints, endpoint)
		}
	}

	return apiDef, nil
}
