package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-arcade/arcade/internal/engine/conf"
	"github.com/go-arcade/arcade/pkg/cache"
	"github.com/go-arcade/arcade/pkg/ctx"
	"github.com/go-arcade/arcade/pkg/database"
	"github.com/go-arcade/arcade/pkg/log"
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

	// 加载配置
	appConf := conf.NewConf(configFile)

	// 初始化日志
	logger, err := log.NewLog(&appConf.Log)
	if err != nil {
		panic(err)
	}

	// 初始化 Redis、数据库、context
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

	// 插件管理器已在 wire 中初始化完成
	// 可选：启动心跳检查
	app.PluginMgr.StartHeartbeat(30 * time.Second)

	// 启动 gRPC 服务
	if app.GrpcServer != nil && appConf.Grpc.Port > 0 {
		go func() {
			if err := app.GrpcServer.Start(appConf.Grpc); err != nil {
				logger.Sugar().Errorf("gRPC server failed: %v", err)
			}
		}()
	}

	// 设置信号监听（优雅关闭）
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// 启动 HTTP 服务（异步）
	go func() {
		addr := appConf.Http.Host + ":" + fmt.Sprintf("%d", appConf.Http.Port)
		logger.Sugar().Infof("HTTP server starting at %s", addr)
		if err := app.HttpApp.Listen(addr); err != nil {
			logger.Sugar().Errorf("HTTP server failed: %v", err)
		}
	}()

	// 等待退出信号
	sig := <-quit
	logger.Sugar().Infof("Received signal: %v, shutting down gracefully...", sig)

	// 按顺序关闭各个组件
	// 1. 关闭 HTTP 服务器
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	if err := app.HttpApp.ShutdownWithContext(shutdownCtx); err != nil {
		logger.Sugar().Errorf("HTTP server shutdown error: %v", err)
	} else {
		logger.Info("HTTP server shut down gracefully")
	}

	// 2. 关闭插件管理器和其他资源
	cleanup()

	logger.Info("Server shutdown complete")
}
