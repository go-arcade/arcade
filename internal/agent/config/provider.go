package config

import (
	"github.com/go-arcade/arcade/pkg/http"
	"github.com/go-arcade/arcade/pkg/log"
	"github.com/google/wire"
)

// ProviderSet is a Wire provider set for configuration
var ProviderSet = wire.NewSet(
	NewConf,
	ProvideHttpConfig,
	ProvideLogConfig,
)

// ProvideHttpConfig 提供 HTTP 配置
func ProvideHttpConfig(agentConf AgentConfig) *http.Http {
	httpConfig := &agentConf.Http
	httpConfig.SetDefaults()
	return httpConfig
}

// ProvideLogConfig 提供日志配置
func ProvideLogConfig(agentConf AgentConfig) *log.Conf {
	return &agentConf.Log
}
