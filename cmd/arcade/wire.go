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

//go:build wireinject
// +build wireinject

package main

import (
	"context"
	"fmt"

	"github.com/go-arcade/arcade/internal/engine/bootstrap"
	"github.com/go-arcade/arcade/internal/engine/config"
	"github.com/go-arcade/arcade/internal/engine/repo"
	"github.com/go-arcade/arcade/internal/engine/router"
	"github.com/go-arcade/arcade/internal/engine/service"
	"github.com/go-arcade/arcade/internal/pkg/grpc"
	"github.com/go-arcade/arcade/internal/pkg/queue"
	"github.com/go-arcade/arcade/internal/pkg/storage"
	"github.com/go-arcade/arcade/pkg/cache"
	"github.com/go-arcade/arcade/pkg/database"
	"github.com/go-arcade/arcade/pkg/log"
	"github.com/go-arcade/arcade/pkg/metrics"
	"github.com/go-arcade/arcade/pkg/plugin"
	"github.com/go-arcade/arcade/pkg/pprof"
	"github.com/google/wire"
)

// ProvideContext 提供基础 context.Context
func ProvideContext() context.Context {
	return context.Background()
}

// ProvidePluginConfig 提供插件配置，直接从 conf.d/plugins.toml 读取
func ProvidePluginConfig() map[string]any {
	configs, err := plugin.LoadPluginConfig("conf.d/plugins.toml")
	if err != nil {
		panic(fmt.Sprintf("failed to load plugin config: %v", err))
	}
	return configs
}

func initApp(configPath string) (*bootstrap.App, func(), error) {
	panic(wire.Build(
		// 提供插件配置
		ProvidePluginConfig,
		// 配置层
		config.ProviderSet,
		// 上下文层
		ProvideContext,
		// 日志层（依赖 config）
		log.ProviderSet,
		// 数据库层（依赖 config, log, ctx）
		database.ProviderSet,
		// 缓存层（依赖 config）
		cache.ProviderSet,
		// 任务队列层（依赖 config, cache）
		queue.ProviderSet,
		// 指标层（依赖 config, queue）
		metrics.ProviderSet,
		// pprof层（依赖 config, log）
		pprof.ProviderSet,
		// 仓储层（依赖 database）
		repo.ProviderSet,
		// 存储层（依赖 repo）
		storage.ProviderSet,
		// 插件层（依赖 config, database）
		plugin.ProviderSet,
		// 服务层（依赖 repo, storage, plugin, database, cache）
		service.ProviderSet,
		// 路由层（依赖 config, repo, service, storage, plugin）
		router.ProviderSet,
		// gRPC 服务层
		grpc.ProviderSet,
		// 应用层
		bootstrap.NewApp,
	))
}
