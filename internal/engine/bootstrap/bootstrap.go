// Copyright 2025 Arcade Team
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	"github.com/go-arcade/arcade/internal/pkg/grpc"
	"github.com/go-arcade/arcade/internal/pkg/queue"
	"github.com/go-arcade/arcade/internal/pkg/storage"
	"github.com/go-arcade/arcade/pkg/database"
	"github.com/go-arcade/arcade/pkg/log"
	"github.com/go-arcade/arcade/pkg/metrics"
	"github.com/go-arcade/arcade/pkg/plugin"
	"github.com/go-arcade/arcade/pkg/pprof"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type App struct {
	HttpApp       *fiber.App
	PluginMgr     *plugin.Manager
	GrpcServer    *grpc.ServerWrapper
	QueueServer   *queue.Server // 队列服务器（任务发布者）
	MetricsServer *metrics.Server
	PprofServer   *pprof.Server
	Logger        *log.Logger
	Storage       storage.StorageProvider
	AppConf       *config.AppConfig
}

// InitAppFunc init app function type
type InitAppFunc func(configPath string) (*App, func(), error)

func NewApp(
	rt *router.Router,
	logger *log.Logger,
	pluginMgr *plugin.Manager,
	grpcServer *grpc.ServerWrapper,
	queueServer *queue.Server,
	metricsServer *metrics.Server,
	pprofServer *pprof.Server,
	storage storage.StorageProvider,
	mongoDB database.MongoDB,
	appConf *config.AppConfig,
	db database.IDatabase,
) (*App, func(), error) {
	httpApp := rt.Router()

	// 主程序作为 queue server，只发布任务，不执行任务
	// 不需要注册任务处理器

	// 设置 AppConf
	app := &App{
		HttpApp:       httpApp,
		PluginMgr:     pluginMgr,
		GrpcServer:    grpcServer,
		QueueServer:   queueServer,
		MetricsServer: metricsServer,
		PprofServer:   pprofServer,
		Logger:        logger,
		Storage:       storage,
		AppConf:       appConf,
	}

	cleanup := func() {
		// stop queue server
		if queueServer != nil {
			queueServer.Shutdown()
		}

		// stop pprof server
		if pprofServer != nil {
			log.Info("Shutting down pprof server...")
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := pprofServer.Stop(shutdownCtx); err != nil {
				log.Errorw("Failed to stop pprof server", zap.Error(err))
			}
		}

		// stop metrics server
		if metricsServer != nil {
			log.Info("Shutting down metrics server...")
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := metricsServer.Stop(shutdownCtx); err != nil {
				log.Errorw("Failed to stop metrics server", zap.Error(err))
			}
		}

		// stop all plugins
		if pluginMgr != nil {
			log.Info("Shutting down plugin manager...")
			if err := pluginMgr.Close(); err != nil {
				log.Errorw("Failed to close plugin manager", zap.Error(err))
			}
		}

		// stop gRPC server
		if grpcServer != nil {
			log.Info("Shutting down gRPC server...")
			grpcServer.Stop()
		}
	}

	return app, cleanup, nil
}

// Bootstrap init app, return App instance and cleanup function
func Bootstrap(configFile string, initApp InitAppFunc) (*App, func(), *config.AppConfig, error) {
	// Wire build App (所有依赖都由 wire 自动注入)
	app, cleanup, err := initApp(configFile)
	if err != nil {
		return nil, nil, nil, err
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

	// Register Task Queue metrics if queue server is available
	if app.MetricsServer != nil && app.QueueServer != nil {
		metrics.RegisterAsynqMetricsFromQueueServer(app.MetricsServer.GetRegistry(), app.QueueServer)
	}

	// start metrics server
	if app.MetricsServer != nil {
		if err := app.MetricsServer.Start(); err != nil {
			logger.Errorw("Metrics server failed: %v", err)
		}
	}

	// start pprof server
	if app.PprofServer != nil {
		if err := app.PprofServer.Start(); err != nil {
			logger.Errorw("Pprof server failed: %v", err)
		}
	}

	// start gRPC server
	if app.GrpcServer != nil && appConf.Grpc.Port > 0 {
		go func() {
			if err := app.GrpcServer.Start(appConf.Grpc); err != nil {
				logger.Errorw("gRPC server failed: %v", err)
			}
		}()
	}

	// set signal listener (graceful shutdown)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// start HTTP server (async)
	go func() {
		addr := appConf.Http.Host + ":" + fmt.Sprintf("%d", appConf.Http.Port)
		log.Infow("HTTP listener started",
			"address", addr,
		)
		if err := app.HttpApp.Listen(addr); err != nil {
			log.Errorw("HTTP listener failed",
				"address", addr,
				zap.Error(err),
			)
		}
	}()

	// wait for exit signal
	sig := <-quit
	log.Infow("Received signal, shutting down gracefully...", "signal", sig)

	// close components in order
	// close HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	if err := app.HttpApp.ShutdownWithContext(shutdownCtx); err != nil {
		log.Errorw("HTTP server shutdown error: %v", err)
	} else {
		log.Info("HTTP server shut down gracefully")
	}

	// close plugin manager and other resources
	cleanup()

	log.Info("Server shutdown complete")
}
