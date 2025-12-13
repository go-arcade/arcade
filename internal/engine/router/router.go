package router

import (
	"embed"
	"io/fs"
	"net/http"
	"strconv"
	"time"

	"github.com/go-arcade/arcade/pkg/log"

	"github.com/go-arcade/arcade/internal/engine/service"
	"github.com/go-arcade/arcade/pkg/cache"
	httpx "github.com/go-arcade/arcade/pkg/http"
	"github.com/go-arcade/arcade/pkg/http/middleware"
	"github.com/go-arcade/arcade/pkg/version"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type Router struct {
	Http     *httpx.Http
	Cache    cache.ICache
	Services *service.Services
}

const (
	apiContextPath = "/api/v1"
	// openApiPath    = "/openapi"
)

//go:embed all:static
var web embed.FS

func NewRouter(
	httpConf *httpx.Http,
	cache cache.ICache,
	services *service.Services,
) *Router {
	return &Router{
		Http:     httpConf,
		Cache:    cache,
		Services: services,
	}
}

func (rt *Router) Router() *fiber.App {
	// 设置默认的 BodyLimit（100MB）
	bodyLimit := rt.Http.BodyLimit
	if bodyLimit <= 0 {
		bodyLimit = 100 * 1024 * 1024 // 100MB 默认值
	}

	app := fiber.New(fiber.Config{
		AppName: "Arcade",
		// DisableStartupMessage: true,
		ReadTimeout:  time.Duration(rt.Http.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(rt.Http.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(rt.Http.IdleTimeout) * time.Second,
		BodyLimit:    bodyLimit, // 请求体大小限制，用于插件上传等
	})

	app.Use(httpx.AccessLogFormat(rt.Http))

	// 中间件
	app.Use(
		recover.New(),
		cors.New(),
		middleware.UnifiedResponseMiddleware(),
	)

	// 静态文件
	if rt.Http.UseFileAssets {
		staticFS, err := fs.Sub(web, "static")
		if err != nil {
			log.Fatalw("embed FS subdir error", "error", err)
		}

		app.Use("/", filesystem.New(filesystem.Config{
			Root:   http.FS(staticFS),
			Index:  "index.html",
			Browse: false,
		}))

		app.Use(func(c *fiber.Ctx) error {
			if c.Method() != fiber.MethodGet {
				return c.Next()
			}
			file, err := staticFS.Open("index.html")
			if err != nil {
				return fiber.ErrNotFound
			}
			stat, _ := file.Stat()
			return c.Type("html").Status(fiber.StatusOK).SendStream(file, int(stat.Size()))
		})
	}

	// 健康检查
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	// 版本信息
	app.Get("/version", func(c *fiber.Ctx) error {
		return c.JSON(version.GetVersion())
	})

	// API路由
	api := app.Group(apiContextPath)
	{
		// WebSocket
		// api.Post("/ws", ws.Handle)

		// 核心路由
		rt.routerGroup(api)
	}

	// openapi
	// openApi := app.Group("/openapi")
	// {
	// 	openApi.Get("/swagger.json", func(c *fiber.Ctx) error {
	// 		return c.SendFile("docs/swagger.json")
	// 	})
	// }

	// 找不到路径时的处理 - 必须在所有路由注册之后
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).
			JSON(httpx.WithRepErr(c, fiber.StatusNotFound, "request path not found", c.Path()))
	})

	return app
}

func (rt *Router) routerGroup(r fiber.Router) {
	auth := middleware.AuthorizationMiddleware(rt.Http.Auth.SecretKey, rt.Cache)

	// user
	rt.userRouter(r, auth)
	rt.userExtensionRouter(r, auth)

	// identity
	rt.identityRouter(r, auth)

	// role
	rt.roleRouter(r, auth)

	// agent
	rt.agentRouter(r, auth)

	// team
	rt.teamRouter(r, auth)

	// storag
	rt.storageRouter(r, auth)

	// plugin
	rt.pluginRouter(r, auth)

	// general settings
	rt.generalSettingsRouter(r, auth)

	// secrets
	rt.secretRouter(r, auth)
}

func queryInt(c *fiber.Ctx, key string) int {
	value := c.Query(key)
	if value == "" {
		return 0
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return intValue
}
