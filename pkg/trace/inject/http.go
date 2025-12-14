package inject

import (
	"context"
	"time"

	"github.com/go-arcade/arcade/pkg/trace"
	tracecontext "github.com/go-arcade/arcade/pkg/trace/context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// HTTPRequest 对 HTTP 请求进行埋点
// ctx: 上下文
// method: HTTP 方法
// url: 请求 URL
// fn: 执行请求的函数，返回响应状态码、响应大小和错误
func HTTPRequest(ctx context.Context, method, url string, fn func(ctx context.Context) (statusCode int, responseSize int64, err error)) (int, int64, error) {
	ctx, span := trace.StartSpan(ctx, "http.request",
		oteltrace.WithSpanKind(oteltrace.SpanKindClient))
	defer span.End()

	// 将包含 span 的 context 设置到 goroutine context，以便日志能获取到 trace 信息
	tracecontext.SetContext(ctx)
	defer tracecontext.ClearContext()

	startTime := time.Now()

	// 添加请求属性
	trace.AddSpanAttributes(span,
		attribute.String("http.method", method),
		attribute.String("http.url", url),
	)

	// 执行请求
	statusCode, responseSize, err := fn(ctx)

	duration := time.Since(startTime)

	// 添加响应属性
	trace.AddSpanAttributes(span,
		attribute.Int("http.status_code", statusCode),
		attribute.Int64("http.response.size", responseSize),
		attribute.Int64("http.duration_ms", duration.Milliseconds()),
	)

	if err != nil {
		trace.RecordError(span, err)
		return statusCode, responseSize, err
	}

	if statusCode >= 400 {
		trace.SetSpanStatus(span, codes.Error, "")
	} else {
		trace.SetSpanStatus(span, codes.Ok, "")
	}

	return statusCode, responseSize, err
}

// HTTPServerRequest 对 HTTP 服务器请求进行埋点
// ctx: 上下文
// method: HTTP 方法
// path: 请求路径
// fn: 处理请求的函数，返回状态码和错误
func HTTPServerRequest(ctx context.Context, method, path string, fn func(ctx context.Context) (statusCode int, err error)) (int, error) {
	ctx, span := trace.StartSpan(ctx, "http.server.request",
		oteltrace.WithSpanKind(oteltrace.SpanKindServer))
	defer span.End()

	// 将包含 span 的 context 设置到 goroutine context，以便日志能获取到 trace 信息
	// 注意：不在这里清除 context，让调用者管理 context 生命周期
	tracecontext.SetContext(ctx)

	startTime := time.Now()

	// 添加请求属性
	trace.AddSpanAttributes(span,
		attribute.String("http.method", method),
		attribute.String("http.route", path),
		attribute.String("http.scheme", "http"), // 可以根据实际情况设置
	)

	// 处理请求
	statusCode, err := fn(ctx)

	duration := time.Since(startTime)

	// 添加响应属性
	trace.AddSpanAttributes(span,
		attribute.Int("http.status_code", statusCode),
		attribute.Int64("http.duration_ms", duration.Milliseconds()),
	)

	if err != nil {
		trace.RecordError(span, err)
	} else {
		if statusCode >= 400 {
			trace.SetSpanStatus(span, codes.Error, "")
		} else {
			trace.SetSpanStatus(span, codes.Ok, "")
		}
	}

	// 在返回前，确保 context 仍然在 goroutine context 中
	// 这样 AccessLogFormat 中间件就能获取到 trace 信息
	// 注意：即使 span.End() 被调用，context 中的 span 仍然可以用于获取 trace ID 和 span ID
	tracecontext.SetContext(ctx)

	return statusCode, err
}
