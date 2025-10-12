package router

import (
	"github.com/google/wire"
	"github.com/observabil/arcade/pkg/ctx"
	"github.com/observabil/arcade/pkg/http"
)

// ProviderSet 提供路由相关的依赖
var ProviderSet = wire.NewSet(ProvideRouter)

// ProvideRouter 提供路由实例
func ProvideRouter(httpConf *http.Http, ctx *ctx.Context) *Router {
	return NewRouter(httpConf, ctx)
}
