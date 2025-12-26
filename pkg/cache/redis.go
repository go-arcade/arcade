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
	"crypto/tls"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-arcade/arcade/pkg/log"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
)

// ProviderSet 提供缓存相关的依赖（支持单节点、Sentinel 和集群模式）
var ProviderSet = wire.NewSet(ProvideRedisCmdable, ProvideICache)

// ProviderSetLegacy 提供缓存相关的依赖（兼容旧版本，仅支持单节点和 Sentinel）
var ProviderSetLegacy = wire.NewSet(ProvideRedis, ProvideICacheFromClient)

// ProvideRedis 提供 Redis 实例（兼容旧版本，返回单节点客户端）
// 注意：集群模式请使用 ProvideRedisCmdable
func ProvideRedis(conf Redis) (*redis.Client, error) {
	cmdable, err := NewRedisCmdable(conf)
	if err != nil {
		return nil, err
	}
	// 如果是单节点客户端，返回 *redis.Client
	if client, ok := cmdable.(*redis.Client); ok {
		return client, nil
	}
	// 集群模式不支持返回 *redis.Client，返回错误
	return nil, fmt.Errorf("cluster mode does not support *redis.Client, use ProvideRedisCmdable instead")
}

// ProvideRedisCmdable 提供 Redis 实例（支持单节点、Sentinel 和集群模式）
func ProvideRedisCmdable(conf Redis) (redis.Cmdable, error) {
	return NewRedisCmdable(conf)
}

// ProvideICache 提供 ICache 接口实例（支持单节点、Sentinel 和集群模式）
func ProvideICache(cmdable redis.Cmdable) ICache {
	return NewRedisCache(cmdable)
}

// ProvideICacheFromClient 提供 ICache 接口实例（从 *redis.Client 创建，兼容旧版本）
func ProvideICacheFromClient(client *redis.Client) ICache {
	return NewRedisCache(client)
}

type Redis struct {
	Mode             string
	Address          string
	Password         string
	DB               int
	PoolSize         int
	UseTLS           bool
	MasterName       string
	SentinelUsername string
	SentinelPassword string
	DialTimeout      time.Duration // 连接超时
	ReadTimeout      time.Duration // 读超时
	WriteTimeout     time.Duration // 写超时
	MaxRedirects     int           // 集群模式最大重定向次数
}

// NewRedis 创建 Redis 客户端（兼容旧版本，仅支持单节点和 Sentinel）
// 注意：集群模式请使用 NewRedisCmdable
func NewRedis(cfg Redis) (*redis.Client, error) {
	cmdable, err := NewRedisCmdable(cfg)
	if err != nil {
		return nil, err
	}
	// 如果是单节点客户端，返回 *redis.Client
	if client, ok := cmdable.(*redis.Client); ok {
		return client, nil
	}
	// 集群模式不支持返回 *redis.Client
	return nil, fmt.Errorf("cluster mode does not support *redis.Client, use NewRedisCmdable instead")
}

// NewRedisCmdable 创建 Redis 客户端（支持单节点、Sentinel 和集群模式）
func NewRedisCmdable(cfg Redis) (redis.Cmdable, error) {
	var cmdable redis.Cmdable

	switch cfg.Mode {
	case "single":
		redisOptions := &redis.Options{
			Addr:         cfg.Address,
			Password:     cfg.Password,
			DB:           cfg.DB,
			PoolSize:     cfg.PoolSize,
			DialTimeout:  cfg.DialTimeout * time.Second,
			ReadTimeout:  cfg.ReadTimeout * time.Second,
			WriteTimeout: cfg.WriteTimeout * time.Second,
		}
		if cfg.UseTLS {
			redisOptions.TLSConfig = &tls.Config{}
		}
		cmdable = redis.NewClient(redisOptions)
	case "sentinel":
		redisOptions := &redis.FailoverOptions{
			MasterName:       cfg.MasterName,
			SentinelAddrs:    strings.Split(cfg.Address, ","),
			Password:         cfg.Password,
			DB:               cfg.DB,
			PoolSize:         cfg.PoolSize,
			SentinelUsername: cfg.SentinelUsername,
			SentinelPassword: cfg.SentinelPassword,
			DialTimeout:      cfg.DialTimeout * time.Second,
			ReadTimeout:      cfg.ReadTimeout * time.Second,
			WriteTimeout:     cfg.WriteTimeout * time.Second,
		}
		if cfg.UseTLS {
			redisOptions.TLSConfig = &tls.Config{}
		}
		cmdable = redis.NewFailoverClient(redisOptions)
	case "cluster":
		maxRedirects := cfg.MaxRedirects
		if maxRedirects == 0 {
			maxRedirects = 3 // 默认最大重定向次数
		}
		clusterOptions := &redis.ClusterOptions{
			Addrs:        strings.Split(cfg.Address, ","),
			Password:     cfg.Password,
			PoolSize:     cfg.PoolSize,
			DialTimeout:  cfg.DialTimeout * time.Second,
			ReadTimeout:  cfg.ReadTimeout * time.Second,
			WriteTimeout: cfg.WriteTimeout * time.Second,
			MaxRedirects: maxRedirects,
		}
		if cfg.UseTLS {
			clusterOptions.TLSConfig = &tls.Config{}
		}
		cmdable = redis.NewClusterClient(clusterOptions)
	default:
		log.Errorw("failed to init redis, redis type is illegal", "mode", cfg.Mode)
		os.Exit(1)
	}

	err := cmdable.Ping(context.Background()).Err()
	if err != nil {
		log.Errorw("failed to connect redis", "error", err, "mode", cfg.Mode)
		return nil, err
	}

	log.Infow("redis connected",
		"mode", cfg.Mode,
	)

	return cmdable, nil
}
