package config

import (
	"github.com/go-arcade/arcade/internal/engine/service"
	"github.com/go-arcade/arcade/internal/pkg/grpc"
	"github.com/go-arcade/arcade/pkg/cache"
	"github.com/go-arcade/arcade/pkg/database"
	"github.com/go-arcade/arcade/pkg/http"
	"github.com/go-arcade/arcade/pkg/log"
	"github.com/go-arcade/arcade/pkg/metrics"
	"github.com/go-arcade/arcade/pkg/pprof"
	"github.com/google/wire"
)

// ProviderSet 提供配置层相关的依赖
var ProviderSet = wire.NewSet(
	ProvideConf,
	ProvideTaskPoolConfig,
	ProvideHttpConfig,
	ProvideGrpcConfig,
	ProvideLogConfig,
	ProvideDatabaseConfig,
	ProvideRedisConfig,
	ProvideMetricsConfig,
	ProvidePprofConfig,
)

// ProvideTaskPoolConfig 提供任务池配置
func ProvideTaskPoolConfig(appConf AppConfig) service.TaskPoolConfig {
	return service.TaskPoolConfig{
		MaxWorkers:    appConf.Task.MaxWorkers,
		QueueSize:     appConf.Task.QueueSize,
		WorkerTimeout: appConf.Task.WorkerTimeout,
	}
}

// ProvideConf 提供应用配置
func ProvideConf(configPath string) AppConfig {
	return NewConf(configPath)
}

// ProvideHttpConfig 提供 HTTP 配置
func ProvideHttpConfig(appConf AppConfig) *http.Http {
	httpConfig := &appConf.Http
	httpConfig.SetDefaults()
	return httpConfig
}

// ProvideGrpcConfig 提供 gRPC 配置
func ProvideGrpcConfig(appConf AppConfig) *grpc.Conf {
	return &appConf.Grpc
}

// ProvideLogConfig 提供日志配置
func ProvideLogConfig(appConf AppConfig) *log.Conf {
	return &appConf.Log
}

// ProvideDatabaseConfig 提供数据库配置
func ProvideDatabaseConfig(appConf AppConfig) database.Database {
	return appConf.Database
}

// ProvideRedisConfig 提供 Redis 配置
func ProvideRedisConfig(appConf AppConfig) cache.Redis {
	return appConf.Redis
}

// ProvideMetricsConfig 提供 Metrics 配置
func ProvideMetricsConfig(appConf AppConfig) metrics.MetricsConfig {
	return appConf.Metrics
}

// ProvidePprofConfig 提供 Pprof 配置
func ProvidePprofConfig(appConf AppConfig) pprof.PprofConfig {
	return appConf.Pprof
}
