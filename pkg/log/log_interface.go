package log

type ILogger interface {
	Info(args ...any)
	Infow(msg string, keysAndValues ...any)

	Debug(args ...any)
	Debugw(msg string, keysAndValues ...any)

	Warn(args ...any)
	Warnw(msg string, keysAndValues ...any)

	Error(args ...any)
	Errorw(msg string, keysAndValues ...any)
}
