package service

import (
	taskv1 "github.com/go-arcade/arcade/api/task/v1"
)

type TaskServiceImpl struct {
	taskv1.UnimplementedTaskServiceServer
}
