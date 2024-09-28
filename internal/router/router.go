package router

import (
	"embed"
	"github.com/gin-contrib/pprof"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/go-arcade/arcade/internal/app/engine/server"
	"github.com/go-arcade/arcade/pkg/ctx"
	"github.com/go-arcade/arcade/pkg/httpx"
	"github.com/go-arcade/arcade/pkg/httpx/interceptor"
	"github.com/go-arcade/arcade/pkg/httpx/ws"
	"github.com/go-arcade/arcade/pkg/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/8 15:48
 * @file: router.go
 * @description: router
 */

type Router struct {
	Http *server.Http
	Ctx  *ctx.Context
}

//go:embed static
var web embed.FS

func NewRouter(cfg *server.Http, ctx *ctx.Context) *Router {
	return &Router{
		Http: cfg,
		Ctx:  ctx,
	}
}

func (ar *Router) Router() *gin.Engine {

	gin.SetMode(ar.Http.Mode)

	r := gin.New()

	// panic recover
	r.Use(interceptor.ExceptionInterceptor)

	// unified response interceptor
	r.Use(interceptor.UnifiedResponseInterceptor())

	// web static resource
	r.Use(static.Serve("/", static.EmbedFolder(web, "static")))
	r.NoRoute(func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/")
	})

	if ar.Http.AccessLog {
		r.Use(gin.LoggerWithFormatter(httpx.AccessLogFormat))
	}

	if ar.Http.PProf {
		pprof.Register(r, "/debug/pprof")
	}

	if ar.Http.ExposeMetrics {
		r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	}

	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	r.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, version.GetVersion())
	})

	// core router, internal api router
	core := r.Group(ar.Http.InternalContextPath)
	{
		// engine router
		ar.RouterGroup(core)
	}

	return r
}

func (ar *Router) RouterGroup(r *gin.RouterGroup) *gin.RouterGroup {

	r.POST("/ws", ws.Handle)

	route := r.Group("/agent")
	{
		route.POST("/add", ar.addAgent)
		//r.POST("delete", deleteAgent)
		//r.POST("update", updateAgent)
		route.GET("/list", ar.listAgent)
	}

	return route
}
