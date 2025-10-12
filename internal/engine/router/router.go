package router

import (
	"embed"
	"io/fs"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/observabil/arcade/pkg/ctx"
	httpx "github.com/observabil/arcade/pkg/http"
	"github.com/observabil/arcade/pkg/http/middleware"
	"github.com/observabil/arcade/pkg/http/ws"
	"github.com/observabil/arcade/pkg/version"
	"go.uber.org/zap"
)

type Router struct {
	Http *httpx.Http
	Ctx  *ctx.Context
}

const (
	apiContextPath = "/api/v1"
	// openApiPath    = "/openapi"
)

//go:embed all:static
var web embed.FS

func NewRouter(httpConf *httpx.Http, ctx *ctx.Context) *Router {
	return &Router{
		Http: httpConf,
		Ctx:  ctx,
	}
}

func (rt *Router) Router(log *zap.Logger) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: "Arcade",
	})

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
			log.Fatal("embed FS subdir error:", zap.Error(err))
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

	// 访问日志
	if rt.Http.AccessLog {
		app.Use(logger.New())
	}

	// pprof
	if rt.Http.Pprof {
		pprofGroup := app.Group("/debug/pprof")
		rt.debugRouter(pprofGroup)
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
		api.Post("/ws", ws.Handle)

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

	return app
}

func (rt *Router) routerGroup(r fiber.Router) {
	auth := middleware.AuthorizationMiddleware(
		rt.Http.Auth.SecretKey,
		rt.Http.Auth.RedisKeyPrefix,
		*rt.Ctx.GetRedis(),
	)

	// user
	rt.userRouter(r, auth)

	// auth
	rt.authRouter(r, auth)

	// agent
	rt.agentRouter(r, auth)
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
