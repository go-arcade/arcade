package job

import (
	"fmt"
	"runtime"

	"github.com/google/wire"
	"github.com/observabil/arcade/pkg/log"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2025/01/13
 * @file: provider.go
 * @description: Job service provider for dependency injection
 */

// JobPoolConfig Job 池配置
type JobPoolConfig struct {
	MaxWorkers    int
	QueueSize     int
	WorkerTimeout int
}

// ProviderSet 提供 Job 服务相关依赖
var ProviderSet = wire.NewSet(
	ProvideJobWorkerPool,
)

// ProvideJobWorkerPool 提供 Job 协程池实例
func ProvideJobWorkerPool(config JobPoolConfig) *JobWorkerPool {
	// 如果配置中 maxWorkers 为 0，则根据 CPU 核心数计算
	maxWorkers := config.MaxWorkers
	if maxWorkers <= 0 {
		numCPU := runtime.NumCPU()
		maxWorkers = max(5, min(50, numCPU*2)) // 范围: [5, 50]
		log.Infof("maxWorkers not configured, calculated based on CPU: %d (CPU cores: %d)", maxWorkers, numCPU)
	}

	// 如果配置中 queueSize 为 0，则根据 maxWorkers 计算
	queueSize := config.QueueSize
	if queueSize <= 0 {
		queueSize = maxWorkers * 10 // 队列大小为工作协程数的 10 倍
		log.Infof("queueSize not configured, calculated as maxWorkers * 10: %d", queueSize)
	}

	log.Infof("initializing job worker pool: workers=%d, queue_size=%d, timeout=%ds",
		maxWorkers, queueSize, config.WorkerTimeout)

	pool := NewJobWorkerPool(maxWorkers, queueSize)

	// 启动协程池
	if err := pool.Start(); err != nil {
		panic(fmt.Sprintf("failed to start job worker pool: %v", err))
	}

	log.Info("job worker pool started successfully")

	return pool
}
