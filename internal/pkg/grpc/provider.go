package grpc

import (
	"github.com/go-arcade/arcade/internal/engine/service"
	"github.com/google/wire"
)

// ProviderSet 提供 gRPC 服务层相关的依赖（主程序使用）
var ProviderSet = wire.NewSet(
	ProvideGrpcServer,
	ProvideGrpcClient,
)

// ProvideGrpcServer 提供 gRPC 服务器实例
func ProvideGrpcServer(cfg *Conf, services *service.Services) *ServerWrapper {
	server := NewGrpcServer(*cfg)
	server.Register(services)
	return server
}

// ProvideGrpcClient 提供 gRPC 客户端实例
func ProvideGrpcClient(cfg *ClientConf) (*ClientWrapper, error) {
	return NewGrpcClient(*cfg)
}
