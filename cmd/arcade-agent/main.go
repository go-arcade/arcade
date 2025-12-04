package main

import (
	"flag"

	"github.com/go-arcade/arcade/internal/agent/bootstrap"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "conf.d/agent.toml", "config file path, e.g. -config ./conf.d")
}

func main() {
	flag.Parse()

	// Bootstrap 初始化应用
	app, cleanup, _, err := bootstrap.Bootstrap(configFile, initAgent)
	if err != nil {
		panic(err)
	}

	// 启动应用并等待退出信号
	bootstrap.Run(app, cleanup)
}
