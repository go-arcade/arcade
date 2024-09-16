package log

import (
	"fmt"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/6/8 1:21
 * @file: log_file.go
 * @description: logger writer file
 */
const filename string = "arcade.LOG"

// getFileLogWriter returns the WriteSyncer for logging to a file.
func getFileLogWriter(config *LogConfig) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   fmt.Sprintf("%s/%s", config.Path, filename),
		MaxSize:    config.rotateSize,
		MaxBackups: config.rotateNum,
		MaxAge:     config.keepHours,
		Compress:   true,
	}
	return zapcore.AddSync(lumberJackLogger)
}
