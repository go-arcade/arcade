package queue

import (
	"github.com/go-arcade/arcade/pkg/log"
)

// asynqLoggerAdapter 适配器，将 asynq.Logger 接口适配到 pkg/log
type asynqLoggerAdapter struct{}

// Debug 实现 asynq.Logger 接口
func (l *asynqLoggerAdapter) Debug(args ...interface{}) {
	log.Debug(args...)
}

// Info 实现 asynq.Logger 接口
func (l *asynqLoggerAdapter) Info(args ...interface{}) {
	log.Info(args...)
}

// Warn 实现 asynq.Logger 接口
func (l *asynqLoggerAdapter) Warn(args ...interface{}) {
	log.Warn(args...)
}

// Error 实现 asynq.Logger 接口
func (l *asynqLoggerAdapter) Error(args ...interface{}) {
	log.Error(args...)
}

// Fatal 实现 asynq.Logger 接口
func (l *asynqLoggerAdapter) Fatal(args ...interface{}) {
	log.Fatal(args...)
}
