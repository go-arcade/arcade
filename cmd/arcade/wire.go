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
	"go.uber.org/zap"
)

func initApp(configPath string, appCtx *ctx.Context, logger *zap.Logger) (*app.App, func(), error) {
	panic(wire.Build(
		conf.ProviderSet,
		router.ProviderSet,
		plugin.ProviderSet,
		grpc.ProviderSet,
		provideHttpConfig,
		provideGrpcConfig,
		app.NewApp,
	))
}

func provideHttpConfig(appConf conf.AppConfig) *http.Http {
	return &appConf.Http
}

func provideGrpcConfig(appConf conf.AppConfig) *grpc.GrpcConf {
	return &appConf.Grpc
}
