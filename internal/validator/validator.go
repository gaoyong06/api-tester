package validator

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gaoyong06/api-tester/internal/parser"
	"github.com/gaoyong06/api-tester/internal/types"
	"github.com/gaoyong06/api-tester/pkg/client"
	"github.com/xeipuuv/gojsonschema"
)

// ValidateResponse 验证API响应
func ValidateResponse(endpoint *parser.Endpoint, response *client.Response) *types.ValidationResult {
	// 如果请求失败，直接返回失败结果
	if response.Error != nil {
		return &types.ValidationResult{
			Passed:        false,
			FailureReason: response.Error.Error(),
			ActualStatus: 0,
			ResponseTime: response.ResponseTime,
		}
	}

	// 验证状态码
	expectedStatus := ""
	for status := range endpoint.Responses {
		expectedStatus = status
		break // 只取第一个预期状态码
	}

	statusOK := false
	if expectedStatus != "" {
		// 将预期状态码转换为整数
		expectedStatusInt := 0
		fmt.Sscanf(expectedStatus, "%d", &expectedStatusInt)
		statusOK = (expectedStatusInt == response.StatusCode)
	} else {
		// 如果没有预期状态码，则假设2xx为成功
		statusOK = (response.StatusCode >= 200 && response.StatusCode < 300)
	}

	if !statusOK {
		return &types.ValidationResult{
			Passed:         false,
			FailureReason:  fmt.Sprintf("状态码不匹配: 预期 %s, 实际 %d", expectedStatus, response.StatusCode),
			ExpectedStatus: expectedStatus,
			ActualStatus:   response.StatusCode,
			ResponseTime:   response.ResponseTime,
			ResponseBody:   string(response.Body),
		}
	}

	// 验证响应体结构（如果有预期响应）
	expectedResponse := endpoint.Responses[expectedStatus]
	if expectedResponse != "" && len(response.Body) > 0 {
		// 尝试解析响应体为JSON
		var actualJSON interface{}
		if err := json.Unmarshal(response.Body, &actualJSON); err != nil {
			return &types.ValidationResult{
				Passed:         false,
				FailureReason:  fmt.Sprintf("响应体不是有效的JSON: %v", err),
				ExpectedStatus: expectedStatus,
				ActualStatus:   response.StatusCode,
				ResponseTime:   response.ResponseTime,
				ResponseBody:   string(response.Body),
			}
		}

		// 尝试解析预期响应为JSON
		var expectedJSON interface{}
		if err := json.Unmarshal([]byte(expectedResponse), &expectedJSON); err != nil {
			// 如果预期响应不是有效的JSON，跳过结构验证
			return &types.ValidationResult{
				Passed:         true,
				ExpectedStatus: expectedStatus,
				ActualStatus:   response.StatusCode,
				ResponseTime:   response.ResponseTime,
				ResponseBody:   client.PrettyJSON(response.Body),
			}
		}

		// 验证响应体结构
		schemaLoader := gojsonschema.NewGoLoader(expectedJSON)
		documentLoader := gojsonschema.NewGoLoader(actualJSON)

		result, err := gojsonschema.Validate(schemaLoader, documentLoader)
		if err != nil {
			return &types.ValidationResult{
				Passed:         false,
				FailureReason:  fmt.Sprintf("验证响应体结构失败: %v", err),
				ExpectedStatus: expectedStatus,
				ActualStatus:   response.StatusCode,
				ResponseTime:   response.ResponseTime,
				ResponseBody:   client.PrettyJSON(response.Body),
			}
		}

		if !result.Valid() {
			// 收集验证错误
			errors := make([]string, 0, len(result.Errors()))
			for _, err := range result.Errors() {
				errors = append(errors, fmt.Sprintf("- %s", err.String()))
			}

			return &types.ValidationResult{
				Passed:         false,
				FailureReason:  fmt.Sprintf("响应体结构不匹配:\n%s", strings.Join(errors, "\n")),
				ExpectedStatus: expectedStatus,
				ActualStatus:   response.StatusCode,
				ResponseTime:   response.ResponseTime,
				ResponseBody:   client.PrettyJSON(response.Body),
			}
		}
	}

	// 验证通过
	return &types.ValidationResult{
		Passed:         true,
		ExpectedStatus: expectedStatus,
		ActualStatus:   response.StatusCode,
		ResponseTime:   response.ResponseTime,
		ResponseBody:   client.PrettyJSON(response.Body),
	}
}
