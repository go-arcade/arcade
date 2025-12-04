package bootstrap

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-arcade/arcade/internal/engine/config"
	"github.com/go-arcade/arcade/internal/engine/router"
	"github.com/go-arcade/arcade/internal/engine/service"
	"github.com/go-arcade/arcade/internal/pkg/grpc"
	"github.com/go-arcade/arcade/internal/pkg/storage"
	"github.com/go-arcade/arcade/pkg/ctx"
	"github.com/go-arcade/arcade/pkg/database"
	"github.com/go-arcade/arcade/pkg/log"
	"github.com/go-arcade/arcade/pkg/plugin"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type App struct {
	HttpApp    *fiber.App
	PluginMgr  *plugin.Manager
	GrpcServer *grpc.ServerWrapper
	Logger     *log.Logger
	Storage    storage.StorageProvider
	AppConf    config.AppConfig
}

// InitAppFunc init app function type
type InitAppFunc func(configPath string) (*App, func(), error)

func NewApp(
	rt *router.Router,
	logger *log.Logger,
	pluginMgr *plugin.Manager,
	grpcServer *grpc.ServerWrapper,
	storage storage.StorageProvider,
	appCtx *ctx.Context,
	mongoDB database.MongoDB,
	appConf config.AppConfig,
	db database.IDatabase,
) (*App, func(), error) {
	zapLogger := logger.Log.Desugar()
	httpApp := rt.Router(zapLogger)

	// init plugin task manager
	service.InitTaskManager(mongoDB)
	zapLogger.Info("Plugin task manager initialized with MongoDB persistence")

	// 设置 AppConf
	app := &App{
		HttpApp:    httpApp,
		PluginMgr:  pluginMgr,
		GrpcServer: grpcServer,
		Logger:     logger,
		Storage:    storage,
		AppConf:    appConf,
	}

	cleanup := func() {
		// stop all plugins
		if pluginMgr != nil {
			zapLogger.Info("Shutting down plugin manager...")
			if err := pluginMgr.Close(); err != nil {
				zapLogger.Error("Failed to close plugin manager", zap.Error(err))
			} else {
				zapLogger.Info("Plugin manager stopped successfully")
			}
		}

		// stop gRPC server
		if grpcServer != nil {
			zapLogger.Info("Shutting down gRPC server...")
			grpcServer.Stop()
		}
	}

	return app, cleanup, nil
}

// Bootstrap init app, return App instance and cleanup function
func Bootstrap(configFile string, initApp InitAppFunc) (*App, func(), config.AppConfig, error) {
	// Wire build App (所有依赖都由 wire 自动注入)
	app, cleanup, err := initApp(configFile)
	if err != nil {
		return nil, nil, config.AppConfig{}, err
	}

	// 获取配置（从 app 中获取）
	appConf := app.AppConf

	return app, cleanup, appConf, nil
}

// Run start app and wait for exit signal, then gracefully shutdown
func Run(app *App, cleanup func()) {
	logger := app.Logger.Log
	appConf := app.AppConf

	// plugin manager is initialized in wire
	// optional: start heartbeat check
	app.PluginMgr.StartHeartbeat(30 * time.Second)

	// start gRPC server
	if app.GrpcServer != nil && appConf.Grpc.Port > 0 {
		go func() {
			if err := app.GrpcServer.Start(appConf.Grpc); err != nil {
				logger.Errorf("gRPC server failed: %v", err)
			}
		}()
	}

	// set signal listener (graceful shutdown)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// start HTTP server (async)
	go func() {
		glog := app.Logger.Log.Desugar().WithOptions(zap.AddCallerSkip(-1)).Sugar()
		addr := appConf.Http.Host + ":" + fmt.Sprintf("%d", appConf.Http.Port)
		glog.Infow("HTTP listener started",
			"address", addr,
		)
		if err := app.HttpApp.Listen(addr); err != nil {
			glog.Errorw("HTTP listener failed",
				"address", addr,
				zap.Error(err),
			)
		}
	}()

	// wait for exit signal
	sig := <-quit
	logger.Infof("Received signal: %v, shutting down gracefully...", sig)

	// close components in order
	// close HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	if err := app.HttpApp.ShutdownWithContext(shutdownCtx); err != nil {
		logger.Errorf("HTTP server shutdown error: %v", err)
	} else {
		logger.Info("HTTP server shut down gracefully")
	}

	// close plugin manager and other resources
	cleanup()

	logger.Info("Server shutdown complete")
}
