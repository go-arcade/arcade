package service

import (
	"context"
	"time"

	agentv1 "github.com/go-arcade/arcade/api/agent/v1"
)

type AgentServiceImpl struct {
	agentv1.UnimplementedAgentServiceServer
}

func (a *AgentServiceImpl) Heartbeat(ctx context.Context, req *agentv1.HeartbeatRequest) (*agentv1.HeartbeatResponse, error) {
	return &agentv1.HeartbeatResponse{
		Success:   true,
		Message:   "pong",
		Timestamp: time.Now().Unix(),
	}, nil
}
