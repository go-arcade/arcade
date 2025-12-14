// Copyright 2025 Arcade Team
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package repo

import (
	"github.com/go-arcade/arcade/internal/engine/model"
	"github.com/go-arcade/arcade/pkg/database"
)

type IPluginRepository interface {
	GetEnabledPlugins() ([]model.Plugin, error)
	GetPluginByID(pluginID string) (*model.Plugin, error)
	GetPluginsByType(pluginType string) ([]model.Plugin, error)
	GetPluginConfig(pluginID string) (*model.PluginConfig, error)
	GetTaskPlugins(taskID string) ([]model.TaskPlugin, error)
	CreateTaskPlugin(taskPlugin *model.TaskPlugin) error
	UpdateTaskPluginStatus(id int, status int, result, errorMsg string) error
	CreatePlugin(p *model.Plugin) error
	UpdatePlugin(pluginID string, updates map[string]interface{}) error
	DeletePlugin(pluginID string) error
	UpdatePluginStatus(pluginID string, status int) error
	ListPlugins(pluginType string, isEnabled int) ([]model.Plugin, error)
	GetPluginByName(name string) (*model.Plugin, error)
	DeletePluginConfigs(pluginID string) error
	CreatePluginConfig(config *model.PluginConfig) error
	UpdatePluginConfig(pluginID string, updates map[string]interface{}) error
	UpdatePluginRegistrationStatus(pluginID string, status int, errorMsg string) error
}

type PluginRepo struct {
	database.IDatabase
}

func NewPluginRepo(db database.IDatabase) IPluginRepository {
	return &PluginRepo{
		IDatabase: db,
	}
}

// GetEnabledPlugins 获取所有启用的插件
func (r *PluginRepo) GetEnabledPlugins() ([]model.Plugin, error) {
	var plugins []model.Plugin
	var plugin model.Plugin
	err := r.Database().Table(plugin.TableName()).
		Select("id", "plugin_id", "name", "display_name", "description", "plugin_type", "version", "author", "repo_url", "icon", "is_enabled", "install_time", "created_at", "updated_at").
		Where("is_enabled = ?", 1).
		Order("plugin_type ASC, id ASC").
		Find(&plugins).Error
	return plugins, err
}

// GetPluginByID 根据plugin_id获取插件
func (r *PluginRepo) GetPluginByID(pluginID string) (*model.Plugin, error) {
	var plugin model.Plugin
	err := r.Database().Table(plugin.TableName()).
		Select("id", "plugin_id", "name", "version", "description", "author", "plugin_type", "entry_point", "icon", "repository", "documentation", "is_enabled", "register_status", "register_error", "checksum", "source", "s3_path", "manifest", "install_time", "created_at", "updated_at").
		Where("plugin_id = ? AND is_enabled = ?", pluginID, 1).
		First(&plugin).Error
	if err != nil {
		return nil, err
	}
	return &plugin, nil
}

// GetPluginsByType 根据类型获取插件列表
func (r *PluginRepo) GetPluginsByType(pluginType string) ([]model.Plugin, error) {
	var plugins []model.Plugin
	var plugin model.Plugin
	err := r.Database().Table(plugin.TableName()).
		Select("id", "plugin_id", "name", "version", "description", "author", "plugin_type", "entry_point", "icon", "repository", "documentation", "is_enabled", "register_status", "register_error", "checksum", "source", "s3_path", "manifest", "install_time", "created_at", "updated_at").
		Where("plugin_type = ? AND is_enabled = ?", pluginType, 1).
		Order("id ASC").
		Find(&plugins).Error
	return plugins, err
}

