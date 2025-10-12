package ctx

import (
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"gorm.io/gorm"
)

// ProviderSet 提供上下文相关的依赖
var ProviderSet = wire.NewSet(ProvideContext, ProvideBaseContext)

// ProvideBaseContext 提供基础 context.Context
func ProvideBaseContext() context.Context {
	return context.Background()
}

// ProvideContext 提供应用上下文
func ProvideContext(
	baseCtx context.Context,
	mongodb *mongo.Database,
	redis *redis.Client,
	db *gorm.DB,
	logger *zap.SugaredLogger,
) *Context {
	return NewContext(baseCtx, mongodb.Client(), redis, db, logger)
}

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
