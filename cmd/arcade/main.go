package main

import (
	"flag"

	"github.com/go-arcade/arcade/internal/engine/bootstrap"
)

var (
	configFile string
)

func init() {
	flag.StringVar(&configFile, "conf", "conf.d/config.toml", "config file path, e.g. -conf ./conf.d")
}

func main() {
	flag.Parse()

	// Bootstrap 初始化应用
	app, cleanup, _, err := bootstrap.Bootstrap(configFile, initApp)
	if err != nil {
		panic(err)
	}

	// 启动应用并等待退出信号
	bootstrap.Run(app, cleanup)
}
