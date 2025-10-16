package grpc

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/observabil/arcade/pkg/log"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	agentapi "github.com/observabil/arcade/api/agent/v1"
	pipelineapi "github.com/observabil/arcade/api/pipeline/v1"
	streamapi "github.com/observabil/arcade/api/stream/v1"
	taskapi "github.com/observabil/arcade/api/task/v1"
	"github.com/observabil/arcade/internal/engine/service/agent"
	"github.com/observabil/arcade/internal/engine/service/pipeline"
	"github.com/observabil/arcade/internal/engine/service/stream"
	"github.com/observabil/arcade/internal/engine/service/task"
	"github.com/observabil/arcade/internal/pkg/grpc/middleware"
)

// gRPC 配置
type GrpcConf struct {
	Host             string
	Port             int
	MaxConnections   int
	ReadWriteTimeout int
}

type ServerWrapper struct {
	svr *grpc.Server
}

// NewGrpcServer 创建 gRPC 服务
func NewGrpcServer(cfg GrpcConf, log *zap.Logger) *ServerWrapper {
	opts := []grpc.ServerOption{
		grpc.MaxConcurrentStreams(uint32(cfg.MaxConnections)),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			// 注意顺序，先 tags，再 logging，再 auth，最后 recovery
			grpc_ctxtags.StreamServerInterceptor(),
			middleware.LoggingStreamInterceptor(log), // 使用自定义日志拦截器，可过滤心跳接口
			middleware.AuthStreamInterceptor(),       // 使用自定义认证拦截器，可跳过心跳接口
			grpc_recovery.StreamServerInterceptor(),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_ctxtags.UnaryServerInterceptor(),
			middleware.LoggingUnaryInterceptor(log), // 使用自定义日志拦截器，可过滤心跳接口
			middleware.AuthUnaryInterceptor(),       // 使用自定义认证拦截器，可跳过心跳接口
			grpc_recovery.UnaryServerInterceptor(),
		)),
	}

	s := grpc.NewServer(opts...)
	return &ServerWrapper{svr: s}
}

// RegisterAll 注册所有 gRPC 服务
func (s *ServerWrapper) Register() {
	agentapi.RegisterAgentServer(s.svr, &agent.AgentServiceImpl{})
	taskapi.RegisterTaskServer(s.svr, &task.TaskServiceImpl{})
	streamapi.RegisterStreamServer(s.svr, &stream.StreamServiceImpl{})
	pipelineapi.RegisterPipelineServer(s.svr, &pipeline.PipelineServiceImpl{})
	// reflection（调试）
	reflection.Register(s.svr)
}

// Start 启动服务
func (s *ServerWrapper) Start(cfg GrpcConf) error {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	go func() {
		log.Infof("gRPC server started at %s", addr)
		if err := s.svr.Serve(lis); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	// 优雅退出
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	log.Info("Shutting down gRPC server...")
	s.svr.GracefulStop()
	return nil
}

// Stop 手动停止
func (s *ServerWrapper) Stop() {
	s.svr.GracefulStop()
}
