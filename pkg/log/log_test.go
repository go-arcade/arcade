package log

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap/zapcore"
)

func TestDefaultConf(t *testing.T) {
	conf := SetDefaults()

	if conf.Output != "stdout" {
		t.Errorf("expected output to be stdout, got %s", conf.Output)
	}

	if conf.Level != "INFO" {
		t.Errorf("expected level to be INFO, got %s", conf.Level)
	}

	if conf.KeepHours != 7 {
		t.Errorf("expected KeepHours to be 7, got %d", conf.KeepHours)
	}
}

func TestConf_Validate(t *testing.T) {
	tests := []struct {
		name    string
		conf    *Conf
		wantErr bool
	}{
		{
			name: "valid stdout config",
			conf: &Conf{
				Output: "stdout",
				Level:  "INFO",
			},
			wantErr: false,
		},
		{
			name: "valid file config",
			conf: &Conf{
				Output:     "file",
				Path:       "/tmp/logs",
				Level:      "DEBUG",
				KeepHours:  7,
				RotateSize: 100,
				RotateNum:  10,
			},
			wantErr: false,
		},
		{
			name: "invalid file config - missing path",
			conf: &Conf{
				Output: "file",
				Level:  "INFO",
			},
			wantErr: true,
		},
		{
			name: "file config with auto-correction",
			conf: &Conf{
				Output: "file",
				Path:   "/tmp/logs",
				Level:  "INFO",
				// 未设置 KeepHours, RotateSize, RotateNum
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.conf.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 验证自动修正
			if !tt.wantErr && tt.conf.Output == "file" {
				if tt.conf.RotateSize <= 0 {
					t.Error("RotateSize should be auto-corrected to positive value")
				}
				if tt.conf.RotateNum <= 0 {
					t.Error("RotateNum should be auto-corrected to positive value")
				}
				if tt.conf.KeepHours <= 0 {
					t.Error("KeepHours should be auto-corrected to positive value")
				}
			}
		})
	}
}

func TestNewLog_Stdout(t *testing.T) {
	conf := &Conf{
		Output: "stdout",
		Level:  "DEBUG",
	}

	logger, err := NewLog(conf)
	if err != nil {
		t.Fatalf("NewLog() error = %v", err)
	}

	if logger == nil {
		t.Fatal("expected non-nil logger")
	}

	// 测试日志输出
	logger.Info("test message")
}

func TestNewLog_File(t *testing.T) {
	tmpDir := t.TempDir()

	conf := &Conf{
		Output:     "file",
		Path:       tmpDir,
		Level:      "INFO",
		KeepHours:  1,
		RotateSize: 1,
		RotateNum:  3,
	}

	logger, err := NewLog(conf)
	if err != nil {
		t.Fatalf("NewLog() error = %v", err)
	}

	if logger == nil {
		t.Fatal("expected non-nil logger")
	}

	// 写入一些日志
	logger.Info("test message 1")
	logger.Debug("test message 2")
	logger.Warn("test message 3")

	// 确保日志被写入
	logger.Sync()

	// 验证日志文件存在
	logFile := filepath.Join(tmpDir, "arcade.LOG")
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Errorf("log file should exist at %s", logFile)
	}
}

func TestInit(t *testing.T) {
	conf := SetDefaults()
	err := Init(conf)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// 验证全局 logger 已初始化
	if sugar == nil {
		t.Error("global sugar logger should be initialized")
	}
}

func TestGlobalLogFunctions(t *testing.T) {
	// 重置全局变量
	mu.Lock()
	sugar = nil
	logger = nil
	mu.Unlock()
	once = *new(sync.Once)

	// 测试在未初始化的情况下调用（应该自动初始化）
	Info("test info message")
	Debug("test debug message")
	Warn("test warn message")
	Error("test error message")

	// 验证已自动初始化
	if sugar == nil {
		t.Error("global logger should be auto-initialized")
	}
}

func TestGlobalLogFunctions_Formatted(t *testing.T) {
	conf := SetDefaults()
	Init(conf)

	Infof("formatted info: %s", "test")
	Debugf("formatted debug: %d", 123)
	Warnf("formatted warn: %v", true)
	Errorf("formatted error: %f", 3.14)
}

func TestGlobalLogFunctions_WithFields(t *testing.T) {
	conf := SetDefaults()
	Init(conf)

	Infow("structured info", "key1", "value1", "key2", 123)
	Debugw("structured debug", "user", "alice", "action", "login")
	Warnw("structured warn", "count", 5)
	Errorw("structured error", "error", "something went wrong")
}

