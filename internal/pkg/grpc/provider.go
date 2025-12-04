package grpc

import (
	"github.com/go-arcade/arcade/pkg/log"
	"github.com/google/wire"
)

// ProviderSet 提供 gRPC 服务层相关的依赖
var ProviderSet = wire.NewSet(
	ProvideGrpcServer,
)

// ProvideGrpcServer 提供 gRPC 服务器实例
func ProvideGrpcServer(cfg *Conf, logger *log.Logger) *ServerWrapper {
	zapLogger := logger.Log.Desugar()
	server := NewGrpcServer(*cfg, zapLogger)
	server.Register()
	return server
}
