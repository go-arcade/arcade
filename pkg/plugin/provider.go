package plugin

import (
	"time"

	"github.com/go-arcade/arcade/internal/engine/conf"
	"github.com/go-arcade/arcade/pkg/database"
	"github.com/go-arcade/arcade/pkg/log"
	"github.com/google/wire"
)

// ProviderSet 提供插件层相关的依赖
var ProviderSet = wire.NewSet(
	ProvideDatabaseAccessor,
	ProvidePluginManager,
)

// ProvideDatabaseAccessor 提供数据库访问器
// 从 database.DB 创建适配器，实现 DatabaseAccessor 接口
func ProvideDatabaseAccessor(db database.DB) DatabaseAccessor {
	return NewPluginDBAccessorAdapter(db)
}

// ProvidePluginManager 提供插件管理器实例
// dbAccessor 由 ProvideDatabaseAccessor 提供
func ProvidePluginManager(appConf conf.AppConfig, dbAccessor DatabaseAccessor) *Manager {
	// Get plugin directory from configuration (use default if not set)
	pluginDir := "/var/lib/arcade/plugins"
	if appConf.Plugin.CacheDir != "" {
		pluginDir = appConf.Plugin.CacheDir
	}

	// Create plugin manager configuration
	config := &ManagerConfig{
		PluginDir:       pluginDir,
		HandshakeConfig: RPCHandshake,
		PluginConfig:    make(map[string]any),
		Timeout:         30 * time.Second,
		MaxRetries:      3,
	}

	// Create plugin manager
	m := NewManager(config)

	// 设置数据库访问器（由 internal/repo 实现）
	m.SetDatabaseAccessor(dbAccessor)

	// Auto-load plugins from directory on startup
	if err := m.LoadPluginsFromDir(); err != nil {
		log.Warnf("failed to auto-load plugins: %v", err)
	}

	log.Infof("plugin manager initialized with directory: %s", pluginDir)
	return m
}
