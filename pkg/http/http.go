package http

import (
	"context"
	"errors"
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
	InternalContextPath string
	ExternalContextPath string
	Heartbeat           int64
	PProf               bool
	ExposeMetrics       bool
	AccessLog           bool
	UseFileAssets       bool
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

func NewHttp(cfg Http, engine *gin.Engine) func() {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	srv := &http.Server{
		Addr:         addr,
		Handler:      engine,
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.IdleTimeout) * time.Second,
	}

	go func() {
		fmt.Printf("[Init] http server start at: %s\n", srv.Addr)
		var err error
		if cfg.TLS.CertFile != "" && cfg.TLS.KeyFile != "" {
			if err := srv.ListenAndServeTLS(cfg.TLS.CertFile, cfg.TLS.KeyFile); err != nil {
				panic(err)
			}
		} else {
			if err := srv.ListenAndServe(); err != nil {
				panic(err)
			}
		}
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("[Error] HTTP server error: %v\n", err)
			os.Exit(1)
		}
	}()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	return createShutdownHook(srv, cfg.ShutdownTimeout, sc)
}

func createShutdownHook(server *http.Server, shutdownTimeout int, signalChan chan os.Signal) func() {
	signal.Notify(signalChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	return func() {
		<-signalChan
		fmt.Println("[Shutdown] HTTP server shutting down...")

		ctx2, cancel := context.WithTimeout(context.Background(), time.Duration(shutdownTimeout)*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx2); err != nil {
			fmt.Printf("[Error] Server shutdown error: %v\n", err)
		} else {
			fmt.Println("[Shutdown] HTTP server shut down gracefully.")
		}
	}
}
