package plugin

import (
	"github.com/go-arcade/arcade/internal/engine/model/plugin"
	"github.com/go-arcade/arcade/pkg/database"
)

type IPluginRepository interface {
	GetEnabledPlugins() ([]plugin.Plugin, error)
	GetPluginByID(pluginID string) (*plugin.Plugin, error)
	GetPluginsByType(pluginType string) ([]plugin.Plugin, error)
	GetPluginConfig(pluginID string) (*plugin.PluginConfig, error)
	GetTaskPlugins(taskID string) ([]plugin.TaskPlugin, error)
	CreateTaskPlugin(taskPlugin *plugin.TaskPlugin) error
	UpdateTaskPluginStatus(id int, status int, result, errorMsg string) error
	CreatePlugin(p *plugin.Plugin) error
	UpdatePlugin(pluginID string, updates map[string]interface{}) error
	DeletePlugin(pluginID string) error
	UpdatePluginStatus(pluginID string, status int) error
	ListPlugins(pluginType string, isEnabled int) ([]plugin.Plugin, error)
	GetPluginByName(name string) (*plugin.Plugin, error)
	DeletePluginConfigs(pluginID string) error
	CreatePluginConfig(config *plugin.PluginConfig) error
	UpdatePluginConfig(pluginID string, updates map[string]interface{}) error
	UpdatePluginRegistrationStatus(pluginID string, status int, errorMsg string) error
}

type PluginRepo struct {
	db          database.IDatabase
	PluginModel plugin.Plugin
}

func NewPluginRepo(db database.IDatabase) IPluginRepository {
	return &PluginRepo{
		db:          db,
		PluginModel: plugin.Plugin{},
	}
}

// GetEnabledPlugins 获取所有启用的插件
func (r *PluginRepo) GetEnabledPlugins() ([]plugin.Plugin, error) {
	var plugins []plugin.Plugin
	err := r.db.Database().Table(r.PluginModel.TableName()).
		Where("is_enabled = ?", 1).
		Order("plugin_type ASC, id ASC").
		Find(&plugins).Error
	return plugins, err
}

// GetPluginByID 根据plugin_id获取插件
func (r *PluginRepo) GetPluginByID(pluginID string) (*plugin.Plugin, error) {
	var p plugin.Plugin
	err := r.db.Database().Table(r.PluginModel.TableName()).
		Where("plugin_id = ? AND is_enabled = ?", pluginID, 1).
		First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// GetPluginsByType 根据类型获取插件列表
func (r *PluginRepo) GetPluginsByType(pluginType string) ([]plugin.Plugin, error) {
	var plugins []plugin.Plugin
	err := r.db.Database().Table(r.PluginModel.TableName()).
		Where("plugin_type = ? AND is_enabled = ?", pluginType, 1).
		Order("id ASC").
		Find(&plugins).Error
	return plugins, err
}

// GetPluginConfig 根据插件ID获取配置
func (r *PluginRepo) GetPluginConfig(pluginID string) (*plugin.PluginConfig, error) {
	var config plugin.PluginConfig
	err := r.db.Database().Table("t_plugin_config").
		Where("plugin_id = ?", pluginID).
		First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// GetTaskPlugins 获取任务关联的插件列表
func (r *PluginRepo) GetTaskPlugins(taskID string) ([]plugin.TaskPlugin, error) {
	var taskPlugins []plugin.TaskPlugin
	err := r.db.Database().Table("t_task_plugin").
		Where("task_id = ?", taskID).
		Order("execution_order ASC").
		Find(&taskPlugins).Error
	return taskPlugins, err
}

// CreateTaskPlugin 创建任务插件关联
func (r *PluginRepo) CreateTaskPlugin(taskPlugin *plugin.TaskPlugin) error {
	return r.db.Database().Table("t_task_plugin").Create(taskPlugin).Error
}

// UpdateTaskPluginStatus 更新任务插件执行状态
func (r *PluginRepo) UpdateTaskPluginStatus(id int, status int, result, errorMsg string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if result != "" {
		updates["result"] = result
	}
	if errorMsg != "" {
		updates["error_message"] = errorMsg
	}
	return r.db.Database().Table("t_task_plugin").
		Where("id = ?", id).
		Updates(updates).Error
}

// CreatePlugin 创建插件
func (r *PluginRepo) CreatePlugin(p *plugin.Plugin) error {
	return r.db.Database().Table(r.PluginModel.TableName()).Create(p).Error
}

// UpdatePlugin 更新插件
func (r *PluginRepo) UpdatePlugin(pluginID string, updates map[string]interface{}) error {
	return r.db.Database().Table(r.PluginModel.TableName()).
		Where("plugin_id = ?", pluginID).
		Updates(updates).Error
}

// DeletePlugin 删除插件
func (r *PluginRepo) DeletePlugin(pluginID string) error {
	return r.db.Database().Table(r.PluginModel.TableName()).
		Where("plugin_id = ?", pluginID).
		Delete(&plugin.Plugin{}).Error
}

// UpdatePluginStatus 更新插件状态
func (r *PluginRepo) UpdatePluginStatus(pluginID string, status int) error {
	return r.db.Database().Table(r.PluginModel.TableName()).
		Where("plugin_id = ?", pluginID).
		Update("is_enabled", status).Error
}

// ListPlugins 列出插件（支持过滤）
func (r *PluginRepo) ListPlugins(pluginType string, isEnabled int) ([]plugin.Plugin, error) {
	var plugins []plugin.Plugin
	query := r.db.Database().Table(r.PluginModel.TableName())

	if pluginType != "" {
		query = query.Where("plugin_type = ?", pluginType)
	}
	if isEnabled != 0 {
		query = query.Where("is_enabled = ?", isEnabled)
	}

	err := query.Order("install_time DESC").Find(&plugins).Error
	return plugins, err
}

// GetPluginByName 根据名称获取插件
func (r *PluginRepo) GetPluginByName(name string) (*plugin.Plugin, error) {
	var p plugin.Plugin
	err := r.db.Database().Table(r.PluginModel.TableName()).
		Where("name = ?", name).
		First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// DeletePluginConfigs 删除插件的所有配置
func (r *PluginRepo) DeletePluginConfigs(pluginID string) error {
	return r.db.Database().Table("t_plugin_config").
		Where("plugin_id = ?", pluginID).
		Delete(&plugin.PluginConfig{}).Error
}

// CreatePluginConfig 创建插件配置
func (r *PluginRepo) CreatePluginConfig(config *plugin.PluginConfig) error {
	return r.db.Database().Table("t_plugin_config").Create(config).Error
}

// UpdatePluginConfig 更新插件配置
func (r *PluginRepo) UpdatePluginConfig(pluginID string, updates map[string]interface{}) error {
	return r.db.Database().Table("t_plugin_config").
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
	return r.db.Database().Table(r.PluginModel.TableName()).
		Where("plugin_id = ?", pluginID).
		Updates(updates).Error
}
