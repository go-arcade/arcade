package log

import (
	"fmt"
	"os"
	"time"

	"github.com/google/wire"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const key = iota

var (
	logger      *zap.Logger
	sugar       *zap.SugaredLogger
	ProviderSet = wire.NewSet(NewLog)
)

// LogConfig holds LogConfig configuration options.
type LogConfig struct {
	Output       string
	Path         string
	Level        string
	keepHours    int
	rotateSize   int
	rotateNum    int
	KafkaBrokers string
	KafkaTopic   string
}

type Logger struct {
	Log  *zap.SugaredLogger
	conf *LogConfig
}

type Option func(*Logger)

// NewLog initializes the LogConfig and returns a sugared LogConfig.
func NewLog(conf *LogConfig) *zap.Logger {
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
		writeSyncer = getFileLogWriter(conf)
	case "kafka":
		kafkaSyncer, err := getKafkaLogWriter(conf)
		if err != nil {
			return nil
		}
		writeSyncer = kafkaSyncer
	default:
		writeSyncer = zapcore.AddSync(os.Stdout)
	}

	switch conf.Level {
	case "DEBUG":
		core = zapcore.NewCore(encoder, writeSyncer, zapcore.DebugLevel)
	case "INFO":
		core = zapcore.NewCore(encoder, writeSyncer, zapcore.InfoLevel)
	case "WARN":
		core = zapcore.NewCore(encoder, writeSyncer, zapcore.WarnLevel)
	case "ERROR":
		core = zapcore.NewCore(encoder, writeSyncer, zapcore.ErrorLevel)
	case "FATAL":
		core = zapcore.NewCore(encoder, writeSyncer, zapcore.FatalLevel)
	default:
		core = zapcore.NewCore(encoder, writeSyncer, zapcore.InfoLevel)
	}

	logger = zap.New(core, zap.AddCallerSkip(1), zap.AddCaller())
	fmt.Printf("[Init] log output: %s\n", conf.Output)

	sugar = logger.Sugar()

	return logger
}

// getEncoder returns the appropriate encoder based on the mode.
func getEncoder() zapcore.Encoder {
	var encoderConfig zapcore.EncoderConfig

	encoderConfig = zap.NewDevelopmentEncoderConfig()

	encoderConfig.TimeKey = "time"
	encoderConfig.LevelKey = "level"
	encoderConfig.NameKey = "LogConfig"
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
