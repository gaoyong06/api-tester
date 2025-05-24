package template

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/google/uuid"
)

// Processor 模板处理器
type Processor struct {
	// 模板函数
	Functions template.FuncMap
	// 变量存储
	Variables map[string]interface{}
}

// NewProcessor 创建新的模板处理器
func NewProcessor() *Processor {
	p := &Processor{
		Variables: make(map[string]interface{}),
	}
	p.RegisterFunctions()
	return p
}

// RegisterFunctions 注册模板函数
func (tp *Processor) RegisterFunctions() {
	tp.Functions = template.FuncMap{
		// 时间函数
		"now": time.Now,
		"formatTime": func(format string) string {
			return time.Now().Format(format)
		},
		"addDays": func(days int) string {
			return time.Now().AddDate(0, 0, days).Format("2006-01-02")
		},

		// 随机数函数
		"uuid": func() string {
			return uuid.New().String()
		},
		"random": func(min, max int) int {
			return rand.Intn(max-min) + min
		},
		"randomString": func(length int) string {
			const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
			b := make([]byte, length)
			for i := range b {
				b[i] = charset[rand.Intn(len(charset))]
			}
			return string(b)
		},

		// 编码函数
		"base64": func(s string) string {
			return base64.StdEncoding.EncodeToString([]byte(s))
		},
		"base64decode": func(s string) string {
			data, err := base64.StdEncoding.DecodeString(s)
			if err != nil {
				return err.Error()
			}
			return string(data)
		},

		// 字符串函数
		"lower": strings.ToLower,
		"upper": strings.ToUpper,
		"title": strings.Title,
		"trim": strings.TrimSpace,
		"replace": strings.ReplaceAll,
		"contains": strings.Contains,
		"hasPrefix": strings.HasPrefix,
		"hasSuffix": strings.HasSuffix,
		"join": strings.Join,
		"split": strings.Split,
		"substr": func(s string, start, length int) string {
			if start < 0 || start >= len(s) {
				return ""
			}
			end := start + length
			if end > len(s) {
				end = len(s)
			}
			return s[start:end]
		},

		// 数学函数
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		"mul": func(a, b int) int { return a * b },
		"div": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a / b
		},
		"mod": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a % b
		},

		// 条件函数
		"ifThen": func(cond bool, a, b interface{}) interface{} {
			if cond {
				return a
			}
			return b
		},
		"eq": func(a, b interface{}) bool { return a == b },
		"ne": func(a, b interface{}) bool { return a != b },
		"lt": func(a, b int) bool { return a < b },
		"le": func(a, b int) bool { return a <= b },
		"gt": func(a, b int) bool { return a > b },
		"ge": func(a, b int) bool { return a >= b },

		// 正则表达式
		"regexMatch": func(pattern, s string) bool {
			match, _ := regexp.MatchString(pattern, s)
			return match
		},
		"regexReplace": func(pattern, replacement, s string) string {
			reg, err := regexp.Compile(pattern)
			if err != nil {
				return s
			}
			return reg.ReplaceAllString(s, replacement)
		},
	}
}

// Process 处理模板
func (tp *Processor) Process(input string) (string, error) {
	// 创建模板
	tmpl, err := template.New("").Funcs(tp.Functions).Parse(input)
	if err != nil {
		return "", fmt.Errorf("无法解析模板: %v", err)
	}

	// 执行模板
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, tp.Variables); err != nil {
		return "", fmt.Errorf("无法执行模板: %v", err)
	}

	return buf.String(), nil
}

// SetVariable 设置变量
func (tp *Processor) SetVariable(name string, value interface{}) {
	tp.Variables[name] = value
}

// GetVariable 获取变量
func (tp *Processor) GetVariable(name string) (interface{}, bool) {
	value, exists := tp.Variables[name]
	return value, exists
}

// SetVariables 批量设置变量
func (tp *Processor) SetVariables(variables map[string]interface{}) {
	for name, value := range variables {
		tp.Variables[name] = value
	}
}

// ProcessSimple 简单变量替换
func (tp *Processor) ProcessSimple(input string) string {
	// 匹配 {{.variable}} 格式的变量
	result := input
	for key, value := range tp.Variables {
		placeholder := "{{." + key + "}}"
		strValue := fmt.Sprintf("%v", value)
		result = strings.ReplaceAll(result, placeholder, strValue)
	}

	return result
}

// AddCustomFunction 添加自定义函数
func (tp *Processor) AddCustomFunction(name string, function interface{}) {
	tp.Functions[name] = function
}
