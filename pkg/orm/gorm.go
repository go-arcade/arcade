package orm

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"time"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/16 11:29
 * @file: gorm.go
 * @description: gorm orm
 */

var db *gorm.DB

type Database struct {
	Host              string
	Port              int
	Username          string
	Password          string
	DB                string
	MaxOpenConns      int
	MaxIdleConns      int
	MaxLifetime       int
	MaxIdleTime       int
	EnableAutoMigrate bool
}

func NewDatabase(cfg Database) *gorm.DB {

	gormLogger := logger.New(
		log.New(log.Writer(), "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // 慢 SQL 阈值
			LogLevel:                  logger.Info, // Log level
			Colorful:                  false,       // 禁用彩色打印
			IgnoreRecordNotFoundError: true,        // 忽略ErrRecordNotFound（记录未找到）错误
			ParameterizedQueries:      true,        // 启用参数化查询
		},
	)

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DB)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormLogger,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
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

	var result int
	err = db.Raw("SELECT 1").Scan(&result).Error
	if err != nil {
		panic(err)
	}

	return db
}

func GetConnection() *gorm.DB {
	return db
}
