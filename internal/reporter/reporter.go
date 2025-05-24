package reporter

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gaoyong06/api-tester/internal/parser"
	"github.com/gaoyong06/api-tester/internal/types"
)

// ReportData 包含生成报告所需的数据
type ReportData struct {
	// API标题
	Title string
	// API版本
	Version string
	// API描述
	Description string
	// 测试时间
	Timestamp string
	// 测试结果
	Results []*types.EndpointTestResult
	// 总测试数
	Total int
	// 通过测试数
	Passed int
	// 失败测试数
	Failed int
	// 通过率
	PassRate float64
	// 总响应时间
	TotalResponseTime int64
	// 平均响应时间
	AvgResponseTime float64
}

// GenerateReport 生成测试报告
func GenerateReport(apiDef *parser.APIDefinition, results []*types.EndpointTestResult, outputDir string) (string, error) {
	// 创建输出目录
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("无法创建输出目录: %v", err)
	}

	// 准备报告数据
	total := len(results)
	passed := 0
	failed := 0
	totalResponseTime := int64(0)

	for _, result := range results {
		if result.Validation.Passed {
			passed++
		} else {
			failed++
		}
		totalResponseTime += result.Validation.ResponseTime
	}

	// 计算通过率和平均响应时间
	passRate := 0.0
	avgResponseTime := 0.0

	if total > 0 {
		passRate = float64(passed) / float64(total) * 100
		avgResponseTime = float64(totalResponseTime) / float64(total)
	}

	// 创建报告数据
	reportData := &ReportData{
		Title:             apiDef.Title,
		Version:           apiDef.Version,
		Description:       apiDef.Description,
		Timestamp:         time.Now().Format("2006-01-02 15:04:05"),
		Results:           results,
		Total:             total,
		Passed:            passed,
		Failed:            failed,
		PassRate:          passRate,
		TotalResponseTime: totalResponseTime,
		AvgResponseTime:   avgResponseTime,
	}

	// 生成报告文件名
	reportFileName := fmt.Sprintf("api-test-report-%s.html", time.Now().Format("20060102-150405"))
	reportPath := filepath.Join(outputDir, reportFileName)

	// 创建报告文件
	reportFile, err := os.Create(reportPath)
	if err != nil {
		return "", fmt.Errorf("无法创建报告文件: %v", err)
	}
	defer reportFile.Close()

	// 使用模板生成报告
	tmpl := template.New("report")

	// 添加自定义函数
	tmpl.Funcs(template.FuncMap{
		"lower": strings.ToLower,
		"statusClass": func(status int) string {
			if status >= 200 && status < 300 {
				return "2xx"
			} else if status >= 400 && status < 500 {
				return "4xx"
			} else if status >= 500 {
				return "5xx"
			}
			return ""
		},
		"join": strings.Join,
	})

	// 解析模板
	tmpl, err = tmpl.Parse(reportTemplate)
	if err != nil {
		return "", fmt.Errorf("无法解析报告模板: %v", err)
	}

	if err := tmpl.Execute(reportFile, reportData); err != nil {
		return "", fmt.Errorf("无法生成报告: %v", err)
	}

	return reportPath, nil
}

