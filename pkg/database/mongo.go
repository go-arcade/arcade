package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	Uri         string
	DB          string
	Compressors []string
	PoolSize    uint64
}

// MongoClient 包装MongoDB客户端和数据库
type MongoClient struct {
	Client *mongo.Client
	DB     *mongo.Database
}

func NewMongoDB(cfg MongoDB, ctx context.Context) (*MongoClient, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	clientOption := options.Client().ApplyURI(cfg.Uri)
	clientOption.SetCompressors(cfg.Compressors)
	clientOption.SetMaxPoolSize(cfg.PoolSize)
	client, err := mongo.Connect(ctx, clientOption)
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	// 获取数据库实例
	database := client.Database(cfg.DB)

	return &MongoClient{
		Client: client,
		DB:     database,
	}, nil
}

// GetCollection 获取集合，无需再指定数据库
func (mc *MongoClient) GetCollection(name string) *mongo.Collection {
	return mc.DB.Collection(name)
}

// Close 关闭连接
func (mc *MongoClient) Close(ctx context.Context) error {
	return mc.Client.Disconnect(ctx)
}
