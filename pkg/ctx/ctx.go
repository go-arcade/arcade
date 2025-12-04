package ctx

import (
	"github.com/go-arcade/arcade/pkg/log"
	"github.com/google/wire"
	"go.uber.org/zap"
	"golang.org/x/net/context"
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
	logger *log.Logger,
) *Context {
	return NewContext(baseCtx, logger.Log)
}

type Context struct {
	ctx context.Context
	log *zap.SugaredLogger
}

func NewContext(ctx context.Context, log *zap.SugaredLogger) *Context {
	return &Context{
		ctx: ctx,
		log: log,
	}
}

func (c *Context) ContextIns() context.Context {
	return c.ctx
}

func (c *Context) Log() *zap.SugaredLogger {
	return c.log
}
