package grpc

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-arcade/arcade/internal/engine/service/agent"
	"github.com/go-arcade/arcade/internal/engine/service/pipeline"
	"github.com/go-arcade/arcade/internal/engine/service/stream"
	"github.com/go-arcade/arcade/internal/engine/service/task"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/go-arcade/arcade/pkg/log"

	agentv1 "github.com/go-arcade/arcade/api/agent/v1"
	pipelinev1 "github.com/go-arcade/arcade/api/pipeline/v1"
	streamv1 "github.com/go-arcade/arcade/api/stream/v1"
	taskv1 "github.com/go-arcade/arcade/api/task/v1"
	"github.com/go-arcade/arcade/internal/pkg/grpc/middleware"
	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcrecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpcctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
)

// Conf 配置
type Conf struct {
	Host             string
	Port             int
	MaxConnections   int
	ReadWriteTimeout int
}

type ServerWrapper struct {
	svr *grpc.Server
}

// NewGrpcServer 创建 gRPC 服务
func NewGrpcServer(cfg Conf, log *zap.Logger) *ServerWrapper {
	opts := []grpc.ServerOption{
		grpc.MaxConcurrentStreams(uint32(cfg.MaxConnections)),
		grpc.StreamInterceptor(grpcmiddleware.ChainStreamServer(
			// 注意顺序，先 tags，再 logging，再 auth，最后 recovery
			grpcctxtags.StreamServerInterceptor(),
			middleware.LoggingStreamInterceptor(log), // 使用自定义日志拦截器，可过滤心跳接口
			middleware.AuthStreamInterceptor(),       // 使用自定义认证拦截器，可跳过心跳接口
			grpcrecovery.StreamServerInterceptor(),
		)),
		grpc.UnaryInterceptor(grpcmiddleware.ChainUnaryServer(
			grpcctxtags.UnaryServerInterceptor(),
			middleware.LoggingUnaryInterceptor(log), // 使用自定义日志拦截器，可过滤心跳接口
			middleware.AuthUnaryInterceptor(),       // 使用自定义认证拦截器，可跳过心跳接口
			grpcrecovery.UnaryServerInterceptor(),
		)),
	}

	s := grpc.NewServer(opts...)
	return &ServerWrapper{svr: s}
}

// Register 注册所有 gRPC 服务
func (s *ServerWrapper) Register() {
	agentv1.RegisterAgentServiceServer(s.svr, &agent.AgentServiceImpl{})
	taskv1.RegisterTaskServiceServer(s.svr, &task.TaskServiceImpl{})
	streamv1.RegisterStreamServiceServer(s.svr, &stream.StreamServiceImpl{})
	pipelinev1.RegisterPipelineServiceServer(s.svr, &pipeline.PipelineServiceImpl{})
	// reflection（调试）
	reflection.Register(s.svr)
}

// Start 启动服务
func (s *ServerWrapper) Start(cfg Conf) error {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	go func() {
		log.Infow("gRPC listener started",
			"address", addr,
		)
		if err := s.svr.Serve(lis); err != nil {
			log.Errorw("gRPC listener failed",
				"address", addr,
				"error", err,
			)
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
