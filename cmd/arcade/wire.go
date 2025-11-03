//go:build wireinject
// +build wireinject

package main

import (
	"time"

	"github.com/go-arcade/arcade/internal/app"
	"github.com/go-arcade/arcade/internal/engine/conf"
	"github.com/go-arcade/arcade/internal/engine/repo"
	"github.com/go-arcade/arcade/internal/engine/router"
	"github.com/go-arcade/arcade/internal/engine/service/task"
	"github.com/go-arcade/arcade/internal/pkg/grpc"
	"github.com/go-arcade/arcade/pkg/ctx"
	"github.com/go-arcade/arcade/pkg/http"
	"github.com/go-arcade/arcade/pkg/log"
	"github.com/go-arcade/arcade/pkg/plugin"
	"github.com/go-arcade/arcade/pkg/storage"
	"github.com/google/wire"
	"go.uber.org/zap"
)

func initApp(configPath string, appCtx *ctx.Context, logger *zap.Logger) (*app.App, func(), error) {
	panic(wire.Build(
		// 配置层
		confProviderSet,
		// 仓储层
		repoProviderSet,
		// 路由层
		routerProviderSet,
		// 存储层
		storageProviderSet,
		// gRPC 服务层
		grpcProviderSet,
		// 插件层
		pluginProviderSet,
		// 应用层
		app.NewApp,
	))
}

// confProviderSet 配置层 ProviderSet
var confProviderSet = wire.NewSet(
	provideConf,
	provideTaskPoolConfig,
	provideHttpConfig,
	provideGrpcConfig,
)

func provideConf(configPath string) conf.AppConfig {
	return conf.NewConf(configPath)
}

func provideTaskPoolConfig(appConf conf.AppConfig) task.TaskPoolConfig {
	return task.TaskPoolConfig{
		MaxWorkers:    appConf.Task.MaxWorkers,
		QueueSize:     appConf.Task.QueueSize,
		WorkerTimeout: appConf.Task.WorkerTimeout,
	}
}

func provideHttpConfig(appConf conf.AppConfig) *http.Http {
	return &appConf.Http
}

func provideGrpcConfig(appConf conf.AppConfig) *grpc.Conf {
	return &appConf.Grpc
}

// repoProviderSet 仓储层 ProviderSet
var repoProviderSet = wire.NewSet(
	provideAgentRepo,
	provideUserRepo,
	providePluginRepo,
	provideIdentityIntegrationRepo,
	provideStorageRepo,
)

func provideAgentRepo(appCtx *ctx.Context) *repo.AgentRepo {
	return repo.NewAgentRepo(appCtx)
}

func provideUserRepo(appCtx *ctx.Context) *repo.UserRepo {
	return repo.NewUserRepo(appCtx)
}

func providePluginRepo(appCtx *ctx.Context) *repo.PluginRepo {
	return repo.NewPluginRepo(appCtx)
}

func provideIdentityIntegrationRepo(appCtx *ctx.Context) *repo.IdentityIntegrationRepo {
	return repo.NewIdentityIntegrationRepo(appCtx)
}

func provideStorageRepo(appCtx *ctx.Context) *repo.StorageRepo {
	return repo.NewStorageRepo(appCtx)
}

// routerProviderSet 路由层 ProviderSet
var routerProviderSet = wire.NewSet(
	provideRouter,
	providePluginConfig,
)

func providePluginConfig(appConf conf.AppConfig) *conf.PluginConfig {
	return &appConf.Plugin
}

func provideRouter(httpConf *http.Http, appCtx *ctx.Context, pluginConfig *conf.PluginConfig, pluginManager *plugin.Manager) *router.Router {
	return router.NewRouter(httpConf, appCtx, pluginConfig, pluginManager)
}

// storageProviderSet 存储层 ProviderSet
var storageProviderSet = wire.NewSet(
	provideStorageFromDB,
)

func provideStorageFromDB(appCtx *ctx.Context, storageRepo *repo.StorageRepo) (storage.StorageProvider, error) {
	dbProvider, err := storage.NewStorageDBProvider(appCtx, storageRepo)
	if err != nil {
		return nil, err
	}
	return dbProvider.GetStorageProvider()
}

// grpcProviderSet gRPC 服务层 ProviderSet
var grpcProviderSet = wire.NewSet(
	provideGrpcServer,
)

func provideGrpcServer(cfg *grpc.Conf, logger *zap.Logger) *grpc.ServerWrapper {
	server := grpc.NewGrpcServer(*cfg, logger)
	server.Register()
	return server
}

// pluginProviderSet 插件层 ProviderSet
var pluginProviderSet = wire.NewSet(
	providePluginManager,
)

func providePluginManager(appConf conf.AppConfig) *plugin.Manager {
	// Get plugin directory from configuration (use default if not set)
	pluginDir := "/var/lib/arcade/plugins"
	if appConf.Plugin.CacheDir != "" {
		pluginDir = appConf.Plugin.CacheDir
	}

	// Create plugin manager configuration
	config := &plugin.ManagerConfig{
		PluginDir:       pluginDir,
		HandshakeConfig: plugin.RPCHandshake,
		PluginConfig:    make(map[string]any),
		Timeout:         30 * time.Second,
		MaxRetries:      3,
	}

	// Create plugin manager
	m := plugin.NewManager(config)

	// Auto-load plugins from directory on startup
	if err := m.LoadPluginsFromDir(); err != nil {
		log.Warnf("failed to auto-load plugins: %v", err)
	}

	log.Infof("plugin manager initialized with directory: %s", pluginDir)
	return m
}
