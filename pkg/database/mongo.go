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
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB 定义 MongoDB 接口（抽象）
type MongoDB interface {
	// GetCollection 获取集合
	GetCollection(name string) *mongo.Collection
}

// MongoClientWrapper MongoDB 客户端包装器实现
type MongoClientWrapper struct {
	client *MongoClient
}

// NewMongoDBWrapper 创建 MongoDB 包装器实例
func NewMongoDBWrapper(client *MongoClient) MongoDB {
	return &MongoClientWrapper{client: client}
}

// GetCollection 获取集合
func (m *MongoClientWrapper) GetCollection(name string) *mongo.Collection {
	return m.client.GetCollection(name)
}

type MongoConfig struct {
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

func NewMongoDB(cfg MongoConfig, ctx context.Context) (*MongoClient, error) {
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
