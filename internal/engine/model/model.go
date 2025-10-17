package model

import "time"

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/28 21:55
 * @file: model.go
 * @description: base model
 */

type BaseModel struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}
