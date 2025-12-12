package queue

import (
	agentconfig "github.com/go-arcade/arcade/internal/agent/config"
	"github.com/go-arcade/arcade/internal/engine/config"
	"github.com/go-arcade/arcade/pkg/database"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
)

// ProviderSet 提供 queue 相关的依赖（主程序使用）
var ProviderSet = wire.NewSet(
	ProvideConfig,
	ProvideQueueServer,
)

// AgentProviderSet 提供 queue 相关的依赖（Agent 使用）
var AgentProviderSet = wire.NewSet(
	ProvideAgentConfig,
	ProvideQueueClient,
	ProvidePipelineTaskHandler,
	ProvideJobTaskHandler,
	ProvideStepTaskHandler,
)

// ProvideConfig 提供 queue 配置（主程序使用）
func ProvideConfig(appConf *config.AppConfig, mongoDB database.MongoDB, redisClient *redis.Client) *Config {
	taskQueueConf := appConf.TaskQueue

	// 默认队列配置
	queues := taskQueueConf.Priority
	if len(queues) == 0 {
		queues = map[string]int{
			Critical: 6,
			Default:  3,
			Low:      1,
		}
	}

	// 默认并发数
	concurrency := taskQueueConf.Concurrency
	if concurrency <= 0 {
		concurrency = 10
	}

	return &Config{
		RedisClient:      redisClient, // 复用已有的 Redis 客户端
		MongoDB:          mongoDB,     // MongoDB 用于任务记录
		Concurrency:      concurrency,
		StrictPriority:   taskQueueConf.StrictPriority,
		Queues:           queues,
		DefaultQueue:     Default,
		LogLevel:         taskQueueConf.LogLevel,
		ShutdownTimeout:  taskQueueConf.ShutdownTimeout,
		GroupGracePeriod: taskQueueConf.GroupGracePeriod,
		GroupMaxDelay:    taskQueueConf.GroupMaxDelay,
		GroupMaxSize:     taskQueueConf.GroupMaxSize,
	}
}

// ProvideAgentConfig 提供 queue 配置（Agent 使用）
func ProvideAgentConfig(agentConf *agentconfig.AgentConfig, redisClient *redis.Client) *Config {
	taskQueueConf := agentConf.TaskQueue

	// 默认队列配置
	queues := taskQueueConf.Priority
	if len(queues) == 0 {
		queues = map[string]int{
			Critical: 6,
			Default:  3,
			Low:      1,
		}
	}

	// 默认并发数
	concurrency := taskQueueConf.Concurrency
	if concurrency <= 0 {
		concurrency = 10
	}

	return &Config{
		RedisClient:      redisClient, // 复用已有的 Redis 客户端
		MongoDB:          nil,         // Agent 不需要 MongoDB 实例
		Concurrency:      concurrency,
		StrictPriority:   taskQueueConf.StrictPriority,
		Queues:           queues,
		DefaultQueue:     Default,
		LogLevel:         taskQueueConf.LogLevel,
		ShutdownTimeout:  taskQueueConf.ShutdownTimeout,
		GroupGracePeriod: taskQueueConf.GroupGracePeriod,
		GroupMaxDelay:    taskQueueConf.GroupMaxDelay,
		GroupMaxSize:     taskQueueConf.GroupMaxSize,
	}
}

// ProvideQueueServer 提供队列服务器（主程序使用）
func ProvideQueueServer(taskQueueConfig *Config) (*Server, error) {
	return NewQueueServer(taskQueueConfig)
}

// ProvideQueueClient 提供队列客户端（Agent 使用）
func ProvideQueueClient(taskQueueConfig *Config) (*Client, error) {
	return NewQueueClient(taskQueueConfig)
}

// ProvidePipelineTaskHandler 提供流水线任务处理器
func ProvidePipelineTaskHandler() *PipelineTaskHandler {
	return NewPipelineTaskHandler()
}

// ProvideJobTaskHandler 提供作业任务处理器
func ProvideJobTaskHandler() *JobTaskHandler {
	return NewJobTaskHandler()
}

// ProvideStepTaskHandler 提供步骤任务处理器
func ProvideStepTaskHandler() *StepTaskHandler {
	return NewStepTaskHandler()
}
