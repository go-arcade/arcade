package database

import (
	"context"

	"github.com/google/wire"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ProviderSet 提供数据库相关的依赖
var ProviderSet = wire.NewSet(ProvideDatabase, ProvideMongoDB)

// ProvideDatabase 提供 MySQL 数据库实例
func ProvideDatabase(conf Database, logger *zap.Logger) (*gorm.DB, error) {
	return NewDatabase(conf, *logger)
}

// ProvideMongoDB 提供 MongoDB 实例
func ProvideMongoDB(conf Database, ctx context.Context) (*mongo.Database, error) {
	client, err := NewMongoDB(conf.MongoDB, ctx)
	if err != nil {
		return nil, err
	}
	return client.Database(conf.MongoDB.DB), nil
}
