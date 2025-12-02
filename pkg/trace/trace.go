package trace

import (
	"context"

	tracectx "github.com/go-arcade/arcade/pkg/trace/context"
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
