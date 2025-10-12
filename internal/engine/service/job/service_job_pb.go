package job

import (
	"context"
	"time"

	jobapi "github.com/observabil/arcade/api/job/v1"
)

type JobServiceImpl struct {
	jobapi.UnimplementedJobServer
}

func (a *JobServiceImpl) Ping(ctx context.Context, req *jobapi.PingRequest) (*jobapi.PingResponse, error) {
	return &jobapi.PingResponse{Message: "pong", Timestamp: time.Now().Unix()}, nil
}
