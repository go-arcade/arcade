package ctx

import (
	"github.com/go-arcade/arcade/pkg/database"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
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
	mongodb *database.MongoClient,
	redis *redis.Client,
	db *gorm.DB,
	logger *zap.SugaredLogger,
) *Context {
	return NewContext(baseCtx, mongodb, redis, db, logger)
}

type Context struct {
	DB    *gorm.DB
	Mongo *database.MongoClient
	Redis *redis.Client
	Ctx   context.Context
	Log   *zap.SugaredLogger
}

func NewContext(ctx context.Context, mongodb *database.MongoClient, redis *redis.Client, mysql *gorm.DB, log *zap.SugaredLogger) *Context {
	return &Context{
		DB:    mysql,
		Mongo: mongodb,
		Redis: redis,
		Ctx:   ctx,
		Log:   log,
	}
}

func (c *Context) ContextIns() context.Context {
	return c.Ctx
}

func (c *Context) SetDBSession(db *gorm.DB) {
	c.DB = db
}

func (c *Context) DBSession() *gorm.DB {
	return c.DB
}

func (c *Context) SetMongoSession(client *database.MongoClient) {
	c.Mongo = client
}

func (c *Context) MongoSession() *database.MongoClient {
	return c.Mongo
}

func (c *Context) SetRedisSession(redis *redis.Client) {
	c.Redis = redis
}

func (c *Context) RedisSession() *redis.Client {
	return c.Redis
}
