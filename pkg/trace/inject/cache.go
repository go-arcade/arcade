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

// CacheOperation 对缓存操作进行埋点
// ctx: 上下文
// operation: 操作类型，如 "GET", "SET", "DEL", "EXISTS"
// key: 缓存键
// fn: 执行缓存操作的函数，返回错误
func CacheOperation(ctx context.Context, operation, key string, fn func(ctx context.Context) error) error {
	ctx, span := trace.StartSpan(ctx, "cache.operation",
		oteltrace.WithSpanKind(oteltrace.SpanKindClient))
	defer span.End()

	// 将包含 span 的 context 设置到 goroutine context，以便日志能获取到 trace 信息
	tracecontext.SetContext(ctx)
	defer tracecontext.ClearContext()

	startTime := time.Now()

	// 添加操作属性
	trace.AddSpanAttributes(span,
		attribute.String("cache.operation", operation),
		attribute.String("cache.key", key),
	)

	// 执行缓存操作
	err := fn(ctx)

	duration := time.Since(startTime)

	// 添加性能指标
	trace.AddSpanAttributes(span,
		attribute.Int64("cache.duration_ms", duration.Milliseconds()),
	)

	if err != nil {
		trace.RecordError(span, err)
		return err
	}

	trace.SetSpanStatus(span, codes.Ok, "")
	return nil
}

// CacheGet 对缓存 GET 操作进行埋点
func CacheGet(ctx context.Context, key string, fn func(ctx context.Context) (found bool, err error)) (bool, error) {
	ctx, span := trace.StartSpan(ctx, "cache.get",
		oteltrace.WithSpanKind(oteltrace.SpanKindClient))
	defer span.End()

	// 将包含 span 的 context 设置到 goroutine context，以便日志能获取到 trace 信息
	tracecontext.SetContext(ctx)
	defer tracecontext.ClearContext()

	startTime := time.Now()

	trace.AddSpanAttributes(span,
		attribute.String("cache.operation", "GET"),
		attribute.String("cache.key", key),
	)

	found, err := fn(ctx)

	duration := time.Since(startTime)

	trace.AddSpanAttributes(span,
		attribute.Bool("cache.hit", found),
		attribute.Int64("cache.duration_ms", duration.Milliseconds()),
	)

	if err != nil {
		trace.RecordError(span, err)
		return found, err
	}

	trace.SetSpanStatus(span, codes.Ok, "")
	return found, err
}

// CacheSet 对缓存 SET 操作进行埋点
func CacheSet(ctx context.Context, key string, expiration time.Duration, fn func(ctx context.Context) error) error {
	ctx, span := trace.StartSpan(ctx, "cache.set",
		oteltrace.WithSpanKind(oteltrace.SpanKindClient))
	defer span.End()

	// 将包含 span 的 context 设置到 goroutine context，以便日志能获取到 trace 信息
	tracecontext.SetContext(ctx)
	defer tracecontext.ClearContext()

	startTime := time.Now()

	trace.AddSpanAttributes(span,
		attribute.String("cache.operation", "SET"),
		attribute.String("cache.key", key),
		attribute.Int64("cache.expiration_ms", expiration.Milliseconds()),
	)

	err := fn(ctx)

	duration := time.Since(startTime)

	trace.AddSpanAttributes(span,
		attribute.Int64("cache.duration_ms", duration.Milliseconds()),
	)

	if err != nil {
		trace.RecordError(span, err)
		return err
	}

	trace.SetSpanStatus(span, codes.Ok, "")
	return nil
}

// CacheDel 对缓存 DEL 操作进行埋点
func CacheDel(ctx context.Context, keys []string, fn func(ctx context.Context) (deleted int64, err error)) (int64, error) {
	ctx, span := trace.StartSpan(ctx, "cache.del",
		oteltrace.WithSpanKind(oteltrace.SpanKindClient))
	defer span.End()

	startTime := time.Now()

	trace.AddSpanAttributes(span,
		attribute.String("cache.operation", "DEL"),
		attribute.Int("cache.keys.count", len(keys)),
	)

	deleted, err := fn(ctx)

	duration := time.Since(startTime)

	trace.AddSpanAttributes(span,
		attribute.Int64("cache.deleted", deleted),
		attribute.Int64("cache.duration_ms", duration.Milliseconds()),
	)

	if err != nil {
		trace.RecordError(span, err)
		return deleted, err
	}

	trace.SetSpanStatus(span, codes.Ok, "")
	return deleted, err
}
