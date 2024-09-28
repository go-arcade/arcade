package ctx

import (
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
	Ctx      context.Context
	Log      *zap.SugaredLogger
}

func NewContext(ctx context.Context, db *gorm.DB, log *zap.SugaredLogger) *Context {
	return &Context{
		MySQLIns: db,
		Ctx:      ctx,
		Log:      log,
	}
}

func (c *Context) GetCtx() context.Context {
	return c.Ctx
}

func (c *Context) SetDB(db *gorm.DB) {
	c.MySQLIns = db
}

func (c *Context) GetDB() *gorm.DB {
	return c.MySQLIns
}
