package agent

import (
	"context"
	"time"

	agentapi "github.com/go-arcade/arcade/api/agent/v1"
)

type AgentServiceImpl struct {
	agentapi.UnimplementedAgentServiceServer
}

func (a *AgentServiceImpl) Heartbeat(ctx context.Context, req *agentapi.HeartbeatRequest) (*agentapi.HeartbeatResponse, error) {
	return &agentapi.HeartbeatResponse{
		Success:   true,
		Message:   "pong",
		Timestamp: time.Now().Unix(),
	}, nil
}
