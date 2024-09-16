package log

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/16 15:34
 * @file: log_level.go
 * @description: LogConfig level
 */

// LogLevel defines the severity of a LogConfig internal.
type LogLevel int8

const (
	// DebugLevel logs debug messages.
	DebugLevel LogLevel = iota - 1
	// InfoLevel logs informational messages.
	InfoLevel
	// WarnLevel logs warning messages.
	WarnLevel
	// ErrorLevel logs error messages.
	ErrorLevel
	// FatalLevel logs fatal messages.
	FatalLevel
)

func (l LogLevel) String() string {
	switch l {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	case FatalLevel:
		return "fatal"
	default:
		return "unknown"
	}
}

// ParseLogLevel converts a string level to a LogLevel.
func ParseLogLevel(level string) LogLevel {
	switch level {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn":
		return WarnLevel
	case "error":
		return ErrorLevel
	case "fatal":
		return FatalLevel
	default:
		return InfoLevel
	}
}
