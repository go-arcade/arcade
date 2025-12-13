package service

import (
	pipelinev1 "github.com/go-arcade/arcade/api/pipeline/v1"
)

type PipelineServiceImpl struct {
	pipelinev1.UnimplementedPipelineServiceServer
}
