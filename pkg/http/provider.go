package http

import (
	"github.com/gofiber/fiber/v2"
)

// ProvideHttpServer 提供 HTTP 服务器启动和清理函数
func ProvideHttpServer(cfg Http, app *fiber.App) func() {
	return NewHttp(cfg, app)
}
