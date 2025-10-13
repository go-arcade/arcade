package model

/**
 * @author: gagral.x@gmail.com
 * @time: 2025/01/13
 * @file: model_config.go
 * @description: config model
 */

// SystemConfig 系统配置表
type SystemConfig struct {
	BaseModel
	ConfigKey   string `gorm:"column:config_key" json:"configKey"`
	ConfigValue string `gorm:"column:config_value;type:text" json:"configValue"`
	ConfigType  string `gorm:"column:config_type" json:"configType"` // string/json/int/bool
	Description string `gorm:"column:description" json:"description"`
	IsEncrypted int    `gorm:"column:is_encrypted" json:"isEncrypted"` // 0:否 1:是
}

func (SystemConfig) TableName() string {
	return "t_system_config"
}

// Secret 密钥管理表
type Secret struct {
	BaseModel
	SecretId    string `gorm:"column:secret_id" json:"secretId"`
	Name        string `gorm:"column:name" json:"name"`
	SecretType  string `gorm:"column:secret_type" json:"secretType"`             // password/token/ssh_key/env
	SecretValue string `gorm:"column:secret_value;type:text" json:"secretValue"` // 加密存储
	Description string `gorm:"column:description" json:"description"`
	Scope       string `gorm:"column:scope" json:"scope"` // global/pipeline/user
	ScopeId     string `gorm:"column:scope_id" json:"scopeId"`
	CreatedBy   string `gorm:"column:created_by" json:"createdBy"`
}

func (Secret) TableName() string {
	return "t_secret"
}
