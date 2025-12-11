package database

import (
	"context"
	"errors"
	"time"

	"github.com/go-arcade/arcade/pkg/log"
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
