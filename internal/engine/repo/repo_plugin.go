package repo

import (
	"github.com/observabil/arcade/internal/engine/model"
	"github.com/observabil/arcade/pkg/ctx"
)

type PluginRepo struct {
	Ctx         *ctx.Context
	PluginModel model.Plugin
}

func NewPluginRepo(ctx *ctx.Context) *PluginRepo {
	return &PluginRepo{
		Ctx:         ctx,
		PluginModel: model.Plugin{},
	}
}

// GetEnabledPlugins 获取所有启用的插件
func (r *PluginRepo) GetEnabledPlugins() ([]model.Plugin, error) {
	var plugins []model.Plugin
	err := r.Ctx.DBSession().Table(r.PluginModel.TableName()).
		Where("is_enabled = ?", 1).
		Order("plugin_type ASC, id ASC").
		Find(&plugins).Error
	return plugins, err
}

// GetPluginByID 根据plugin_id获取插件
func (r *PluginRepo) GetPluginByID(pluginID string) (*model.Plugin, error) {
	var plugin model.Plugin
	err := r.Ctx.DBSession().Table(r.PluginModel.TableName()).
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
	err := r.Ctx.DBSession().Table(r.PluginModel.TableName()).
		Where("plugin_type = ? AND is_enabled = ?", pluginType, 1).
		Order("id ASC").
		Find(&plugins).Error
	return plugins, err
}

// GetPluginConfig 根据插件ID获取配置
func (r *PluginRepo) GetPluginConfig(pluginID string) (*model.PluginConfig, error) {
	var config model.PluginConfig
	err := r.Ctx.DBSession().Table("t_plugin_config").
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
	err := r.Ctx.DBSession().Table("t_task_plugin").
		Where("task_id = ?", taskID).
		Order("execution_order ASC").
		Find(&taskPlugins).Error
	return taskPlugins, err
}

// CreateTaskPlugin 创建任务插件关联
func (r *PluginRepo) CreateTaskPlugin(taskPlugin *model.TaskPlugin) error {
	return r.Ctx.DBSession().Table("t_task_plugin").Create(taskPlugin).Error
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
	return r.Ctx.DBSession().Table("t_task_plugin").
		Where("id = ?", id).
		Updates(updates).Error
}

// ======================== 插件管理方法 ========================

// CreatePlugin 创建插件
func (r *PluginRepo) CreatePlugin(plugin *model.Plugin) error {
	return r.Ctx.DBSession().Table(r.PluginModel.TableName()).Create(plugin).Error
}

// UpdatePlugin 更新插件
func (r *PluginRepo) UpdatePlugin(pluginID string, updates map[string]interface{}) error {
	return r.Ctx.DBSession().Table(r.PluginModel.TableName()).
		Where("plugin_id = ?", pluginID).
		Updates(updates).Error
}

// DeletePlugin 删除插件
func (r *PluginRepo) DeletePlugin(pluginID string) error {
	return r.Ctx.DBSession().Table(r.PluginModel.TableName()).
		Where("plugin_id = ?", pluginID).
		Delete(&model.Plugin{}).Error
}

// UpdatePluginStatus 更新插件状态
func (r *PluginRepo) UpdatePluginStatus(pluginID string, status int) error {
	return r.Ctx.DBSession().Table(r.PluginModel.TableName()).
		Where("plugin_id = ?", pluginID).
		Update("is_enabled", status).Error
}

// ListPlugins 列出插件（支持过滤）
func (r *PluginRepo) ListPlugins(pluginType string, isEnabled int) ([]model.Plugin, error) {
	var plugins []model.Plugin
	query := r.Ctx.DBSession().Table(r.PluginModel.TableName())

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
func (r *PluginRepo) GetPluginByName(name string) (*model.Plugin, error) {
	var plugin model.Plugin
	err := r.Ctx.DBSession().Table(r.PluginModel.TableName()).
		Where("name = ?", name).
		First(&plugin).Error
	if err != nil {
		return nil, err
	}
	return &plugin, nil
}

// DeletePluginConfigs 删除插件的所有配置
func (r *PluginRepo) DeletePluginConfigs(pluginID string) error {
	return r.Ctx.DBSession().Table("t_plugin_config").
		Where("plugin_id = ?", pluginID).
		Delete(&model.PluginConfig{}).Error
}

// CreatePluginConfig 创建插件配置
func (r *PluginRepo) CreatePluginConfig(config *model.PluginConfig) error {
	return r.Ctx.DBSession().Table("t_plugin_config").Create(config).Error
}

// UpdatePluginConfig 更新插件配置
func (r *PluginRepo) UpdatePluginConfig(pluginID string, updates map[string]interface{}) error {
	return r.Ctx.DBSession().Table("t_plugin_config").
		Where("plugin_id = ?", pluginID).
		Updates(updates).Error
}
