# API-Tester

## 项目简介

API-Tester 是一个基于 Go 语言开发的通用 API 自动化测试工具，可以根据 OpenAPI/Swagger 规范文件自动测试 API 接口并生成详细的测试报告。本工具支持模板处理、测试数据管理、测试场景定义、模拟数据生成和多种格式的测试报告输出，可以无缝对接到其他项目的测试流程中。

## 主要功能

- 自动解析 OpenAPI/Swagger 规范文件
- 自动构建和发送 API 请求
- 验证 API 响应是否符合预期
- 支持复杂的测试场景和数据依赖
- 支持测试数据初始化和清理
- 支持模板变量和函数
- 支持模拟数据生成
- 生成多种格式的测试报告（HTML、JSON、XML、JUnit）
- 支持 API 规范格式转换
- 详细的命令行输出和日志

## 安装方法

### 前置条件

- Go 1.16 或更高版本

### 安装步骤

```bash
# 克隆仓库
git clone [https://github.com/gaoyong06/api-tester.git](https://github.com/gaoyong06/api-tester.git)

# 进入项目目录
cd api-tester

# 安装依赖
go mod tidy

# 编译项目
go build -o bin/api-tester cmd/api-tester/main.go
```

## 快速开始

### 基本使用

API-Tester 提供了多个子命令，每个子命令负责不同的功能。以下是基本的使用方法：

```bash

# 运行 API 测试
./bin/api-tester run --spec /path/to/openapi.yaml --url http://api.example.com

# 初始化测试数据
./bin/api-tester init-data --config /path/to/config.yaml

# 生成测试报告
./bin/api-tester report --spec /path/to/openapi.yaml --results /path/to/results.json

# 生成模拟数据
./bin/api-tester mock --spec /path/to/openapi.yaml --schema User --count 10

# 转换 API 规范格式
./bin/api-tester convert --input /path/to/swagger.json --format yaml --version 3
```

### 详细命令说明

#### 全局选项

| 选项          | 描述                              | 默认值        |
| ------------- | --------------------------------- | ------------- |
| --config      | 配置文件路径                      | ./config.yaml |
| --verbose     | 启用详细输出                      | false         |
| --output      | 输出目录路径                      | ./reports     |
| --report-type | 报告类型 (html, json, xml, junit) | html          |

#### run 命令

运行 API 测试并生成报告。

```bash
./bin/api-tester run --spec /path/to/openapi.yaml --url http://api.example.com
```

| 选项             | 描述                         | 是否必填               | 默认值 |
| ---------------- | ---------------------------- | ---------------------- | ------ |
| --spec           | OpenAPI/Swagger 规范文件路径 | 是（如果没有配置文件） | -      |
| --url            | API 基础 URL                 | 是（如果没有配置文件） | -      |
| --headers        | 请求头 (JSON 格式)           | 否                     | -      |
| --timeout        | 请求超时时间 (秒)            | 否                     | 30     |
| --path-params    | 路径参数文件 (JSON 格式)     | 否                     | -      |
| --request-bodies | 请求体模板文件 (JSON 格式)   | 否                     | -      |
| --scenario       | 测试场景文件 (YAML 格式)     | 否                     | -      |

#### init-data 命令

初始化测试数据，支持多种数据源。

```bash
./bin/api-tester init-data --data-source /path/to/datasources.yaml
```

| 选项          | 描述                       | 是否必填               | 默认值 |
| ------------- | -------------------------- | ---------------------- | ------ |
| --data-source | 数据源定义文件 (YAML 格式) | 是（如果没有配置文件） | -      |
| --cleanup     | 测试完成后清理数据         | 否                     | false  |

#### report 命令

从已有测试结果生成报告。

```bash
./bin/api-tester report --results /path/to/results.json
```

| 选项          | 描述             | 是否必填               | 默认值 |
| ------------- | ---------------- | ---------------------- | ------ |
| --results     | 测试结果文件路径 | 是（如果没有配置文件） | -      |
| --title       | 报告标题         | 否                     | -      |
| --description | 报告描述         | 否                     | -      |

#### mock 命令

生成模拟数据，基于 OpenAPI/Swagger 规范

```bash
./bin/api-tester mock --spec /path/to/openapi.yaml --schema User --count 5
```

| 选项          | 描述                           | 是否必填             | 默认值   |
| ------------- | ------------------------------ | -------------------- | -------- |
| --spec        | OpenAPI/Swagger 规范文件路径   | 是                   | -        |
| --schema      | 要生成数据的模式名称或路径     | 否（如果使用 --all） | -        |
| --output-file | 输出文件路径                   | 否                   | 标准输出 |
| --count       | 生成的数据数量                 | 否                   | 1        |
| --rules       | 自定义生成规则文件 (JSON 格式) | 否                   | -        |
| --all         | 生成所有模式的数据             | 否                   | false    |
| --pretty      | 美化输出 JSON                  | 否                   | true     |

