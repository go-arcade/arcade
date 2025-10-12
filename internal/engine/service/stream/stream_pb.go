package stream

import (
	"context"
	"time"

	streamapi "github.com/observabil/arcade/api/stream/v1"
)

type StreamServiceImpl struct {
	streamapi.UnimplementedStreamServer
}

func (a *StreamServiceImpl) Ping(ctx context.Context, req *streamapi.PingRequest) (*streamapi.PingResponse, error) {
	return &streamapi.PingResponse{Message: "pong", Timestamp: time.Now().Unix()}, nil
}