// GetPluginConfig 根据插件ID获取配置
func (r *PluginRepo) GetPluginConfig(pluginID string) (*model.PluginConfig, error) {
	var config model.PluginConfig
	err := r.Database().Table(config.TableName()).
		Select("id", "plugin_id", "args", "config", "created_at", "updated_at").
		Where("plugin_id = ?", pluginID).
		First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// GetTaskPlugins 获取任务关联的插件列表
func (r *PluginRepo) GetTaskPlugins(taskID string) ([]model.TaskPlugin, error) {
	var taskPlugins []model.TaskPlugin
	var taskPlugin model.TaskPlugin

	err := r.Database().Table(taskPlugin.TableName()).
		Select("id", "task_id", "plugin_id", "plugin_config_id", "args", "execution_order", "execution_stage", "status", "result", "error_message", "started_at", "completed_at", "created_at", "updated_at").
		Where("task_id = ?", taskID).
		Order("execution_order ASC").
		Find(&taskPlugins).Error
	return taskPlugins, err
}

// CreateTaskPlugin 创建任务插件关联
func (r *PluginRepo) CreateTaskPlugin(taskPlugin *model.TaskPlugin) error {
	return r.Database().Table(taskPlugin.TableName()).Create(taskPlugin).Error
}

// UpdateTaskPluginStatus 更新任务插件执行状态
func (r *PluginRepo) UpdateTaskPluginStatus(id int, status int, result, errorMsg string) error {
	var taskPlugin model.TaskPlugin

	updates := map[string]interface{}{
		"status": status,
	}
	if result != "" {
		updates["result"] = result
	}
	if errorMsg != "" {
		updates["error_message"] = errorMsg
	}
	return r.Database().Table(taskPlugin.TableName()).
		Where("id = ?", id).
		Updates(updates).Error
}

// CreatePlugin 创建插件
func (r *PluginRepo) CreatePlugin(p *model.Plugin) error {
	var plugin model.Plugin

	return r.Database().Table(plugin.TableName()).Create(p).Error
}

// UpdatePlugin 更新插件
func (r *PluginRepo) UpdatePlugin(pluginID string, updates map[string]interface{}) error {
	var plugin model.Plugin
	return r.Database().Table(plugin.TableName()).
		Where("plugin_id = ?", pluginID).
		Updates(updates).Error
}

// DeletePlugin 删除插件
func (r *PluginRepo) DeletePlugin(pluginID string) error {
	var plugin model.Plugin
	return r.Database().Table(plugin.TableName()).
		Where("plugin_id = ?", pluginID).
		Delete(&model.Plugin{}).Error
}

// UpdatePluginStatus 更新插件状态
func (r *PluginRepo) UpdatePluginStatus(pluginID string, status int) error {
	var plugin model.Plugin
	return r.Database().Table(plugin.TableName()).
		Where("plugin_id = ?", pluginID).
		Update("is_enabled", status).Error
}

// ListPlugins 列出插件（支持过滤）
func (r *PluginRepo) ListPlugins(pluginType string, isEnabled int) ([]model.Plugin, error) {
	var plugins []model.Plugin
	var plugin model.Plugin
	query := r.Database().Table(plugin.TableName())

	if pluginType != "" {
		query = query.Where("plugin_type = ?", pluginType)
	}
	if isEnabled != 0 {
		query = query.Where("is_enabled = ?", isEnabled)
	}

	err := query.Select("id", "plugin_id", "name", "version", "description", "author", "plugin_type", "entry_point", "icon", "repository", "documentation", "is_enabled", "register_status", "register_error", "checksum", "source", "s3_path", "manifest", "install_time", "created_at", "updated_at").
		Order("install_time DESC").Find(&plugins).Error
	return plugins, err
}

// GetPluginByName 根据名称获取插件
func (r *PluginRepo) GetPluginByName(name string) (*model.Plugin, error) {
	var plugin model.Plugin
	err := r.Database().Table(plugin.TableName()).
		Select("id", "plugin_id", "name", "version", "description", "author", "plugin_type", "entry_point", "icon", "repository", "documentation", "is_enabled", "register_status", "register_error", "checksum", "source", "s3_path", "manifest", "install_time", "created_at", "updated_at").
		Where("name = ?", name).
		First(&plugin).Error
	if err != nil {
		return nil, err
	}
	return &plugin, nil
}

// DeletePluginConfigs 删除插件的所有配置
func (r *PluginRepo) DeletePluginConfigs(pluginID string) error {
	var config model.PluginConfig
	return r.Database().Table(config.TableName()).
		Where("plugin_id = ?", pluginID).
		Delete(&model.PluginConfig{}).Error
}

// CreatePluginConfig 创建插件配置
func (r *PluginRepo) CreatePluginConfig(config *model.PluginConfig) error {
	return r.Database().Table(config.TableName()).Create(config).Error
}

// UpdatePluginConfig 更新插件配置
func (r *PluginRepo) UpdatePluginConfig(pluginID string, updates map[string]interface{}) error {
	var config model.PluginConfig
	return r.Database().Table(config.TableName()).
		Where("plugin_id = ?", pluginID).
		Updates(updates).Error
}

// UpdatePluginRegistrationStatus 更新插件注册状态
func (r *PluginRepo) UpdatePluginRegistrationStatus(pluginID string, status int, errorMsg string) error {
	updates := map[string]interface{}{
		"register_status": status,
	}
	if errorMsg != "" {
		updates["register_error"] = errorMsg
	} else {
		updates["register_error"] = ""
	}
	var plugin model.Plugin
	return r.Database().Table(plugin.TableName()).
		Where("plugin_id = ?", pluginID).
		Updates(updates).Error
}
