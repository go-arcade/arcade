package database

import (
	"fmt"
	"go.uber.org/zap"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/16 11:29
 * @file: mysql.go
 * @description: gorm database
 */

type Database struct {
	Type         string
	Host         string
	Port         string
	User         string
	Password     string
	DB           string
	OutPut       bool
	MaxOpenConns int
	MaxIdleConns int
	MaxLifetime  int
	MaxIdleTime  int
	MongoDB      MongoDB `toml:"mongodb"`
}

const (
	defaultTablePrefix = "t_"
	defaultLogLevel    = logger.Info
	defaultSlowSQL     = time.Second
)

var dbConnection *gorm.DB
var mu sync.Mutex

// NewDatabase initializes and returns a new Gorm database instance.
func NewDatabase(cfg Database, zapLogger zap.Logger) (*gorm.DB, error) {

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
			Logger: NewGormLogger(logConfig, logger.Info, &zapLogger),
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
		return nil, fmt.Errorf("failed to get underlying sql.DB handle: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.MaxLifetime) * time.Second)
	sqlDB.SetConnMaxIdleTime(time.Duration(cfg.MaxIdleTime) * time.Second)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	dbConnection = db
	return db, nil
}

// GetConn returns the global database connection.
func GetConn() *gorm.DB {
	mu.Lock()
	defer mu.Unlock()
	return dbConnection
}
