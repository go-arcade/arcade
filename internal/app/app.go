package app

import (
	"github.com/gofiber/fiber/v2"
	"github.com/observabil/arcade/internal/engine/router"
	service_plugin "github.com/observabil/arcade/internal/engine/service/plugin"
	"github.com/observabil/arcade/internal/pkg/grpc"
	"github.com/observabil/arcade/pkg/ctx"
	"github.com/observabil/arcade/pkg/log"
	"github.com/observabil/arcade/pkg/plugin"
	"github.com/observabil/arcade/pkg/storage"
	"go.uber.org/zap"
)

type App struct {
	HttpApp    *fiber.App
	PluginMgr  *plugin.Manager
	GrpcServer *grpc.ServerWrapper
	Logger     *zap.Logger
	Storage    storage.StorageProvider
}

func NewApp(
	rt *router.Router,
	logger *zap.Logger,
	pluginMgr *plugin.Manager,
	grpcServer *grpc.ServerWrapper,
	storage storage.StorageProvider,
	appCtx *ctx.Context,
) (*App, func(), error) {
	httpApp := rt.Router(logger)

	// 初始化插件任务管理器（MongoDB持久化）
	service_plugin.InitTaskManager(appCtx)
	logger.Info("Plugin task manager initialized with MongoDB persistence")

	cleanup := func() {
		// 停止所有插件
		if pluginMgr != nil {
			logger.Info("Shutting down plugin manager...")
			if err := pluginMgr.Close(); err != nil {
				log.Errorf("Failed to close plugin manager: %v", err)
			} else {
				logger.Info("Plugin manager stopped successfully")
			}
		}

		// 停止 gRPC 服务器
		if grpcServer != nil {
			logger.Info("Shutting down gRPC server...")
			grpcServer.Stop()
		}
	}

	app := &App{
		HttpApp:    httpApp,
		PluginMgr:  pluginMgr,
		GrpcServer: grpcServer,
		Logger:     logger,
		Storage:    storage,
	}
	return app, cleanup, nil
}
