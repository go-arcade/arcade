package model

import "time"

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/28 21:55
 * @file: model.go
 * @description: base model
 */

type BaseModel struct {
	ID         int       `gorm:"primaryKey" json:"id"`
	CreateTime time.Time `gorm:"column:create_time" json:"createTime,omitempty"`
	UpdateTime time.Time `gorm:"column:update_time" json:"updateTime,omitempty"`
}
