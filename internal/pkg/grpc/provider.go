package grpc

import (
	"github.com/google/wire"
)

// ProviderSet 提供 gRPC 相关依赖
var ProviderSet = wire.NewSet(ProvideGrpcServer)

// ProvideGrpcServer 提供 gRPC 服务实例
func ProvideGrpcServer(cfg *GrpcConf) *ServerWrapper {
	server := NewGrpcServer(*cfg)
	server.Register()
	return server
}
