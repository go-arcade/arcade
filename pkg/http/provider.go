package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/wire"
)

// ProviderSet 提供 HTTP 相关的依赖
var ProviderSet = wire.NewSet(ProvideHttpServer)

// ProvideHttpServer 提供 HTTP 服务器启动和清理函数
func ProvideHttpServer(cfg Http, app *fiber.App) func() {
	return NewHttp(cfg, app)
}
