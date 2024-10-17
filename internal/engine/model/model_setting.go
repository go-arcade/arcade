package model

import "gorm.io/datatypes"

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/10/11 21:59
 * @file: model_setting.go
 * @description: setting model
 */

type ConfigItems struct {
	BaseModel
	ConfigItemId     string         `gorm:"column:config_item_id" json:"configItemId"`
	ConfigItemIdName string         `gorm:"column:config_item_name" json:"configItemIdName"`
	ConfigItemIdVal  datatypes.JSON `gorm:"column:config_item_val" json:"configItemIdVal"`
	ConfigItemIdDes  string         `gorm:"column:config_item_des" json:"configItemIdDes"`
}

func (s *ConfigItems) TableName() string {
	return "config_items"
}