#### convert 命令

转换 API 规范格式，支持不同版本和格式之间的转换。

```bash
./bin/api-tester convert --input /path/to/swagger.json --format yaml --version 3
```

| 选项      | 描述                    | 是否必填 | 默认值             |
| --------- | ----------------------- | -------- | ------------------ |
| --input   | 输入文件路径            | 是       | -                  |
| --output  | 输出文件路径            | 否       | 根据输入文件名生成 |
| --format  | 输出格式 (json 或 yaml) | 否       | yaml               |
| --version | OpenAPI 版本 (2 或 3)   | 否       | 3                  |

## 配置文件

API-Tester 支持使用 YAML 格式的配置文件来管理测试配置。以下是一个完整的配置文件示例：

```yaml
# API 配置
api:
  # OpenAPI/Swagger 规范文件路径
  specFile: ./api/openapi.yaml
  # API 基础 URL
  baseURL: http://api.example.com
  # 请求头
  headers:
    Authorization: Bearer token
    Content-Type: application/json
  # 请求超时时间（秒）
  timeout: 30
  # 路径参数
  pathParams:
    userId: "12345"
    productId: "67890"
  # 请求体模板
  requestBodies:
    createUser:
      name: "{{.name}}"
      email: "{{.email}}"
      age: "{{.age}}"

# 输出配置
output:
  # 输出目录
  directory: ./reports
  # 报告类型
  reportType: html
  # 详细日志
  verbose: true

# 测试数据配置
testData:
  # 数据源
  dataSources:
    - name: users
      type: file
      path: ./testdata/users.json
    - name: products
      type: sql
      path: ./testdata/init_products.sql
  # 初始化脚本
  initScripts:
    - name: create-tables
      type: sql
      content: |
        CREATE TABLE IF NOT EXISTS users (
          id VARCHAR(36) PRIMARY KEY,
          name VARCHAR(100) NOT NULL,
          email VARCHAR(100) UNIQUE NOT NULL
        );
  # 清理脚本
  cleanupScripts:
    - name: drop-tables
      type: sql
      content: DROP TABLE IF EXISTS users;

# 测试场景配置
scenarios:
  - name: 用户管理测试
    description: 测试用户创建、查询和删除功能
    steps:
      - name: 创建用户
        path: /users
        method: POST
        body:
          name: "测试用户"
          email: "test@example.com"
        extract:
          userId: $.id
        assert:
          status: 201
      - name: 查询用户
        path: /users/{{.userId}}
        method: GET
        depends: [创建用户]
        assert:
          status: 200
          body:
            $.name: "测试用户"
      - name: 删除用户
        path: /users/{{.userId}}
        method: DELETE
        depends: [查询用户]
        assert:
          status: 204
```

## 模板使用指南
API-Tester 支持在请求路径、请求参数和请求体中使用模板变量和函数。模板使用 Go 的模板语法，用 {{}} 包裹。

### 基本变量替换

```yaml
# 在路径中使用变量
path: /users/{{.userId}}

# 在请求体中使用变量
body:
  name: "{{.userName}}"
  email: "{{.userEmail}}"
```

### 内置函数

API-Tester 提供了多种内置函数，可以在模板中使用：

#### 时间函数

```yaml
# 当前时间
"{{now}}"

# 格式化时间
"{{formatTime "2006-01-02"}}"

# 添加天数
"{{addDays 7}}"
```

#### 随机数函数

```yaml
# 生成 UUID
"{{uuid}}"

# 生成随机数（范围内）
"{{random 1 100}}"

# 生成随机字符串
"{{randomString 10}}"
```

#### 编码函数

```yaml
# Base64 编码
"{{base64 "hello"}}"

# Base64 解码
"{{base64decode "aGVsbG8="}}"
```

#### 字符串函数

```yaml
# 转小写
"{{lower "TEXT"}}"

# 转大写
"{{upper "text"}}"

# 替换
"{{replace "hello" "l" "x"}}"
```

#### 数学函数

```yaml
# 加法
"{{add 5 3}}"

# 减法
"{{sub 10 5}}"

# 乘法
"{{mul 2 4}}"
```

#### 条件函数

```yaml
# 条件判断
"{{ifThen (eq .status "active") "激活" "未激活"}}"
```

#### 自定义函数

您可以在测试代码中添加自定义函数，然后在模板中使用。

## 测试数据管理

