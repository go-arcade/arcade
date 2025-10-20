package log

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

var once sync.Once

// ensureLogger 确保 logger 已初始化，如果没有则使用默认配置
func ensureLogger() {
	mu.RLock()
	initialized := sugar != nil
	mu.RUnlock()

	if !initialized {
		once.Do(func() {
			// 使用默认配置初始化
			if err := Init(DefaultConf()); err != nil {
				// 如果初始化失败，创建一个基本的 stdout logger
				zapLogger, _ := zap.NewProduction()
				mu.Lock()
				logger = zapLogger
				sugar = zapLogger.Sugar()
				mu.Unlock()
			}
		})
	}
}

// getSugar 线程安全地获取 sugar logger
func getSugar() *zap.SugaredLogger {
	ensureLogger()
	mu.RLock()
	defer mu.RUnlock()
	return sugar
}

func Info(args ...interface{}) {
	getSugar().Info(args...)
}

func Infof(format string, args ...interface{}) {
	getSugar().Infof(format, args...)
}

func Infow(msg string, keysAndValues ...interface{}) {
	getSugar().Infow(msg, keysAndValues...)
}

// WithContext 从 context 中提取字段并返回带有这些字段的 logger
// 注意：这里假设你的 context 中存储了 requestID、userID 等字段
func WithContext(ctx context.Context) *zap.SugaredLogger {
	logger := getSugar()

	// 从 context 中提取常见字段
	if requestID, ok := ctx.Value("request_id").(string); ok {
		logger = logger.With("request_id", requestID)
	}
	if traceID, ok := ctx.Value("trace_id").(string); ok {
		logger = logger.With("trace_id", traceID)
	}

	return logger
}

// With 添加结构化字段
func With(args ...interface{}) *zap.SugaredLogger {
	return getSugar().With(args...)
}

func Debug(args ...interface{}) {
	getSugar().Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	getSugar().Debugf(format, args...)
}

func Debugw(msg string, keysAndValues ...interface{}) {
	getSugar().Debugw(msg, keysAndValues...)
}

func Warn(args ...interface{}) {
	getSugar().Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	getSugar().Warnf(format, args...)
}

func Warnw(msg string, keysAndValues ...interface{}) {
	getSugar().Warnw(msg, keysAndValues...)
}

func Error(args ...interface{}) {
	getSugar().Error(args...)
}

func Errorf(format string, args ...interface{}) {
	getSugar().Errorf(format, args...)
}

func Errorw(msg string, keysAndValues ...interface{}) {
	getSugar().Errorw(msg, keysAndValues...)
}

func Fatal(args ...interface{}) {
	getSugar().Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	getSugar().Fatalf(format, args...)
}

func Fatalw(msg string, keysAndValues ...interface{}) {
	getSugar().Fatalw(msg, keysAndValues...)
}

// Sync 刷新缓冲区
func Sync() error {
	mu.RLock()
	defer mu.RUnlock()
	if logger != nil {
		return logger.Sync()
	}
	return nil
}
