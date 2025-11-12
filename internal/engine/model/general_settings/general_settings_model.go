package general_settings

import (
	"github.com/go-arcade/arcade/internal/engine/model"
	"gorm.io/datatypes"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2025/01/15
 * @file: model_general_settings.go
 * @description: general settings and secret model
 */

// GeneralSettings 通用配置表
type GeneralSettings struct {
	model.BaseModel
	Category    string         `gorm:"column:category;type:varchar(64);not null" json:"category"`         // 配置类别，如 system、security、pipeline
	Name        string         `gorm:"column:name;type:varchar(128);not null" json:"name"`                // 配置名称（业务语义字段）
	DisplayName string         `gorm:"column:display_name;type:varchar(128);not null" json:"displayName"` // 展示名，如 JWT 密钥
	Data        datatypes.JSON `gorm:"column:data;type:json;not null" json:"data"`                        // 配置内容（结构化 JSON）
	Schema      datatypes.JSON `gorm:"column:schema;type:json" json:"schema"`                             // 配置结构定义（JSON Schema）
	Description string         `gorm:"column:description;type:varchar(255)" json:"description"`           // 配置说明
}

func (GeneralSettings) TableName() string {
	return "t_general_settings"
}

// Secret 密钥管理表
type Secret struct {
	model.BaseModel
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
