package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gaoyong06/api-tester/internal/config"
	"github.com/gaoyong06/api-tester/internal/runner"
)

func main() {
	// 命令行参数
	specFile := flag.String("spec", "", "OpenAPI/Swagger 规范文件路径 (必填)")
	baseURL := flag.String("url", "", "API 基础 URL (必填)")
	headers := flag.String("headers", "", "请求头 (JSON 格式，可选)")
	outputDir := flag.String("output", "./reports", "测试报告输出目录 (可选)")
	verbose := flag.Bool("verbose", false, "显示详细日志 (可选)")
	timeout := flag.Int("timeout", 30, "请求超时时间 (秒) (可选)")
	pathParams := flag.String("path-params", "", "路径参数文件 (JSON 格式，可选)")
	requestBodies := flag.String("request-bodies", "", "请求体模板文件 (JSON 格式，可选)")

	// 解析命令行参数
	flag.Parse()

	// 验证必填参数
	if *specFile == "" || *baseURL == "" {
		flag.Usage()
		os.Exit(1)
	}

	// 创建配置
	cfg, err := config.NewConfig(*specFile, *baseURL, *headers, *outputDir, *verbose, *timeout, *pathParams, *requestBodies)
	if err != nil {
		log.Fatalf("配置错误: %v", err)
	}

	// 创建并运行测试
	r := runner.NewRunner(cfg)
	results, err := r.Run()
	if err != nil {
		log.Fatalf("测试运行失败: %v", err)
	}

	// 输出测试结果摘要
	fmt.Printf("\n测试完成! 总计: %d, 通过: %d, 失败: %d\n", 
		results.Total, results.Passed, results.Failed)
	fmt.Printf("详细报告已保存到: %s\n", results.ReportPath)

	// 如果有测试失败，返回非零退出码
	if results.Failed > 0 {
		os.Exit(1)
	}
}
