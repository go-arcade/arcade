package config

import (
	grpcclient "github.com/go-arcade/arcade/internal/pkg/grpc"
	"github.com/go-arcade/arcade/pkg/http"
	"github.com/go-arcade/arcade/pkg/log"
	"github.com/go-arcade/arcade/pkg/metrics"
	"github.com/go-arcade/arcade/pkg/pprof"
	"github.com/google/wire"
)

// ProviderSet is a Wire provider set for configuration
var ProviderSet = wire.NewSet(
	NewConf,
	ProvideHttpConfig,
	ProvideLogConfig,
	ProvideGrpcClientConfig,
	ProvideMetricsConfig,
	ProvidePprofConfig,
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

// ProvideGrpcClientConfig 提供 gRPC 客户端配置
func ProvideGrpcClientConfig(agentConf AgentConfig) *grpcclient.ClientConf {
	// build server address from host and port
	return &grpcclient.ClientConf{
		ServerAddr:           agentConf.Grpc.ServerAddr,
		Token:                agentConf.Grpc.Token,
		ReadWriteTimeout:     agentConf.Grpc.ReadWriteTimeout,
		MaxMsgSize:           agentConf.Grpc.MaxMsgSize,
		MaxReconnectAttempts: agentConf.Grpc.MaxReconnectAttempts,
	}
}

// ProvideMetricsConfig 提供 Metrics 配置
func ProvideMetricsConfig(agentConf AgentConfig) metrics.MetricsConfig {
	return agentConf.Metrics
}

// ProvidePprofConfig 提供 Pprof 配置
func ProvidePprofConfig(agentConf AgentConfig) pprof.PprofConfig {
	return agentConf.Pprof
}
