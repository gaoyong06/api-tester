#!/bin/bash

# 微服务 API 自动化测试脚本
# 用于测试 Passport、Payment 和 Subscription 三个服务

set -e

echo "========================================="
echo "微服务 API 自动化测试"
echo "========================================="
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查服务是否运行
check_service() {
    local service_name=$1
    local port=$2
    
    echo -n "检查 $service_name (端口 $port)... "
    if nc -z localhost $port 2>/dev/null; then
        echo -e "${GREEN}✓ 运行中${NC}"
        return 0
    else
        echo -e "${RED}✗ 未运行${NC}"
        return 1
    fi
}

# 等待服务启动
wait_for_service() {
    local service_name=$1
    local port=$2
    local max_attempts=30
    local attempt=1
    
    echo "等待 $service_name 启动..."
    while [ $attempt -le $max_attempts ]; do
        if nc -z localhost $port 2>/dev/null; then
            echo -e "${GREEN}$service_name 已启动${NC}"
            return 0
        fi
        echo "尝试 $attempt/$max_attempts..."
        sleep 1
        ((attempt++))
    done
    
    echo -e "${RED}$service_name 启动超时${NC}"
    return 1
}

# 步骤 1: 检查所有服务状态
echo "步骤 1: 检查服务状态"
echo "-----------------------------------"

all_services_running=true

if ! check_service "Passport Service" 9000; then
    all_services_running=false
fi

if ! check_service "Payment Service" 9001; then
    all_services_running=false
fi

if ! check_service "Subscription Service" 9002; then
    all_services_running=false
fi

if ! check_service "MySQL" 3306; then
    all_services_running=false
fi

echo ""

# 如果服务未运行，提示启动
if [ "$all_services_running" = false ]; then
    echo -e "${YELLOW}警告: 部分服务未运行${NC}"
    echo ""
    echo "请先启动所有服务："
    echo "  方式1: docker-compose up -d"
    echo "  方式2: 手动启动各服务"
    echo ""
    read -p "是否继续测试? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# 步骤 2: 运行测试
echo "步骤 2: 运行 API 测试"
echo "-----------------------------------"

# 确保 api-tester 已编译
if [ ! -f "./bin/api-tester" ]; then
    echo "编译 api-tester..."
    go build -o bin/api-tester ./cmd/api-tester
fi

# 创建报告目录
mkdir -p test-reports

# 运行测试
echo "开始执行测试..."
echo ""

./bin/api-tester run \
    --config test-config.yaml \
    --output ./test-reports \
    --report-type html \
    --verbose

test_exit_code=$?

echo ""
echo "========================================="
echo "测试完成"
echo "========================================="
echo ""

# 步骤 3: 显示测试结果
if [ $test_exit_code -eq 0 ]; then
    echo -e "${GREEN}✓ 所有测试通过${NC}"
    echo ""
    echo "测试报告已生成:"
    echo "  HTML: ./test-reports/report.html"
    echo "  JSON: ./test-reports/results.json"
    echo ""
    echo "查看报告:"
    echo "  open ./test-reports/report.html"
else
    echo -e "${RED}✗ 部分测试失败${NC}"
    echo ""
    echo "请查看详细报告:"
    echo "  ./test-reports/report.html"
    exit 1
fi
