package cache

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/redis/go-redis/v9"
	"os"
	"strings"
	"time"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/8 23:46
 * @file: redis.go
 * @description: redis
 */

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
		fmt.Println("failed to init redis , redis type is illegal", cfg.Mode)
		os.Exit(1)
	}

	err := redisClient.Ping(context.Background()).Err()
	if err != nil {
		fmt.Println("failed to connect redis", err)
		return nil, err
	}

	fmt.Printf("[Init] redis connected, mode: %s\n", cfg.Mode)

	return redisClient, nil
}
