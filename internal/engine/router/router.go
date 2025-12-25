// Copyright 2025 Arcade Team
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package router

import (
	"embed"
	"io/fs"
	"net/http"
	"strconv"
	"time"

	"github.com/go-arcade/arcade/internal/engine/service"
	"github.com/go-arcade/arcade/pkg/cache"
	httpx "github.com/go-arcade/arcade/pkg/http"
	"github.com/go-arcade/arcade/pkg/http/middleware"
	"github.com/go-arcade/arcade/pkg/log"
	"github.com/go-arcade/arcade/pkg/shutdown"
	"github.com/go-arcade/arcade/pkg/version"
	fiberi18n "github.com/gofiber/contrib/fiberi18n/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"golang.org/x/text/language"
)

type Router struct {
	Http        *httpx.Http
	Cache       cache.ICache
	Services    *service.Services
	ShutdownMgr *shutdown.Manager
}

const (
	apiContextPath = "/api/v1"
	// openApiPath    = "/openapi"
)

//go:embed all:static
var web embed.FS

//go:embed localize/*
var localizeFS embed.FS

func NewRouter(
	httpConf *httpx.Http,
	cache cache.ICache,
	services *service.Services,
	shutdownMgr *shutdown.Manager,
) *Router {
	return &Router{
		Http:        httpConf,
		Cache:       cache,
		Services:    services,
		ShutdownMgr: shutdownMgr,
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

	// Configure i18n middleware with embedded filesystem
	app.Use(fiberi18n.New(&fiberi18n.Config{
		RootPath:         "localize",
		FormatBundleFile: "yaml",
		AcceptLanguages: []language.Tag{
			language.English,
			language.Chinese,
		},
		DefaultLanguage: language.English,
		Loader:          &fiberi18n.EmbedLoader{FS: localizeFS},
	}))

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

	// 健康检查 - 在下线时返回 503，用于 Kubernetes readiness probe
	app.Get("/health", func(c *fiber.Ctx) error {
		if rt.ShutdownMgr != nil && rt.ShutdownMgr.IsShuttingDown() {
			return c.Status(fiber.StatusServiceUnavailable).SendString("shutting down")
		}
		return c.SendString("ok")
	})

	// 优雅下线接口 - 触发服务优雅关闭
	app.Post("/shutdown", func(c *fiber.Ctx) error {
		if rt.ShutdownMgr == nil {
			return c.Status(fiber.StatusInternalServerError).
				JSON(httpx.WithRepErr(c, fiber.StatusInternalServerError, "shutdown manager not initialized", c.Path()))
		}

		if rt.ShutdownMgr.Shutdown() {
			log.Info("Graceful shutdown triggered via HTTP endpoint")
			return c.JSON(httpx.WithRepDetail(c, fiber.StatusOK, "shutdown initiated", nil))
		}

		return c.Status(fiber.StatusConflict).
			JSON(httpx.WithRepErr(c, fiber.StatusConflict, "shutdown already in progress", c.Path()))
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
	rt.userExtRouter(r, auth)

	// identity
	rt.identityRouter(r, auth)

	// agent
	rt.agentRouter(r, auth)

	// team
	rt.teamRouter(r, auth)

	// storag
	rt.storageRouter(r, auth)

	// general settings
	rt.generalSettingsRouter(r, auth)

	// project
	rt.projectRouter(r, auth)

	// secrets
	rt.secretRouter(r, auth)

	// role
	rt.roleRouter(r, auth)

	// plugin
	rt.pluginRouter(r, auth)
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
