package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/observabil/arcade/internal/engine/conf"
	"github.com/observabil/arcade/internal/engine/router"
	"github.com/observabil/arcade/pkg/cache"
	"github.com/observabil/arcade/pkg/ctx"
	"github.com/observabil/arcade/pkg/database"
	httpx "github.com/observabil/arcade/pkg/http"
	"github.com/observabil/arcade/pkg/log"
	"github.com/observabil/arcade/pkg/plugin"
	"github.com/observabil/arcade/pkg/runner"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/4 19:51
 * @file: main.go
 * @description: arcade program
 */

var (
	configFile       string
	pluginConfigFile string
	pluginDir        string
)

func init() {
	flag.StringVar(&configFile, "conf", "conf.d/config.toml", "conf file path, e.g. -conf ./conf.d")
	flag.StringVar(&pluginConfigFile, "plugin-config", "conf.d/plugins.yaml", "plugin config file path, e.g. -plugin-config ./conf.d")
	flag.StringVar(&pluginDir, "plugin", "plugins", "plugin dir path, e.g. -plugin ./plugins")
}

func main() {
	flag.Parse()
	printRunner()

	appConf := conf.NewConf(configFile)

	logger := log.NewLog(&appConf.Log)

	redis, err := cache.NewRedis(appConf.Redis)
	if err != nil {
		panic(err)
	}

	// db
	db, err := database.NewDatabase(appConf.Database, *logger)
	if err != nil {
		panic(err)
	}
	mongo, err := database.NewMongoDB(appConf.Database.MongoDB, context.Background())
	if err != nil {
		panic(err)
	}
	Ctx := ctx.NewContext(context.Background(), mongo, redis, db, logger.Sugar())

	manager := plugin.NewManager()
	manager.SetContext(context.Background())
	manager.LoadPluginsFromConfig(pluginConfigFile)
	manager.Init(context.Background())

	// 启动自动监控
	manager.StartAutoWatch([]string{pluginDir}, pluginConfigFile)
	defer manager.StopAutoWatch()

	route := router.NewRouter(&appConf.Http, Ctx)
	// http srv
	app := route.Router(logger)
	cleanup := httpx.NewHttp(appConf.Http, app)
	cleanup()
}

func printRunner() {
	fmt.Println("runner.pwd:", runner.Pwd)
	fmt.Println("runner.hostname:", runner.Hostname)
}
