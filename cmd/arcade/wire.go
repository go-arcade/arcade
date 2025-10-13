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
		conf.ProviderSet,
		repo.ProviderSet,
		router.ProviderSet,
		pluginProviderSet,
		storage.ProviderSet,
		grpc.ProviderSet,
		provideHttpConfig,
		provideGrpcConfig,
		providePluginRepository,
		app.NewApp,
	))
}

func provideHttpConfig(appConf conf.AppConfig) *http.Http {
	return &appConf.Http
}

func provideGrpcConfig(appConf conf.AppConfig, logger *zap.Logger) *grpc.GrpcConf {
	return &appConf.Grpc
}

func providePluginRepository(adapter *repo.PluginRepoAdapter) plugin.PluginRepository {
	return adapter
}

// ProviderSet 提供插件相关的依赖
var pluginProviderSet = wire.NewSet(
	plugin.ProvidePluginManager, // 只声明构造函数
)
// ProvidePluginManager 提供插件管理器实例
// 如果提供了 PluginRepository，则从数据库加载插件
func ProvidePluginManager(repo plugin.PluginRepository) *plugin.Manager {
	m := plugin.NewManager()

	// 设置数据库仓库
	m.SetPluginRepository(repo)

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
