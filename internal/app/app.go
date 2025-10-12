package app

import (
	"github.com/gofiber/fiber/v2"
	"github.com/observabil/arcade/internal/engine/router"
	"github.com/observabil/arcade/internal/pkg/grpc"
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
) (*App, func(), error) {
	httpApp := rt.Router(logger)

	cleanup := func() {
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
