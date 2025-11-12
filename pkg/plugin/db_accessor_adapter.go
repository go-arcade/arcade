// Package plugin provides database accessor adapter
package plugin

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-arcade/arcade/pkg/database"
)

// DatabaseAccessor 定义插件可访问的数据库能力
// 插件通过此接口访问数据库，而不直接依赖 internal 包
type DatabaseAccessor interface {
	// QueryConfig 查询插件配置
	// pluginID: 插件ID
	// 返回插件配置的 JSON 字符串
	QueryConfig(ctx context.Context, pluginID string) (string, error)

	// QueryConfigByKey 根据配置键查询配置值
	// pluginID: 插件ID
	// key: 配置键名
	// 返回配置值的 JSON 字符串
	QueryConfigByKey(ctx context.Context, pluginID string, key string) (string, error)

	// ListConfigs 列出所有插件配置
	// 返回所有插件配置的 JSON 数组字符串
	ListConfigs(ctx context.Context) (string, error)
}

// PluginDBAccessorAdapter 适配器：将 internal/repo 的实现适配到 pkg/plugin.DatabaseAccessor 接口
// 这样避免了 internal/repo 直接 import pkg/plugin，防止循环依赖
type PluginDBAccessorAdapter struct {
	db database.DB
}

// NewPluginDBAccessorAdapter 创建插件数据库访问器适配器
// 直接接收 database.DB，避免循环依赖
func NewPluginDBAccessorAdapter(db database.DB) DatabaseAccessor {
	return &PluginDBAccessorAdapter{
		db: db,
	}
}

// QueryConfig 查询插件配置
func (a *PluginDBAccessorAdapter) QueryConfig(ctx context.Context, pluginID string) (string, error) {
	if a.db == nil {
		return "", fmt.Errorf("database is not initialized")
	}

	var config struct {
		PluginID string          `gorm:"column:plugin_id" json:"pluginId"`
		Params   json.RawMessage `gorm:"column:params;type:json" json:"params"`
		Config   json.RawMessage `gorm:"column:config;type:json" json:"config"`
	}

	err := a.db.DB().WithContext(ctx).
		Table("t_plugin_config").
		Where("plugin_id = ?", pluginID).
		First(&config).Error
	if err != nil {
		return "", err
	}

	// 序列化为 JSON
	configJSON, err := json.Marshal(config)
	if err != nil {
		return "", err
	}

	return string(configJSON), nil
}

// QueryConfigByKey 根据配置键查询配置值
func (a *PluginDBAccessorAdapter) QueryConfigByKey(ctx context.Context, pluginID string, key string) (string, error) {
	if a.db == nil {
		return "", fmt.Errorf("database is not initialized")
	}

	var config struct {
		Config json.RawMessage `gorm:"column:config;type:json" json:"config"`
	}

	err := a.db.DB().WithContext(ctx).
		Table("t_plugin_config").
		Where("plugin_id = ?", pluginID).
		First(&config).Error
	if err != nil {
		return "", err
	}

	// 解析 Config JSON
	var configMap map[string]interface{}
	if err := json.Unmarshal(config.Config, &configMap); err != nil {
		return "", err
	}

	// 获取指定键的值
	value, exists := configMap[key]
	if !exists {
		return "", nil // 键不存在，返回空字符串
	}

	// 序列化值
	valueJSON, err := json.Marshal(value)
	if err != nil {
		return "", err
	}

	return string(valueJSON), nil
}

// ListConfigs 列出所有插件配置
func (a *PluginDBAccessorAdapter) ListConfigs(ctx context.Context) (string, error) {
	if a.db == nil {
		return "", fmt.Errorf("database is not initialized")
	}

	var configs []struct {
		PluginID string          `gorm:"column:plugin_id" json:"pluginId"`
		Params   json.RawMessage `gorm:"column:params;type:json" json:"params"`
		Config   json.RawMessage `gorm:"column:config;type:json" json:"config"`
	}

	err := a.db.DB().WithContext(ctx).
		Table("t_plugin_config").
		Find(&configs).Error
	if err != nil {
		return "", err
	}

	// 序列化为 JSON 数组
	configsJSON, err := json.Marshal(configs)
	if err != nil {
		return "", err
	}

	return string(configsJSON), nil
}
