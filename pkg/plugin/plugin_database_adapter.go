// Package plugin provides database accessor adapter
package plugin

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-arcade/arcade/pkg/database"
)

// DB defines database capabilities accessible by plugins
// Plugins access database through this interface without directly depending on internal packages
type DB interface {
	// QueryConfig queries plugin configuration
	// pluginID: plugin ID
	// Returns JSON string of plugin configuration
	QueryConfig(ctx context.Context, pluginID string) (string, error)

	// QueryConfigByKey queries configuration value by key
	// pluginID: plugin ID
	// key: configuration key name
	// Returns JSON string of configuration value
	QueryConfigByKey(ctx context.Context, pluginID string, key string) (string, error)

	// ListConfigs lists all plugin configurations
	// Returns JSON array string of all plugin configurations
	ListConfigs(ctx context.Context) (string, error)
}

// PluginDBAdapter is an adapter that adapts internal/repo implementation to pkg/plugin.DatabaseAccessor interface
// This avoids direct import of pkg/plugin by internal/repo, preventing circular dependencies
type PluginDBAdapter struct {
	db database.IDatabase
}

// NewPluginDBAdapter creates a plugin database accessor adapter
// Directly receives database.IDatabase to avoid circular dependencies
func NewPluginDBAdapter(db database.IDatabase) DB {
	return &PluginDBAdapter{
		db: db,
	}
}

// QueryConfig queries plugin configuration
func (a *PluginDBAdapter) QueryConfig(ctx context.Context, pluginID string) (string, error) {
	if a.db == nil {
		return "", fmt.Errorf("database is not initialized")
	}

	var config struct {
		PluginID string          `gorm:"column:plugin_id" json:"pluginId"`
		Params   json.RawMessage `gorm:"column:params;type:json" json:"params"`
		Config   json.RawMessage `gorm:"column:config;type:json" json:"config"`
	}

	err := a.db.Database().WithContext(ctx).
		Table("t_plugin_config").
		Where("plugin_id = ?", pluginID).
		First(&config).Error
	if err != nil {
		return "", err
	}

	// Serialize to JSON
	configJSON, err := json.Marshal(config)
	if err != nil {
		return "", err
	}

	return string(configJSON), nil
}

// QueryConfigByKey queries configuration value by key
func (a *PluginDBAdapter) QueryConfigByKey(ctx context.Context, pluginID string, key string) (string, error) {
	if a.db == nil {
		return "", fmt.Errorf("database is not initialized")
	}

	var config struct {
		Config json.RawMessage `gorm:"column:config;type:json" json:"config"`
	}

	err := a.db.Database().WithContext(ctx).
		Table("t_plugin_config").
		Where("plugin_id = ?", pluginID).
		First(&config).Error
	if err != nil {
		return "", err
	}

	// Parse Config JSON
	var configMap map[string]interface{}
	if err := json.Unmarshal(config.Config, &configMap); err != nil {
		return "", err
	}

	// Get value for specified key
	value, exists := configMap[key]
	if !exists {
		return "", nil // Key does not exist, return empty string
	}

	// Serialize value
	valueJSON, err := json.Marshal(value)
	if err != nil {
		return "", err
	}

	return string(valueJSON), nil
}

// ListConfigs lists all plugin configurations
func (a *PluginDBAdapter) ListConfigs(ctx context.Context) (string, error) {
	if a.db == nil {
		return "", fmt.Errorf("database is not initialized")
	}

	var configs []struct {
		PluginID string          `gorm:"column:plugin_id" json:"pluginId"`
		Params   json.RawMessage `gorm:"column:params;type:json" json:"params"`
		Config   json.RawMessage `gorm:"column:config;type:json" json:"config"`
	}

	err := a.db.Database().WithContext(ctx).
		Table("t_plugin_config").
		Find(&configs).Error
	if err != nil {
		return "", err
	}

	// Serialize to JSON array
	configsJSON, err := json.Marshal(configs)
	if err != nil {
		return "", err
	}

	return string(configsJSON), nil
}

