package mock

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v6"
)

// MockDataGenerator 模拟数据生成器
type MockDataGenerator struct {
	// 自定义生成规则
	Rules map[string]interface{}
	// 随机种子
	Seed int64
}

// NewMockDataGenerator 创建一个新的模拟数据生成器
func NewMockDataGenerator() *MockDataGenerator {
	return &MockDataGenerator{
		Rules: make(map[string]interface{}),
		Seed:  time.Now().UnixNano(),
	}
}

// SetSeed 设置随机种子
func (g *MockDataGenerator) SetSeed(seed int64) *MockDataGenerator {
	g.Seed = seed
	rand.Seed(seed)
	gofakeit.Seed(seed)
	return g
}

// SetCustomRules 设置自定义规则
func (g *MockDataGenerator) SetCustomRules(rules map[string]interface{}) *MockDataGenerator {
	g.Rules = rules
	return g
}

// GenerateFromSchema 根据 Schema 生成数据
func (g *MockDataGenerator) GenerateFromSchema(schema interface{}) (interface{}, error) {
	if schema == nil {
		return nil, fmt.Errorf("无效的 Schema")
	}

	// 如果 schema 是 map，则处理为对象
	schemaMap, isMap := schema.(map[string]interface{})
	if isMap {
		// 检查是否有 type 字段
		if typeVal, ok := schemaMap["type"]; ok {
			schemaType, isString := typeVal.(string)
			if isString {
				switch schemaType {
				case "object":
					return g.generateObject(schemaMap)
				case "array":
					return g.generateArray(schemaMap)
				case "string":
					return g.generateString(schemaMap, "")
				case "number", "integer":
					return g.generateNumber(schemaMap)
				case "boolean":
					return gofakeit.Bool(), nil
				default:
					return nil, fmt.Errorf("不支持的类型: %s", schemaType)
				}
			}
		}

		// 如果没有 type 字段，则默认处理为对象
		return g.generateObject(schemaMap)
	}

	// 如果是字符串，则根据内容猜测类型
	schemaStr, isString := schema.(string)
	if isString {
		switch schemaStr {
		case "string":
			return gofakeit.Sentence(5), nil
		case "number", "integer":
			return gofakeit.Number(1, 1000), nil
		case "boolean":
			return gofakeit.Bool(), nil
		default:
			// 尝试根据字符串内容猜测类型
			if strings.Contains(strings.ToLower(schemaStr), "id") {
				return gofakeit.UUID(), nil
			} else if strings.Contains(strings.ToLower(schemaStr), "name") {
				return gofakeit.Name(), nil
			} else if strings.Contains(strings.ToLower(schemaStr), "email") {
				return gofakeit.Email(), nil
			} else if strings.Contains(strings.ToLower(schemaStr), "phone") {
				return gofakeit.Phone(), nil
			} else if strings.Contains(strings.ToLower(schemaStr), "address") {
				return gofakeit.Address().Address, nil
			} else if strings.Contains(strings.ToLower(schemaStr), "date") {
				return gofakeit.Date().Format("2006-01-02"), nil
			} else {
				return gofakeit.Sentence(3), nil
			}
		}
	}

	// 如果是其他类型，则返回原样
	return schema, nil
}

