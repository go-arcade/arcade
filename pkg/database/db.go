package database

import (
	"gorm.io/gorm"
)

// IDatabase 定义数据库接口（抽象）
type IDatabase interface {
	// Database 返回底层的 *gorm.DB
	Database() *gorm.DB
}

// GormDB GORM 数据库实现
type GormDB struct {
	db *gorm.DB
}

// NewGormDB 创建 GORM 数据库实例
func NewGormDB(db *gorm.DB) IDatabase {
	return &GormDB{db: db}
}

// Database 返回底层的 *gorm.DB
func (g *GormDB) Database() *gorm.DB {
	return g.db
}
