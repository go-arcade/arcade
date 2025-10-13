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

func NewMongoDB(cfg MongoDB, ctx context.Context) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	clientOption := options.Client().ApplyURI(cfg.Uri)
	clientOption.SetCompressors(cfg.Compressors)
	clientOption.SetMaxPoolSize(cfg.PoolSize)
	client, err := mongo.Connect(ctx, clientOption)
	if err != nil {
		return client, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return client, err
	}

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	client.Database(cfg.DB)

	return client, nil
}
