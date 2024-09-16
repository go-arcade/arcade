package main

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"os"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/16 14:38
 * @file: main.go
 * @description:
 */

func main() {

	l := log.NewHelper(logger())
	l.Info("world")
}

func logger() log.Logger {
	l := log.With(log.NewStdLogger(os.Stdout),
		log.FilterLevel(log.LevelInfo),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"trace_id", tracing.TraceID(),
		"span_id", tracing.SpanID(),
	)
	return l
}
