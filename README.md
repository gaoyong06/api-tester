# API-Tester

## 项目简介

API-Tester 是一个基于 Go 语言开发的 API 自动化测试工具。

**核心功能：**
- 根据 YAML 配置文件自动执行 API 测试
- 支持测试场景编排和步骤依赖
- 支持变量提取和数据传递
- 支持多种断言验证
- 生成详细的测试报告

**解决的问题：**
- 自动化 API 接口测试，减少手工测试工作量
- 支持复杂的测试场景和数据依赖关系
- 可集成到 CI/CD 流程中实现持续测试

## 安装

### 方式一：使用 go install（推荐）

```bash
go install github.com/gaoyong06/api-tester/cmd/api-tester@latest
```

### 方式二：从源码编译

```bash
git clone https://github.com/gaoyong06/api-tester.git
cd api-tester
go build -o bin/api-tester cmd/api-tester/main.go
```

## 快速开始

### 1. 创建配置文件

创建 `config.yaml` 文件：

```yaml
# API 基础配置
spec: ./openapi.yaml          # OpenAPI 规范文件路径
base_url: http://localhost:8080  # API 基础 URL
timeout: 30                   # 请求超时时间（秒）
verbose: true                 # 是否显示详细日志
output_dir: ./test/reports    # 测试报告输出目录

# 全局变量（可选）
variables:
  user_id: "1001"
  api_key: "test-key"

# 测试场景
scenarios:
  - name: 用户管理测试
    description: 测试用户的创建、查询和更新
    steps:
      - name: 创建用户
        endpoint: /users
        method: POST
        request_body:
          name: "测试用户"
          email: "test@example.com"
        assert:
          status: 201
        extract:
          user_id: $.id

      - name: 查询用户
        endpoint: /users/{id}
        method: GET
        dependencies: [创建用户]
        path_params:
          id: "{{.user_id}}"
        assert:
          status: 200
          body:
            $.name: "测试用户"
```

### 2. 运行测试

```bash
api-tester run --config config.yaml
```

## 配置文件说明

### 顶层配置

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `spec` | 字符串 | 是 | OpenAPI/Swagger 规范文件路径 |
| `base_url` | 字符串 | 是 | API 基础 URL |
| `timeout` | 整数 | 否 | 请求超时时间（秒），默认 30 |
| `verbose` | 布尔 | 否 | 是否显示详细日志，默认 false |
| `output_dir` | 字符串 | 否 | 测试报告输出目录，默认 `./test-reports` |
| `variables` | 对象 | 否 | 全局变量，可在测试中使用 |
| `scenarios` | 数组 | 是 | 测试场景列表 |

### 测试场景配置

每个场景包含以下字段：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `name` | 字符串 | 是 | 场景名称 |
| `description` | 字符串 | 否 | 场景描述 |
| `steps` | 数组 | 是 | 测试步骤列表 |

### 测试步骤配置

每个步骤包含以下字段：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `name` | 字符串 | 是 | 步骤名称 |
| `endpoint` | 字符串 | 是 | API 端点路径，支持占位符如 `/users/{id}` |
| `method` | 字符串 | 是 | HTTP 方法（GET、POST、PUT、DELETE 等） |
| `request_body` | 对象 | 否 | 请求体，支持模板变量 |
| `headers` | 对象 | 否 | 请求头，支持模板变量 |
| `path_params` | 对象 | 否 | 路径参数，用于替换 endpoint 中的占位符 |
| `query_params` | 对象 | 否 | 查询参数 |
| `dependencies` | 数组 | 否 | 依赖的步骤名称列表 |
| `extract` | 对象 | 否 | 从响应中提取变量，格式：`变量名: JSONPath表达式` |
| `assert` | 对象 | 否 | 断言规则 |

### 断言配置

| 字段 | 类型 | 说明 |
|------|------|------|
| `status` | 整数/数组 | 期望的 HTTP 状态码，如 `200` 或 `[200, 201]` |
| `body` | 对象 | 响应体断言，格式：`JSONPath: 期望值` |
| `headers` | 对象 | 响应头断言 |

## 使用示例

### 示例 1：简单的 API 测试

```yaml
spec: ./api.yaml
base_url: http://localhost:8080
scenarios:
  - name: 健康检查
    steps:
      - name: 检查服务状态
        endpoint: /health
        method: GET
        assert:
          status: 200
```

### 示例 2：带变量提取和传递

