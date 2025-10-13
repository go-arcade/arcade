package main

import (
	"context"
	"flag"

	"github.com/observabil/arcade/internal/engine/conf"
	"github.com/observabil/arcade/pkg/cache"
	"github.com/observabil/arcade/pkg/ctx"
	"github.com/observabil/arcade/pkg/database"
	"github.com/observabil/arcade/pkg/http"
	"github.com/observabil/arcade/pkg/log"
)

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

	// 加载配置
	appConf := conf.NewConf(configFile)

	// 初始化日志
	logger := log.NewLog(&appConf.Log)

	// 初始化 Redis、数据库
	redis, err := cache.NewRedis(appConf.Redis)
	if err != nil {
		panic(err)
	}
	db, err := database.NewDatabase(appConf.Database, *logger)
	if err != nil {
		panic(err)
	}
	mongo, err := database.NewMongoDB(appConf.Database.MongoDB, context.Background())
	if err != nil {
		panic(err)
	}
	appCtx := ctx.NewContext(context.Background(), mongo, redis, db, logger.Sugar())

	// Wire 构建 App
	app, cleanup, err := initApp(configFile, appCtx, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// 启动插件系统
	app.PluginMgr.SetContext(context.Background())
	app.PluginMgr.LoadPluginsFromConfig(pluginConfigFile)
	app.PluginMgr.Init(context.Background())
	app.PluginMgr.StartAutoWatch([]string{pluginDir}, pluginConfigFile)
	defer app.PluginMgr.StopAutoWatch()

	// 启动 gRPC 服务
	if app.GrpcServer != nil && appConf.Grpc.Port > 0 {
		go func() {
			if err := app.GrpcServer.Start(appConf.Grpc); err != nil {
				logger.Sugar().Errorf("gRPC server failed: %v", err)
			}
		}()
	}

	// 启动 HTTP 服务
	httpCleanup := http.NewHttp(appConf.Http, app.HttpApp)
	httpCleanup()
}
