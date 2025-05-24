package main

import (
	"math/rand"
	"time"

	"github.com/gaoyong06/api-tester/cmd/api-tester/cmd"
)

func main() {
	// 设置随机数种子
	rand.Seed(time.Now().UnixNano())

	// 执行根命令
	cmd.Execute()
}
