//go:build wireinject
// +build wireinject

package main

import (
	"github.com/go-arcade/arcade/internal/engine/bootstrap"
	"github.com/go-arcade/arcade/internal/engine/config"
	"github.com/go-arcade/arcade/internal/engine/repo"
	"github.com/go-arcade/arcade/internal/engine/router"
	"github.com/go-arcade/arcade/internal/engine/service"
	"github.com/go-arcade/arcade/internal/pkg/grpc"
	"github.com/go-arcade/arcade/internal/pkg/queue"
	"github.com/go-arcade/arcade/internal/pkg/storage"
	"github.com/go-arcade/arcade/pkg/cache"
	"github.com/go-arcade/arcade/pkg/ctx"
	"github.com/go-arcade/arcade/pkg/database"
	"github.com/go-arcade/arcade/pkg/log"
	"github.com/go-arcade/arcade/pkg/metrics"
	"github.com/go-arcade/arcade/pkg/plugin"
	"github.com/go-arcade/arcade/pkg/pprof"
	"github.com/google/wire"
)

func initApp(configPath string) (*bootstrap.App, func(), error) {
	panic(wire.Build(
		// 配置层
		config.ProviderSet,
		// 日志层（依赖 config）
		log.ProviderSet,
		// 上下文层（依赖 log）
		ctx.ProviderSet,
		// 数据库层（依赖 config, log, ctx）
		database.ProviderSet,
		// 缓存层（依赖 config）
		cache.ProviderSet,
		// 任务队列层（依赖 config, cache）
		queue.ProviderSet,
		// 指标层（依赖 config, log）
		metrics.ProviderSet,
		// pprof层（依赖 config, log）
		pprof.ProviderSet,
		// 仓储层（依赖 database）
		repo.ProviderSet,
		// 存储层（依赖 repo）
		storage.ProviderSet,
		// 插件层（依赖 config, database）
		plugin.ProviderSet,
		// 服务层（依赖 repo, storage, plugin, database, cache, ctx）
		service.ProviderSet,
		// 路由层（依赖 config, repo, service, storage, plugin）
		router.ProviderSet,
		// gRPC 服务层
		grpc.ProviderSet,
		// 应用层
		bootstrap.NewApp,
	))
}
