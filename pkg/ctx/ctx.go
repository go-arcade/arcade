package ctx

import (
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"gorm.io/gorm"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/9 0:12
 * @file: ctx.go
 * @description: Global context
 */

type Context struct {
	DB    *gorm.DB
	Mongo *mongo.Client
	Redis *redis.Client
	Ctx   context.Context
	Log   *zap.SugaredLogger
}

func NewContext(ctx context.Context, mongodb *mongo.Client, redis *redis.Client, mysql *gorm.DB, log *zap.SugaredLogger) *Context {
	return &Context{
		DB:    mysql,
		Mongo: mongodb,
		Redis: redis,
		Ctx:   ctx,
		Log:   log,
	}
}

func (c *Context) GetCtx() context.Context {
	return c.Ctx
}

func (c *Context) SetDB(db *gorm.DB) {
	c.DB = db
}

func (c *Context) GetDB() *gorm.DB {
	return c.DB
}

func (c *Context) SetMongoIns(client *mongo.Client) {
	c.Mongo = client
}

func (c *Context) GetMongoIns() *mongo.Client {
	return c.Mongo
}

func (c *Context) SetRedis(redis *redis.Client) {
	c.Redis = redis
}

func (c *Context) GetRedis() *redis.Client {
	return c.Redis
}
