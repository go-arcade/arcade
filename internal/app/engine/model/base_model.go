package model

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/28 21:55
 * @file: base_model.go
 * @description: base model
 */

type BaseModel struct {
	ID        int `gorm:"primaryKey" json:"id"`
	CreatedAt int `gorm:"column:created_at" json:"CreatedAt,omitempty"`
	UpdatedAt int `gorm:"column:updated_at" json:"UpdatedAt,omitempty"`
}
