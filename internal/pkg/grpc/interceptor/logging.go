package interceptor

import (
	"context"
	"time"

	"github.com/go-arcade/arcade/pkg/log"
	"google.golang.org/grpc"
)

// excluded methods that don't need logging
var excludedMethods = map[string]bool{
	// "/api.agent.v1.Agent/Heartbeat":  true,
	// "/api.job.v1.Job/Ping":           true,
	// "/api.stream.v1.Stream/Ping":     true,
	// "/api.pipeline.v1.Pipeline/Ping": true,
	"/agent.v1.AgentService/Heartbeat": true,
}

// logCall logs a gRPC server call with common format
func logCall(callType string, method string, duration time.Duration, err error) {
	if err != nil {
		log.Errorw("gRPC call failed", "type", callType, "error", err)
	} else {
		log.Debugw("gRPC call", "type", callType, "method", method, "duration", duration.String())
	}
}

// logClientCall logs a gRPC client call with common format
func logClientCall(callType string, method string, duration time.Duration, err error) {
	if err != nil {
		log.Errorw("gRPC client call failed", "type", callType, "error", err)
	} else {
		log.Debugw("gRPC client call", "type", callType, "method", method, "duration", duration.String())
	}
}

// LoggingUnaryInterceptor unary server interceptor (can filter heartbeat interface)
func LoggingUnaryInterceptor() grpc.UnaryServerInterceptor {

	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if excludedMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)

		logCall("unary", info.FullMethod, duration, err)

		return resp, err
	}
}

// LoggingStreamInterceptor stream server interceptor (can filter heartbeat interface)
func LoggingStreamInterceptor() grpc.StreamServerInterceptor {

	return func(
		srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		if excludedMethods[info.FullMethod] {
			return handler(srv, ss)
		}

		start := time.Now()
		err := handler(srv, ss)
		duration := time.Since(start)

		logCall("stream", info.FullMethod, duration, err)

		return err
	}
}

// LoggingUnaryClientInterceptor unary client interceptor
func LoggingUnaryClientInterceptor() grpc.UnaryClientInterceptor {

	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		start := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		duration := time.Since(start)

		logClientCall("unary", method, duration, err)

		return err
	}
}

// LoggingStreamClientInterceptor stream client interceptor
func LoggingStreamClientInterceptor() grpc.StreamClientInterceptor {

	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		start := time.Now()
		stream, err := streamer(ctx, desc, cc, method, opts...)
		duration := time.Since(start)

		logClientCall("stream", method, duration, err)

		return stream, err
	}
}
