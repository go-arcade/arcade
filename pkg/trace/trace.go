package trace

import (
	"context"

	tracectx "github.com/go-arcade/arcade/pkg/trace/context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Go .
func Go(ctx context.Context, fn func()) {
	GoWithContext(ctx, func(ctx context.Context) {
		fn()
	})
}

// GoWithContext .
func GoWithContext(ctx context.Context, fn func(ctx context.Context)) {
	if ctx == nil {
		ctx = context.Background()
	}
	if span := trace.SpanFromContext(ctx); !span.SpanContext().IsValid() {
		pct := tracectx.GetContext()
		if pct != nil {
			if span := trace.SpanFromContext(pct); span.SpanContext().IsValid() {
				ctx = trace.ContextWithSpan(ctx, span)
			}
		}
	}
	go tracectx.RunWithContext(ctx, fn)
}

// ContextWithSpan .
func ContextWithSpan(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if span := trace.SpanFromContext(ctx); !span.SpanContext().IsValid() {
		pct := tracectx.GetContext()
		if pct != nil {
			if span := trace.SpanFromContext(pct); span.SpanContext().IsValid() {
				ctx = trace.ContextWithSpan(ctx, span)
			}
		}
	}
	return ctx
}

// RunWithTrace .
func RunWithTrace(ctx context.Context, fn func(ctx context.Context)) {
	tracectx.RunWithContext(ctx, fn)
}

// StartSpan 手动创建一个新的 span
// name: span 的名称
// opts: 可选的 span 选项，如 trace.WithSpanKind() 等
// 返回新的 context 和 span
func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	ctx = ContextWithSpan(ctx)
	tracer := otel.Tracer("github.com/go-arcade/arcade/pkg/trace")
	return tracer.Start(ctx, name, opts...)
}

// EndSpan 结束 span，如果发生错误则记录错误状态
func EndSpan(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}
	span.End()
}

// AddSpanAttributes 向 span 添加属性
func AddSpanAttributes(span trace.Span, attrs ...attribute.KeyValue) {
	span.SetAttributes(attrs...)
}

// RecordError 记录错误到 span
func RecordError(span trace.Span, err error) {
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

// SetSpanStatus 设置 span 状态
func SetSpanStatus(span trace.Span, code codes.Code, description string) {
	span.SetStatus(code, description)
}
