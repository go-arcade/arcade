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
	db    *gorm.DB
	mongo *database.MongoClient
	redis *redis.Client
	ctx   context.Context
	log   *zap.SugaredLogger
}

func NewContext(ctx context.Context, mongodb *database.MongoClient, redis *redis.Client, mysql *gorm.DB, log *zap.SugaredLogger) *Context {
	return &Context{
		db:    mysql,
		mongo: mongodb,
		redis: redis,
		ctx:   ctx,
		log:   log,
	}
}

func (c *Context) ContextIns() context.Context {
	return c.ctx
}

func (c *Context) DBSession() *gorm.DB {
	return c.db
}

func (c *Context) MongoSession() *database.MongoClient {
	return c.mongo
}

func (c *Context) RedisSession() *redis.Client {
	return c.redis
}

func (c *Context) Log() *zap.SugaredLogger {
	return c.log
}
