package service

import (
	"context"
	"errors"
	"time"

	"github.com/bytedance/sonic"
	agentv1 "github.com/go-arcade/arcade/api/agent/v1"
	agentmodel "github.com/go-arcade/arcade/internal/engine/model"
	"github.com/go-arcade/arcade/pkg/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type AgentServiceImpl struct {
	agentv1.UnimplementedAgentServiceServer
	agentService *AgentService
}

func NewAgentServiceImpl(agentService *AgentService) *AgentServiceImpl {
	return &AgentServiceImpl{
		agentService: agentService,
	}
}

func (a *AgentServiceImpl) Heartbeat(ctx context.Context, req *agentv1.HeartbeatRequest) (*agentv1.HeartbeatResponse, error) {

	resp := &agentv1.HeartbeatResponse{
		Success:   true,
		Message:   "pong",
		Timestamp: time.Now().Unix(),
	}

	return resp, nil
}

func (a *AgentServiceImpl) Register(ctx context.Context, req *agentv1.RegisterRequest) (*agentv1.RegisterResponse, error) {
	// Validate token
	if req.Token == "" {
		return nil, status.Errorf(codes.InvalidArgument, "token is required")
	}

	agentRepo := a.agentService.agentRepo
	var agentId string
	var existingAgent *agentmodel.Agent
	var err error

	// Extract agentId from request (if provided)
	if req.AgentId != "" {
		agentId = req.AgentId
		// Verify token by regenerating it and comparing
		expectedToken, err := a.agentService.generateAgentToken(agentId)
		if err != nil {
			log.Errorw("failed to generate token for verification", "agentId", agentId, "error", err)
			return nil, status.Errorf(codes.Internal, "failed to verify token")
		}

		if req.Token != expectedToken {
			log.Warnw("token verification failed", "agentId", agentId)
			return nil, status.Errorf(codes.Unauthenticated, "invalid token")
		}

		// Check if agent exists
		existingAgent, err = agentRepo.GetAgentByAgentId(agentId)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, status.Errorf(codes.NotFound, "agent not found: %s", agentId)
			}
			log.Errorw("failed to get agent", "agentId", agentId, "error", err)
			return nil, status.Errorf(codes.Internal, "failed to get agent")
		}
	} else {
		// If agentId is not provided, try to find agent by verifying token against all agents
		// This is less efficient but allows registration with only token
		agents, _, err := agentRepo.ListAgent(1, 1000) // Get up to 1000 agents
		if err != nil {
			log.Errorw("failed to list agents for token verification", "error", err)
			return nil, status.Errorf(codes.Internal, "failed to verify token")
		}

		// Try to find matching agent by verifying token
		found := false
		for _, agent := range agents {
			expectedToken, err := a.agentService.generateAgentToken(agent.AgentId)
			if err != nil {
				log.Debugw("failed to generate token for agent", "agentId", agent.AgentId, "error", err)
				continue
			}

			if req.Token == expectedToken {
				agentId = agent.AgentId
				existingAgent = &agent
				found = true
				break
			}
		}

		if !found {
			log.Warnw("token verification failed: no matching agent found")
			return nil, status.Errorf(codes.Unauthenticated, "invalid token")
		}
	}

	// Update agent information from registration request
	updates := make(map[string]any)

	if req.Ip != "" {
		updates["address"] = req.Ip
	}
	if req.Os != "" {
		updates["os"] = req.Os
	}
	if req.Arch != "" {
		updates["arch"] = req.Arch
	}
	if req.Version != "" {
		updates["version"] = req.Version
	}
	if len(req.Labels) > 0 {
		updates["labels"] = req.Labels
	}
	updates["status"] = 1 // Set status to online
	updates["last_heartbeat"] = time.Now()

	if len(updates) > 0 {
		if err := agentRepo.UpdateAgentById(existingAgent.ID, updates); err != nil {
			log.Errorw("failed to update agent during registration", "agentId", agentId, "error", err)
			return nil, status.Errorf(codes.Internal, "failed to update agent")
		}
	}

	// Get agent detail to return heartbeat interval
	detail, err := agentRepo.GetAgentDetailById(existingAgent.ID)
	if err != nil {
		log.Errorw("failed to get agent detail", "agentId", agentId, "error", err)
		return nil, status.Errorf(codes.Internal, "failed to get agent detail")
	}

	heartbeatInterval := int64(60) // default

	// Parse labels from JSON
	labels := make(map[string]string)
	if len(detail.Labels) > 0 {
		if err := sonic.Unmarshal(detail.Labels, &labels); err != nil {
			log.Warnw("failed to parse labels", "agentId", agentId, "error", err)
			// Continue with empty labels if parsing fails
		}
	}

	return &agentv1.RegisterResponse{
		Success:           true,
		Message:           "registration successful",
		AgentId:           agentId,
		HeartbeatInterval: heartbeatInterval,
		Labels:            labels,
	}, nil
}
