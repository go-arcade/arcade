// Copyright 2025 Arcade Team
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package database

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-arcade/arcade/pkg/log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// IDatabase define database interface (abstract)
type IDatabase interface {
	// Database return the underlying *gorm.DB
	Database() *gorm.DB
}

// GormDB GORM database implementation
type GormDB struct {
	db *gorm.DB
}

// NewGormDB create GORM database instance
func NewGormDB(db *gorm.DB) IDatabase {
	return &GormDB{db: db}
}

// Database return the underlying *gorm.DB
func (g *GormDB) Database() *gorm.DB {
	return g.db
}

type Database struct {
	Type         string
	Host         string
	Port         string
	User         string
	Password     string
	DB           string
	OutPut       bool        `mapstructure:"output"`
	MaxOpenConns int         `mapstructure:"maxOpenConns"`
	MaxIdleConns int         `mapstructure:"maxIdleConns"`
	MaxLifetime  int         `mapstructure:"maxLifeTime"`
	MaxIdleTime  int         `mapstructure:"maxIdleTime"`
	MongoDB      MongoConfig `mapstructure:"mongodb"`
}

const (
	defaultTablePrefix = "t_"
	defaultLogLevel    = logger.Info
	defaultSlowSQL     = time.Second
)

var dbConnection *gorm.DB
var mu sync.Mutex

// NewDatabase initializes and returns a new Gorm database instance.
func NewDatabase(cfg Database) (*gorm.DB, error) {

	if cfg.Type != "mysql" {
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Type)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DB)

	var db *gorm.DB
	var err error

	logConfig := logger.Config{
		SlowThreshold:             defaultSlowSQL,
		LogLevel:                  defaultLogLevel,
		Colorful:                  false,
		IgnoreRecordNotFoundError: true,
		ParameterizedQueries:      true,
	}

	if cfg.OutPut {
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: NewGormLoggerAdapter(logConfig, logger.Info),
			NamingStrategy: schema.NamingStrategy{
				TablePrefix:   defaultTablePrefix,
				SingularTable: true,
			},
		})
	} else {
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
			NamingStrategy: schema.NamingStrategy{
				TablePrefix:   defaultTablePrefix,
				SingularTable: true,
			},
		})
	}
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.IDatabase handle: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.MaxLifetime) * time.Second)
	sqlDB.SetConnMaxIdleTime(time.Duration(cfg.MaxIdleTime) * time.Second)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info("database connected successfully")

	dbConnection = db
	return db, nil
}

// GetConn returns the global database connection.
func GetConn() *gorm.DB {
	mu.Lock()
	defer mu.Unlock()
	return dbConnection
}
