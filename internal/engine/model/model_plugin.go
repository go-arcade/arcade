package model

import "github.com/observabil/arcade/pkg/datatype"

/**
 * @author: gagral.x@gmail.com
 * @time: 2025/01/13
 * @file: model_plugin.go
 * @description: plugin model
 */

// Plugin 插件表
type Plugin struct {
	BaseModel
	PluginId      string        `gorm:"column:plugin_id" json:"pluginId"`
	Name          string        `gorm:"column:name" json:"name"`
	Version       string        `gorm:"column:version" json:"version"`
	Description   string        `gorm:"column:description;type:text" json:"description"`
	Author        string        `gorm:"column:author" json:"author"`
	PluginType    string        `gorm:"column:plugin_type" json:"pluginType"` // notify/deploy/test/build/custom
	EntryPoint    string        `gorm:"column:entry_point" json:"entryPoint"`
	ConfigSchema  datatype.JSON `gorm:"column:config_schema;type:json" json:"configSchema"`   // JSON Schema
	ParamsSchema  datatype.JSON `gorm:"column:params_schema;type:json" json:"paramsSchema"`   // JSON Schema
	DefaultConfig datatype.JSON `gorm:"column:default_config;type:json" json:"defaultConfig"` // 默认配置
	Icon          string        `gorm:"column:icon" json:"icon"`
	Repository    string        `gorm:"column:repository" json:"repository"`
	Documentation string        `gorm:"column:documentation;type:text" json:"documentation"`
	IsEnabled     int           `gorm:"column:is_enabled" json:"isEnabled"` // 0:禁用 1:启用
	IsBuiltin     int           `gorm:"column:is_builtin" json:"isBuiltin"` // 0:否 1:是
	InstallPath   string        `gorm:"column:install_path" json:"installPath"`
	Checksum      string        `gorm:"column:checksum" json:"checksum"`
}

func (Plugin) TableName() string {
	return "t_plugin"
}

// PluginConfig 插件配置表
type PluginConfig struct {
	BaseModel
	ConfigId    string        `gorm:"column:config_id" json:"configId"`
	PluginId    string        `gorm:"column:plugin_id" json:"pluginId"`
	Name        string        `gorm:"column:name" json:"name"`
	ConfigItems datatype.JSON `gorm:"column:config_items;type:json" json:"configItems"` // key-value配置
	Description string        `gorm:"column:description" json:"description"`
	Scope       string        `gorm:"column:scope" json:"scope"`          // global/pipeline/user
	ScopeId     string        `gorm:"column:scope_id" json:"scopeId"`     // 作用域ID
	IsDefault   int           `gorm:"column:is_default" json:"isDefault"` // 0:否 1:是
	CreatedBy   string        `gorm:"column:created_by" json:"createdBy"`
}

func (PluginConfig) TableName() string {
	return "t_plugin_config"
}

// JobPlugin 任务插件关联表
type JobPlugin struct {
	BaseModel
	JobId          string        `gorm:"column:job_id" json:"jobId"`
	PluginId       string        `gorm:"column:plugin_id" json:"pluginId"`
	PluginConfigId string        `gorm:"column:plugin_config_id" json:"pluginConfigId"`
	Params         datatype.JSON `gorm:"column:params;type:json" json:"params"` // 任务特定参数
	ExecutionOrder int           `gorm:"column:execution_order" json:"executionOrder"`
	ExecutionStage string        `gorm:"column:execution_stage" json:"executionStage"` // before/after/on_success/on_failure
	Status         int           `gorm:"column:status" json:"status"`                  // 0:未执行 1:执行中 2:成功 3:失败
	Result         string        `gorm:"column:result;type:text" json:"result"`
	ErrorMessage   string        `gorm:"column:error_message;type:text" json:"errorMessage"`
	StartedAt      *string       `gorm:"column:started_at" json:"startedAt"`
	CompletedAt    *string       `gorm:"column:completed_at" json:"completedAt"`
}

func (JobPlugin) TableName() string {
	return "t_job_plugin"
}

// PluginSchema 插件Schema结构（用于解析config_schema和params_schema）
type PluginSchema struct {
	Type       string                    `json:"type"`
	Properties map[string]SchemaProperty `json:"properties"`
	Required   []string                  `json:"required,omitempty"`
}

// SchemaProperty Schema属性定义
type SchemaProperty struct {
	Type        string          `json:"type"`
	Description string          `json:"description,omitempty"`
	Default     interface{}     `json:"default,omitempty"`
	Items       *SchemaProperty `json:"items,omitempty"` // 用于array类型
}

// PluginExecutionStage 插件执行阶段常量
const (
	PluginStageBefore    = "before"     // 任务执行前
	PluginStageAfter     = "after"      // 任务执行后
	PluginStageOnSuccess = "on_success" // 任务成功后
	PluginStageOnFailure = "on_failure" // 任务失败后
)

// PluginType 插件类型常量
const (
	PluginTypeNotify = "notify" // 通知类插件
	PluginTypeDeploy = "deploy" // 部署类插件
	PluginTypeTest   = "test"   // 测试类插件
	PluginTypeBuild  = "build"  // 构建类插件
	PluginTypeCustom = "custom" // 自定义插件
)

// JobPluginStatus 任务插件执行状态
const (
	JobPluginStatusPending = 0 // 未执行
	JobPluginStatusRunning = 1 // 执行中
	JobPluginStatusSuccess = 2 // 成功
	JobPluginStatusFailed  = 3 // 失败
)
