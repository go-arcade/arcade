package config

import (
	"github.com/go-arcade/arcade/internal/server/http"
	"github.com/go-arcade/arcade/pkg/cache"
	"github.com/go-arcade/arcade/pkg/log"
	"github.com/go-arcade/arcade/pkg/orm"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/8 23:20
 * @file: config.go
 * @description:
 */

type AppConfig struct {
	Log      log.LogConfig
	Http     http.HTTP
	Database orm.Database
	Redis    cache.Redis
}
