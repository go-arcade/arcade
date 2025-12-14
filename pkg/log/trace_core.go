package log

import (
	tracectx "github.com/go-arcade/arcade/pkg/trace/context"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// traceCore 是一个 zap Core wrapper，自动在日志中添加 trace 信息
type traceCore struct {
	zapcore.Core
}

// With 返回一个新的 traceCore，添加字段
func (c *traceCore) With(fields []zapcore.Field) zapcore.Core {
	return &traceCore{
		Core: c.Core.With(fields),
	}
}

// Write 在写入日志时自动添加 trace 信息
func (c *traceCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	// 从 goroutine context 中获取 trace 信息
	ctx := tracectx.GetContext()
	if ctx == nil {
		// 如果没有 context，直接写入日志
		return c.Core.Write(entry, fields)
	}

	span := trace.SpanFromContext(ctx)
	if span == nil {
		// 如果 context 中没有 span，直接写入日志
		return c.Core.Write(entry, fields)
	}

	spanCtx := span.SpanContext()
	if !spanCtx.IsValid() {
		// 如果 span context 无效，直接写入日志
		return c.Core.Write(entry, fields)
	}

	traceID := spanCtx.TraceID()
	spanID := spanCtx.SpanID()

	// 只有当 trace ID 和 span ID 都有效时才添加
	// 注意：即使 span 被 End() 了，SpanContext() 仍然有效，可以获取 trace ID 和 span ID
	if traceID.IsValid() && spanID.IsValid() {
		// 添加 trace ID 和 span ID
		traceFields := []zapcore.Field{
			zap.String("trace_id", traceID.String()),
			zap.String("span_id", spanID.String()),
		}

		// 如果有 trace flags，也可以添加
		if spanCtx.TraceFlags() != 0 {
			traceFields = append(traceFields, zap.Uint8("trace_flags", uint8(spanCtx.TraceFlags())))
		}

		// 将 trace 字段添加到现有字段前面
		fields = append(traceFields, fields...)
	}

	// 调用底层的 Core.Write
	return c.Core.Write(entry, fields)
}

// Check 检查是否应该记录该级别的日志
func (c *traceCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	return c.Core.Check(ent, ce)
}

// Enabled 检查是否启用某个级别
func (c *traceCore) Enabled(level zapcore.Level) bool {
	return c.Core.Enabled(level)
}

// Sync 同步日志
func (c *traceCore) Sync() error {
	return c.Core.Sync()
}

// wrapCoreWithTrace 包装一个 Core，添加 trace 信息支持
func wrapCoreWithTrace(core zapcore.Core) zapcore.Core {
	return &traceCore{Core: core}
}
