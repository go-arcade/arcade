package plugin

import (
	"path/filepath"
	"time"

	"github.com/go-arcade/arcade/pkg/database"
	"github.com/go-arcade/arcade/pkg/log"
	"github.com/go-arcade/arcade/pkg/runner"
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
func ProvidePluginManager(dbAccessor DB) *Manager {
	// 固定使用当前目录下的plugins目录
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
		log.Warnf("failed to auto-load plugins: %v", err)
	}

	// Create and start plugin watcher
	watcher, err := NewWatcher(m)
	if err != nil {
		log.Warnf("failed to create plugin watcher: %v", err)
	} else {
		// Add plugin directory to watch
		if err := watcher.AddWatchDir(pluginDir); err != nil {
			log.Warnf("failed to add watch dir %s: %v", pluginDir, err)
		} else {
			// Start watcher
			watcher.Start()
			m.watcher = watcher
			log.Infof("plugin watcher started for directory: %s", pluginDir)
		}
	}

	log.Infof("plugin manager initialized with directory: %s", pluginDir)
	return m
}
