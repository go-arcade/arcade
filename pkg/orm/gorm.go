package orm

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"sync"
	"time"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/16 11:29
 * @file: gorm.go
 * @description: gorm orm
 */

type Database struct {
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
	PrintSQL     bool
}

var conn *gorm.DB
var mu sync.Mutex

func NewDatabase(cfg Database) *gorm.DB {

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DB)

	conf := logger.Config{
		SlowThreshold:             time.Second, // 慢 SQL 阈值
		LogLevel:                  logger.Info, // Log level
		Colorful:                  false,       // 禁用彩色打印
		IgnoreRecordNotFoundError: true,        // 忽略ErrRecordNotFound（记录未找到）错误
		ParameterizedQueries:      true,        // 启用参数化查询
	}

	var db *gorm.DB
	var err error

	if cfg.OutPut {
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: NewGormLogger(conf, logger.Info),
			NamingStrategy: schema.NamingStrategy{
				SingularTable: true,
			},
		})
	} else {
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
			NamingStrategy: schema.NamingStrategy{
				SingularTable: true,
			},
		})
	}
	if err != nil {
		panic(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.MaxLifetime) * time.Second)
	sqlDB.SetConnMaxIdleTime(time.Duration(cfg.MaxIdleTime) * time.Second)

	err = sqlDB.Ping()
	if err != nil {
		panic(err)
	}

	var version string
	err = db.Raw("SELECT VERSION()").Scan(&version).Error
	if err != nil {
		return nil
	}

	fmt.Println("[Init] mysql version:", version)
	conn = db
	return db
}

func GetConn() *gorm.DB {
	mu.Lock()
	defer mu.Unlock()
	return conn
}
