package conf

import (
	"github.com/go-arcade/arcade/internal/engine/service/task"
	"github.com/go-arcade/arcade/internal/pkg/grpc"
	"github.com/go-arcade/arcade/pkg/http"
	"github.com/google/wire"
)

// ProviderSet 提供配置层相关的依赖
var ProviderSet = wire.NewSet(
	ProvideConf,
	ProvideTaskPoolConfig,
	ProvideHttpConfig,
	ProvideGrpcConfig,
)

// ProvideTaskPoolConfig 提供任务池配置
func ProvideTaskPoolConfig(appConf AppConfig) task.TaskPoolConfig {
	return task.TaskPoolConfig{
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
	return &appConf.Http
}

// ProvideGrpcConfig 提供 gRPC 配置
func ProvideGrpcConfig(appConf AppConfig) *grpc.Conf {
	return &appConf.Grpc
}
