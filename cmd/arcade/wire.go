//go:build wireinject
// +build wireinject

package main

import (
	"context"

	"github.com/google/wire"
	"github.com/observabil/arcade/internal/app"
	"github.com/observabil/arcade/internal/engine/conf"
	"github.com/observabil/arcade/internal/engine/repo"
	"github.com/observabil/arcade/internal/engine/router"
	"github.com/observabil/arcade/internal/engine/service/job"
	"github.com/observabil/arcade/internal/pkg/grpc"
	"github.com/observabil/arcade/pkg/ctx"
	"github.com/observabil/arcade/pkg/http"
	"github.com/observabil/arcade/pkg/log"
	"github.com/observabil/arcade/pkg/plugin"
	"github.com/observabil/arcade/pkg/storage"
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
	provideJobPoolConfig,
	provideHttpConfig,
	provideGrpcConfig,
)

func provideConf(configPath string) conf.AppConfig {
	return conf.NewConf(configPath)
}

func provideJobPoolConfig(appConf conf.AppConfig) job.JobPoolConfig {
	return job.JobPoolConfig{
		MaxWorkers:    appConf.Job.MaxWorkers,
		QueueSize:     appConf.Job.QueueSize,
		WorkerTimeout: appConf.Job.WorkerTimeout,
	}
}

func provideHttpConfig(appConf conf.AppConfig) *http.Http {
	return &appConf.Http
}

func provideGrpcConfig(appConf conf.AppConfig) *grpc.GrpcConf {
	return &appConf.Grpc
}

// repoProviderSet 仓储层 ProviderSet
var repoProviderSet = wire.NewSet(
	provideAgentRepo,
	provideUserRepo,
	providePluginRepo,
	providePluginRepoAdapter,
	provideSSORepo,
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

func providePluginRepoAdapter(pluginRepo *repo.PluginRepo) *repo.PluginRepoAdapter {
	return repo.NewPluginRepoAdapter(pluginRepo)
}

func provideSSORepo(appCtx *ctx.Context) *repo.SSORepo {
	return repo.NewSSORepo(appCtx)
}

func provideStorageRepo(appCtx *ctx.Context) *repo.StorageRepo {
	return repo.NewStorageRepo(appCtx)
}

// routerProviderSet 路由层 ProviderSet
var routerProviderSet = wire.NewSet(
	provideRouter,
)

func provideRouter(httpConf *http.Http, appCtx *ctx.Context) *router.Router {
	return router.NewRouter(httpConf, appCtx)
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

func provideGrpcServer(cfg *grpc.GrpcConf, logger *zap.Logger) *grpc.ServerWrapper {
	server := grpc.NewGrpcServer(*cfg, logger)
	server.Register()
	return server
}

// pluginProviderSet 插件层 ProviderSet
var pluginProviderSet = wire.NewSet(
	providePluginManager,
	providePluginRepository,
)

func providePluginRepository(adapter *repo.PluginRepoAdapter) plugin.PluginRepository {
	return adapter
}

func providePluginManager(pluginRepo plugin.PluginRepository) *plugin.Manager {
	m := plugin.NewManager()
	m.SetPluginRepository(pluginRepo)

	// 从数据库加载插件
	if err := m.LoadPluginsFromDatabase(); err != nil {
		log.Warnf("failed to load plugins from database: %v, will try file system", err)
		// 如果数据库加载失败，尝试从文件系统加载（向后兼容）
		if err := m.LoadPluginsFromDir("./plugins"); err != nil {
			log.Warnf("failed to load plugins from directory: %v", err)
		}
	}

	// 初始化所有插件
	if err := m.Init(context.Background()); err != nil {
		log.Errorf("failed to initialize plugins: %v", err)
	}

	return m
}
