//go:build wireinject
// +build wireinject

package main

import (
	"github.com/go-arcade/arcade/internal/app"
	"github.com/go-arcade/arcade/internal/engine/conf"
	"github.com/go-arcade/arcade/internal/engine/repo"
	"github.com/go-arcade/arcade/internal/engine/router"
	"github.com/go-arcade/arcade/internal/engine/service"
	"github.com/go-arcade/arcade/internal/pkg/grpc"
	"github.com/go-arcade/arcade/internal/pkg/storage"
	"github.com/go-arcade/arcade/pkg/cache"
	"github.com/go-arcade/arcade/pkg/ctx"
	"github.com/go-arcade/arcade/pkg/database"
	"github.com/go-arcade/arcade/pkg/plugin"
	"github.com/google/wire"
	"go.uber.org/zap"
)

func initApp(configPath string, appCtx *ctx.Context, logger *zap.Logger, db database.DB, mongo database.MongoDB, cache cache.Cache) (*app.App, func(), error) {
	panic(wire.Build(
		// 配置层
		conf.ProviderSet,
		// 仓储层
		repo.ProviderSet,
		// 存储层（依赖 repo）
		storage.ProviderSet,
		// 插件层（依赖 conf）
		plugin.ProviderSet,
		// 服务层（依赖 repo, storage, plugin）
		service.ProviderSet,
		// 路由层（依赖 conf, repo, service, storage, plugin）
		router.ProviderSet,
		// gRPC 服务层
		grpc.ProviderSet,
		// 应用层
		app.NewApp,
	))
}
