package main

import (
	"flag"

	"github.com/go-arcade/arcade/internal/bootstrap"
)

var (
	configFile string
	pluginDir  string
)

func init() {
	flag.StringVar(&configFile, "conf", "conf.d/config.toml", "conf file path, e.g. -conf ./conf.d")
	flag.StringVar(&pluginDir, "plugin", "plugins", "plugin dir path, e.g. -plugin ./plugins")
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
