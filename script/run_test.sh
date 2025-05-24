#!/bin/bash

# API-Tester 运行脚本
# 此脚本用于简化 API-Tester 的使用

# 颜色定义
GREEN="\033[0;32m"
YELLOW="\033[0;33m"
RED="\033[0;31m"
NC="\033[0m" # 无颜色

# 获取脚本所在目录的绝对路径
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# 帮助信息
show_help() {
    echo -e "${GREEN}API-Tester 使用帮助${NC}"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -s, --spec <文件路径>     OpenAPI/Swagger 规范文件路径 (必填)"
    echo "  -u, --url <URL>           API 基础 URL (必填)"
    echo "  -h, --headers <JSON>      请求头 (JSON 格式)"
    echo "  -o, --output <目录>       测试报告输出目录"
    echo "  -v, --verbose             显示详细日志"
    echo "  -t, --timeout <秒>        请求超时时间 (秒)"
    echo "  --help                    显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 -s ./api/openapi.yaml -u http://localhost:8000"
    echo "  $0 -s ./api/openapi.yaml -u http://localhost:8000 -h '{\"X-API-Key\":\"test-key\"}' -v"
    echo ""
}

# 检查是否已编译
check_binary() {
    if [ ! -f "$PROJECT_ROOT/bin/api-tester" ]; then
        echo -e "${YELLOW}API-Tester 二进制文件不存在，正在编译...${NC}"
        mkdir -p "$PROJECT_ROOT/bin"
        cd "$PROJECT_ROOT" && go build -o bin/api-tester cmd/api-tester/main.go
        if [ $? -ne 0 ]; then
            echo -e "${RED}编译失败！${NC}"
            exit 1
        fi
        echo -e "${GREEN}编译成功！${NC}"
    fi
}

# 参数解析
SPEC_FILE=""
BASE_URL=""
HEADERS=""
OUTPUT_DIR=""
VERBOSE=false
TIMEOUT=""

while [[ $# -gt 0 ]]; do
    case $1 in
        -s|--spec)
            SPEC_FILE="$2"
            shift 2
            ;;
        -u|--url)
            BASE_URL="$2"
            shift 2
            ;;
        -h|--headers)
            HEADERS="$2"
            shift 2
            ;;
        -o|--output)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -t|--timeout)
            TIMEOUT="$2"
            shift 2
            ;;
        --help)
            show_help
            exit 0
            ;;
        *)
            echo -e "${RED}未知选项: $1${NC}"
            show_help
            exit 1
            ;;
    esac
done

# 验证必填参数
if [ -z "$SPEC_FILE" ] || [ -z "$BASE_URL" ]; then
    echo -e "${RED}错误: 规范文件路径和 API 基础 URL 为必填项${NC}"
    show_help
    exit 1
fi

# 检查规范文件是否存在
if [ ! -f "$SPEC_FILE" ]; then
    echo -e "${RED}错误: 规范文件 '$SPEC_FILE' 不存在${NC}"
    exit 1
fi

# 构建命令
CMD="$PROJECT_ROOT/bin/api-tester -spec $SPEC_FILE -url $BASE_URL"

if [ ! -z "$HEADERS" ]; then
    CMD="$CMD -headers '$HEADERS'"
fi

if [ ! -z "$OUTPUT_DIR" ]; then
    CMD="$CMD -output $OUTPUT_DIR"
fi

if [ "$VERBOSE" = true ]; then
    CMD="$CMD -verbose"
fi

if [ ! -z "$TIMEOUT" ]; then
    CMD="$CMD -timeout $TIMEOUT"
fi

# 检查并编译二进制文件
check_binary

# 运行测试
echo -e "${GREEN}开始运行 API 测试...${NC}"
echo "执行命令: $CMD"
echo ""

eval $CMD

if [ $? -eq 0 ]; then
    echo -e "\n${GREEN}测试完成！${NC}"
else
    echo -e "\n${RED}测试过程中出现错误！${NC}"
    exit 1
fi