type contextKey string

const (
	requestIDKey contextKey = "request_id"
	userIDKey    contextKey = "user_id"
	traceIDKey   contextKey = "trace_id"
)

func TestWithContext(t *testing.T) {
	conf := SetDefaults()
	Init(conf)

	// 创建带有字段的 context
	ctx := context.Background()
	ctx = context.WithValue(ctx, requestIDKey, "req-123")
	ctx = context.WithValue(ctx, userIDKey, "user-456")
	ctx = context.WithValue(ctx, traceIDKey, "trace-789")

	// 使用 WithContext 创建带上下文的 logger
	ctxLogger := WithContext(ctx)
	ctxLogger.Info("message with context")

	// 验证返回的是 logger
	if ctxLogger == nil {
		t.Error("WithContext should return non-nil logger")
	}
}

func TestWith(t *testing.T) {
	conf := SetDefaults()
	Init(conf)

	// 使用 With 添加字段
	logger := With("component", "test", "version", "1.0")
	logger.Info("message with fields")

	if logger == nil {
		t.Error("With should return non-nil logger")
	}
}

func TestSync(t *testing.T) {
	conf := SetDefaults()
	Init(conf)

	Info("test message")

	err := Sync()
	if err != nil {
		// Sync 可能会返回错误（比如 stdout 无法 sync），这是正常的
		t.Logf("Sync returned error (this is often expected): %v", err)
	}
}

func TestConcurrentLogging(t *testing.T) {
	conf := SetDefaults()
	Init(conf)

	// 并发写日志
	done := make(chan bool, 100)

	for i := 0; i < 100; i++ {
		go func(n int) {
			Infof("concurrent message %d", n)
			Debugf("debug message %d", n)
			Warnf("warn message %d", n)
			done <- true
		}(i)
	}

	// 等待所有 goroutine 完成
	for range 100 {
		<-done
	}
}

func TestMultipleInit(t *testing.T) {
	// 测试多次初始化
	for i := 0; i < 3; i++ {
		conf := SetDefaults()
		err := Init(conf)
		if err != nil {
			t.Fatalf("Init() iteration %d error = %v", i, err)
		}
	}

	// 验证最后一次初始化生效
	Info("test after multiple init")
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected zapcore.Level
	}{
		{"DEBUG", zapcore.DebugLevel},
		{"INFO", zapcore.InfoLevel},
		{"WARN", zapcore.WarnLevel},
		{"ERROR", zapcore.ErrorLevel},
		{"FATAL", zapcore.FatalLevel},
		{"INVALID", zapcore.InfoLevel}, // 默认值
		{"", zapcore.InfoLevel},        // 默认值
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseLogLevel(tt.input)
			if result != tt.expected {
				t.Errorf("parseLogLevel(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestLogRotation(t *testing.T) {
	tmpDir := t.TempDir()

	conf := &Conf{
		Output:     "file",
		Path:       tmpDir,
		Level:      "INFO",
		KeepHours:  1,
		RotateSize: 1, // 1MB，容易触发轮转
		RotateNum:  3,
	}

	logger, err := NewLog(conf)
	if err != nil {
		t.Fatalf("NewLog() error = %v", err)
	}

	// 写入大量日志以触发轮转
	sugar := logger.Sugar()
	for i := 0; i < 10000; i++ {
		sugar.Info("This is a test message to trigger log rotation. Message number:", i)
	}

	logger.Sync()

	// 验证日志文件存在
	logFile := filepath.Join(tmpDir, "arcade.LOG")
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Errorf("log file should exist at %s", logFile)
	}

	// 检查文件内容
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	if len(content) == 0 {
		t.Error("log file should not be empty")
	}
}

func BenchmarkInfo(b *testing.B) {
	conf := SetDefaults()
	Init(conf)

	
	for b.Loop() {
		Info("benchmark message")
	}
}

func BenchmarkInfof(b *testing.B) {
	conf := SetDefaults()
	Init(conf)

	
	for i := 0; b.Loop(); i++ {
		Infof("benchmark message %d", i)
	}
}

func BenchmarkInfow(b *testing.B) {
	conf := SetDefaults()
	Init(conf)

	
	for i := 0; b.Loop(); i++ {
		Infow("benchmark message", "iteration", i, "timestamp", time.Now())
	}
}
