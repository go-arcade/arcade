package config

import (
	"fmt"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/go-arcade/arcade/internal/pkg/grpc"
	"github.com/go-arcade/arcade/pkg/http"
	"github.com/spf13/viper"

	"github.com/go-arcade/arcade/pkg/cache"
	"github.com/go-arcade/arcade/pkg/database"
	"github.com/go-arcade/arcade/pkg/log"
)

type TaskConfig struct {
	MaxWorkers    int
	QueueSize     int
	WorkerTimeout int
}

type PluginConfig struct {
}

type AppConfig struct {
	Log      log.Conf
	Grpc     grpc.Conf
	Http     http.Http
	Database database.Database
	Redis    cache.Redis
	Task     TaskConfig
	Plugin   PluginConfig
}

var (
	cfg  AppConfig
	once sync.Once
)

func NewConf(confDir string) AppConfig {
	once.Do(func() {
		var err error
		cfg, err = LoadConfigFile(confDir)
		if err != nil {
			panic(fmt.Sprintf("load config file error: %s", err))
		}
	})
	return cfg
}

// LoadConfigFile load config file
func LoadConfigFile(confDir string) (AppConfig, error) {

	config := viper.New()
	config.SetConfigFile(confDir) //文件名
	if err := config.ReadInConfig(); err != nil {
		return cfg, fmt.Errorf("failed to read configuration file: %v", err)
	}

	config.WatchConfig()
	config.OnConfigChange(func(e fsnotify.Event) {
		log.Infof("The configuration changes, re -analyze the configuration file: %s", e.Name)
		if err := config.Unmarshal(&cfg); err != nil {
			_ = fmt.Errorf("failed to unmarshal configuration file: %v", err)
		}
	})
	if err := config.Unmarshal(&cfg); err != nil {
		return cfg, fmt.Errorf("failed to unmarshal configuration file: %v", err)
	}
	log.Infow("config file loaded",
		"path", confDir,
	)

	return cfg, nil
}
