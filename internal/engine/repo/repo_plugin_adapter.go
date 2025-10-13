package repo

import (
	"encoding/json"

	"github.com/observabil/arcade/pkg/plugin"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2025/01/13
 * @file: repo_plugin_adapter.go
 * @description: plugin repository adapter for plugin manager
 */

// PluginRepoAdapter 适配器，让 PluginRepo 实现 plugin.PluginRepository 接口
type PluginRepoAdapter struct {
	repo *PluginRepo
}

func NewPluginRepoAdapter(repo *PluginRepo) *PluginRepoAdapter {
	return &PluginRepoAdapter{repo: repo}
}

// GetEnabledPlugins 实现 plugin.PluginRepository 接口
func (a *PluginRepoAdapter) GetEnabledPlugins() ([]plugin.PluginModel, error) {
	plugins, err := a.repo.GetEnabledPlugins()
	if err != nil {
		return nil, err
	}

	result := make([]plugin.PluginModel, len(plugins))
	for i, p := range plugins {
		result[i] = plugin.PluginModel{
			PluginId:      p.PluginId,
			Name:          p.Name,
			Version:       p.Version,
			PluginType:    p.PluginType,
			EntryPoint:    p.EntryPoint,
			ConfigSchema:  json.RawMessage(p.ConfigSchema),
			DefaultConfig: json.RawMessage(p.DefaultConfig),
			IsEnabled:     p.IsEnabled,
			InstallPath:   p.InstallPath,
			Checksum:      p.Checksum,
		}
	}

	return result, nil
}

// GetPluginByID 实现 plugin.PluginRepository 接口
func (a *PluginRepoAdapter) GetPluginByID(pluginID string) (*plugin.PluginModel, error) {
	p, err := a.repo.GetPluginByID(pluginID)
	if err != nil {
		return nil, err
	}

	return &plugin.PluginModel{
		PluginId:      p.PluginId,
		Name:          p.Name,
		Version:       p.Version,
		PluginType:    p.PluginType,
		EntryPoint:    p.EntryPoint,
		ConfigSchema:  json.RawMessage(p.ConfigSchema),
		DefaultConfig: json.RawMessage(p.DefaultConfig),
		IsEnabled:     p.IsEnabled,
		InstallPath:   p.InstallPath,
		Checksum:      p.Checksum,
	}, nil
}

// GetDefaultPluginConfig 实现 plugin.PluginRepository 接口
func (a *PluginRepoAdapter) GetDefaultPluginConfig(pluginID string) (*plugin.PluginConfigModel, error) {
	cfg, err := a.repo.GetDefaultPluginConfig(pluginID)
	if err != nil {
		return nil, err
	}

	return &plugin.PluginConfigModel{
		ConfigId:    cfg.ConfigId,
		PluginId:    cfg.PluginId,
		ConfigItems: json.RawMessage(cfg.ConfigItems),
	}, nil
}

// Ensure PluginRepoAdapter implements plugin.PluginRepository
var _ plugin.PluginRepository = (*PluginRepoAdapter)(nil)
