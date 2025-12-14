package inject

import (
	"context"
	"time"

	"github.com/go-arcade/arcade/pkg/trace"
	tracecontext "github.com/go-arcade/arcade/pkg/trace/context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// GRPCUnaryCall 对 gRPC 一元调用进行埋点
// ctx: 上下文
// method: gRPC 方法名（完整路径，如 "/package.Service/Method"）
// fn: 执行 gRPC 调用的函数，返回响应和错误
func GRPCUnaryCall[T any](ctx context.Context, method string, fn func(ctx context.Context) (T, error)) (T, error) {
	ctx, span := trace.StartSpan(ctx, method,
		oteltrace.WithSpanKind(oteltrace.SpanKindClient))
	defer span.End()

	// 将包含 span 的 context 设置到 goroutine context，以便日志能获取到 trace 信息
	tracecontext.SetContext(ctx)
	defer tracecontext.ClearContext()

	startTime := time.Now()

	// 添加请求属性
	trace.AddSpanAttributes(span,
		attribute.String("rpc.system", "grpc"),
		attribute.String("rpc.service", method),
		attribute.String("rpc.method", method),
	)

	// 执行 gRPC 调用
	response, err := fn(ctx)

	duration := time.Since(startTime)

	// 添加性能指标
	trace.AddSpanAttributes(span,
		attribute.Int64("rpc.duration_ms", duration.Milliseconds()),
	)

	if err != nil {
		// 尝试从 gRPC 错误中提取状态码
		if s, ok := status.FromError(err); ok {
			trace.AddSpanAttributes(span,
				attribute.Int("rpc.grpc.status_code", int(s.Code())),
				attribute.String("rpc.error.message", s.Message()),
			)
			trace.SetSpanStatus(span, codes.Error, s.Message())
		} else {
			trace.RecordError(span, err)
		}
		return response, err
	}

	trace.SetSpanStatus(span, codes.Ok, "")
	return response, err
}

// GRPCStreamCall 对 gRPC 流式调用进行埋点
// ctx: 上下文
// method: gRPC 方法名（完整路径）
// fn: 执行 gRPC 流式调用的函数，返回流和错误
// 注意：对于流式调用，span 会在流创建时开始，但不会立即结束
// 调用者需要在流关闭时手动结束 span
func GRPCStreamCall[T grpc.ClientStream](ctx context.Context, method string, fn func(ctx context.Context) (T, error)) (T, error) {
	ctx, span := trace.StartSpan(ctx, method,
		oteltrace.WithSpanKind(oteltrace.SpanKindClient))
	// 注意：不在这里 defer span.End()，因为流式调用的 span 应该在流关闭时结束

	// 将包含 span 的 context 设置到 goroutine context，以便日志能获取到 trace 信息
	tracecontext.SetContext(ctx)
	defer tracecontext.ClearContext()

	startTime := time.Now()

	// 添加请求属性
	trace.AddSpanAttributes(span,
		attribute.String("rpc.system", "grpc"),
		attribute.String("rpc.service", method),
		attribute.String("rpc.method", method),
		attribute.String("rpc.type", "stream"),
	)

	// 执行 gRPC 流式调用
	stream, err := fn(ctx)

	if err != nil {
		// 如果创建流失败，立即结束 span
		duration := time.Since(startTime)
		trace.AddSpanAttributes(span,
			attribute.Int64("rpc.duration_ms", duration.Milliseconds()),
		)

		// 尝试从 gRPC 错误中提取状态码
		if s, ok := status.FromError(err); ok {
			trace.AddSpanAttributes(span,
				attribute.Int("rpc.grpc.status_code", int(s.Code())),
				attribute.String("rpc.error.message", s.Message()),
			)
			trace.SetSpanStatus(span, codes.Error, s.Message())
		} else {
			trace.RecordError(span, err)
		}
		span.End()
		return stream, err
	}

	// 对于流式调用，span 会在流关闭时结束
	// 这里先设置初始状态，但保持 span 打开
	// 注意：实际的 span 结束应该在流关闭时处理
	trace.SetSpanStatus(span, codes.Ok, "")
	return stream, nil
}

// GRPCServerRequest 对 gRPC 服务器请求进行埋点
// ctx: 上下文
// method: gRPC 方法名（完整路径）
// fn: 处理 gRPC 请求的函数，返回响应和错误
func GRPCServerRequest[T any](ctx context.Context, method string, fn func(ctx context.Context) (T, error)) (T, error) {
	ctx, span := trace.StartSpan(ctx, method,
		oteltrace.WithSpanKind(oteltrace.SpanKindServer))
	defer span.End()

	// 将包含 span 的 context 设置到 goroutine context，以便日志能获取到 trace 信息
	tracecontext.SetContext(ctx)
	defer tracecontext.ClearContext()

	startTime := time.Now()

	// 添加请求属性
	trace.AddSpanAttributes(span,
		attribute.String("rpc.system", "grpc"),
		attribute.String("rpc.service", method),
		attribute.String("rpc.method", method),
	)

	// 处理请求
	response, err := fn(ctx)

	duration := time.Since(startTime)

	// 添加性能指标
	trace.AddSpanAttributes(span,
		attribute.Int64("rpc.duration_ms", duration.Milliseconds()),
	)

	if err != nil {
		// 尝试从 gRPC 错误中提取状态码
		if s, ok := status.FromError(err); ok {
			trace.AddSpanAttributes(span,
				attribute.Int("rpc.grpc.status_code", int(s.Code())),
				attribute.String("rpc.error.message", s.Message()),
			)
			trace.SetSpanStatus(span, codes.Error, s.Message())
		} else {
			trace.RecordError(span, err)
		}
		return response, err
	}

	trace.SetSpanStatus(span, codes.Ok, "")
	return response, err
}
