package mock

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/getkin/kin-openapi/openapi3"
)

// Generator 是 Mock 数据生成器
type Generator struct {
	// OpenAPI 文档
	Doc *openapi3.T
	// 生成策略
	Strategy Strategy
	// 自定义生成规则
	Rules map[string]*Rule
	// 随机种子
	Seed int64
}

// Strategy 表示数据生成策略
type Strategy string

const (
	// RandomStrategy 随机生成所有字段
	RandomStrategy Strategy = "random"
	// RequiredOnlyStrategy 仅生成必填字段
	RequiredOnlyStrategy Strategy = "required_only"
	// ExampleStrategy 使用示例值
	ExampleStrategy Strategy = "example"
	// DefaultStrategy 使用默认值
	DefaultStrategy Strategy = "default"
	// BoundaryStrategy 边界值测试
	BoundaryStrategy Strategy = "boundary"
)

// Rule 定义数据生成规则
type Rule struct {
	// 字段模式（正则表达式）
	Pattern *regexp.Regexp
	// 生成函数
	Generator func() interface{}
}

// NewGenerator 创建一个新的 Mock 数据生成器
func NewGenerator(doc *openapi3.T) *Generator {
	return &Generator{
		Doc:      doc,
		Strategy: RandomStrategy,
		Rules:    make(map[string]*Rule),
		Seed:     time.Now().UnixNano(),
	}
}

// WithStrategy 设置生成策略
func (g *Generator) WithStrategy(strategy Strategy) *Generator {
	g.Strategy = strategy
	return g
}

// WithSeed 设置随机种子
func (g *Generator) WithSeed(seed int64) *Generator {
	g.Seed = seed
	return g
}

// AddRule 添加自定义生成规则
func (g *Generator) AddRule(pattern string, generator func() interface{}) error {
	reg, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("无效的正则表达式: %v", err)
	}

	g.Rules[pattern] = &Rule{
		Pattern:   reg,
		Generator: generator,
	}

	return nil
}

// GenerateFromSchema 根据 Schema 生成数据
func (g *Generator) GenerateFromSchema(schemaRef *openapi3.SchemaRef) (interface{}, error) {
	if schemaRef == nil || schemaRef.Value == nil {
		return nil, fmt.Errorf("无效的 Schema")
	}

	schema := schemaRef.Value

	// 如果使用示例策略并且有示例值，则返回示例值
	if g.Strategy == ExampleStrategy && schema.Example != nil {
		return schema.Example, nil
	}

	// 根据 schema 类型生成数据
	if schema.Type == nil {
		return nil, fmt.Errorf("未指定类型")
	}

	// 检查 schema 类型
	if schema.Type.Is("object") {
		return g.generateObject(schema)
	} else if schema.Type.Is("array") {
		return g.generateArray(schema)
	} else if schema.Type.Is("string") {
		return g.generateString(schema, "")
	} else if schema.Type.Is("number") || schema.Type.Is("integer") {
		return g.generateNumber(schema)
	} else if schema.Type.Is("boolean") {
		return gofakeit.Bool(), nil
	} else {
		return nil, fmt.Errorf("不支持的类型: %s", schema.Type)
	}
}

// generateObject 生成对象数据
func (g *Generator) generateObject(schema *openapi3.Schema) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for propName, propSchema := range schema.Properties {
		// 检查是否是必填字段
		required := false
		for _, req := range schema.Required {
			if req == propName {
				required = true
				break
			}
		}

		// 如果使用 RequiredOnlyStrategy 并且不是必填字段，则跳过
		if g.Strategy == RequiredOnlyStrategy && !required {
			continue
		}

		// 检查是否有匹配的自定义规则
		value, err := g.applyCustomRule(propName, propSchema.Value)
		if err != nil {
			return nil, err
		}

		// 如果没有匹配的自定义规则，则使用 schema 生成
		if value == nil {
			value, err = g.GenerateFromSchema(propSchema)
			if err != nil {
				return nil, err
			}
		}

		result[propName] = value
	}

	return result, nil
}

