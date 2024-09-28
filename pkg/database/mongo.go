package database

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/21 17:01
 * @file: mongo.go
 * @description: mongo database
 */

type MongoDB struct {
	Uri         string
	DB          string
	Compressors []string
	PoolSize    uint64
}

func NewMongoDB(cfg MongoDB) *MongoDB {
	return &MongoDB{
		Uri:         cfg.Uri,
		DB:          cfg.DB,
		Compressors: cfg.Compressors,
		PoolSize:    cfg.PoolSize,
	}
}

func (m *MongoDB) Connect(ctx context.Context) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	clientOption := options.Client().ApplyURI(m.Uri)
	clientOption.SetCompressors(m.Compressors)
	clientOption.SetMaxPoolSize(m.PoolSize)
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

	client.Database(m.DB)

	return client, nil
}
