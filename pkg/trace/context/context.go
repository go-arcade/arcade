package context

import (
	"context"
	"runtime"
	"sync"

	"github.com/go-arcade/arcade/pkg/num"
	"github.com/timandy/routine"
	"go.opentelemetry.io/otel/trace"
)

const bucketsSize = 128
const armSystem = "arm64"

type (
	contextBucket struct {
		lock sync.RWMutex
		data map[int64]context.Context
	}
	contextBuckets struct {
		buckets [bucketsSize]*contextBucket
	}
)

var goroutineContext contextBuckets

func init() {
	for i := range goroutineContext.buckets {
		goroutineContext.buckets[i] = &contextBucket{
			data: make(map[int64]context.Context),
		}
	}
}

// GetContext .
func GetContext() context.Context {
	if runtime.GOARCH == armSystem {
		return context.Background()
	}
	god := routine.Goid()
	idx := god % bucketsSize
	bucket := goroutineContext.buckets[idx]
	bucket.lock.RLock()
	ctx := bucket.data[num.MustInt64(god)]
	bucket.lock.RUnlock()
	return ctx
}

// SetContext .
func SetContext(ctx context.Context) {
	if runtime.GOARCH == armSystem {
		return
	}
	god := routine.Goid()
	idx := god % bucketsSize
	bucket := goroutineContext.buckets[idx]
	bucket.lock.Lock()
	defer bucket.lock.Unlock()
	bucket.data[num.MustInt64(god)] = ctx
}

// ClearContext .
func ClearContext() {
	if runtime.GOARCH == armSystem {
		return
	}
	god := routine.Goid()
	idx := god % bucketsSize
	bucket := goroutineContext.buckets[idx]
	bucket.lock.Lock()
	defer bucket.lock.Unlock()
	delete(bucket.data, num.MustInt64(god))
}

// RunWithContext .
func RunWithContext(ctx context.Context, fn func(ctx context.Context)) {
	SetContext(ctx)
	defer ClearContext()
	fn(ctx)
}

// ContextWithSpan .
func ContextWithSpan(ctx context.Context) context.Context {
	if span := trace.SpanFromContext(ctx); !span.SpanContext().IsValid() {
		pct := GetContext()
		if pct != nil {
			if span := trace.SpanFromContext(pct); span.SpanContext().IsValid() {
				ctx = trace.ContextWithSpan(ctx, span)
			}
		}
	}
	return ctx
}
