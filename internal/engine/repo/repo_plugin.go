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

// GetPluginConfig 获取插件配置
func (r *PluginRepo) GetPluginConfig(configID string) (*model.PluginConfig, error) {
	var config model.PluginConfig
	err := r.Ctx.DBSession().Table("t_plugin_config").
		Where("config_id = ?", configID).
		First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// GetDefaultPluginConfig 获取插件的默认配置
func (r *PluginRepo) GetDefaultPluginConfig(pluginID string) (*model.PluginConfig, error) {
	var config model.PluginConfig
	err := r.Ctx.DBSession().Table("t_plugin_config").
		Where("plugin_id = ? AND is_default = ?", pluginID, 1).
		First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// GetJobPlugins 获取任务关联的插件列表
func (r *PluginRepo) GetJobPlugins(jobID string) ([]model.JobPlugin, error) {
	var jobPlugins []model.JobPlugin
	err := r.Ctx.DBSession().Table("t_job_plugin").
		Where("job_id = ?", jobID).
		Order("execution_order ASC").
		Find(&jobPlugins).Error
	return jobPlugins, err
}

// CreateJobPlugin 创建任务插件关联
func (r *PluginRepo) CreateJobPlugin(jobPlugin *model.JobPlugin) error {
	return r.Ctx.DBSession().Table("t_job_plugin").Create(jobPlugin).Error
}

// UpdateJobPluginStatus 更新任务插件执行状态
func (r *PluginRepo) UpdateJobPluginStatus(id int, status int, result, errorMsg string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if result != "" {
		updates["result"] = result
	}
	if errorMsg != "" {
		updates["error_message"] = errorMsg
	}
	return r.Ctx.DBSession().Table("t_job_plugin").
		Where("id = ?", id).
		Updates(updates).Error
}
