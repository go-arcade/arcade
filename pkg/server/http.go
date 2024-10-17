package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-arcade/arcade/pkg/ctx"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/8 15:38
 * @file: http.go
 * @description: http server
 */

type Http struct {
	Host                string
	Port                int
	Mode                string
	InternalContextPath string `mapstructure:"internalContextPath"`
	ExternalContextPath string
	Heartbeat           int64
	PProf               bool
	ExposeMetrics       bool
	AccessLog           bool
	ReadTimeout         int
	WriteTimeout        int
	IdleTimeout         int
	ShutdownTimeout     int
	TLS                 TLS
	Auth                Auth
	Ctx                 ctx.Context
}

type TLS struct {
	CertFile string
	KeyFile  string
}

type Auth struct {
	SecretKey      string
	AccessExpire   time.Duration
	RefreshExpire  time.Duration
	RedisKeyPrefix string
}

func NewHttp(cfg Http) *Http {
	return &Http{
		Host:                cfg.Host,
		Port:                cfg.Port,
		Mode:                cfg.Mode,
		InternalContextPath: cfg.InternalContextPath,
		ExternalContextPath: cfg.ExternalContextPath,
		Heartbeat:           cfg.Heartbeat,
		PProf:               cfg.PProf,
		ExposeMetrics:       cfg.ExposeMetrics,
		AccessLog:           cfg.AccessLog,
		ReadTimeout:         cfg.ReadTimeout,
		WriteTimeout:        cfg.WriteTimeout,
		IdleTimeout:         cfg.IdleTimeout,
		ShutdownTimeout:     cfg.ShutdownTimeout,
		TLS:                 cfg.TLS,
		Auth:                cfg.Auth,
		Ctx:                 cfg.Ctx,
	}
}

func (h *Http) Server(engine *gin.Engine) func() {
	addr := fmt.Sprintf("%s:%d", h.Host, h.Port)

	srv := &http.Server{
		Addr:         addr,
		Handler:      engine,
		ReadTimeout:  time.Duration(h.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(h.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(h.IdleTimeout) * time.Second,
	}

	go func() {
		fmt.Printf("[Init] http server start at: %s\n", addr)

		if h.TLS.CertFile != "" && h.TLS.KeyFile != "" {
			if err := srv.ListenAndServeTLS(h.TLS.CertFile, h.TLS.KeyFile); err != nil {
				panic(err)
			}
		} else {
			if err := srv.ListenAndServe(); err != nil {
				panic(err)
			}
		}
	}()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	return func() {

		<-sc
		fmt.Println("[shutdown] server is shutting down...")

		c, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(h.ShutdownTimeout))
		defer cancel()

		srv.SetKeepAlivesEnabled(false)
		if err := srv.Shutdown(c); err != nil {
			fmt.Println("[shutdown] server shutdown error: ", err)
		}

		select {
		case <-c.Done():
			fmt.Println("[shutdown] server shutdown timeout of ", h.ShutdownTimeout, " seconds")
		default:
			fmt.Println("[shutdown] http exit...")
		}
	}
}
