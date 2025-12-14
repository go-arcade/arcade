package router

import (
	"time"

	"github.com/go-arcade/arcade/pkg/http"
	"github.com/go-arcade/arcade/pkg/http/middleware"
	"github.com/go-arcade/arcade/pkg/version"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberrecover "github.com/gofiber/fiber/v2/middleware/recover"
)

type Router struct {
	Http *http.Http
}

func NewRouter(
	httpConf *http.Http,
) *Router {
	return &Router{
		Http: httpConf,
	}
}

func (rt *Router) Router() *fiber.App {
	// 设置默认的 BodyLimit（100MB）
	bodyLimit := rt.Http.BodyLimit
	if bodyLimit <= 0 {
		bodyLimit = 100 * 1024 * 1024 // 100MB 默认值
	}

	app := fiber.New(fiber.Config{
		AppName: "Arcade Agent",
		// DisableStartupMessage: true,
		ReadTimeout:  time.Duration(rt.Http.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(rt.Http.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(rt.Http.IdleTimeout) * time.Second,
		BodyLimit:    bodyLimit, // 请求体大小限制，用于插件上传等
	})

	app.Use(http.AccessLogFormat(rt.Http))

	// 中间件
	app.Use(
		fiberrecover.New(),
		cors.New(),
		middleware.TraceMiddleware(), // 链路追踪中间件
		middleware.UnifiedResponseMiddleware(),
	)

	// 健康检查
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	// 版本信息
	app.Get("/version", func(c *fiber.Ctx) error {
		return c.JSON(version.GetVersion())
	})

	// 找不到路径时的处理 - 必须在所有路由注册之后
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).
			JSON(http.WithRepErr(c, fiber.StatusNotFound, "request path not found", c.Path()))
	})

	return app
}