### 数据源定义

API-Tester 支持多种数据源类型：

#### 文件数据源

```yaml
dataSources:
  - name: users
    type: file
    path: ./testdata/users.json
```

文件内容示例（JSON）：

```json
[
  {
    "id": "1",
    "name": "张三",
    "email": "zhangsan@example.com"
  },
  {
    "id": "2",
    "name": "李四",
    "email": "lisi@example.com"
  }
]
```

#### SQL 数据源

```yaml
dataSources:
  - name: products
    type: sql
    path: ./testdata/init_products.sql
```

SQL 文件内容示例：

```sql
INSERT INTO products (id, name, price) VALUES ('1', '产品1', 99.99);
INSERT INTO products (id, name, price) VALUES ('2', '产品2', 199.99);
```

#### API 数据源

```yaml
dataSources:
  - name: categories
    type: api
    path: http://api.example.com/categories
    headers:
      Authorization: Bearer token
```

## 初始化和清理脚本

```yaml
# 初始化脚本
initScripts:
  - name: create-tables
    type: sql
    content: |
      CREATE TABLE IF NOT EXISTS users (
        id VARCHAR(36) PRIMARY KEY,
        name VARCHAR(100) NOT NULL,
        email VARCHAR(100) UNIQUE NOT NULL
      );

# 清理脚本
cleanupScripts:
  - name: drop-tables
    type: sql
    content: DROP TABLE IF EXISTS users;
```

## 测试场景定义

测试场景允许您定义一系列相互依赖的测试步骤，每个步骤可以从前面的步骤中提取数据。

```yaml
scenarios:
  - name: 用户注册和登录
    description: 测试用户注册和登录流程
    steps:
      - name: 注册用户
        path: /register
        method: POST
        body:
          username: "testuser"
          email: "test@example.com"
          password: "password123"
        extract:
          userId: $.id
          activationCode: $.activationCode
        assert:
          status: 201

      - name: 激活账户
        path: /activate
        method: POST
        depends: [注册用户]
        body:
          userId: "{{.userId}}"
          activationCode: "{{.activationCode}}"
        assert:
          status: 200

      - name: 登录
        path: /login
        method: POST
        depends: [激活账户]
        body:
          username: "testuser"
          password: "password123"
        extract:
          token: $.token
        assert:
          status: 200
          body:
            $.success: true

      - name: 获取用户信息
        path: /users/{{.userId}}
        method: GET
        depends: [登录]
        headers:
          Authorization: "Bearer {{.token}}"
        assert:
          status: 200
          body:
            $.username: "testuser"
            $.email: "test@example.com"
```

### 步骤定义

每个测试步骤包含以下部分：

- name：步骤名称，用于在依赖关系中引用
- path：API 路径，支持模板变量
- method：HTTP 方法（GET、POST、PUT、DELETE 等）
- depends：依赖的步骤列表，只有当所有依赖步骤成功后才会执行当前步骤
- headers：请求头，支持模板变量
- params：查询参数，支持模板变量
- body：请求体，支持模板变量
- extract：从响应中提取数据的规则，使用 JSONPath 表达式
- assert：断言规则，用于验证响应是否符合预期

### 数据提取

使用 extract 字段可以从响应中提取数据，然后在后续步骤中使用

```yaml
extract:
  token: $.token
  userId: $.user.id
```

提取的数据可以在后续步骤中通过模板变量引用：

```yaml
path: /users/{{.userId}}
headers:
  Authorization: "Bearer {{.token}}"
```

### 断言

使用 assert 字段可以验证响应是否符合预期：

```yaml
assert:
  status: 200
  headers:
    Content-Type: application/json
  body:
    $.success: true
    $.user.name: "测试用户"
    $.items.length: 5
```

## 模拟数据生成

API-Tester 可以根据 OpenAPI/Swagger 规范生成模拟数据，用于测试或开发。

### 生成单个模式的数据

```bash
./bin/api-tester mock --spec /path/to/openapi.yaml --schema User --count 5 --output-file ./mock/users.json
```

### 生成所有模式的数据

```bash
./bin/api-tester mock --spec /path/to/openapi.yaml --all --output-file ./mock/all.json
```

### 自定义生成规则

您可以创建一个 JSON 文件来定义自定义生成规则：

```json
{
  "email": "test@example.com",
  "regex:.*Id$": "fixed-id-12345",
  "regex:.*name": "测试名称"
}
```

然后在命令中使用该规则文件：

```
./bin/api-tester mock --spec /path/to/openapi.yaml --schema User --rules ./rules.json
```

## 测试报告

API-Tester 支持生成多种格式的测试报告：

