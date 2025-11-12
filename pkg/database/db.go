package database

import (
	"gorm.io/gorm"
)

// DB 定义数据库接口（抽象）
type DB interface {
	// DB 返回底层的 *gorm.DB
	DB() *gorm.DB
}

// GormDB GORM 数据库实现
type GormDB struct {
	db *gorm.DB
}

// NewGormDB 创建 GORM 数据库实例
func NewGormDB(db *gorm.DB) DB {
	return &GormDB{db: db}
}

// DB 返回底层的 *gorm.DB
func (g *GormDB) DB() *gorm.DB {
	return g.db
}
