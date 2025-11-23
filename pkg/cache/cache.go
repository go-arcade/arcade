package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// ICache 定义缓存接口（抽象）
type ICache interface {
	// Get 获取缓存值
	Get(ctx context.Context, key string) *redis.StringCmd
	// Set 设置缓存值
	Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd
	// Del 删除缓存
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	// Pipeline 创建管道
	Pipeline() redis.Pipeliner
}

// RedisClientGetter 获取 Redis 客户端的接口（用于需要直接访问 redis.Client 的场景）
type RedisClientGetter interface {
	// GetClient 获取底层的 redis.Client
	GetClient() *redis.Client
}
