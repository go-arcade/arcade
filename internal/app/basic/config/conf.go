package config

import (
	"github.com/arcade/arcade/internal/server/http"
	"github.com/arcade/arcade/pkg/cache"
	"github.com/arcade/arcade/pkg/log"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/8 23:20
 * @file: config.go
 * @description:
 */

type Config struct {
	Log  log.Log
	Http http.HTTP

	Redis cache.RedisConf
}
