package bootstrap

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-arcade/arcade/internal/agent/config"
	"github.com/go-arcade/arcade/internal/agent/router"
	"github.com/go-arcade/arcade/pkg/log"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type Agent struct {
	HttpApp   *fiber.App
	Logger    *log.Logger
	AgentConf config.AgentConfig
}

type InitAppFunc func(configPath string) (*Agent, func(), error)

func NewAgent(
	rt *router.Router,
	logger *log.Logger,
	agentConf config.AgentConfig,
) (*Agent, func(), error) {
	zapLogger := logger.Log.Desugar()
	httpApp := rt.Router(zapLogger)

	cleanup := func() {

	}

	app := &Agent{
		HttpApp:   httpApp,
		Logger:    logger,
		AgentConf: agentConf,
	}
	return app, cleanup, nil
}

// Bootstrap init app, return App instance and cleanup function
func Bootstrap(configFile string, initApp InitAppFunc) (*Agent, func(), config.AgentConfig, error) {
	// Wire build App (所有依赖都由 wire 自动注入)
	app, cleanup, err := initApp(configFile)
	if err != nil {
		return nil, nil, config.AgentConfig{}, err
	}

	// 获取配置（从 app 中获取）
	agentConf := app.AgentConf

	return app, cleanup, agentConf, nil
}

// Run start app and wait for exit signal, then gracefully shutdown
func Run(app *Agent, cleanup func()) {
	logger := app.Logger.Log
	appConf := app.AgentConf

	// set signal listener (graceful shutdown)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// start HTTP server (async)
	go func() {
		glog := app.Logger.Log.Desugar().WithOptions(zap.AddCallerSkip(-1)).Sugar()
		addr := appConf.Http.Host + ":" + fmt.Sprintf("%d", appConf.Http.Port)
		glog.Infow("HTTP listener started",
			"address", addr,
		)
		if err := app.HttpApp.Listen(addr); err != nil {
			glog.Errorw("HTTP listener failed",
				"address", addr,
				zap.Error(err),
			)
		}
	}()

	// wait for exit signal
	sig := <-quit
	logger.Infof("Received signal: %v, shutting down gracefully...", sig)

	// close components in order
	// close HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	if err := app.HttpApp.ShutdownWithContext(shutdownCtx); err != nil {
		logger.Errorf("HTTP server shutdown error: %v", err)
	} else {
		logger.Info("HTTP server shut down gracefully")
	}

	// close plugin manager and other resources
	cleanup()

	logger.Info("Server shutdown complete")
}
