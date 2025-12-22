// Copyright 2025 Arcade Team
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cache

import (
	"context"
	"time"

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
	return r.client.Get(ctx, key)
}

// Set 设置缓存值
func (r *RedisCache) Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd {
	return r.client.Set(ctx, key, value, expiration)
}

// Del 删除缓存
func (r *RedisCache) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	return r.client.Del(ctx, keys...)
}

// Pipeline 创建管道
func (r *RedisCache) Pipeline() redis.Pipeliner {
	return r.client.Pipeline()
}

// GetClient 获取底层的 redis.Client
func (r *RedisCache) GetClient() *redis.Client {
	return r.client
}

// HSet 设置 Hash 字段
func (r *RedisCache) HSet(ctx context.Context, key string, values ...any) *redis.IntCmd {
	return r.client.HSet(ctx, key, values...)
}

// HGetAll 获取 Hash 所有字段
func (r *RedisCache) HGetAll(ctx context.Context, key string) *redis.MapStringStringCmd {
	return r.client.HGetAll(ctx, key)
}

// HDel 删除 Hash 字段
func (r *RedisCache) HDel(ctx context.Context, key string, fields ...string) *redis.IntCmd {
	return r.client.HDel(ctx, key, fields...)
}

// Expire 设置过期时间
func (r *RedisCache) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	return r.client.Expire(ctx, key, expiration)
}
