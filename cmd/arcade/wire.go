//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/observabil/arcade/internal/app"
	"github.com/observabil/arcade/internal/engine/conf"
	"github.com/observabil/arcade/internal/engine/repo"
	"github.com/observabil/arcade/internal/engine/router"
	"github.com/observabil/arcade/internal/pkg/grpc"
	"github.com/observabil/arcade/pkg/ctx"
	"github.com/observabil/arcade/pkg/http"
	"github.com/observabil/arcade/pkg/plugin"
	"github.com/observabil/arcade/pkg/storage"
	"go.uber.org/zap"
)

func initApp(configPath string, appCtx *ctx.Context, logger *zap.Logger) (*app.App, func(), error) {
	panic(wire.Build(
		conf.ProviderSet,
		repo.ProviderSet,
		router.ProviderSet,
		plugin.ProviderSet,
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
