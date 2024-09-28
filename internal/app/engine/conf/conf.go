package conf

import (
	"fmt"
	"github.com/go-arcade/arcade/pkg/server"
	"path/filepath"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/go-arcade/arcade/pkg/cache"
	"github.com/go-arcade/arcade/pkg/database"
	"github.com/go-arcade/arcade/pkg/log"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/8 23:20
 * @file: conf.go
 * @description:
 */

type AppConfig struct {
	Log      log.LogConfig
	Http     server.Http
	Database database.Database
	Redis    cache.Redis
}

var (
	cfg  AppConfig
	once sync.Once
)

func NewConf(confDir string) AppConfig {
	once.Do(func() {
		cfg, _ = LoadConfigFile(confDir)
	})
	return cfg
}

// LoadConfigFile load conf file
func LoadConfigFile(confDir string) (AppConfig, error) {

	filePath, err := filepath.Abs(confDir)
	if err != nil {
		panic(fmt.Sprintf("conf file path error: %s", err))
	}

	c := new(AppConfig)
	if _, err := toml.DecodeFile(filePath, c); err != nil {
		panic(fmt.Sprintf("conf file error: %s", err))
	}

	return *c, nil
}
