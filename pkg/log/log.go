package log

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/google/wire"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	mu     sync.RWMutex
	logger *zap.Logger
	sugar  *zap.SugaredLogger
)

// ProviderSet 提供日志相关的依赖
var ProviderSet = wire.NewSet(ProvideLogger)

// ProvideLogger 提供 Logger 实例
func ProvideLogger(conf *Conf) (*Logger, error) {
	zapLogger, err := NewLog(conf)
	if err != nil {
		return nil, err
	}
	return &Logger{Log: zapLogger.Sugar()}, nil
}

// Conf holds Conf configuration options.
type Conf struct {
	Output       string
	Path         string
	Filename     string // 日志文件名，为空时使用默认值
	Level        string
	KeepHours    int // 日志保留天数（改为导出）
	RotateSize   int // 单个日志文件最大大小（MB）
	RotateNum    int // 保留的日志文件数量
	KafkaBrokers string
	KafkaTopic   string
}

// DefaultConf 返回默认配置
func DefaultConf() *Conf {
	return &Conf{
		Output:     "stdout",
		Level:      "INFO",
		KeepHours:  7,   // 默认保留7天
		RotateSize: 100, // 默认100MB
		RotateNum:  10,  // 默认保留10个文件
	}
}

// Validate 验证配置
func (c *Conf) Validate() error {
	if c.Output == "file" {
		if c.Path == "" {
			return fmt.Errorf("log path is required when output is 'file'")
		}
		if c.RotateSize <= 0 {
			c.RotateSize = 100
		}
		if c.RotateNum <= 0 {
			c.RotateNum = 10
		}
		if c.KeepHours <= 0 {
			c.KeepHours = 7
		}
	}
	return nil
}

type Logger struct {
	Log *zap.SugaredLogger
}

type Option func(*Logger)

// NewLog initializes the logger and returns a zap.Logger.
func NewLog(conf *Conf) (*zap.Logger, error) {
	// 验证配置
	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("invalid log config: %w", err)
	}

	var (
		writeSyncer zapcore.WriteSyncer
		encoder     zapcore.Encoder
		core        zapcore.Core
	)

	encoder = getEncoder()

	switch conf.Output {
	case "stdout":
		writeSyncer = zapcore.AddSync(os.Stdout)
	case "file":
		var err error
		writeSyncer, err = getFileLogWriter(conf)
		if err != nil {
			return nil, fmt.Errorf("failed to create file log writer: %w", err)
		}
	default:
		writeSyncer = zapcore.AddSync(os.Stdout)
	}

	core = zapcore.NewCore(encoder, writeSyncer, parseLogLevel(conf.Level))

	newLogger := zap.New(core, zap.AddCallerSkip(1), zap.AddCaller())

	// 线程安全地更新全局变量
	mu.Lock()
	logger = newLogger
	sugar = newLogger.Sugar()
	mu.Unlock()

	fmt.Printf("[Init] log initialized - output: %s, level: %s\n", conf.Output, conf.Level)

	return newLogger, nil
}

// Init 初始化全局日志实例（便捷方法）
func Init(conf *Conf) error {
	_, err := NewLog(conf)
	return err
}

// MustInit 初始化全局日志实例，失败则 panic
func MustInit(conf *Conf) {
	if err := Init(conf); err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
}

// getEncoder returns the appropriate encoder based on the mode.
func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewDevelopmentEncoderConfig()

	encoderConfig.TimeKey = "time"
	encoderConfig.LevelKey = "level"
	encoderConfig.NameKey = "Conf"
	encoderConfig.CallerKey = "caller"
	encoderConfig.MessageKey = "msg"
	encoderConfig.StacktraceKey = "stacktrace"
	encoderConfig.LineEnding = zapcore.DefaultLineEnding
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder // 大写编码器
	encoderConfig.EncodeTime = customTimeEncoder            // ISO8601 UTC 时间格式
	encoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder // 相对路径编码器
	encoderConfig.EncodeName = zapcore.FullNameEncoder

	return zapcore.NewConsoleEncoder(encoderConfig)
}

// customTimeEncoder formats the time as 2024-06-08 00:51:55.
func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}

// parseLogLevel converts a string level to a zapcore.Level.
func parseLogLevel(level string) zapcore.Level {
	switch level {
	case "DEBUG":
		return zapcore.DebugLevel
	case "INFO":
		return zapcore.InfoLevel
	case "WARN":
		return zapcore.WarnLevel
	case "ERROR":
		return zapcore.ErrorLevel
	case "FATAL":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}
