package ctx

import (
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
	DB  *gorm.DB
	Ctx context.Context
}

func NewContext(ctx context.Context, db *gorm.DB) *Context {
	return &Context{
		DB:  db,
		Ctx: ctx,
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
