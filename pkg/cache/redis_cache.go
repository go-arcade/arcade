package cache

import (
	"context"
	"time"

	"github.com/go-arcade/arcade/pkg/trace/inject"
	"github.com/redis/go-redis/v9"
)

// RedisCache Redis 缓存实现
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache 创建 Redis 缓存实例
func NewRedisCache(client *redis.Client) ICache {
	return &RedisCache{client: client}
}

// Get 获取缓存值
func (r *RedisCache) Get(ctx context.Context, key string) *redis.StringCmd {
	var cmd *redis.StringCmd
	_, _ = inject.CacheGet(ctx, key, func(ctx context.Context) (bool, error) {
		cmd = r.client.Get(ctx, key)
		err := cmd.Err()
		if err == redis.Nil {
			return false, nil
		}
		return err == nil, err
	})
	return cmd
}

// Set 设置缓存值
func (r *RedisCache) Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd {
	var cmd *redis.StatusCmd
	_ = inject.CacheSet(ctx, key, expiration, func(ctx context.Context) error {
		cmd = r.client.Set(ctx, key, value, expiration)
		return cmd.Err()
	})
	return cmd
}

// Del 删除缓存
func (r *RedisCache) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	var cmd *redis.IntCmd
	_, _ = inject.CacheDel(ctx, keys, func(ctx context.Context) (int64, error) {
		cmd = r.client.Del(ctx, keys...)
		return cmd.Val(), cmd.Err()
	})
	return cmd
}

// Pipeline 创建管道
func (r *RedisCache) Pipeline() redis.Pipeliner {
	return r.client.Pipeline()
}

// GetClient 获取底层的 redis.Client
func (r *RedisCache) GetClient() *redis.Client {
	return r.client
}
