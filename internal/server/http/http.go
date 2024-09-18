package http

import (
	"context"
	"fmt"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/go-arcade/arcade/internal/router"
	"github.com/go-arcade/arcade/pkg/httpx"
	"github.com/go-arcade/arcade/pkg/httpx/interceptor"
	"github.com/go-arcade/arcade/pkg/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"time"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/8 15:38
 * @file: httpx.go
 * @description: http server
 */

type HTTP struct {
	Host            string
	Port            int
	Mode            string
	ContextPath     string
	Heartbeat       string
	PProf           bool
	ExposeMetrics   bool
	AccessLog       bool
	ReadTimeout     int
	WriteTimeout    int
	IdleTimeout     int
	ShutdownTimeout int
	TLS             TLS
	Auth            Auth
}

type TLS struct {
	CertFile string
	KeyFile  string
}

type Auth struct {
	AccessExpire   int64
	RefreshExpire  int64
	RedisKeyPrefix string
}

func NewHTTPEngine(cfg HTTP) *gin.Engine {

	gin.SetMode(cfg.Mode)

	r := gin.New()
	// panic recover
	r.Use(gin.Recovery())

	// response interceptor
	r.Use(interceptor.UnifiedResponseInterceptor())

	if cfg.AccessLog {
		r.Use(gin.LoggerWithFormatter(httpx.AccessLogFormat))
	}

	if cfg.PProf {
		pprof.Register(r, "/debug/pprof")
	}

	if cfg.ExposeMetrics {
		r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	}

	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	r.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, version.GetVersion())
	})

	core := r.Group(cfg.ContextPath)
	{
		// engine router
		router.EngineRouter(core)
	}

	return r
}

func NewHTTP(cfg HTTP, handler http.Handler) func() {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:      handler,
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.IdleTimeout) * time.Second,
	}

	go func() {
		fmt.Println("[Init] httpx server start at:", addr)

		if cfg.TLS.CertFile != "" && cfg.TLS.KeyFile != "" {
			if err := srv.ListenAndServeTLS(cfg.TLS.CertFile, cfg.TLS.KeyFile); err != nil {
				panic(err)
			}
		} else {
			if err := srv.ListenAndServe(); err != nil {
				panic(err)
			}
		}
	}()

	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(cfg.ShutdownTimeout))
		defer cancel()

		srv.SetKeepAlivesEnabled(false)
		if err := srv.Shutdown(ctx); err != nil {
			fmt.Println("server shutdown error: ", err)
		}

		select {
		case <-ctx.Done():
			fmt.Println("server shutdown timeout of ", cfg.ShutdownTimeout, " seconds")
		default:
			fmt.Println("[Done] httpx exit...")
		}
	}
}
