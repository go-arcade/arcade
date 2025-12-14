// Copyright 2025 Arcade Team
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package service

import (
	"maps"
	"context"
	"time"

	agentv1 "github.com/go-arcade/arcade/api/agent/v1"
	"github.com/go-arcade/arcade/internal/agent/config"
	grpcclient "github.com/go-arcade/arcade/internal/pkg/grpc"
	"github.com/go-arcade/arcade/pkg/log"
)

// AgentService implements agent.v1.AgentServiceServer
type AgentService struct {
	agentv1.UnimplementedAgentServiceServer
	agentConf  *config.AgentConfig
	grpcClient *grpcclient.ClientWrapper
}

// NewAgentService creates a new AgentService instance
func NewAgentService(agentConf *config.AgentConfig, grpcClient *grpcclient.ClientWrapper) *AgentService {
	return &AgentService{
		agentConf:  agentConf,
		grpcClient: grpcClient,
	}
}

// Heartbeat handles heartbeat requests from server
func (s *AgentService) Heartbeat(ctx context.Context, req *agentv1.HeartbeatRequest) (*agentv1.HeartbeatResponse, error) {
	log.Debugw("Heartbeat received", "agent_id", req.AgentId, "status", req.Status.String())

	return &agentv1.HeartbeatResponse{
		Success:   true,
		Message:   "heartbeat acknowledged",
		Timestamp: time.Now().Unix(),
	}, nil
}

// Register handles agent registration requests
func (s *AgentService) Register(ctx context.Context, req *agentv1.RegisterRequest) (*agentv1.RegisterResponse, error) {
	log.Infow("Register request received", "agent_id", req.AgentId, "ip", req.Ip, "version", req.Version)

	// TODO: Implement registration logic
	// For now, return success with default heartbeat interval
	return &agentv1.RegisterResponse{
		Success:           true,
		Message:           "registration successful",
		AgentId:           req.AgentId,
		HeartbeatInterval: int64(s.agentConf.Agent.Interval),
	}, nil
}

// Unregister handles agent unregistration requests
func (s *AgentService) Unregister(ctx context.Context, req *agentv1.UnregisterRequest) (*agentv1.UnregisterResponse, error) {
	log.Infow("Unregister request received", "agent_id", req.AgentId, "reason", req.Reason)

	// TODO: Implement unregistration logic
	return &agentv1.UnregisterResponse{
		Success: true,
		Message: "unregistration successful",
	}, nil
}

// ReportTaskStatus handles task status reporting requests
func (s *AgentService) ReportTaskStatus(ctx context.Context, req *agentv1.ReportTaskStatusRequest) (*agentv1.ReportTaskStatusResponse, error) {
	log.Debugw("ReportTaskStatus request received", "agent_id", req.AgentId, "task_id", req.TaskId, "status", req.Status.String())

	// TODO: Implement task status reporting logic
	return &agentv1.ReportTaskStatusResponse{
		Success: true,
		Message: "task status reported successfully",
	}, nil
}

// ReportTaskLog handles task log reporting requests
func (s *AgentService) ReportTaskLog(ctx context.Context, req *agentv1.ReportTaskLogRequest) (*agentv1.ReportTaskLogResponse, error) {
	log.Debugw("ReportTaskLog request received", "agent_id", req.AgentId, "task_id", req.TaskId, "log_count", len(req.Logs))

	// TODO: Implement task log reporting logic
	return &agentv1.ReportTaskLogResponse{
		Success: true,
		Message: "task logs reported successfully",
	}, nil
}

// CancelTask handles task cancellation requests from server
func (s *AgentService) CancelTask(ctx context.Context, req *agentv1.CancelTaskRequest) (*agentv1.CancelTaskResponse, error) {
	log.Infow("CancelTask request received", "agent_id", req.AgentId, "job_id", req.JobId, "reason", req.Reason)

	// TODO: Implement task cancellation logic
	// This should cancel the running task identified by job_id
	return &agentv1.CancelTaskResponse{
		Success: true,
		Message: "task cancellation request received",
	}, nil
}

// UpdateLabels handles agent labels update requests
func (s *AgentService) UpdateLabels(ctx context.Context, req *agentv1.UpdateLabelsRequest) (*agentv1.UpdateLabelsResponse, error) {
	log.Infow("UpdateLabels request received", "agent_id", req.AgentId, "merge", req.Merge, "labels", req.Labels)

	// TODO: Implement labels update logic
	// Update agent labels based on merge flag
	updatedLabels := make(map[string]string)
	if req.Merge {
		// Merge with existing labels
		maps.Copy(updatedLabels, s.agentConf.Agent.Labels)
	}
	maps.Copy(updatedLabels, req.Labels)

	return &agentv1.UpdateLabelsResponse{
		Success: true,
		Message: "labels updated successfully",
		Labels:  updatedLabels,
	}, nil
}

// DownloadPlugin handles plugin download requests
func (s *AgentService) DownloadPlugin(ctx context.Context, req *agentv1.DownloadPluginRequest) (*agentv1.DownloadPluginResponse, error) {
	log.Infow("DownloadPlugin request received", "agent_id", req.AgentId, "plugin_id", req.PluginId, "version", req.Version)

	// TODO: Implement plugin download logic
	return &agentv1.DownloadPluginResponse{
		Success: true,
		Message: "plugin download initiated",
	}, nil
}

// ListAvailablePlugins handles available plugins listing requests
func (s *AgentService) ListAvailablePlugins(ctx context.Context, req *agentv1.ListAvailablePluginsRequest) (*agentv1.ListAvailablePluginsResponse, error) {
	log.Debugw("ListAvailablePlugins request received", "agent_id", req.AgentId, "plugin_type", req.PluginType)

	// TODO: Implement plugin listing logic
	// For now, return empty plugin list
	return &agentv1.ListAvailablePluginsResponse{
		Success: true,
		Message: "no plugins available",
		Plugins: []*agentv1.PluginInfo{},
	}, nil
}
