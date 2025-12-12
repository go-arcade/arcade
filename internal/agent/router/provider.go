package router

import (
	"github.com/go-arcade/arcade/pkg/http"
	"github.com/google/wire"
)

// ProviderSet 提供路由层相关的依赖
var ProviderSet = wire.NewSet(
	ProvideRouter,
)

// ProvideRouter 提供路由实例
func ProvideRouter(
	httpConf *http.Http,
) *Router {
	return NewRouter(
		httpConf,
	)
}
