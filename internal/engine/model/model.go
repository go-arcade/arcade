package model

import (
	"time"
)

type BaseModel struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"-"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"-"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"-"`
}
