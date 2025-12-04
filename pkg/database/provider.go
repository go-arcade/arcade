package database

import (
	"context"

	"github.com/go-arcade/arcade/pkg/log"
	"github.com/google/wire"
	"gorm.io/gorm"
)

// ProviderSet 提供数据库相关的依赖
var ProviderSet = wire.NewSet(
	ProvideDatabase,
	ProvideMongoDB,
	ProvideIDatabase,
	ProvideMongoDBInterface,
)

// ProvideDatabase 提供 MySQL 数据库实例
func ProvideDatabase(conf Database, logger *log.Logger) (*gorm.DB, error) {
	zapLogger := logger.Log.Desugar()
	return NewDatabase(conf, *zapLogger)
}

// ProvideMongoDB 提供 MongoDB 实例
func ProvideMongoDB(conf Database, ctx context.Context) (*MongoClient, error) {
	client, err := NewMongoDB(conf.MongoDB, ctx)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// ProvideIDatabase 提供 IDatabase 接口实例
func ProvideIDatabase(db *gorm.DB) IDatabase {
	return NewGormDB(db)
}

// ProvideMongoDBInterface 提供 MongoDB 接口实例
func ProvideMongoDBInterface(client *MongoClient) MongoDB {
	return NewMongoDBWrapper(client)
}
