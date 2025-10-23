package pipeline

import (
	"context"
	"time"

	pipelineapi "github.com/go-arcade/arcade/api/pipeline/v1"
)

type PipelineServiceImpl struct {
	pipelineapi.UnimplementedPipelineServiceServer
}

func (a *PipelineServiceImpl) Ping(ctx context.Context, req *pipelineapi.PingRequest) (*pipelineapi.PingResponse, error) {
	return &pipelineapi.PingResponse{Message: "pong", Timestamp: time.Now().Unix()}, nil
}
