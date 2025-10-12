package agent

import (
	"context"
	"time"

	agentapi "github.com/observabil/arcade/api/agent/v1"
)

type AgentServiceImpl struct {
	agentapi.UnimplementedAgentServer
}

func (a *AgentServiceImpl) Heartbeat(ctx context.Context, req *agentapi.HeartbeatRequest) (*agentapi.HeartbeatResponse, error) {
	return &agentapi.HeartbeatResponse{
		Message:   "pong",
		Timestamp: time.Now().Unix(),
	}, nil
}