// 报告HTML模板
const reportTemplate = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>API测试报告 - {{.Title}}</title>
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
        }
        h1, h2, h3 {
            color: #2c3e50;
        }
        .header {
            border-bottom: 2px solid #eee;
            padding-bottom: 10px;
            margin-bottom: 20px;
        }
        .summary {
            display: flex;
            flex-wrap: wrap;
            gap: 20px;
            margin-bottom: 30px;
        }
        .summary-card {
            flex: 1;
            min-width: 200px;
            padding: 15px;
            border-radius: 8px;
            box-shadow: 0 2px 5px rgba(0,0,0,0.1);
        }
        .total {
            background-color: #f8f9fa;
        }
        .passed {
            background-color: #d4edda;
            color: #155724;
        }
        .failed {
            background-color: #f8d7da;
            color: #721c24;
        }
        .response-time {
            background-color: #e2e3e5;
            color: #383d41;
        }
        .endpoint {
            margin-bottom: 25px;
            border: 1px solid #ddd;
            border-radius: 8px;
            overflow: hidden;
        }
        .endpoint-header {
            padding: 12px 15px;
            display: flex;
            justify-content: space-between;
            align-items: center;
            cursor: pointer;
        }
        .endpoint-passed {
            background-color: #d4edda;
        }
        .endpoint-failed {
            background-color: #f8d7da;
        }
        .method {
            font-weight: bold;
            padding: 5px 10px;
            border-radius: 4px;
            color: white;
        }
        .get { background-color: #61affe; }
        .post { background-color: #49cc90; }
        .put { background-color: #fca130; }
        .delete { background-color: #f93e3e; }
        .patch { background-color: #50e3c2; }
        .head { background-color: #9012fe; }
        .options { background-color: #0d5aa7; }
        .endpoint-details {
            padding: 15px;
            border-top: 1px solid #ddd;
            background-color: #f8f9fa;
            display: none;
        }
        .show {
            display: block;
        }
        .response {
            background-color: #272822;
            color: #f8f8f2;
            padding: 10px;
            border-radius: 4px;
            overflow-x: auto;
            white-space: pre-wrap;
        }
        .status-code {
            font-weight: bold;
        }
        .status-2xx { color: #49cc90; }
        .status-4xx, .status-5xx { color: #f93e3e; }
        .footer {
            margin-top: 30px;
            text-align: center;
            color: #6c757d;
            font-size: 0.9em;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>API测试报告</h1>
        <p><strong>API名称:</strong> {{.Title}} <strong>版本:</strong> {{.Version}}</p>
        <p><strong>测试时间:</strong> {{.Timestamp}}</p>
        <p>{{.Description}}</p>
    </div>

    <div class="summary">
        <div class="summary-card total">
            <h3>总测试数</h3>
            <p>{{.Total}}</p>
        </div>
        <div class="summary-card passed">
            <h3>通过</h3>
            <p>{{.Passed}} ({{printf "%.1f" .PassRate}}%)</p>
        </div>
        <div class="summary-card failed">
            <h3>失败</h3>
            <p>{{.Failed}}</p>
        </div>
        <div class="summary-card response-time">
            <h3>平均响应时间</h3>
            <p>{{printf "%.2f" .AvgResponseTime}} ms</p>
        </div>
    </div>

    <h2>测试详情</h2>
    {{range $index, $result := .Results}}
    <div class="endpoint">
        <div class="endpoint-header {{if $result.Validation.Passed}}endpoint-passed{{else}}endpoint-failed{{end}}" onclick="toggleDetails({{$index}})">
            <div>
                <span class="method {{lower $result.Endpoint.Method}}">{{$result.Endpoint.Method}}</span>
                <span>{{$result.Endpoint.Path}}</span>
            </div>
            <div>
                <span class="status-code status-{{statusClass $result.Validation.ActualStatus}}">{{$result.Validation.ActualStatus}}</span>
                <span>{{$result.Validation.ResponseTime}} ms</span>
            </div>
        </div>
        <div id="details-{{$index}}" class="endpoint-details">
            <p><strong>操作ID:</strong> {{$result.Endpoint.OperationID}}</p>
            <p><strong>描述:</strong> {{$result.Endpoint.Description}}</p>
            {{if $result.Endpoint.Tags}}
            <p><strong>标签:</strong> {{join $result.Endpoint.Tags ", "}}</p>
            {{end}}
            
            {{if not $result.Validation.Passed}}
            <div>
                <h4>失败原因</h4>
                <p style="color: #721c24;">{{$result.Validation.FailureReason}}</p>
            </div>
            {{end}}
            
            <h4>响应详情</h4>
            <p><strong>状态码:</strong> <span class="status-code status-{{statusClass $result.Validation.ActualStatus}}">{{$result.Validation.ActualStatus}}</span></p>
            <p><strong>响应时间:</strong> {{$result.Validation.ResponseTime}} ms</p>
            <div>
                <h4>响应体</h4>
                <pre class="response">{{$result.Validation.ResponseBody}}</pre>
            </div>
        </div>
    </div>
    {{end}}

    <div class="footer">
        <p>由 API-Tester 生成 | {{.Timestamp}}</p>
    </div>

    <script>
        function toggleDetails(index) {
            const details = document.getElementById('details-' + index);
            details.classList.toggle('show');
        }

        // 默认展开失败的测试
        document.addEventListener('DOMContentLoaded', function() {
            {{range $index, $result := .Results}}
            {{if not $result.Validation.Passed}}
            document.getElementById('details-' + {{$index}}).classList.add('show');
            {{end}}
            {{end}}
        });
    </script>
</body>
</html>
`
