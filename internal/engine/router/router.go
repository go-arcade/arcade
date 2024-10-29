package router

import (
	"embed"
	"fmt"
	"github.com/cnlesscode/gotool/gintool"
	"github.com/gin-contrib/pprof"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/go-arcade/arcade/pkg/ctx"
	httpx "github.com/go-arcade/arcade/pkg/http"
	"github.com/go-arcade/arcade/pkg/http/interceptor"
	"github.com/go-arcade/arcade/pkg/http/ws"
	"github.com/go-arcade/arcade/pkg/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/8 15:48
 * @file: router.go
 * @description: setup router
 *  		     internal api router, use by web
 */

type Router struct {
	Http *httpx.Http
	Ctx  *ctx.Context
}

//go:embed static
var web embed.FS

func NewRouter(httpConf *httpx.Http, ctx *ctx.Context) *Router {
	return &Router{
		Http: httpConf,
		Ctx:  ctx,
	}
}

func (rt *Router) Router() *gin.Engine {

	gin.SetMode(rt.Http.Mode)

	r := gin.New()

	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
		fmt.Printf("[Arcade] %-6s %-25s --> %s (%d handlers) \n", httpMethod, absolutePath, handlerName, nuHandlers)
	}

	// cors interceptor
	r.Use(interceptor.CorsInterceptor())

	// panic recover
	r.Use(interceptor.ExceptionInterceptor)

	// unified response interceptor
	r.Use(interceptor.UnifiedResponseInterceptor())

	// r.Use(interceptor.AuthorizationInterceptor(rt.Http.Auth.SecretKey, rt.Http.Auth))

	// web static resource
	if rt.Http.UseFileAssets {
		r.Use(static.Serve("/", static.EmbedFolder(web, "static")))
		r.NoRoute(func(c *gin.Context) {
			c.Redirect(http.StatusMovedPermanently, "/")
		})
	}

	if rt.Http.AccessLog {
		r.Use(gin.LoggerWithFormatter(httpx.AccessLogFormat))
	}

	if rt.Http.PProf {
		pprof.Register(r, "/debug/pprof")
	}

	if rt.Http.ExposeMetrics {
		r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	}

	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	r.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, version.GetVersion())
	})

	// engine router, internal api router
	engine := r.Group(rt.Http.InternalContextPath)
	{
		// ws
		engine.POST("/ws", ws.Handle)

		// core
		rt.routerGroup(engine)
	}

	return r
}

func (rt *Router) routerGroup(r *gin.RouterGroup) *gin.RouterGroup {

	auth := interceptor.AuthorizationInterceptor(rt.Http.Auth.SecretKey, rt.Http.Auth.RedisKeyPrefix, *rt.Ctx.GetRedis())

	// user
	user := r.Group("/user")
	{
		// not auth
		user.POST("/login", rt.login)
		user.POST("/register", rt.register)
		user.POST("/oauth", rt.oauth)

		// auth
		user.POST("/logout", rt.logout, auth)
		user.GET("/refresh", rt.refresh, auth)
		user.POST("/invite", rt.addUser, auth)
		user.POST("/revise", rt.updateUser, auth)
		user.GET("/getUserInfo", rt.getUserInfo, auth)
		//user.GET("/getUserList", rt.getUserList)
	}

	// agent
	route := r.Group("/agent", auth)
	{
		route.POST("/add", rt.addAgent)
		//r.POST("delete", deleteAgent)
		//r.POST("update", updateAgent)
		route.GET("/list", rt.listAgent)
	}

	return route
}

func queryInt(r *gin.Context, key string) int {
	value, ok := gintool.QueryInt(r, key)
	if !ok {
		return 0
	}
	return value
}
