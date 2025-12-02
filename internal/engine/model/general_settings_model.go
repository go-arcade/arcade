package model

import (
	"gorm.io/datatypes"
)

// GeneralSettings 通用配置表
type GeneralSettings struct {
	BaseModel
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