### HTML 报告

默认生成 HTML 格式的报告，包含测试概要、详细结果和图表。

```bash
./bin/api-tester run --spec /path/to/openapi.yaml --url http://api.example.com --report-type html
```

### JSON/XML 报告

生成机器可读的 JSON 或 XML 格式报告，便于与其他系统集成。

```bash
./bin/api-tester run --spec /path/to/openapi.yaml --url http://api.example.com --report-type json
```

```bash
./bin/api-tester run --spec /path/to/openapi.yaml --url http://api.example.com --report-type xml
```

### JUnit 报告


生成 JUnit 格式的报告，便于与 CI/CD 系统集成。

```bash
./bin/api-tester run --spec /path/to/openapi.yaml --url http://api.example.com --report-type junit
```

## 与其他项目集成

### 步骤 1：准备 OpenAPI/Swagger 规范文件

确保您的项目有一个有效的 OpenAPI/Swagger 规范文件。如果没有，可以使用 API-Tester 的转换功能从其他格式转换：

```bash
./bin/api-tester convert --input /path/to/api-docs.json --format yaml --version 3
```

### 步骤 2：创建配置文件

创建一个 config.yaml 文件，包含 API 测试的所有配置：

```yaml
api:
  specFile: ./api/openapi.yaml
  baseURL: http://localhost:8080
  headers:
    Authorization: Bearer {{.token}}
    Content-Type: application/json

output:
  directory: ./test-reports
  reportType: junit
  verbose: true

testData:
  dataSources:
    - name: testdata
      type: file
      path: ./testdata/data.json
  initScripts:
    - name: setup
      type: sql
      content: |
        INSERT INTO test_config (key, value) VALUES ('test_mode', 'true');

scenarios:
  - name: 基本功能测试
    steps:
      - name: 登录
        path: /auth/login
        method: POST
        body:
          username: "admin"
          password: "admin123"
        extract:
          token: $.token
        assert:
          status: 200

      # 更多测试步骤...
```

### 步骤 3：初始化测试数据

```bash
./bin/api-tester init-data --config ./config.yaml
```

### 步骤 4：运行测试

```bash
./bin/api-tester run --config ./config.yaml
```

### 步骤 5：生成测试报告

```bash
./bin/api-tester report --spec ./api/openapi.yaml --results ./test-reports/results.json
```

### 步骤 6：在 CI/CD 流程中集成

在您的 CI/CD 配置文件中添加测试步骤：

```yaml
# GitHub Actions 示例
jobs:
  api-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2

      - name: Set up Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.16

      - name: Install API-Tester
        run: |
          git clone https://github.com/gaoyong06/api-tester.git
          cd api-tester
          go mod tidy
          go build -o bin/api-tester cmd/api-tester/main.go

      - name: Initialize test data
        run: ./api-tester/bin/api-tester init-data --config ./test/config.yaml

      - name: Run API tests
        run: ./api-tester/bin/api-tester run --config ./test/config.yaml

      - name: Upload test reports
        uses: actions/upload-artifact@v2
        with:
          name: api-test-reports
          path: ./test-reports
```

## 常见问题解决

### 模板变量不生效

- 确保变量名称正确，区分大小写
- 检查变量是否已在前面的步骤中提取
- 检查模板语法是否正确，变量应该用 {{.variableName}} 格式

### 数据初始化失败

- 检查数据源路径是否正确
- 确保数据文件格式正确（JSON、SQL 等）
- 检查数据库连接配置（如果使用 SQL 数据源）

### 测试场景依赖关系错误

- 确保依赖的步骤名称正确，区分大小写
- 检查是否有循环依赖
- 确保所有依赖的步骤都能成功执行

### 断言失败

- 检查 API 响应是否符合预期
- 确保 JSONPath 表达式正确
- 使用 --verbose 选项查看详细日志

## 项目结构

```
api-tester/
├── cmd/                 # 命令行入口
│   └── api-tester/      # 主程序
│       ├── cmd/         # 子命令实现
│       └── main.go      # 主入口
├── internal/            # 内部包
│   ├── config/          # 配置管理
│   │   └── yaml/        # YAML 配置
│   ├── mock/            # 模拟数据生成
│   ├── parser/          # OpenAPI 解析
│   ├── reporter/        # 报告生成
│   │   └── machine/     # 机器可读报告
│   ├── runner/          # 测试运行器
│   ├── scenario/        # 测试场景
│   ├── template/        # 模板处理
│   └── types/           # 通用类型
├── pkg/                 # 公共包
│   ├── client/          # HTTP 客户端
│   └── utils/           # 工具函数
└── reports/             # 测试报告输出目录
```
