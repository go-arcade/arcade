package model

import "gorm.io/datatypes"

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/10/11 21:59
 * @file: model_config_security_items.go
 * @description: config security items
 */

type ConfigSecurityItems struct {
	BaseModel
	ConfigSecurityItemId     string         `gorm:"column:config_security_item_id" json:"configSecurityItemId"`
	ConfigSecurityItemIdName string         `gorm:"column:config_security_item_name" json:"configSecurityItemIdName"`
	ConfigSecurityItemIdVal  datatypes.JSON `gorm:"column:config_security_item_val" json:"configSecurityItemIdVal"`
	ConfigSecurityItemIdDes  string         `gorm:"column:config_security_item_des" json:"configSecurityItemIdDes"`
}

func (s *ConfigSecurityItems) TableName() string {
	return "config_security_items"
}
