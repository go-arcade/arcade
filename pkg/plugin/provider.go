package plugin

import (
	"time"

	"github.com/go-arcade/arcade/pkg/database"
	"github.com/go-arcade/arcade/pkg/log"
	"github.com/google/wire"
)

// PluginCacheDir is the plugin cache directory path type
type PluginCacheDir string

// ProviderSet provides plugin layer related dependencies
var ProviderSet = wire.NewSet(
	ProvideDatabaseAccessor,
	ProvidePluginManager,
)

// ProvideDatabaseAccessor provides database accessor
// Creates an adapter from database.IDatabase, implementing DatabaseAccessor interface
func ProvideDatabaseAccessor(db database.IDatabase) DB {
	return NewPluginDBAdapter(db)
}

// ProvidePluginManager provides plugin manager instance
// dbAccessor is provided by ProvideDatabaseAccessor
// pluginCacheDir is the plugin cache directory path (can be empty to use default)
func ProvidePluginManager(pluginCacheDir PluginCacheDir, dbAccessor DB) *Manager {
	// Get plugin directory from configuration (use default if not set)
	pluginDir := "/var/lib/arcade/plugins"
	if string(pluginCacheDir) != "" {
		pluginDir = string(pluginCacheDir)
	}

	// Create plugin manager configuration
	config := &ManagerConfig{
		PluginDir:       pluginDir,
		HandshakeConfig: PluginHandshake,
		PluginConfig:    make(map[string]any),
		Timeout:         30 * time.Second,
		MaxRetries:      3,
	}

	// Create plugin manager
	m := NewManager(config)

	// Set database accessor (implemented by internal/repo)
	m.SetDB(dbAccessor)

	// Auto-load plugins from directory on startup
	if err := m.LoadPluginsFromDir(); err != nil {
		log.Warnf("failed to auto-load plugins: %v", err)
	}

	log.Infof("plugin manager initialized with directory: %s", pluginDir)
	return m
}