```yaml
spec: ./api.yaml
base_url: http://localhost:8080
scenarios:
  - name: 用户注册登录流程
    steps:
      - name: 注册用户
        endpoint: /register
        method: POST
        request_body:
          username: "testuser"
          password: "password123"
        assert:
          status: 201
        extract:
          user_id: $.id
          token: $.token

      - name: 使用 token 获取用户信息
        endpoint: /users/{id}
        method: GET
        dependencies: [注册用户]
        headers:
          Authorization: "Bearer {{.token}}"
        path_params:
          id: "{{.user_id}}"
        assert:
          status: 200
          body:
            $.username: "testuser"
```

### 示例 3：使用全局变量

```yaml
spec: ./api.yaml
base_url: http://localhost:8080
variables:
  admin_token: "admin-secret-token"
  page_size: "20"

scenarios:
  - name: 管理员操作
    steps:
      - name: 获取用户列表
        endpoint: /admin/users
        method: GET
        headers:
          Authorization: "Bearer {{.admin_token}}"
        query_params:
          limit: "{{.page_size}}"
        assert:
          status: 200
```

## 命令行参数

### run 命令

运行 API 测试：

```bash
api-tester run --config <配置文件路径>
```

**参数：**
- `--config`：配置文件路径（必填）
- `--verbose`：启用详细输出（可选）
- `--output`：输出目录路径（可选，默认 `./reports`）

### 其他命令

```bash
# 查看帮助
api-tester --help

# 查看版本
api-tester version

# 生成测试报告
api-tester report --results <结果文件路径>
```

## 变量和模板

### 变量来源

变量按以下优先级使用：
1. 步骤中显式定义的参数（`path_params`、`query_params`）
2. 从前面步骤提取的变量（`extract`）
3. 全局变量（`variables`）

### 模板语法

在配置中使用 `{{.变量名}}` 引用变量：

```yaml
endpoint: /users/{{.user_id}}
headers:
  Authorization: "Bearer {{.token}}"
request_body:
  name: "{{.username}}"
```

### 内置函数

```yaml
# 生成 UUID
user_id: "{{uuid}}"

# 当前时间戳
timestamp: "{{now}}"

# 随机字符串
random_str: "{{randomString 10}}"
```

## 断言说明

### 状态码断言

```yaml
# 单个状态码
assert:
  status: 200

# 多个可接受的状态码
assert:
  status: [200, 201, 204]
```

### 响应体断言

```yaml
assert:
  status: 200
  body:
    $.success: true           # 精确匹配
    $.data.id: "!null"        # 值不为 null
    $.data.name: "测试用户"    # 字符串匹配
    $.data.items.length: 5    # 数组长度
```

### 响应头断言

```yaml
assert:
  status: 200
  headers:
    Content-Type: "application/json"
    X-Request-ID: "!null"
```

## CI/CD 集成

### GitHub Actions 示例

```yaml
name: API Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      
      - name: Set up Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.22
      
      - name: Install API-Tester
        run: go install github.com/gaoyong06/api-tester/cmd/api-tester@latest
      
      - name: Run API tests
        run: api-tester run --config ./test-config.yaml
      
      - name: Upload test reports
        uses: actions/upload-artifact@v2
        with:
          name: test-reports
          path: ./test/reports
```

## 常见问题

### Q: 如何处理认证？

A: 在全局变量或步骤的 `headers` 中设置认证信息：

```yaml
variables:
  auth_token: "your-token"

scenarios:
  - name: 认证测试
    steps:
      - name: 访问受保护的资源
        endpoint: /protected
        method: GET
        headers:
          Authorization: "Bearer {{.auth_token}}"
```

### Q: 如何处理动态数据？

A: 使用 `extract` 从响应中提取数据，然后在后续步骤中使用：

```yaml
steps:
  - name: 创建资源
    endpoint: /resources
    method: POST
    extract:
      resource_id: $.id
  
  - name: 使用创建的资源
    endpoint: /resources/{{.resource_id}}
    method: GET
    dependencies: [创建资源]
```

### Q: 如何调试测试失败？

A: 使用 `--verbose` 参数查看详细日志：

```bash
api-tester run --config config.yaml --verbose
```

## 项目结构

```
api-tester/
├── cmd/api-tester/     # 命令行入口
├── internal/           # 内部实现
│   ├── config/        # 配置管理
│   ├── runner/        # 测试运行器
│   ├── scenario/      # 场景管理
│   └── reporter/      # 报告生成
└── pkg/               # 公共包
```

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！
