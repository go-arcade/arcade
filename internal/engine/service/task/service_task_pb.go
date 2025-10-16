package task

import (
	"context"
	"time"

	taskapi "github.com/observabil/arcade/api/task/v1"
)

type TaskServiceImpl struct {
	taskapi.UnimplementedTaskServiceServer
}

func (a *TaskServiceImpl) Ping(ctx context.Context, req *taskapi.PingRequest) (*taskapi.PingResponse, error) {
	return &taskapi.PingResponse{Message: "pong", Timestamp: time.Now().Unix()}, nil
}
