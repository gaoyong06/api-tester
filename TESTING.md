# 微服务 API 测试指南

## 概述

本文档说明如何使用 api-tester 对 Passport、Payment 和 Subscription 三个微服务进行自动化测试。

## Bug 修复记录

### 已修复的问题

1. **废弃的 rand.Seed() 调用**
   - **问题**: 使用了 Go 1.20+ 已废弃的 `rand.Seed()` 函数
   - **修复**: 移除 `rand.Seed()` 调用，Go 1.20+ 会自动初始化随机数种子
   - **文件**: `cmd/api-tester/main.go`

2. **Go Workspace 配置**
   - **问题**: api-tester 未添加到 workspace
   - **修复**: 在 `go.work` 中添加 `./api-tester`
   - **影响**: 现在可以正常编译和运行

## 测试覆盖范围

### Passport Service (端口 9000)
- ✅ 用户注册
- ✅ 用户登录
- ✅ 获取用户信息
- ✅ JWT Token 验证

### Payment Service (端口 9001)
- ✅ 创建支付订单
- ✅ 查询支付状态
- ✅ 支付回调处理
- ✅ 支付状态更新

### Subscription Service (端口 9002)
- ✅ 获取套餐列表
- ✅ 查询订阅状态
- ✅ 创建订阅订单
- ✅ 处理支付成功
- ✅ 验证订阅激活

### 集成测试
- ✅ 完整用户流程（注册 → 查看套餐 → 购买 → 支付 → 验证）

## 快速开始

### 前置条件

1. 所有微服务已启动
2. MySQL 数据库已运行
3. api-tester 已编译

### 方式一：使用测试脚本（推荐）

```bash
cd api-tester
./run-tests.sh
```

脚本会自动：
- 检查所有服务状态
- 编译 api-tester（如果需要）
- 运行所有测试
- 生成测试报告

### 方式二：手动运行

```bash
# 1. 编译 api-tester
cd api-tester
go build -o bin/api-tester ./cmd/api-tester

# 2. 运行测试
./bin/api-tester run \
    --config test-config.yaml \
    --output ./test-reports \
    --report-type html \
    --verbose
```

## 测试配置

测试配置文件位于 `test-config.yaml`，包含以下场景：

### 1. Passport Service 测试
```yaml
scenarios:
  - name: Passport Service - 用户注册登录流程
    steps:
      - 注册新用户
      - 用户登录
      - 获取用户信息
```

### 2. Payment Service 测试
```yaml
scenarios:
  - name: Payment Service - 支付流程测试
    steps:
      - 创建支付订单
      - 查询支付状态
      - 模拟支付成功回调
      - 验证支付成功
```

### 3. Subscription Service 测试
```yaml
scenarios:
  - name: Subscription Service - 订阅管理测试
    steps:
      - 获取套餐列表
      - 查询用户订阅状态
      - 创建订阅订单
      - 模拟支付成功
      - 验证订阅已激活
```

### 4. 完整流程测试
```yaml
scenarios:
  - name: 完整用户流程测试
    steps:
      - 注册新用户
      - 查看订阅套餐
      - 购买订阅
      - 完成支付
      - 验证会员状态
```

## 测试报告

测试完成后会生成以下报告：

- **HTML 报告**: `test-reports/report.html`（可视化报告）
- **JSON 报告**: `test-reports/results.json`（机器可读）

查看 HTML 报告：
```bash
open test-reports/report.html
```

## 自定义测试

### 添加新测试场景

编辑 `test-config.yaml`：

```yaml
scenarios:
  - name: 你的测试场景
    description: 场景描述
    base_url: http://localhost:9000
    steps:
      - name: 测试步骤1
        endpoint: /v1/your/endpoint
        method: POST
        request_body:
          field: "value"
        assert:
          status: 200
        extract:
          var_name: $.response.field
```

### 使用变量

在 `variables` 部分定义全局变量：

```yaml
variables:
  test_user: "testuser"
  test_email: "test@example.com"
```

在测试中使用：

```yaml
request_body:
  username: "{{.test_user}}"
  email: "{{.test_email}}"
```

### 数据提取和依赖

从响应中提取数据：

```yaml
extract:
  user_id: $.id
  token: $.access_token
```

在后续步骤中使用：

```yaml
dependencies: [前一个步骤名称]
headers:
  Authorization: "Bearer {{.token}}"
path_params:
  id: "{{.user_id}}"
```

## 故障排查

### 服务未启动

```bash
# 检查服务状态
docker-compose ps

# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f
```

### 测试失败

1. 查看详细日志：
```bash
./bin/api-tester run --config test-config.yaml --verbose
```

2. 查看服务日志：
```bash
docker-compose logs passport-service
docker-compose logs payment-service
docker-compose logs subscription-service
```

3. 检查数据库：
```bash
docker exec -it mysql mysql -uroot -proot
USE passport_service;
SELECT * FROM user;
```

### 端口冲突

如果端口被占用：

```bash
# 查看占用端口的进程
lsof -i :9000
lsof -i :9001
lsof -i :9002

# 停止进程
kill -9 <PID>
```

## CI/CD 集成

### GitHub Actions

```yaml
name: API Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      
      - name: Start services
        run: docker-compose up -d
      
      - name: Wait for services
        run: sleep 30
      
      - name: Run API tests
        run: |
          cd api-tester
          go build -o bin/api-tester ./cmd/api-tester
          ./bin/api-tester run --config test-config.yaml --report-type junit
      
      - name: Publish test results
        uses: EnricoMi/publish-unit-test-result-action@v1
        if: always()
        with:
          files: api-tester/test-reports/*.xml
```

### Jenkins

```groovy
pipeline {
    agent any
    stages {
        stage('Start Services') {
            steps {
                sh 'docker-compose up -d'
                sh 'sleep 30'
            }
        }
        stage('Run Tests') {
            steps {
                dir('api-tester') {
                    sh 'go build -o bin/api-tester ./cmd/api-tester'
                    sh './bin/api-tester run --config test-config.yaml --report-type junit'
                }
            }
        }
    }
    post {
        always {
            junit 'api-tester/test-reports/*.xml'
            sh 'docker-compose down'
        }
    }
}
```

## 最佳实践

1. **测试隔离**: 每个测试使用唯一的测试数据
2. **清理数据**: 测试后清理创建的数据
3. **断言明确**: 使用具体的断言而不是只检查状态码
4. **依赖管理**: 正确设置步骤依赖关系
5. **错误处理**: 为失败场景编写测试

## 性能测试

对于性能测试，可以使用 `--count` 参数：

```bash
./bin/api-tester run \
    --config test-config.yaml \
    --count 100 \
    --concurrent 10
```

## 相关文档

- [api-tester README](README.md)
- [Passport Service API](../passport-service/README.md)
- [Payment Service API](../payment-service/README.md)
- [Subscription Service API](../subscription-service/README.md)
- [微服务总览](../README.md)
