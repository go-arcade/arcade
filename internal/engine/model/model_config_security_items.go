package model

import "gorm.io/datatypes"

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/10/11 21:59
 * @file: model_config_secure_items.go
 * @description: config secure items
 */

type ConfigSecureItems struct {
	BaseModel
	ConfigSecurityItemId     string         `gorm:"column:config_secure_item_id" json:"configSecurityItemId"`
	ConfigSecurityItemIdName string         `gorm:"column:config_secure_item_name" json:"configSecurityItemIdName"`
	ConfigSecurityItemIdVal  datatypes.JSON `gorm:"column:config_secure_item_val" json:"configSecurityItemIdVal"`
	ConfigSecurityItemIdDes  string         `gorm:"column:config_secure_item_des" json:"configSecurityItemIdDes"`
}

func (s *ConfigSecureItems) TableName() string {
	return "t_config_secure_items"
}
