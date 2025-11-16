package plugin

import (
	"time"

	"github.com/go-arcade/arcade/internal/engine/conf"
	"github.com/go-arcade/arcade/pkg/database"
	"github.com/go-arcade/arcade/pkg/log"
	"github.com/google/wire"
)

// ProviderSet provides plugin layer related dependencies
var ProviderSet = wire.NewSet(
	ProvideDatabaseAccessor,
	ProvidePluginManager,
)

// ProvideDatabaseAccessor provides database accessor
// Creates an adapter from database.DB, implementing DatabaseAccessor interface
func ProvideDatabaseAccessor(db database.DB) DatabaseAccessor {
	return NewPluginDBAccessorAdapter(db)
}

// ProvidePluginManager provides plugin manager instance
// dbAccessor is provided by ProvideDatabaseAccessor
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

	// Set database accessor (implemented by internal/repo)
	m.SetDatabaseAccessor(dbAccessor)

	// Auto-load plugins from directory on startup
	if err := m.LoadPluginsFromDir(); err != nil {
		log.Warnf("failed to auto-load plugins: %v", err)
	}

	log.Infof("plugin manager initialized with directory: %s", pluginDir)
	return m
}
