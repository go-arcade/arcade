//go:build wireinject
// +build wireinject

package main

import (
	"github.com/go-arcade/arcade/internal/agent/bootstrap"
	"github.com/go-arcade/arcade/internal/agent/config"
	"github.com/go-arcade/arcade/internal/agent/router"
	"github.com/go-arcade/arcade/pkg/log"
	"github.com/google/wire"
)

func initAgent(configPath string) (*bootstrap.Agent, func(), error) {
	panic(wire.Build(
		// 配置层
		config.ProviderSet,
		// 日志层（依赖 config）
		log.ProviderSet,
		router.ProviderSet,
		// 应用层
		bootstrap.NewAgent,
	))
}
