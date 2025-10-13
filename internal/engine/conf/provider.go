package conf

import (
	"github.com/google/wire"
	"github.com/observabil/arcade/internal/engine/service/job"
	"github.com/observabil/arcade/pkg/ctx"
	"github.com/observabil/arcade/pkg/storage"
)

// ProviderSet 提供配置相关的依赖
var ProviderSet = wire.NewSet(ProvideConf, ProvideStorageConfig, ProvideJobPoolConfig)

// ProvideConf 提供完整配置实例
func ProvideConf(configFile string) AppConfig {
	return NewConf(configFile)
}

// ProvideStorageConfig 提供存储配置
func ProvideStorageConfig(appConf AppConfig, appCtx *ctx.Context) *storage.Storage {
	return &storage.Storage{
		Ctx:       appCtx,
		Provider:  appConf.Storage.Provider,
		AccessKey: appConf.Storage.AccessKey,
		SecretKey: appConf.Storage.SecretKey,
		Endpoint:  appConf.Storage.Endpoint,
		Bucket:    appConf.Storage.Bucket,
		Region:    appConf.Storage.Region,
		UseTLS:    appConf.Storage.UseTLS,
		BasePath:  appConf.Storage.BasePath,
	}
}

// ProvideJobPoolConfig 提供 Job 池配置
func ProvideJobPoolConfig(appConf AppConfig) job.JobPoolConfig {
	return job.JobPoolConfig{
		MaxWorkers:    appConf.Job.MaxWorkers,
		QueueSize:     appConf.Job.QueueSize,
		WorkerTimeout: appConf.Job.WorkerTimeout,
	}
}
