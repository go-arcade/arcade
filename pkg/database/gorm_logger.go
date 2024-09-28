package database

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-arcade/arcade/pkg/runner"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
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
	log    logger.Writer
	//Log    *zap.SugaredLogger
}

func NewGormLogger(config logger.Config, logLevel logger.LogLevel) *GormLogger {
	return &GormLogger{
		Config: config,
		Level:  logLevel,
		log:    log.New(os.Stdout, "", log.LstdFlags),
		//Log:    log,
	}
}

type writer struct{}

func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	l.Level = level
	return l
}

func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.Level < logger.Info {
		return
	}
	l.log.Printf(msg, data...)
}

func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.Level < logger.Warn {
		return
	}
	l.log.Printf(msg, data...)
}

func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.Level < logger.Error {
		return
	}
	l.log.Printf(msg, data...)
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.Level <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	switch {
	case err != nil && l.Config.LogLevel >= logger.Error && (!errors.Is(err, logger.ErrRecordNotFound) || !l.Config.IgnoreRecordNotFoundError):
		sql, rows := fc()
		if rows == -1 {
			l.log.Printf("%s [%s] %s %s", elapsed, FileWithLineNum(), sql, err)
		} else {
			l.log.Printf("%s [%s] %s %v %s", elapsed, FileWithLineNum(), sql, rows, err)
		}
	case elapsed > l.Config.SlowThreshold && l.Config.SlowThreshold != 0 && l.Config.LogLevel >= logger.Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.Config.SlowThreshold)
		if rows == -1 {
			l.log.Printf("%s [%s] %s [%s]", elapsed, FileWithLineNum(), slowLog, sql)
		} else {
			l.log.Printf("%s [%s] %s [%s] %v", elapsed, FileWithLineNum(), slowLog, sql, rows)
		}
	case l.Config.LogLevel == logger.Info:
		sql, rows := fc()
		if rows == -1 {
			l.log.Printf("%s [%s] `%s`", elapsed, FileWithLineNum(), sql)
		} else {
			l.log.Printf("%s [%s] `%s` %v", elapsed, FileWithLineNum(), sql, rows)
		}
	default:
		return
	}
}

func FileWithLineNum() string {

	absBaseDir, err := filepath.Abs(runner.Pwd)
	if err != nil {
		return ""
	}
	for i := 3; i < 15; i++ {
		_, file, line, ok := runtime.Caller(i)
		if ok && !strings.Contains(file, "gorm.io") {
			relFile, err := filepath.Rel(absBaseDir, file)
			if err != nil {
				return ""
			}
			return relFile + ":" + strconv.Itoa(line)
		}
	}
	return ""
}
