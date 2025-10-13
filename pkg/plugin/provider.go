package plugin

import (
	"context"

	"github.com/observabil/arcade/pkg/log"
)

// ProvidePluginManager 提供插件管理器实例
func ProvidePluginManager(repo PluginRepository) *Manager {
	m := NewManager()
	m.SetPluginRepository(repo)

	if err := m.LoadPluginsFromDatabase(); err != nil {
		log.Warnf("failed to load plugins from database: %v, will try file system", err)
		if err := m.LoadPluginsFromDir("./plugins"); err != nil {
			log.Warnf("failed to load plugins from directory: %v", err)
		}
	}

	if err := m.Init(context.Background()); err != nil {
		log.Errorf("failed to initialize plugins: %v", err)
	}

	return m
}
