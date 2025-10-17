package router

import (
	"embed"
	"io/fs"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/observabil/arcade/internal/engine/conf"
	"github.com/observabil/arcade/internal/engine/service"
	"github.com/observabil/arcade/pkg/ctx"
	httpx "github.com/observabil/arcade/pkg/http"
	"github.com/observabil/arcade/pkg/http/middleware"
	"github.com/observabil/arcade/pkg/http/ws"
	pluginpkg "github.com/observabil/arcade/pkg/plugin"
	"github.com/observabil/arcade/pkg/version"
	"go.uber.org/zap"
)

type Router struct {
	Http          *httpx.Http
	Ctx           *ctx.Context
	Plugin        *conf.PluginConfig
	PermService   *service.PermissionService
	PluginManager *pluginpkg.Manager
}

const (
	apiContextPath = "/api/v1"
	// openApiPath    = "/openapi"
)

//go:embed all:static
var web embed.FS

func NewRouter(httpConf *httpx.Http, ctx *ctx.Context, pluginConfig *conf.PluginConfig, pluginManager *pluginpkg.Manager) *Router {
	return &Router{
		Http:          httpConf,
		Ctx:           ctx,
		Plugin:        pluginConfig,
		PluginManager: pluginManager,
	}
}

func (rt *Router) Router(log *zap.Logger) *fiber.App {
	// 设置默认的 BodyLimit（100MB）
	bodyLimit := rt.Http.BodyLimit
	if bodyLimit <= 0 {
		bodyLimit = 100 * 1024 * 1024 // 100MB 默认值
	}

	app := fiber.New(fiber.Config{
		AppName:      "Arcade",
		ReadTimeout:  time.Duration(rt.Http.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(rt.Http.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(rt.Http.IdleTimeout) * time.Second,
		BodyLimit:    bodyLimit, // 请求体大小限制，用于插件上传等
	})

	// 访问日志 - 必须在最前面
	if rt.Http.AccessLog {
		app.Use(httpx.AccessLogFormat(log))
	}

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

	// 找不到路径时的处理 - 必须在所有路由注册之后
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).
			JSON(httpx.WithRepErr(c, fiber.StatusNotFound, "request path not found", c.Path()))
	})

	return app
}

func (rt *Router) routerGroup(r fiber.Router) {
	auth := middleware.AuthorizationMiddleware(
		rt.Http.Auth.SecretKey,
		rt.Http.Auth.RedisKeyPrefix,
		*rt.Ctx.RedisSession(),
	)

	// user
	rt.userRouter(r, auth)

	// auth
	rt.authRouter(r, auth)

	// agent
	rt.agentRouter(r, auth)

	// storag
	rt.storageRouter(r, auth)

	// plugin
	rt.pluginRouter(r, auth)
}

func (rt *Router) storageRouter(r fiber.Router, auth fiber.Handler) {
	// 存储配置管理需要认证
	_ = r.Group("/storage", auth)
	{
		// 这里需要注入 StorageService，暂时留空
		// 实际使用时需要在 router 初始化时注入服务
	}
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
