package bootstrap

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-arcade/arcade/internal/engine/conf"
	"github.com/go-arcade/arcade/internal/engine/router"
	serviceplugin "github.com/go-arcade/arcade/internal/engine/service"
	"github.com/go-arcade/arcade/internal/pkg/grpc"
	"github.com/go-arcade/arcade/internal/pkg/storage"
	"github.com/go-arcade/arcade/pkg/cache"
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
	Logger     *zap.Logger
	Storage    storage.StorageProvider
	AppConf    conf.AppConfig
}

// InitAppFunc init app function type
type InitAppFunc func(configPath string, appCtx *ctx.Context, logger *zap.Logger, db database.IDatabase, mongo database.MongoDB, cache cache.ICache) (*App, func(), error)

func NewApp(
	rt *router.Router,
	logger *zap.Logger,
	pluginMgr *plugin.Manager,
	grpcServer *grpc.ServerWrapper,
	storage storage.StorageProvider,
	appCtx *ctx.Context,
	mongoDB database.MongoDB,
	appConf conf.AppConfig,
) (*App, func(), error) {
	httpApp := rt.Router(logger)

	// init plugin task manager
	serviceplugin.InitTaskManager(mongoDB)
	logger.Info("Plugin task manager initialized with MongoDB persistence")

	cleanup := func() {
		// stop all plugins
		if pluginMgr != nil {
			logger.Info("Shutting down plugin manager...")
			if err := pluginMgr.Close(); err != nil {
				logger.Error("Failed to close plugin manager", zap.Error(err))
			} else {
				logger.Info("Plugin manager stopped successfully")
			}
		}

		// stop gRPC server
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
		AppConf:    appConf,
	}
	return app, cleanup, nil
}

// Bootstrap init app, return App instance and cleanup function
func Bootstrap(configFile string, initApp InitAppFunc) (*App, func(), conf.AppConfig, error) {
	// load config
	appConf := conf.NewConf(configFile)

	// init logger
	logger, err := log.NewLog(&appConf.Log)
	if err != nil {
		return nil, nil, appConf, err
	}

	// init Redis, database, context
	redisClient, err := cache.NewRedis(appConf.Redis)
	if err != nil {
		return nil, nil, appConf, err
	}
	dbClient, err := database.NewDatabase(appConf.Database, *logger)
	if err != nil {
		return nil, nil, appConf, err
	}
	mongoClient, err := database.NewMongoDB(appConf.Database.MongoDB, context.Background())
	if err != nil {
		return nil, nil, appConf, err
	}

	// create interface implementation
	db := database.NewGormDB(dbClient)
	mongo := database.NewMongoDBWrapper(mongoClient)
	redisCache := cache.NewRedisCache(redisClient)

	appCtx := ctx.NewContext(context.Background(), logger.Sugar())

	// Wire build App
	app, cleanup, err := initApp(configFile, appCtx, logger, db, mongo, redisCache)
	if err != nil {
		return nil, nil, appConf, err
	}

	return app, cleanup, appConf, nil
}

// Run start app and wait for exit signal, then gracefully shutdown
func Run(app *App, cleanup func()) {
	logger := app.Logger
	appConf := app.AppConf

	// plugin manager is initialized in wire
	// optional: start heartbeat check
	app.PluginMgr.StartHeartbeat(30 * time.Second)

	// start gRPC server
	if app.GrpcServer != nil && appConf.Grpc.Port > 0 {
		go func() {
			if err := app.GrpcServer.Start(appConf.Grpc); err != nil {
				logger.Sugar().Errorf("gRPC server failed: %v", err)
			}
		}()
	}

	// set signal listener (graceful shutdown)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// start HTTP server (async)
	go func() {
		addr := appConf.Http.Host + ":" + fmt.Sprintf("%d", appConf.Http.Port)
		logger.Sugar().Infow("HTTP listener started",
			"address", addr,
		)
		if err := app.HttpApp.Listen(addr); err != nil {
			logger.Sugar().Errorw("HTTP listener failed",
				"address", addr,
				"error", err,
			)
		}
	}()

	// wait for exit signal
	sig := <-quit
	logger.Sugar().Infof("Received signal: %v, shutting down gracefully...", sig)

	// close components in order
	// close HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	if err := app.HttpApp.ShutdownWithContext(shutdownCtx); err != nil {
		logger.Sugar().Errorf("HTTP server shutdown error: %v", err)
	} else {
		logger.Info("HTTP server shut down gracefully")
	}

	// close plugin manager and other resources
	cleanup()

	logger.Info("Server shutdown complete")
}
