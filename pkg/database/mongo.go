package database

import (
	"context"
	"time"

	"github.com/google/wire"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

type MongoDB struct {
	Uri         string
	DB          string
	Compressors []string
	PoolSize    uint64
}

func NewMongoDB(cfg MongoDB, ctx context.Context) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	clientOption := options.Client().ApplyURI(cfg.Uri)
	clientOption.SetCompressors(cfg.Compressors)
	clientOption.SetMaxPoolSize(cfg.PoolSize)
	client, err := mongo.Connect(context.Background(), clientOption)
	if err != nil {
		return client, err
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		return client, err
	}

	defer func() {
		if err = client.Disconnect(context.Background()); err != nil {
			panic(err)
		}
	}()

	client.Database(cfg.DB)

	return client, nil
}
