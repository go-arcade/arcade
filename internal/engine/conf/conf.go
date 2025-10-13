package conf

import (
	"fmt"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/observabil/arcade/internal/pkg/grpc"
	"github.com/observabil/arcade/pkg/http"
	"github.com/spf13/viper"

	"github.com/observabil/arcade/pkg/cache"
	"github.com/observabil/arcade/pkg/database"
	"github.com/observabil/arcade/pkg/log"
)

type JobConfig struct {
	MaxWorkers    int `mapstructure:"maxWorkers"`
	QueueSize     int `mapstructure:"queueSize"`
	WorkerTimeout int `mapstructure:"workerTimeout"`
}

type AppConfig struct {
	Log      log.LogConfig
	Grpc     grpc.GrpcConf
	Http     http.Http
	Database database.Database
	Redis    cache.Redis
	Job      JobConfig
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
			panic(fmt.Sprintf("load conf file error: %s", err))
		}
	})
	return cfg
}

// LoadConfigFile load conf file
func LoadConfigFile(confDir string) (AppConfig, error) {

	config := viper.New()
	config.SetConfigFile(confDir) //文件名
	config.AddConfigPath("./conf.d")
	config.SetConfigName("config")
	config.SetConfigType("toml")
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
	fmt.Printf("[Init] config file path: %s\n", confDir)

	return cfg, nil
}

func GetString(key string) string {
	return viper.GetString(key)
}
