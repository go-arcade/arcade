package conf

import (
	"github.com/google/wire"
	"github.com/observabil/arcade/internal/engine/service/job"
)

// ProviderSet 提供配置相关的依赖
var ProviderSet = wire.NewSet(ProvideConf, ProvideJobPoolConfig)

// ProvideConf 提供完整配置实例
func ProvideConf(configFile string) AppConfig {
	return NewConf(configFile)
}

// ProvideJobPoolConfig 提供 Job 池配置
func ProvideJobPoolConfig(appConf AppConfig) job.JobPoolConfig {
	return job.JobPoolConfig{
		MaxWorkers:    appConf.Job.MaxWorkers,
		QueueSize:     appConf.Job.QueueSize,
		WorkerTimeout: appConf.Job.WorkerTimeout,
	}
}