// generateObject 生成对象数据
func (g *MockDataGenerator) generateObject(schema map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// 检查是否有 properties 字段
	props, hasProps := schema["properties"].(map[string]interface{})
	if hasProps {
		for propName, propSchema := range props {
			// 检查是否有自定义规则
			value, err := g.applyCustomRule(propName, propSchema)
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

	// 如果没有 properties 字段，则生成一些随机属性
	for i := 0; i < 5; i++ {
		propName := fmt.Sprintf("property%d", i+1)
		result[propName] = gofakeit.Sentence(3)
	}

	return result, nil
}

// generateArray 生成数组数据
func (g *MockDataGenerator) generateArray(schema map[string]interface{}) ([]interface{}, error) {
	// 确定数组长度
	minItems := 1
	maxItems := 5

	// 检查是否有 minItems 和 maxItems 字段
	if minVal, ok := schema["minItems"].(float64); ok {
		minItems = int(minVal)
	}
	if maxVal, ok := schema["maxItems"].(float64); ok {
		maxItems = int(maxVal)
	}

	// 如果最小值大于最大值，则交换
	if minItems > maxItems {
		minItems, maxItems = maxItems, minItems
	}

	// 生成随机数量的元素
	count := rand.Intn(maxItems-minItems+1) + minItems
	result := make([]interface{}, count)

	// 检查是否有 items 字段
	items, hasItems := schema["items"]
	if hasItems {
		for i := 0; i < count; i++ {
			value, err := g.GenerateFromSchema(items)
			if err != nil {
				return nil, err
			}
			result[i] = value
		}
	} else {
		// 如果没有 items 字段，则生成随机字符串
		for i := 0; i < count; i++ {
			result[i] = gofakeit.Sentence(3)
		}
	}

	return result, nil
}

// generateString 生成字符串数据
func (g *MockDataGenerator) generateString(schema map[string]interface{}, propName string) (string, error) {
	// 检查是否有 format 字段
	format, hasFormat := schema["format"].(string)
	if hasFormat {
		switch format {
		case "date":
			return gofakeit.Date().Format("2006-01-02"), nil
		case "date-time":
			return gofakeit.Date().Format(time.RFC3339), nil
		case "email":
			return gofakeit.Email(), nil
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
		case "password":
			return gofakeit.Password(true, true, true, true, false, 12), nil
		}
	}

	// 检查是否有 pattern 字段
	pattern, hasPattern := schema["pattern"].(string)
	if hasPattern {
		// 尝试根据正则表达式生成字符串
		// 这里简化处理，实际上需要更复杂的正则表达式解析
		if strings.Contains(pattern, "^[0-9]{5}$") {
			return gofakeit.Zip(), nil
		} else if strings.Contains(pattern, "^[0-9]{10}$") {
			return fmt.Sprintf("%010d", gofakeit.Number(0, 9999999999)), nil
		} else if strings.Contains(pattern, "^[A-Z]{2}[0-9]{9}$") {
			return fmt.Sprintf("%s%09d", gofakeit.LetterN(2), gofakeit.Number(0, 999999999)), nil
		}
	}

	// 如果没有特殊格式，则根据属性名称生成合适的值
	if propName != "" {
		propNameLower := strings.ToLower(propName)
		if strings.Contains(propNameLower, "name") {
			if strings.Contains(propNameLower, "first") {
				return gofakeit.FirstName(), nil
			} else if strings.Contains(propNameLower, "last") {
				return gofakeit.LastName(), nil
			} else {
				return gofakeit.Name(), nil
			}
		} else if strings.Contains(propNameLower, "email") {
			return gofakeit.Email(), nil
		} else if strings.Contains(propNameLower, "phone") {
			return gofakeit.Phone(), nil
		} else if strings.Contains(propNameLower, "address") {
			return gofakeit.Address().Address, nil
		} else if strings.Contains(propNameLower, "city") {
			return gofakeit.City(), nil
		} else if strings.Contains(propNameLower, "state") {
			return gofakeit.State(), nil
		} else if strings.Contains(propNameLower, "country") {
			return gofakeit.Country(), nil
		} else if strings.Contains(propNameLower, "zip") || strings.Contains(propNameLower, "postal") {
			return gofakeit.Zip(), nil
		} else if strings.Contains(propNameLower, "company") {
			return gofakeit.Company(), nil
		} else if strings.Contains(propNameLower, "job") || strings.Contains(propNameLower, "title") {
			return gofakeit.JobTitle(), nil
		} else if strings.Contains(propNameLower, "description") || strings.Contains(propNameLower, "summary") {
			return gofakeit.Paragraph(2, 5, 10, " "), nil
		} else if strings.Contains(propNameLower, "id") {
			return gofakeit.UUID(), nil
		} else if strings.Contains(propNameLower, "date") || strings.Contains(propNameLower, "time") {
			return gofakeit.Date().Format("2006-01-02"), nil
		} else if strings.Contains(propNameLower, "url") || strings.Contains(propNameLower, "website") {
			return gofakeit.URL(), nil
		} else if strings.Contains(propNameLower, "image") || strings.Contains(propNameLower, "avatar") {
			return gofakeit.ImageURL(300, 300), nil
		} else if strings.Contains(propNameLower, "color") {
			return gofakeit.Color(), nil
		} else if strings.Contains(propNameLower, "password") {
			return gofakeit.Password(true, true, true, true, false, 12), nil
		} else if strings.Contains(propNameLower, "token") {
			return gofakeit.UUID(), nil
		}
	}

	// 默认生成随机句子
	return gofakeit.Sentence(3), nil
}

// generateNumber 生成数字数据
func (g *MockDataGenerator) generateNumber(schema map[string]interface{}) (interface{}, error) {
	// 确定数字范围
	min := 0.0
	max := 1000.0

	// 检查是否有 minimum 和 maximum 字段
	if minVal, ok := schema["minimum"].(float64); ok {
		min = minVal
	}
	if maxVal, ok := schema["maximum"].(float64); ok {
		max = maxVal
	}

	// 如果最小值大于最大值，则交换
	if min > max {
		min, max = max, min
	}

	// 检查是否是整数
	if schemaType, ok := schema["type"].(string); ok && schemaType == "integer" {
		return gofakeit.Number(int(min), int(max)), nil
	}

	// 生成浮点数
	return gofakeit.Float64Range(min, max), nil
}

// applyCustomRule 应用自定义规则
func (g *MockDataGenerator) applyCustomRule(propName string, schema interface{}) (interface{}, error) {
	// 检查是否有完全匹配的规则
	if value, ok := g.Rules[propName]; ok {
		return value, nil
	}

	// 检查是否有正则表达式匹配的规则
	for pattern, value := range g.Rules {

		// 如果键以 "regex:" 开头，则尝试正则表达式匹配
		if strings.HasPrefix(pattern, "regex:") {
			regexPattern := strings.TrimPrefix(pattern, "regex:")
			reg, err := regexp.Compile(regexPattern)
			if err != nil {
				continue
			}

			if reg.MatchString(propName) {
				return value, nil
			}
		}
	}

	return nil, nil
}
