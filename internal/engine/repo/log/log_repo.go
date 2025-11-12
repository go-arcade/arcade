package repo

import "github.com/go-arcade/arcade/pkg/ctx"

type LogRepository struct {
	ctx *ctx.Context
}

func NewLogRepository(ctx *ctx.Context) *LogRepository {
	return &LogRepository{ctx: ctx}
}
