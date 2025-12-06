package database

import (
	"context"
	"errors"
	"time"

	"github.com/go-arcade/arcade/pkg/log"
	"go.uber.org/zap"
	gormLogger "gorm.io/gorm/logger"
)

type GormLoggerAdapter struct {
	Config gormLogger.Config
	Level  gormLogger.LogLevel
	log    *zap.Logger
}

func NewGormLoggerAdapter(config gormLogger.Config, logLevel gormLogger.LogLevel, logger *log.Logger) *GormLoggerAdapter {
	return &GormLoggerAdapter{
		Config: config,
		Level:  logLevel,
		log:    logger.Log.Desugar().WithOptions(zap.AddCallerSkip(2)), // skip 2 frames to get the caller
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
	l.log.Sugar().Infow(msg, data...)
}

func (l *GormLoggerAdapter) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.Level < gormLogger.Warn {
		return
	}
	l.log.Sugar().Warnw(msg, data...)
}

func (l *GormLoggerAdapter) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.Level < gormLogger.Error {
		return
	}
	l.log.Sugar().Errorw(msg, data...)
}

func (l *GormLoggerAdapter) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.Level <= gormLogger.Silent {
		return
	}

	elapsed := time.Since(begin).Seconds() // convert to seconds
	sql, rows := fc()

	if err != nil && l.Config.LogLevel >= gormLogger.Error && (!errors.Is(err, gormLogger.ErrRecordNotFound) || !l.Config.IgnoreRecordNotFoundError) {
		l.log.With().Sugar().Errorw("`%s` [rows: %d, elapsed: %.5f], err: %v", sql, rows, elapsed, err)
		return
	}

	if elapsed > l.Config.SlowThreshold.Seconds() && l.Config.SlowThreshold.Seconds() != 0 && l.Config.LogLevel >= gormLogger.Warn {
		l.log.With().Sugar().Warnw("`%s` [rows: %d, elapsed: %.5f]", sql, rows, elapsed)
		return
	}

	if l.Config.LogLevel == gormLogger.Info {
		l.log.With().Sugar().Debugw("SQL query", "sql", sql, "rows", rows, "elapsed", elapsed)
	}
}
