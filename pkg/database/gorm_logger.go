package database

import (
	"context"
	"errors"
	"go.uber.org/zap"
	"gorm.io/gorm/logger"
	"time"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/16 13:54
 * @file: gorm_logger.go
 * @description: gorm log
 */

type GormLogger struct {
	Config logger.Config
	Level  logger.LogLevel
	log    *zap.Logger
}

func NewGormLogger(config logger.Config, logLevel logger.LogLevel, zapLogger *zap.Logger) *GormLogger {
	return &GormLogger{
		Config: config,
		Level:  logLevel,
		log:    zapLogger.WithOptions(zap.AddCallerSkip(2)), // 调整调用栈深度
	}
}

func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	l.Level = level
	return l
}

func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.Level < logger.Info {
		return
	}
	l.log.Sugar().Infof(msg, data...)
}

func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.Level < logger.Warn {
		return
	}
	l.log.Sugar().Warnf(msg, data...)
}

func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.Level < logger.Error {
		return
	}
	l.log.Sugar().Errorf(msg, data...)
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.Level <= logger.Silent {
		return
	}

	elapsed := time.Since(begin).Seconds() // 转换为秒
	sql, rows := fc()

	if err != nil && l.Config.LogLevel >= logger.Error && (!errors.Is(err, logger.ErrRecordNotFound) || !l.Config.IgnoreRecordNotFoundError) {
		l.log.With().Sugar().Errorf("`%s` [rows: %d, elapsed: %.5f], err: %v", sql, rows, elapsed, err)
		return
	}

	if elapsed > l.Config.SlowThreshold.Seconds() && l.Config.SlowThreshold.Seconds() != 0 && l.Config.LogLevel >= logger.Warn {
		l.log.With().Sugar().Warnf("`%s` [rows: %d, elapsed: %.5f]", sql, rows, elapsed)
		return
	}

	if l.Config.LogLevel == logger.Info {
		l.log.With().Sugar().Debugf("`%s` [rows: %d, elapsed: %.5f]", sql, rows, elapsed)
	}
}
