package config

import (
	"fmt"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/go-arcade/arcade/pkg/http"
	"github.com/go-arcade/arcade/pkg/log"
	"github.com/spf13/viper"
)

// AgentConfig holds all configuration settings
type AgentConfig struct {
	AgentId    string
	ServerAddr string
	APIkey     string
	Log        log.Conf
	Http       http.Http
}

type Agent struct {
	Interval int
	Labels   map[string]string
}

var (
	cfg  AgentConfig
	once sync.Once
)

func NewConf(confDir string) AgentConfig {
	once.Do(func() {
		var err error
		cfg, err = loadConfigFile(confDir)
		if err != nil {
			panic(fmt.Sprintf("load config file error: %s", err))
		}
	})
	return cfg
}

// LoadConfigFile load config file
func loadConfigFile(confDir string) (AgentConfig, error) {

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
