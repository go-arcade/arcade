package log

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const defaultFilename string = "log.LOG"

// getFileLogWriter returns the WriteSyncer for logging to a file.
func getFileLogWriter(config *Conf) (zapcore.WriteSyncer, error) {
	// 确保日志目录存在
	if err := os.MkdirAll(config.Path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// 如果 filename 为空，使用默认值
	filename := config.Filename
	if filename == "" {
		filename = defaultFilename
	}

	logPath := filepath.Join(config.Path, filename)

	lumberJackLogger := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    config.RotateSize, // MB
		MaxBackups: config.RotateNum,
		MaxAge:     config.KeepHours, // days
		Compress:   true,
	}

	return zapcore.AddSync(lumberJackLogger), nil
}
