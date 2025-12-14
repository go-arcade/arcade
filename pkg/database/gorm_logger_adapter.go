package database

import (
	"context"
	"errors"
	"time"

	"github.com/go-arcade/arcade/pkg/log"
	tracecontext "github.com/go-arcade/arcade/pkg/trace/context"
	"github.com/go-arcade/arcade/pkg/trace/inject"
	gormLogger "gorm.io/gorm/logger"
)

type GormLoggerAdapter struct {
	Config gormLogger.Config
	Level  gormLogger.LogLevel
}

func NewGormLoggerAdapter(config gormLogger.Config, logLevel gormLogger.LogLevel) *GormLoggerAdapter {
	return &GormLoggerAdapter{
		Config: config,
		Level:  logLevel,
	}
}

func (l *GormLoggerAdapter) LogMode(level gormLogger.LogLevel) gormLogger.Interface {
	l.Level = level
	return l
}

func (l *GormLoggerAdapter) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.Level < gormLogger.Info {
		return
	}
	log.Infow(msg, data...)
}

func (l *GormLoggerAdapter) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.Level < gormLogger.Warn {
		return
	}
	log.Warnw(msg, data...)
}

func (l *GormLoggerAdapter) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.Level < gormLogger.Error {
		return
	}
	log.Errorw(msg, data...)
}

func (l *GormLoggerAdapter) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.Level <= gormLogger.Silent {
		return
	}

	elapsed := time.Since(begin).Seconds() // convert to seconds
	sql, rows := fc()

	// 将 context 设置到 goroutine context，以便日志能获取到 trace 信息
	tracecontext.SetContext(ctx)

	// 添加数据库查询埋点（会创建新的 span 并更新 goroutine context）
	// DatabaseQuery 会设置包含新 span 的 context 到 goroutine context
	_, _ = inject.DatabaseQueryWithSQL(ctx, "mysql", sql, func(ctx context.Context) (int64, error) {
		// 这里只是用于埋点，实际的查询已经在 GORM 中执行完成
		// 返回影响的行数和错误
		return rows, err
	})

	// 记录日志时，DatabaseQuery 创建的 span context 应该还在 goroutine context 中
	// 在记录完所有日志后再清除 context
	defer tracecontext.ClearContext()

	if err != nil && l.Config.LogLevel >= gormLogger.Error && (!errors.Is(err, gormLogger.ErrRecordNotFound) || !l.Config.IgnoreRecordNotFoundError) {
		log.Errorw("SQL query failed", "sql", sql, "rows", rows, "elapsed", elapsed, "error", err)
		return
	}

	if elapsed > l.Config.SlowThreshold.Seconds() && l.Config.SlowThreshold.Seconds() != 0 && l.Config.LogLevel >= gormLogger.Warn {
		log.Warnw("Slow SQL query", "sql", sql, "rows", rows, "elapsed", elapsed)
		return
	}

	if l.Config.LogLevel == gormLogger.Info {
		log.Debugw("SQL query", "sql", sql, "rows", rows, "elapsed", elapsed)
	}
}
