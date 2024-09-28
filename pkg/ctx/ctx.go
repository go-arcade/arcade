package ctx

import (
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
	MySQLIns *gorm.DB
	MongoIns *mongo.Client
	Ctx      context.Context
	Log      *zap.SugaredLogger
}

func NewContext(ctx context.Context, mongodb *mongo.Client, mysql *gorm.DB, log *zap.SugaredLogger) *Context {
	return &Context{
		MySQLIns: mysql,
		MongoIns: mongodb,
		Ctx:      ctx,
		Log:      log,
	}
}

func (c *Context) GetCtx() context.Context {
	return c.Ctx
}

func (c *Context) SetMySQLIns(db *gorm.DB) {
	c.MySQLIns = db
}

func (c *Context) GetMySQLIns() *gorm.DB {
	return c.MySQLIns
}

func (c *Context) SetMongoIns(client *mongo.Client) {
	c.MongoIns = client
}

func (c *Context) GetMongoIns() *mongo.Client {
	return c.MongoIns
}
