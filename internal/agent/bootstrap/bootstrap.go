package bootstrap

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	agentv1 "github.com/go-arcade/arcade/api/agent/v1"
	"github.com/go-arcade/arcade/internal/agent/config"
	"github.com/go-arcade/arcade/internal/agent/router"
	"github.com/go-arcade/arcade/internal/agent/service"
	grpcclient "github.com/go-arcade/arcade/internal/pkg/grpc"
	"github.com/go-arcade/arcade/pkg/cron"
	"github.com/go-arcade/arcade/pkg/log"
	"github.com/go-arcade/arcade/pkg/metrics"
	"github.com/go-arcade/arcade/pkg/pprof"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type Agent struct {
	HttpApp       *fiber.App
	GrpcClient    *grpcclient.ClientWrapper
	MetricsServer *metrics.Server
	PprofServer   *pprof.Server
	Logger        *log.Logger
	AgentConf     config.AgentConfig
	AgentService  *service.AgentService
}

type InitAppFunc func(configPath string) (*Agent, func(), error)

func NewAgent(
	rt *router.Router,
	grpcClient *grpcclient.ClientWrapper,
	metricsServer *metrics.Server,
	pprofServer *pprof.Server,
	logger *log.Logger,
	agentConf config.AgentConfig,
) (*Agent, func(), error) {
	zaplog := logger.Log.Desugar()
	httpApp := rt.Router(zaplog)

	// Create agent service
	agentService := service.NewAgentService(agentConf, grpcClient)

	cleanup := func() {
		// stop pprof server
		if pprofServer != nil {
			log.Info("Shutting down pprof server...")
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := pprofServer.Stop(shutdownCtx); err != nil {
				log.Error("Failed to stop pprof server", zap.Error(err))
			}
		}

		// stop metrics server
		if metricsServer != nil {
			log.Info("Shutting down metrics server...")
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := metricsServer.Stop(shutdownCtx); err != nil {
				log.Errorw("Failed to stop metrics server", zap.Error(err))
			}
		}

		// stop global cron scheduler
		cron.Stop()
		log.Info("cron scheduler stopped")
		// close gRPC client connection
		if grpcClient != nil {
			if err := grpcClient.Close(); err != nil {
				log.Errorw("failed to close gRPC client", "error", err)
			}
		}
	}

	app := &Agent{
		HttpApp:       httpApp,
		GrpcClient:    grpcClient,
		MetricsServer: metricsServer,
		PprofServer:   pprofServer,
		Logger:        logger,
		AgentConf:     agentConf,
		AgentService:  agentService,
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
	appConf := app.AgentConf

	// Initialize and start global cron scheduler
	cron.Init(app.Logger)
	cron.Start()
	log.Info("Cron scheduler started.")

	// start metrics server
	if app.MetricsServer != nil {
		if err := app.MetricsServer.Start(); err != nil {
			log.Errorw("Metrics server failed", "error", err)
		}
	}

	// start pprof server
	if app.PprofServer != nil {
		if err := app.PprofServer.Start(); err != nil {
			log.Errorw("Pprof server failed", "error", err)
		}
	}

	// start gRPC client
	if app.GrpcClient != nil {
		go func() {
			if err := app.GrpcClient.Start(grpcclient.ClientConf{
				ServerAddr:           appConf.Grpc.ServerAddr,
				Token:                appConf.Grpc.Token,
				ReadWriteTimeout:     appConf.Grpc.ReadWriteTimeout,
				MaxMsgSize:           appConf.Grpc.MaxMsgSize,
				MaxReconnectAttempts: appConf.Grpc.MaxReconnectAttempts,
			}); err != nil {
				log.Errorw("gRPC client failed", "error", err)
			}
		}()

		// start periodic heartbeat using global cron
		app.startPeriodicHeartbeat()
	}

	// set signal listener (graceful shutdown)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// start HTTP server (async)
	go func() {
		addr := appConf.Http.Host + ":" + fmt.Sprintf("%d", appConf.Http.Port)
		log.Infow("HTTP listener started",
			"address", addr,
		)
		if err := app.HttpApp.Listen(addr); err != nil {
			log.Errorw("HTTP listener failed",
				"address", addr,
				"error", err,
			)
		}
	}()

	// wait for exit signal
	sig := <-quit
	log.Infow("Received signal, shutting down gracefully...", "signal", sig)

	// close components in order
	// close HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	if err := app.HttpApp.ShutdownWithContext(shutdownCtx); err != nil {
		log.Errorw("HTTP server shutdown error", "error", err)
	} else {
		log.Info("HTTP server shut down gracefully")
	}

	// stop global cron scheduler
	cron.Stop()

	// close plugin manager and other resources
	cleanup()

	log.Info("Server shutdown complete")
}

// startPeriodicHeartbeat starts a cron job that periodically calls Heartbeat to keep connection alive
func (app *Agent) startPeriodicHeartbeat() {
	appConf := app.AgentConf

	if app.AgentService == nil {
		log.Warn("AgentService is nil, skipping heartbeat setup")
		return
	}

	// Get heartbeat interval from config (default to 60 seconds)
	interval := appConf.Agent.Interval
	if interval <= 0 {
		interval = 60
	}

	// Create cron spec: @every {interval}s
	spec := fmt.Sprintf("@every %ds", interval)

	// Add heartbeat job to global cron
	err := cron.AddFunc(spec, func() {
		// Build heartbeat request
		req := &agentv1.HeartbeatRequest{
			AgentId:           appConf.Agent.ID,
			Status:            agentv1.AgentStatus_AGENT_STATUS_ONLINE,
			RunningJobsCount:  0,  // TODO: track running jobs
			MaxConcurrentJobs: 10, // TODO: get from config or runtime
			Metrics:           make(map[string]string),
			Labels:            appConf.Agent.Labels,
		}

		// Call Heartbeat through gRPC client
		if app.GrpcClient != nil && app.GrpcClient.AgentClient != nil {
			ctx, cancel := app.GrpcClient.WithTimeoutAndAuth(context.Background())
			defer cancel()

			resp, err := app.GrpcClient.AgentClient.Heartbeat(ctx, req)
			if err != nil || !resp.Success {
				log.Warnw("Periodic heartbeat failed", "error", err)
				return
			}
			log.Debugw("Heartbeat Message", "message", resp.Message)
		}
	}, "agent-heartbeat")

	if err != nil {
		log.Errorw("Failed to add heartbeat cron job", "error", err)
		return
	}

	log.Infow("Added periodic heartbeat to crond", "interval", spec)

	// Do initial heartbeat immediately
	req := &agentv1.HeartbeatRequest{
		AgentId:           appConf.Agent.ID,
		Status:            agentv1.AgentStatus_AGENT_STATUS_ONLINE,
		RunningJobsCount:  0,  // TODO: track running jobs
		MaxConcurrentJobs: 10, // TODO: get from config or runtime
		Metrics:           make(map[string]string),
		Labels:            appConf.Agent.Labels,
	}

	if app.GrpcClient != nil && app.GrpcClient.AgentClient != nil {
		ctx, cancel := app.GrpcClient.WithTimeoutAndAuth(context.Background())
		defer cancel()

		resp, err := app.GrpcClient.AgentClient.Heartbeat(ctx, req)
		if err != nil || !resp.Success {
			log.Warnw("Initial heartbeat failed", "error", err)
			return
		}
		log.Infow("Initial heartbeat Message", "message", resp.Message)
	}
}
