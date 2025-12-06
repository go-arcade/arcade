package service

import (
	"fmt"
	"runtime"

	"github.com/go-arcade/arcade/internal/engine/repo"
	"github.com/go-arcade/arcade/internal/pkg/storage"
	"github.com/go-arcade/arcade/pkg/cache"
	"github.com/go-arcade/arcade/pkg/ctx"
	"github.com/go-arcade/arcade/pkg/database"
	"github.com/go-arcade/arcade/pkg/log"
	pluginpkg "github.com/go-arcade/arcade/pkg/plugin"
	"github.com/google/wire"
)

// ProviderSet 提供服务层相关的依赖
var ProviderSet = wire.NewSet(
	ProvideServices,
	ProvideTaskWorkerPool,
)

// ProvideServices 提供统一的 Services 实例
func ProvideServices(
	ctx *ctx.Context,
	db database.IDatabase,
	cache cache.ICache,
	repos *repo.Repositories,
	pluginManager *pluginpkg.Manager,
	storageProvider storage.StorageProvider,
) *Services {
	return NewServices(ctx, db, cache, repos, pluginManager, storageProvider)
}

// TaskPoolConfig Task 池配置
type TaskPoolConfig struct {
	MaxWorkers    int
	QueueSize     int
	WorkerTimeout int
}

// ProvideTaskWorkerPool 提供 Task 协程池实例
func ProvideTaskWorkerPool(config TaskPoolConfig) *TaskWorkerPool {
	// 如果配置中 maxWorkers 为 0，则根据 CPU 核心数计算
	maxWorkers := config.MaxWorkers
	if maxWorkers <= 0 {
		numCPU := runtime.NumCPU()
		calculated := numCPU * 2
		if calculated < 5 {
			maxWorkers = 5
		} else if calculated > 50 {
			maxWorkers = 50
		} else {
			maxWorkers = calculated
		}
		log.Info("maxWorkers not configured, calculated based on CPU: %d (CPU cores: %d)", maxWorkers, numCPU)
	}

	// 如果配置中 queueSize 为 0，则根据 maxWorkers 计算
	queueSize := config.QueueSize
	if queueSize <= 0 {
		queueSize = maxWorkers * 10 // 队列大小为工作协程数的 10 倍
		log.Info("queueSize not configured, calculated as maxWorkers * 10: %d", queueSize)
	}

	log.Info("initializing task worker pool: workers=%d, queue_size=%d, timeout=%ds",
		maxWorkers, queueSize, config.WorkerTimeout)

	pool := NewTaskWorkerPool(maxWorkers, queueSize)

	// 启动协程池
	if err := pool.Start(); err != nil {
		panic(fmt.Sprintf("failed to start task worker pool: %v", err))
	}

	log.Info("task worker pool started successfully")

	return pool
}
