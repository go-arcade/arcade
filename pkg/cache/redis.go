package cache

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/redis/go-redis/v9"
	"os"
	"strings"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/8 23:46
 * @file: redis.go
 * @description: redis
 */

type RedisConf struct {
	Mode             string
	Host             string
	Port             int
	Password         string
	DB               int
	PoolSize         int
	UseTLS           bool
	MasterName       string
	SentinelUsername string
	SentinelPassword string
}

type Redis redis.Cmdable

func NewRedis(cfg RedisConf) (Redis, error) {

	var redisClient Redis
	switch cfg.Mode {
	case "single":
		redisOptions := &redis.Options{
			Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
			Password: cfg.Password,
			DB:       cfg.DB,
			PoolSize: cfg.PoolSize,
		}
		if cfg.UseTLS {
			redisOptions.TLSConfig = &tls.Config{}
		}
		redisClient = redis.NewClient(redisOptions)
	case "cluster":
		redisOptions := &redis.ClusterOptions{
			Addrs:    strings.Split(cfg.Host, ","),
			Password: cfg.Password,
			PoolSize: cfg.PoolSize,
		}
		if cfg.UseTLS {
			redisOptions.TLSConfig = &tls.Config{}
		}
		redisClient = redis.NewClusterClient(redisOptions)
	case "sentinel":
		redisOptions := &redis.FailoverOptions{
			MasterName:       cfg.MasterName,
			SentinelAddrs:    strings.Split(cfg.Host, ","),
			Password:         cfg.Password,
			DB:               cfg.DB,
			PoolSize:         cfg.PoolSize,
			SentinelUsername: cfg.SentinelUsername,
			SentinelPassword: cfg.SentinelPassword,
		}
		if cfg.UseTLS {
			redisOptions.TLSConfig = &tls.Config{}
		}
		redisClient = redis.NewFailoverClient(redisOptions)
	default:
		fmt.Println("failed to init redis , redis type is illegal", cfg.Mode)
		os.Exit(1)
	}

	err := redisClient.Ping(context.Background()).Err()
	if err != nil {
		fmt.Println("failed to connect redis", err)
		return nil, err
	}

	return redisClient, nil
}
