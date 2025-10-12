package grpc

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/observabil/arcade/pkg/log"

	agentapi "github.com/observabil/arcade/api/agent/v1"
	jobapi "github.com/observabil/arcade/api/job/v1"
	streamapi "github.com/observabil/arcade/api/stream/v1"
	"github.com/observabil/arcade/internal/engine/service/agent"
	"github.com/observabil/arcade/internal/engine/service/job"
	"github.com/observabil/arcade/internal/engine/service/stream"
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
func NewGrpcServer(cfg GrpcConf) *ServerWrapper {
	opts := []grpc.ServerOption{
		grpc.MaxConcurrentStreams(uint32(cfg.MaxConnections)),
	}

	s := grpc.NewServer(opts...)
	return &ServerWrapper{svr: s}
}

// RegisterAll 注册所有 gRPC 服务
func (s *ServerWrapper) Register() {
	agentapi.RegisterAgentServer(s.svr, &agent.AgentServiceImpl{})
	jobapi.RegisterJobServer(s.svr, &job.JobServiceImpl{})
	streamapi.RegisterStreamServer(s.svr, &stream.StreamServiceImpl{})
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
