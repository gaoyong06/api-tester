GOPATH:=$(shell go env GOPATH)
VERSION=$(shell git describe --tags --always)
TOOL_NAME=api-tester

.PHONY: build
# 构建 api-tester 工具
build:
	@echo "构建 api-tester 工具..."
	mkdir -p bin/
	go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/api-tester ./cmd/api-tester
	@echo "构建完成: ./bin/api-tester"

.PHONY: install
# 安装到系统路径
install: build
	@echo "安装 api-tester 到系统路径..."
	cp ./bin/api-tester /usr/local/bin/api-tester || sudo cp ./bin/api-tester /usr/local/bin/api-tester
	@echo "安装完成"

.PHONY: clean
# 清理生成的文件
clean:
	rm -rf bin/
	rm -rf reports/

.PHONY: test
# 运行测试
test:
	go test -v ./...

.PHONY: help
# 显示帮助信息
help:
	@echo "API Tester Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build        - 构建 api-tester 工具"
	@echo "  make install      - 安装到系统路径"
	@echo "  make test         - 运行测试"
	@echo "  make clean        - 清理生成的文件"
	@echo "  make help         - 显示此帮助信息"