// generateArray 生成数组数据
func (g *Generator) generateArray(schema *openapi3.Schema) ([]interface{}, error) {
	if schema.Items == nil {
		return []interface{}{}, nil
	}

	// 确定数组长度
	minItems := 1
	maxItems := 5

	// 如果最小值大于最大值，则交换
	if minItems > maxItems {
		minItems, maxItems = maxItems, minItems
	}

	// 生成随机数量的元素
	count := gofakeit.Number(minItems, maxItems)
	result := make([]interface{}, count)

	for i := 0; i < count; i++ {
		value, err := g.GenerateFromSchema(schema.Items)
		if err != nil {
			return nil, err
		}
		result[i] = value
	}

	return result, nil
}

// generateString 生成字符串数据
func (g *Generator) generateString(schema *openapi3.Schema, propName string) (string, error) {
	// 根据格式生成不同类型的字符串
	switch schema.Format {
	case "email":
		return gofakeit.Email(), nil
	case "date-time":
		return gofakeit.Date().Format(time.RFC3339), nil
	case "date":
		return gofakeit.Date().Format("2006-01-02"), nil
	case "uuid":
		return gofakeit.UUID(), nil
	case "uri":
		return gofakeit.URL(), nil
	case "hostname":
		return gofakeit.DomainName(), nil
	case "ipv4":
		return gofakeit.IPv4Address(), nil
	case "ipv6":
		return gofakeit.IPv6Address(), nil
	default:
		// 根据属性名猜测内容类型
		propNameLower := strings.ToLower(propName)
		switch {
		case contains(propNameLower, "name"):
			return gofakeit.Name(), nil
		case contains(propNameLower, "first", "firstname"):
			return gofakeit.FirstName(), nil
		case contains(propNameLower, "last", "lastname"):
			return gofakeit.LastName(), nil
		case contains(propNameLower, "phone"):
			return gofakeit.Phone(), nil
		case contains(propNameLower, "address"):
			return gofakeit.Address().Street, nil
		case contains(propNameLower, "city"):
			return gofakeit.City(), nil
		case contains(propNameLower, "country"):
			return gofakeit.Country(), nil
		case contains(propNameLower, "zip", "postal"):
			return gofakeit.Zip(), nil
		case contains(propNameLower, "company"):
			return gofakeit.Company(), nil
		case contains(propNameLower, "job", "title"):
			return gofakeit.JobTitle(), nil
		case contains(propNameLower, "description"):
			return gofakeit.Paragraph(1, 3, 5, " "), nil
		case contains(propNameLower, "password"):
			return gofakeit.Password(true, true, true, true, false, 10), nil
		case contains(propNameLower, "username"):
			return gofakeit.Username(), nil
		case contains(propNameLower, "url"):
			return gofakeit.URL(), nil
		case contains(propNameLower, "image", "avatar", "photo"):
			return gofakeit.ImageURL(300, 300), nil
		case contains(propNameLower, "color"):
			return gofakeit.Color(), nil
		case contains(propNameLower, "id"):
			return fmt.Sprintf("%d", gofakeit.Number(1, 1000)), nil
		default:
			// 默认生成一个句子
			return gofakeit.Sentence(3), nil
		}
	}
}

// generateNumber 生成数字数据
func (g *Generator) generateNumber(schema *openapi3.Schema) (interface{}, error) {
	min := 0.0
	max := 1000.0

	if schema.Min != nil {
		min = *schema.Min
	}
	if schema.Max != nil {
		max = *schema.Max
	}

	// 如果最小值大于最大值，则交换
	if min > max {
		min, max = max, min
	}

	// 根据策略生成数字
	switch g.Strategy {
	case BoundaryStrategy:
		// 随机选择边界值或中间值
		choice := gofakeit.Number(1, 3)
		switch choice {
		case 1:
			return min, nil
		case 2:
			return max, nil
		case 3:
			return (min + max) / 2, nil
		}
	default:
		// 生成随机值
		if schema.Type != nil && schema.Type.Is("integer") {
			return gofakeit.Number(int(min), int(max)), nil
		}
		return gofakeit.Float64Range(min, max), nil
	}

	return 0, nil
}

