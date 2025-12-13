//go:build wireinject
// +build wireinject

package main

import (
	"github.com/go-arcade/arcade/internal/agent/bootstrap"
	"github.com/go-arcade/arcade/internal/agent/config"
	"github.com/go-arcade/arcade/internal/agent/router"
	"github.com/go-arcade/arcade/internal/pkg/grpc"
	"github.com/go-arcade/arcade/internal/pkg/queue"
	"github.com/go-arcade/arcade/pkg/cache"
	"github.com/go-arcade/arcade/pkg/log"
	"github.com/go-arcade/arcade/pkg/metrics"
	"github.com/go-arcade/arcade/pkg/pprof"
	"github.com/google/wire"
)

func initAgent(configPath string) (*bootstrap.Agent, func(), error) {
	panic(wire.Build(
		// 配置层
		config.ProviderSet,
		// 日志层（依赖 config）
		log.ProviderSet,
		// 缓存层（依赖 config）
		cache.ProviderSet,
		// 任务队列层（依赖 config, cache）
		queue.AgentProviderSet,
		// 指标层（依赖 config）
		metrics.ProviderSet,
		// pprof层（依赖 config, log）
		pprof.ProviderSet,
		// gRPC 客户端层（依赖 config 和 log）
		grpc.ProviderSet,
		router.ProviderSet,
		// 应用层
		bootstrap.NewAgent,
	))
}
