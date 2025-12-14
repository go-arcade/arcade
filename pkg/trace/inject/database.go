package inject

import (
	"context"
	"strings"
	"time"

	"github.com/go-arcade/arcade/pkg/trace"
	tracecontext "github.com/go-arcade/arcade/pkg/trace/context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// DatabaseQuery 对数据库查询进行埋点
// ctx: 上下文
// system: 数据库系统类型，如 "mysql", "postgresql", "mongodb"
// operation: 操作类型，如 "SELECT", "INSERT", "UPDATE", "DELETE"
// table: 表名（可选）
// sql: SQL 语句（可选，敏感信息应脱敏）
// fn: 执行查询的函数，返回影响行数和错误
func DatabaseQuery(ctx context.Context, system, operation, table, sql string, fn func(ctx context.Context) (rowsAffected int64, err error)) (int64, error) {
	ctx, span := trace.StartSpan(ctx, "db.query.execute",
		oteltrace.WithSpanKind(oteltrace.SpanKindClient))
	defer span.End()

	// 将包含 span 的 context 设置到 goroutine context，以便日志能获取到 trace 信息
	// 注意：不在这里清除 context，让调用者管理 context 生命周期
	tracecontext.SetContext(ctx)

	startTime := time.Now()

	attrs := []attribute.KeyValue{
		attribute.String("db.system", system),
		attribute.String("db.operation", operation),
	}

	if table != "" {
		attrs = append(attrs, attribute.String("db.sql.table", table))
	}

	if sql != "" {
		attrs = append(attrs, attribute.String("db.statement", sql))
	}

	trace.AddSpanAttributes(span, attrs...)

	// 执行查询
	rowsAffected, err := fn(ctx)

	duration := time.Since(startTime)

	// 添加性能指标
	trace.AddSpanAttributes(span,
		attribute.Int64("db.duration_ms", duration.Milliseconds()),
		attribute.Int64("db.rows.affected", rowsAffected),
	)

	if err != nil {
		trace.RecordError(span, err)
		return rowsAffected, err
	}

	trace.SetSpanStatus(span, codes.Ok, "")
	return rowsAffected, err
}

// DatabaseQueryWithSQL 对数据库查询进行埋点（带 SQL 语句）
// 这是 DatabaseQuery 的便捷方法，自动解析 SQL 语句提取操作类型和表名
func DatabaseQueryWithSQL(ctx context.Context, system, sql string, fn func(ctx context.Context) (rowsAffected int64, err error)) (int64, error) {
	operation, table := parseSQL(sql)
	return DatabaseQuery(ctx, system, operation, table, sql, fn)
}

// parseSQL 简单解析 SQL 语句，提取操作类型和表名
func parseSQL(sql string) (operation, table string) {
	if len(sql) == 0 {
		return "UNKNOWN", ""
	}

	// 简单的 SQL 解析，提取第一个关键字作为操作类型
	sql = strings.ToUpper(sql)
	if len(sql) > 0 {
		// 找到第一个非空白字符
		start := 0
		for start < len(sql) && (sql[start] == ' ' || sql[start] == '\t' || sql[start] == '\n') {
			start++
		}
		if start < len(sql) {
			// 提取操作类型（如 SELECT, INSERT, UPDATE, DELETE）
			end := start
			for end < len(sql) && sql[end] != ' ' && sql[end] != '\t' && sql[end] != '\n' {
				end++
			}
			operation = sql[start:end]
		}
	}

	// 尝试提取表名（简单实现，可能不准确）
	// 这里可以根据需要实现更复杂的解析逻辑
	table = ""

	return operation, table
}
