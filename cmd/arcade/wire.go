//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/observabil/arcade/internal/app"
	"github.com/observabil/arcade/internal/engine/conf"
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
		router.ProviderSet,
		plugin.ProviderSet,
		storage.ProviderSet,
		grpc.ProviderSet,
		provideHttpConfig,
		provideGrpcConfig,
		provideStorageConfig,
		app.NewApp,
	))
}

func provideStorageConfig(appConf conf.AppConfig, appCtx *ctx.Context) *storage.Storage {
	return &storage.Storage{
		Ctx:       appCtx,
		Provider:  appConf.Storage.Provider,
		AccessKey: appConf.Storage.AccessKey,
		SecretKey: appConf.Storage.SecretKey,
		Endpoint:  appConf.Storage.Endpoint,
		Bucket:    appConf.Storage.Bucket,
		Region:    appConf.Storage.Region,
		UseTLS:    appConf.Storage.UseTLS,
		BasePath:  appConf.Storage.BasePath,
	}
}


func provideHttpConfig(appConf conf.AppConfig) *http.Http {
	return &appConf.Http
}

func provideGrpcConfig(appConf conf.AppConfig) *grpc.GrpcConf {
	return &appConf.Grpc
}