// applyCustomRule 应用自定义规则
func (g *Generator) applyCustomRule(propName string, schema *openapi3.Schema) (interface{}, error) {
	for _, rule := range g.Rules {
		if rule.Pattern.MatchString(propName) {
			return rule.Generator(), nil
		}
	}

	return nil, nil
}

// contains 检查字符串是否包含任何给定的子字符串
func contains(s string, substrings ...string) bool {
	for _, substr := range substrings {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}

// GenerateAll 为所有模型生成数据
func (g *Generator) GenerateAll(count int) (map[string][]interface{}, error) {
	result := make(map[string][]interface{})

	// 为每个模型生成数据
	for name, schema := range g.Doc.Components.Schemas {
		var items []interface{}
		for i := 0; i < count; i++ {
			data, err := g.GenerateFromSchema(schema)
			if err != nil {
				return nil, fmt.Errorf("生成 %s 的 mock 数据失败: %v", name, err)
			}
			items = append(items, data)
		}
		result[name] = items
	}

	return result, nil
}

// GenerateForEndpoint 为指定端点生成请求数据
func (g *Generator) GenerateForEndpoint(path string, method string) (map[string]interface{}, error) {
	// 查找路径项
	pathItem := g.Doc.Paths.Find(path)
	if pathItem == nil {
		return nil, fmt.Errorf("未找到路径: %s", path)
	}

	// 查找操作
	var operation *openapi3.Operation
	switch strings.ToUpper(method) {
	case "GET":
		operation = pathItem.Get
	case "POST":
		operation = pathItem.Post
	case "PUT":
		operation = pathItem.Put
	case "DELETE":
		operation = pathItem.Delete
	case "PATCH":
		operation = pathItem.Patch
	case "HEAD":
		operation = pathItem.Head
	case "OPTIONS":
		operation = pathItem.Options
	default:
		return nil, fmt.Errorf("不支持的 HTTP 方法: %s", method)
	}

	if operation == nil {
		return nil, fmt.Errorf("路径 %s 不支持方法 %s", path, method)
	}

	// 生成请求数据
	result := make(map[string]interface{})

	// 生成路径参数
	pathParams := make(map[string]string)
	// 生成查询参数
	queryParams := make(map[string]string)
	// 生成请求体
	var requestBody interface{}

	// 处理参数
	for _, paramRef := range operation.Parameters {
		if paramRef.Value == nil {
			continue
		}

		param := paramRef.Value
		value, err := g.generateParameterValue(param)
		if err != nil {
			return nil, err
		}

		switch param.In {
		case "path":
			pathParams[param.Name] = fmt.Sprintf("%v", value)
		case "query":
			queryParams[param.Name] = fmt.Sprintf("%v", value)
		}
	}

	// 处理请求体
	if operation.RequestBody != nil && operation.RequestBody.Value != nil {
		for contentType, mediaType := range operation.RequestBody.Value.Content {
			if strings.Contains(contentType, "json") && mediaType.Schema != nil {
				data, err := g.GenerateFromSchema(mediaType.Schema)
				if err != nil {
					return nil, err
				}
				requestBody = data
				break
			}
		}
	}

	// 组装结果
	result["path_params"] = pathParams
	result["query_params"] = queryParams
	result["request_body"] = requestBody

	return result, nil
}

// generateParameterValue 生成参数值
func (g *Generator) generateParameterValue(param *openapi3.Parameter) (interface{}, error) {
	if param.Schema == nil {
		return nil, nil
	}

	// 生成参数值
	return g.GenerateFromSchema(param.Schema)
}
