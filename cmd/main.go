package main

import (
	"gouse/config"
	"gouse/internal/router"
)

func Init() {
	// 调用了 config 包中的 InitConfig 函数，用于初始化日志信息。
	config.InitConfig()
}

func main() {
	Init()                      // 初始化日志信息
	router.InitRouterAndServe() // 初始化路由并启动 HTTP服务器。
}
