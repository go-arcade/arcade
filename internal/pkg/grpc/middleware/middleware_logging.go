package middleware

import (
	"context"
	"path"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// 需要排除的 gRPC 方法（不记录日志）
var excludedMethods = map[string]bool{
	// "/api.agent.v1.Agent/Heartbeat":  true,
	// "/api.job.v1.Job/Ping":           true,
	// "/api.stream.v1.Stream/Ping":     true,
	// "/api.pipeline.v1.Pipeline/Ping": true,
}

// LoggingUnaryInterceptor 一元调用日志拦截器（过滤心跳接口）
func LoggingUnaryInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		// 检查是否需要跳过日志
		if excludedMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)

		// 记录日志
		service := path.Dir(info.FullMethod)[1:]
		method := path.Base(info.FullMethod)

		if err != nil {
			logger.Error("gRPC call failed",
				zap.String("service", service),
				zap.String("method", method),
				zap.Duration("duration", duration),
				zap.Error(err),
			)
		} else {
			logger.Info("gRPC call",
				zap.String("service", service),
				zap.String("method", method),
				zap.Duration("duration", duration),
			)
		}

		return resp, err
	}
}

// LoggingStreamInterceptor 流式调用日志拦截器（过滤心跳接口）
func LoggingStreamInterceptor(logger *zap.Logger) grpc.StreamServerInterceptor {
	return func(
		srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		// 检查是否需要跳过日志
		if excludedMethods[info.FullMethod] {
			return handler(srv, ss)
		}

		start := time.Now()
		err := handler(srv, ss)
		duration := time.Since(start)

		// 记录日志
		service := path.Dir(info.FullMethod)[1:]
		method := path.Base(info.FullMethod)

		if err != nil {
			logger.Error("gRPC stream call failed",
				zap.String("service", service),
				zap.String("method", method),
				zap.Duration("duration", duration),
				zap.Error(err),
			)
		} else {
			logger.Info("gRPC stream call",
				zap.String("service", service),
				zap.String("method", method),
				zap.Duration("duration", duration),
			)
		}

		return err
	}
}
