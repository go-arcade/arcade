package http

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/observabil/arcade/pkg/ctx"
)

type Http struct {
	Host            string
	Port            int
	Mode            string
	Heartbeat       int64
	Pprof           bool
	ExposeMetrics   bool
	AccessLog       bool
	UseFileAssets   bool
	ReadTimeout     int
	WriteTimeout    int
	IdleTimeout     int
	ShutdownTimeout int
	TLS             TLS
	Auth            Auth
	Ctx             ctx.Context
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

// NewHttp 创建并启动 HTTP 服务器
func NewHttp(cfg Http, app *fiber.App) func() {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	config := fiber.Config{
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.IdleTimeout) * time.Second,
	}

	app = fiber.New(config)

	go func() {
		fmt.Printf("[Init] http server start at: %s\n", addr)
		if cfg.TLS.CertFile != "" && cfg.TLS.KeyFile != "" {
			if err := app.ListenTLS(addr, cfg.TLS.CertFile, cfg.TLS.KeyFile); err != nil {
				panic(err)
			}
		} else {
			if err := app.Listen(addr); err != nil {
				panic(err)
			}
		}
	}()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	return createShutdownHook(app, cfg.ShutdownTimeout, sc)
}

// createShutdownHook 创建关闭钩子函数
func createShutdownHook(app *fiber.App, shutdownTimeout int, signalChan chan os.Signal) func() {
	signal.Notify(signalChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	return func() {
		<-signalChan
		fmt.Println("[Shutdown] HTTP server shutting down...")

		ctx2, cancel := context.WithTimeout(context.Background(), time.Duration(shutdownTimeout)*time.Second)
		defer cancel()

		if err := app.ShutdownWithContext(ctx2); err != nil {
			fmt.Printf("[Error] Server shutdown error: %v\n", err)
		} else {
			fmt.Println("[Shutdown] HTTP server shut down gracefully.")
		}
	}
}
