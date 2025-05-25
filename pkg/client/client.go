package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/gaoyong06/api-tester/internal/parser"
)

// APIClient 是一个HTTP客户端，用于测试API端点
type APIClient struct {
	// HTTP客户端
	client *http.Client
	// 基础URL
	baseURL string
	// 全局请求头
	headers map[string]string
	// 是否显示详细日志
	verbose bool
	// 请求体模板
	requestBodies map[string]interface{}
}

// Response 表示API响应
type Response struct {
	// HTTP状态码
	StatusCode int
	// 响应头
	Headers map[string][]string
	// 响应体
	Body []byte
	// 响应时间（毫秒）
	ResponseTime int64
	// 错误信息（如果有）
	Error error
}

// NewAPIClient 创建一个新的API客户端
func NewAPIClient(baseURL string, headers map[string]string, timeout int, verbose bool, requestBodies map[string]interface{}) *APIClient {
	// 确保基础URL以/结尾
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}

	return &APIClient{
		client: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
		baseURL:      baseURL,
		headers:      headers,
		verbose:      verbose,
		requestBodies: requestBodies,
	}
}

// SendRequest 发送API请求
func (c *APIClient) SendRequest(endpoint *parser.Endpoint, pathParams map[string]string, queryParams map[string]string, customRequestBody ...string) (*Response, error) {
	// 构建URL
	url := c.buildURL(endpoint.Path, pathParams, queryParams)

	// 准备请求体
	var reqBody *bytes.Buffer
	
	// 检查是否提供了自定义请求体
	if len(customRequestBody) > 0 && customRequestBody[0] != "" {
		// 使用提供的自定义请求体
		reqBody = bytes.NewBufferString(customRequestBody[0])
		if c.verbose {
			fmt.Printf("使用自定义请求体: %s\n", customRequestBody[0])
		}
	} else if endpoint.Method == "POST" || endpoint.Method == "PUT" || endpoint.Method == "PATCH" {
		// 尝试从请求体模板中获取
		if c.requestBodies != nil {
			// 使用路径作为键名查找请求体模板
			if template, ok := c.requestBodies[endpoint.Path]; ok {
				// 将模板转换为JSON
				jsonData, err := json.Marshal(template)
				if err != nil {
					return nil, fmt.Errorf("序列化请求体模板失败: %v", err)
				}
				reqBody = bytes.NewBuffer(jsonData)
				
				if c.verbose {
					fmt.Printf("使用请求体模板: %s\n", string(jsonData))
				}
			}
		}
	}

	// 如果没有找到模板，使用空请求体
	if reqBody == nil {
		reqBody = bytes.NewBufferString("")
	}

	// 创建请求
	req, err := http.NewRequest(endpoint.Method, url, reqBody)
	if err != nil {
		return &Response{Error: fmt.Errorf("创建请求失败: %v", err)}, nil
	}

	// 添加全局请求头
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	// 添加内容类型头（如果请求体不为空）
	if reqBody != nil && reqBody.Len() > 0 {
		req.Header.Set("Content-Type", "application/json")
	}

	// 添加参数头
	for _, param := range endpoint.Parameters {
		if param.In == "header" && param.Example != "" {
			req.Header.Set(param.Name, param.Example)
		}
	}

	// 记录请求开始时间
	startTime := time.Now()

	// 发送请求
	resp, err := c.client.Do(req)
	if err != nil {
		return &Response{Error: fmt.Errorf("发送请求失败: %v", err)}, nil
	}
	defer resp.Body.Close()

	// 计算响应时间
	responseTime := time.Since(startTime).Milliseconds()

	// 读取响应体
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &Response{Error: fmt.Errorf("读取响应体失败: %v", err)}, nil
	}

	// 打印详细日志
	if c.verbose {
		fmt.Printf("\n> %s %s\n", endpoint.Method, url)
		fmt.Printf("> 请求头: %v\n", req.Header)
		if reqBody != nil && reqBody.Len() > 0 {
			fmt.Printf("> 请求体: %s\n", reqBody.String())
		}
		fmt.Printf("< 状态码: %d\n", resp.StatusCode)
		fmt.Printf("< 响应头: %v\n", resp.Header)
		fmt.Printf("< 响应体: %s\n", string(respBody))
		fmt.Printf("< 响应时间: %d ms\n", responseTime)
	}

	return &Response{
		StatusCode:   resp.StatusCode,
		Headers:      resp.Header,
		Body:         respBody,
		ResponseTime: responseTime,
	}, nil
}

// buildURL 构建完整的请求URL
func (c *APIClient) buildURL(path string, pathParams map[string]string, queryParams map[string]string) string {
	// 替换路径参数
	for name, value := range pathParams {
		path = strings.Replace(path, "{" + name + "}", value, -1)
	}

	// 移除前导斜杠
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}

	// 构建基础URL
	url := c.baseURL + path

	// 添加查询参数
	if len(queryParams) > 0 {
		queries := make([]string, 0, len(queryParams))
		for name, value := range queryParams {
			queries = append(queries, name+"="+value)
		}
		url += "?" + strings.Join(queries, "&")
	}

	return url
}

// ExtractPathParams 从端点路径中提取路径参数
func ExtractPathParams(endpoint *parser.Endpoint) map[string]string {
	pathParams := make(map[string]string)

	for _, param := range endpoint.Parameters {
		if param.In == "path" && param.Example != "" {
			pathParams[param.Name] = param.Example
		}
	}

	return pathParams
}

// ExtractQueryParams 从端点参数中提取查询参数
func ExtractQueryParams(endpoint *parser.Endpoint) map[string]string {
	queryParams := make(map[string]string)

	for _, param := range endpoint.Parameters {
		if param.In == "query" && param.Example != "" {
			queryParams[param.Name] = param.Example
		}
	}

	return queryParams
}

// PrettyJSON 格式化JSON字符串
func PrettyJSON(data []byte) string {
	var out bytes.Buffer
	err := json.Indent(&out, data, "", "  ")
	if err != nil {
		return string(data)
	}
	return out.String()
}
