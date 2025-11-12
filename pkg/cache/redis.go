package cache

import (
	"context"
	"crypto/tls"
	"os"
	"strings"
	"time"

	"github.com/go-arcade/arcade/pkg/log"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
)

// ProviderSet 提供缓存相关的依赖
var ProviderSet = wire.NewSet(ProvideRedis)

// ProvideRedis 提供 Redis 实例
func ProvideRedis(conf Redis) (*redis.Client, error) {
	return NewRedis(conf)
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
}

func NewRedis(cfg Redis) (*redis.Client, error) {

	var redisClient *redis.Client
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
		redisClient = redis.NewClient(redisOptions)
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
		redisClient = redis.NewFailoverClient(redisOptions)
	default:
		log.Errorf("failed to init redis, redis type is illegal: %s", cfg.Mode)
		os.Exit(1)
	}

	err := redisClient.Ping(context.Background()).Err()
	if err != nil {
		log.Errorf("failed to connect redis: %v", err)
		return nil, err
	}

	log.Infow("redis connected",
		"mode", cfg.Mode,
	)

	return redisClient, nil
}
