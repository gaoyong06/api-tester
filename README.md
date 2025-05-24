# API-Tester

## 项目简介

API-Tester 是一个基于 Go 语言开发的通用 API 自动化测试工具，可以根据 OpenAPI/Swagger 规范文件自动测试 API 接口并生成详细的测试报告。

## 主要功能

- 自动解析 OpenAPI/Swagger 规范文件
- 自动构建和发送 API 请求
- 验证 API 响应是否符合预期
- 生成美观的 HTML 测试报告
- 支持自定义请求头和超时设置
- 详细的命令行输出和日志

## 安装方法

### 前置条件

- Go 1.16 或更高版本

### 安装步骤

```bash
# 克隆仓库
git clone https://github.com/gaoyong06/api-tester.git

# 进入项目目录
cd api-tester

# 安装依赖
go mod tidy

# 编译项目
go build -o bin/api-tester cmd/api-tester/main.go
```

## 使用方法

```bash
# 基本用法
./bin/api-tester -spec /path/to/openapi.yaml -url http://api.example.com

# 添加自定义请求头
./bin/api-tester -spec /path/to/openapi.yaml -url http://api.example.com -headers '{"Authorization":"Bearer token"}'

# 指定输出目录
./bin/api-tester -spec /path/to/openapi.yaml -url http://api.example.com -output ./reports

# 显示详细日志
./bin/api-tester -spec /path/to/openapi.yaml -url http://api.example.com -verbose

# 设置请求超时
./bin/api-tester -spec /path/to/openapi.yaml -url http://api.example.com -timeout 60
```

## 命令行参数

| 参数 | 描述 | 是否必填 | 默认值 |
|------|------|----------|--------|
| `-spec` | OpenAPI/Swagger 规范文件路径 | 是 | - |
| `-url` | API 基础 URL | 是 | - |
| `-headers` | 请求头 (JSON 格式) | 否 | - |
| `-output` | 测试报告输出目录 | 否 | `./reports` |
| `-verbose` | 显示详细日志 | 否 | `false` |
| `-timeout` | 请求超时时间 (秒) | 否 | `30` |

## 测试报告

测试完成后，工具会在指定的输出目录生成一个 HTML 格式的测试报告，包含以下内容：

- 测试概要（总测试数、通过数、失败数、通过率）
- 平均响应时间
- 每个端点的详细测试结果
- 失败原因分析（如果有）
- 实际响应内容

## 示例

```bash
# 测试本地开发的桌位安排系统 API
./bin/api-tester -spec ./api/openapi/v1/openapi.yaml -url http://localhost:8000 -headers '{"X-API-Key":"test-api-key"}' -verbose
```

## 项目结构

```
api-tester/
├── cmd/                 # 命令行入口
├── internal/            # 内部包
│   ├── config/          # 配置管理
│   ├── parser/          # OpenAPI 解析
│   ├── reporter/        # 报告生成
│   ├── runner/          # 测试运行器
│   └── validator/       # 响应验证
├── pkg/                 # 公共包
│   ├── client/          # HTTP 客户端
│   └── utils/           # 工具函数
└── reports/             # 测试报告输出目录
```

## 许可证

MIT
