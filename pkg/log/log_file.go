package log

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// getFileLogWriter returns the WriteSyncer for logging to a file.
func getFileLogWriter(config *Conf) (zapcore.WriteSyncer, error) {
	// confirm log directory if not exists
	if err := os.MkdirAll(config.Path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	logPath := filepath.Join(config.Path, config.Filename)

	lumberJackLogger := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    config.RotateSize, // MB
		MaxBackups: config.RotateNum,
		MaxAge:     config.KeepHours, // days
		Compress:   true,
	}

	return zapcore.AddSync(lumberJackLogger), nil
}
