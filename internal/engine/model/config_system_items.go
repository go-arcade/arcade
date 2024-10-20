package model

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/10/20 12:16
 * @file: config_system_items.go
 * @description: config system items
 */

type ConfigSystemItems struct {
	BaseModel
	ConfigSystemItemId     string `gorm:"column:config_system_item_id" json:"configSystemItemId"`
	ConfigSystemItemIdName string `gorm:"column:config_system_item_name" json:"configSystemItemIdName"`
	ConfigSystemItemIdVal  string `gorm:"column:config_system_item_val" json:"configSystemItemIdVal"`
	ConfigSystemItemIdDes  string `gorm:"column:config_system_item_des" json:"configSystemItemIdDes"`
}

func (s *ConfigSystemItems) TableName() string {
	return "config_system_items"
}
