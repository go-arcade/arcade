package grpc

import (
	"github.com/google/wire"
	"go.uber.org/zap"
)

// ProviderSet 提供 gRPC 服务层相关的依赖
var ProviderSet = wire.NewSet(
	ProvideGrpcServer,
)

// ProvideGrpcServer 提供 gRPC 服务器实例
func ProvideGrpcServer(cfg *Conf, logger *zap.Logger) *ServerWrapper {
	server := NewGrpcServer(*cfg, logger)
	server.Register()
	return server
}
