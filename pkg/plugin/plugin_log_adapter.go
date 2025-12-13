package plugin

import (
	"fmt"
	"io"
	stdlog "log"
	"strings"

	"github.com/go-arcade/arcade/pkg/log"
	"github.com/hashicorp/go-hclog"
)

// LogAdapter adapts pkg/log.Logger to hclog.Logger interface for use with go-plugin
type LogAdapter struct {
	logger *log.Logger
	name   string
	level  hclog.Level
	args   []any
}

// NewLogAdapter creates a new LogAdapter that wraps pkg/log.Logger
func NewLogAdapter(logger *log.Logger) hclog.Logger {
	if logger == nil {
		// Fallback to default logger if nil - create a new one
		logger = &log.Logger{Log: log.GetLogger()}
	}
	return &LogAdapter{
		logger: logger,
		name:   "PLUGIN",
		level:  hclog.Level(log.GetLevel()),
		args:   nil,
	}
}

// Log emits a message and key/value pairs at a provided log level
func (l *LogAdapter) Log(level hclog.Level, msg string, args ...any) {
	if !l.IsLevelEnabled(level) {
		return
	}

	// Format message with args
	formattedMsg := l.formatMessage(msg, args...)

	switch level {
	case hclog.Trace:
		l.logger.Log.Debugw(formattedMsg)
	case hclog.Debug:
		l.logger.Log.Debugw(formattedMsg)
	case hclog.Info:
		l.logger.Log.Infow(formattedMsg)
	case hclog.Warn:
		l.logger.Log.Warnw(formattedMsg)
	case hclog.Error:
		l.logger.Log.Errorw(formattedMsg)
	default:
		l.logger.Log.Infow(formattedMsg)
	}
}

// Trace emits a message and key/value pairs at TRACE level
func (l *LogAdapter) Trace(msg string, args ...any) {
	l.Log(hclog.Trace, msg, args...)
}

// Debug emits a message and key/value pairs at DEBUG level
func (l *LogAdapter) Debug(msg string, args ...any) {
	l.Log(hclog.Debug, msg, args...)
}

// Info emits a message and key/value pairs at INFO level
func (l *LogAdapter) Info(msg string, args ...any) {
	l.Log(hclog.Info, msg, args...)
}

// Warn emits a message and key/value pairs at WARN level
func (l *LogAdapter) Warn(msg string, args ...any) {
	l.Log(hclog.Warn, msg, args...)
}

// Error emits a message and key/value pairs at ERROR level
func (l *LogAdapter) Error(msg string, args ...any) {
	l.Log(hclog.Error, msg, args...)
}

// IsTrace returns true if TRACE logs would be emitted
func (l *LogAdapter) IsTrace() bool {
	return l.IsLevelEnabled(hclog.Trace)
}

// IsDebug returns true if DEBUG logs would be emitted
func (l *LogAdapter) IsDebug() bool {
	return l.IsLevelEnabled(hclog.Debug)
}

// IsInfo returns true if INFO logs would be emitted
func (l *LogAdapter) IsInfo() bool {
	return l.IsLevelEnabled(hclog.Info)
}

// IsWarn returns true if WARN logs would be emitted
func (l *LogAdapter) IsWarn() bool {
	return l.IsLevelEnabled(hclog.Warn)
}

// IsError returns true if ERROR logs would be emitted
func (l *LogAdapter) IsError() bool {
	return l.IsLevelEnabled(hclog.Error)
}

// IsLevelEnabled checks if a level is enabled
func (l *LogAdapter) IsLevelEnabled(level hclog.Level) bool {
	return level >= l.level
}

// ImpliedArgs returns With key/value pairs
func (l *LogAdapter) ImpliedArgs() []any {
	return l.args
}

// With creates a sublogger that will always have the given key/value pairs
func (l *LogAdapter) With(args ...any) hclog.Logger {
	newArgs := append(l.args, args...)
	return &LogAdapter{
		logger: l.logger,
		name:   l.name,
		level:  l.level,
		args:   newArgs,
	}
}

// Name returns the name of the logger
func (l *LogAdapter) Name() string {
	return l.name
}

// Named creates a logger that will prepend the name string on the front of all messages
func (l *LogAdapter) Named(name string) hclog.Logger {
	newName := l.name
	if newName != "" {
		newName = newName + "." + name
	} else {
		newName = name
	}
	return &LogAdapter{
		logger: l.logger,
		name:   newName,
		level:  l.level,
		args:   l.args,
	}
}

// ResetNamed creates a logger that will prepend the name string on the front of all messages
// This sets the name of the logger to the value directly, unlike Named which honors the current name as well
func (l *LogAdapter) ResetNamed(name string) hclog.Logger {
	return &LogAdapter{
		logger: l.logger,
		name:   name,
		level:  l.level,
		args:   l.args,
	}
}

// SetLevel updates the level
func (l *LogAdapter) SetLevel(level hclog.Level) {
	l.level = level
}

// GetLevel returns the current level
func (l *LogAdapter) GetLevel() hclog.Level {
	return l.level
}

// StandardLogger returns a value that conforms to the stdlib log.Logger interface
func (l *LogAdapter) StandardLogger(opts *hclog.StandardLoggerOptions) *stdlog.Logger {
	if opts == nil {
		opts = &hclog.StandardLoggerOptions{}
	}
	return stdlog.New(&logWriterAdapter{logger: l, level: opts.ForceLevel}, "", 0)
}

// StandardWriter returns a value that conforms to io.Writer
func (l *LogAdapter) StandardWriter(opts *hclog.StandardLoggerOptions) io.Writer {
	return &logWriterAdapter{logger: l, level: opts.ForceLevel}
}

// formatMessage formats the message with name prefix and key-value pairs
func (l *LogAdapter) formatMessage(msg string, args ...any) string {
	// Add name prefix if present
	if l.name != "" {
		msg = fmt.Sprintf("[%s] %s", l.name, msg)
	}

	// Combine implied args with provided args
	allArgs := append(l.args, args...)

	// Format key-value pairs
	if len(allArgs) > 0 {
		var parts []string
		for i := 0; i < len(allArgs); i += 2 {
			if i+1 < len(allArgs) {
				key := fmt.Sprintf("%v", allArgs[i])
				value := allArgs[i+1]
				parts = append(parts, fmt.Sprintf("%s=%v", key, value))
			} else {
				// Odd number of args, append as is
				parts = append(parts, fmt.Sprintf("%v", allArgs[i]))
			}
		}
		if len(parts) > 0 {
			msg += " " + strings.Join(parts, " ")
		}
	}

	return msg
}

// logWriterAdapter adapts hclog.Logger to io.Writer
type logWriterAdapter struct {
	logger hclog.Logger
	level  hclog.Level
}

func (w *logWriterAdapter) Write(p []byte) (n int, err error) {
	msg := strings.TrimRight(string(p), "\n")
	if w.level == hclog.NoLevel {
		w.logger.Info(msg)
	} else {
		w.logger.Log(w.level, msg)
	}
	return len(p), nil
}
