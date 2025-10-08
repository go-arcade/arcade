package gprc

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cloudwego/kitex/pkg/limit"
	"github.com/cloudwego/kitex/pkg/registry"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/transmeta"
	"github.com/cloudwego/kitex/pkg/utils"
	"github.com/cloudwego/kitex/server"
)

type ServerWrapper struct {
	svr server.Server
}

func NewGrpcServer[T any](handler T, newServer func(T, ...server.Option) server.Server,
	address, service string, reg registry.Registry) *ServerWrapper {
	
		opts := []server.Option{
		server.WithServiceAddr(utils.NewNetAddr("tcp", address)),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
			ServiceName: service,
		}),
		server.WithMetaHandler(transmeta.ServerHTTP2Handler),
		server.WithReadWriteTimeout(3 * time.Second),
		server.WithLimit(&limit.Option{
			MaxConnections: 1000,
			MaxQPS:         100,
		}),
	}

	if reg != nil {
		opts = append(opts, server.WithRegistry(reg))
	}
	return &ServerWrapper{
		svr: newServer(handler, opts...),
	}
}

func (s *ServerWrapper) Start() error {
	go func() {
		if err := s.svr.Run(); err != nil {
			log.Fatalf("server run error: %v", err)
		}
	}()

	// 等待优雅退出信号
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	log.Println("Shutting down server...")
	return s.svr.Stop()
}
