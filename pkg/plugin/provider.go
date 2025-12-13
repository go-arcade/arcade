package plugin

import (
	"path/filepath"
	"time"

	"github.com/go-arcade/arcade/pkg/database"
	"github.com/go-arcade/arcade/pkg/log"
	"github.com/go-arcade/arcade/pkg/runner"
	"github.com/google/wire"
)

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
// Uses default plugin directory: /var/lib/arcade/plugins
func ProvidePluginManager(dbAccessor DB) *Manager {
	// Use default plugin directory
	pluginDir := filepath.Join(runner.Pwd, "plugins")

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

	// Autoload plugins from directory on startup
	if err := m.LoadPluginsFromDir(); err != nil {
		log.Warnw("failed to auto-load plugins", "error", err)
	}

	log.Infow("plugin manager initialized with directory", "directory", pluginDir)
	return m
}
