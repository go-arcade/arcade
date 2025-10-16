package log

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

// parseLogLevel converts a string level to a LogLevel.
func parseLogLevel(level string) LogLevel {
	switch level {
	case "DEBUG":
		return DebugLevel
	case "INFO":
		return InfoLevel
	case "WARN":
		return WarnLevel
	case "ERROR":
		return ErrorLevel
	case "FATAL":
		return FatalLevel
	default:
		return InfoLevel
	}
}
