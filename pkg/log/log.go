package log

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/8 21:12
 * @file: log.go
 * @description: log
 */

type Log struct {
	Output     string
	Level      string
	Path       string
	KeepHours  int
	RotateNum  int
	RotateSize int
}

func NewLog(conf Log) log.Logger {

	logger := log.With(
		log.NewStdLogger(os.Stdout),
		log.FilterLevel(log.LevelInfo),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"trace_id", tracing.TraceID(),
		"span_id", tracing.SpanID(),
	)

	if conf.Output == "file" {
		fileWriter := &lumberjack.Logger{
			Filename: conf.Path,
			MaxSize:  conf.RotateSize, // megabytes
			MaxAge:   conf.KeepHours,  // days
			Compress: true,            // disabled by default
		}
		logger = log.With(
			log.NewStdLogger(io.MultiWriter(os.Stdout, fileWriter)),
			"ts", log.DefaultTimestamp,
			"caller", log.DefaultCaller,
			"trace_id", tracing.TraceID(),
			"span_id", tracing.SpanID(),
		)
	}

	return logger
}
