package pipeline

import (
	"context"
	"fmt"
	"sync"
	"time"

	agentv1 "github.com/go-arcade/arcade/api/agent/v1"
	taskv1 "github.com/go-arcade/arcade/api/task/v1"
	"github.com/go-arcade/arcade/pkg/log"
)

// AgentManager manages agent selection and task assignment for pipeline execution
type AgentManager struct {
	agentClient agentv1.AgentServiceClient
	taskClient  taskv1.TaskServiceClient
	logger      log.Logger
	// Agent status cache (updated from heartbeat)
	agentStatusCache map[string]*AgentStatus
	statusCacheMu    sync.RWMutex
}

// NewAgentManager creates a new agent manager
func NewAgentManager(agentClient agentv1.AgentServiceClient, taskClient taskv1.TaskServiceClient, logger log.Logger) *AgentManager {
	return &AgentManager{
		agentClient:      agentClient,
		taskClient:       taskClient,
		logger:           logger,
		agentStatusCache: make(map[string]*AgentStatus),
	}
}

// StepExecutionRequest represents a request to execute a step on an agent
type StepExecutionRequest struct {
	PipelineID string
	BuildID    string
	JobName    string
	StepName   string
	Step       *Step
	Workspace  string
	Env        map[string]string
	Selector   *AgentSelector
}

// StepExecutionResult represents the result of step execution
type StepExecutionResult struct {
	Success   bool
	ExitCode  int32
	Error     string
	StartTime time.Time
	EndTime   time.Time
	Metrics   map[string]string
}

// ExecuteStepOnAgent executes a pipeline step on a selected agent
func (am *AgentManager) ExecuteStepOnAgent(ctx context.Context, req *StepExecutionRequest) (*StepExecutionResult, error) {
	if am.agentClient == nil {
		return nil, fmt.Errorf("agent client is not initialized")
	}

	// Select an agent based on selector
	agentID, err := am.SelectAgent(ctx, req.Selector)
	if err != nil {
		return nil, fmt.Errorf("select agent: %w", err)
	}

	if am.logger.Log != nil {
		am.logger.Log.Infow("executing step on agent", "job", req.JobName, "step", req.StepName, "agent", agentID)
	}

	// Convert step to agent task
	agentTask, err := am.convertStepToTask(req)
	if err != nil {
		return nil, fmt.Errorf("convert step to task: %w", err)
	}

	taskID, err := am.createTask(ctx, req, agentTask)
	if err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}

	if am.logger.Log != nil {
		am.logger.Log.Infow("created task for step, waiting for agent to fetch", "task", taskID, "job", req.JobName, "step", req.StepName)
	}

	// 2-5. 等待 agent 拉取任务并执行，监听状态更新
	timeout := time.Duration(agentTask.Timeout) * time.Second
	if timeout == 0 {
		timeout = 1 * time.Hour // Default timeout
	}

	result, err := am.WaitForTaskCompletion(ctx, taskID, timeout)
	if err != nil {
		return nil, fmt.Errorf("wait for task completion: %w", err)
	}

	return result, nil
}

// SelectAgent selects an available agent based on the selector criteria
func (am *AgentManager) SelectAgent(ctx context.Context, selector *AgentSelector) (string, error) {
	if selector == nil {
		// No selector specified, use any available agent
		// TODO: Implement agent selection logic
		return "", fmt.Errorf("agent selection not implemented")
	}

	// TODO: Implement agent selection based on:
	// 1. Agent status (online, idle)
	// 2. Label matching (matchLabels and matchExpressions)
	// 3. Resource availability
	// 4. Load balancing

	return "", fmt.Errorf("agent selection not implemented")
}

// convertStepToTask converts a pipeline step to an agent task
func (am *AgentManager) convertStepToTask(req *StepExecutionRequest) (*agentv1.Task, error) {
	// Build commands from step action and params
	commands := am.buildCommands(req.Step)

	// Build label selector
	labelSelector := am.buildLabelSelector(req.Selector)

	// Build plugin info list
	plugins := []*agentv1.PluginInfo{
		{
			PluginId: req.Step.Uses,
			Name:     req.Step.Uses,
			// Version and other fields would be populated from plugin registry
		},
	}

	// Parse timeout
	timeout := int32(3600) // Default 1 hour
	if req.Step.Timeout != "" {
		if duration, err := time.ParseDuration(req.Step.Timeout); err == nil {
			timeout = int32(duration.Seconds())
		}
	}

	task := &agentv1.Task{
		JobId:         fmt.Sprintf("%s-%s-%s", req.PipelineID, req.JobName, req.StepName),
		Name:          req.StepName,
		PipelineId:    req.PipelineID,
		Stage:         0, // Stage number would be determined by step order
		Commands:      commands,
		Env:           req.Env,
		Workspace:     req.Workspace,
		Timeout:       timeout,
		LabelSelector: labelSelector,
		Plugins:       plugins,
	}

	return task, nil
}

// buildCommands builds command list from step action and params
func (am *AgentManager) buildCommands(step *Step) []string {
	commands := []string{}

	// If step uses a plugin, the command would be plugin-specific
	// For now, we'll create a generic command structure
	if step.Action != "" {
		// Build command from action and params
		// This is a simplified version - actual implementation would depend on plugin type
		command := fmt.Sprintf("plugin execute --plugin %s --action %s", step.Uses, step.Action)
		commands = append(commands, command)
	} else {
		// Default action
		command := fmt.Sprintf("plugin execute --plugin %s", step.Uses)
		commands = append(commands, command)
	}

	return commands
}

// buildLabelSelector builds agent label selector from step selector
func (am *AgentManager) buildLabelSelector(selector *AgentSelector) *agentv1.LabelSelector {
	if selector == nil {
		return nil
	}

	labelSelector := &agentv1.LabelSelector{
		MatchLabels: selector.MatchLabels,
	}

	// Convert match expressions (from string operator to agent LabelOperator enum)
	if len(selector.MatchExpressions) > 0 {
		expressions := make([]*agentv1.LabelSelectorRequirement, 0, len(selector.MatchExpressions))
		for _, expr := range selector.MatchExpressions {
			operator := am.convertOperator(expr.Operator)
			expressions = append(expressions, &agentv1.LabelSelectorRequirement{
				Key:      expr.Key,
				Operator: operator,
				Values:   expr.Values,
			})
		}
		labelSelector.MatchExpressions = expressions
	}

	return labelSelector
}

// convertOperator converts string operator to agent LabelOperator enum
func (am *AgentManager) convertOperator(op string) agentv1.LabelOperator {
	switch op {
	case "In":
		return agentv1.LabelOperator_LABEL_OPERATOR_IN
	case "NotIn":
		return agentv1.LabelOperator_LABEL_OPERATOR_NOT_IN
	case "Exists":
		return agentv1.LabelOperator_LABEL_OPERATOR_EXISTS
	case "NotExists":
		return agentv1.LabelOperator_LABEL_OPERATOR_NOT_EXISTS
	case "Gt":
		return agentv1.LabelOperator_LABEL_OPERATOR_GT
	case "Lt":
		return agentv1.LabelOperator_LABEL_OPERATOR_LT
	default:
		return agentv1.LabelOperator_LABEL_OPERATOR_UNSPECIFIED
	}
}

// createTask creates a task in the task service
func (am *AgentManager) createTask(ctx context.Context, req *StepExecutionRequest, agentTask *agentv1.Task) (string, error) {
	if am.taskClient == nil {
		return "", fmt.Errorf("task client is not initialized")
	}

	// Convert agent task to task service request
	createReq := &taskv1.CreateTaskRequest{
		Name:         agentTask.Name,
		PipelineId:   agentTask.PipelineId,
		Stage:        agentTask.Stage,
		Commands:     agentTask.Commands,
		Env:          agentTask.Env,
		Workspace:    agentTask.Workspace,
		Timeout:      agentTask.Timeout,
		AllowFailure: req.Step.ContinueOnError,
		RetryCount:   0, // Retry is handled at step level
	}

	// Convert label selector
	if agentTask.LabelSelector != nil {
		createReq.LabelSelector = &taskv1.LabelSelector{
			MatchLabels: agentTask.LabelSelector.MatchLabels,
		}
		if len(agentTask.LabelSelector.MatchExpressions) > 0 {
			expressions := make([]*taskv1.LabelSelectorRequirement, 0, len(agentTask.LabelSelector.MatchExpressions))
			for _, expr := range agentTask.LabelSelector.MatchExpressions {
				// Convert agent LabelOperator to task LabelOperator
				taskOp := am.convertAgentOperatorToTaskOperator(expr.Operator)
				expressions = append(expressions, &taskv1.LabelSelectorRequirement{
					Key:      expr.Key,
					Operator: taskOp,
					Values:   expr.Values,
				})
			}
			createReq.LabelSelector.MatchExpressions = expressions
		}
	}

	// Create task
	resp, err := am.taskClient.CreateTask(ctx, createReq)
	if err != nil {
		return "", fmt.Errorf("create task via task service: %w", err)
	}

	if !resp.Success {
		return "", fmt.Errorf("create task failed: %s", resp.Message)
	}

	return resp.TaskId, nil
}

// convertAgentOperatorToTaskOperator converts agent LabelOperator to task LabelOperator
func (am *AgentManager) convertAgentOperatorToTaskOperator(op agentv1.LabelOperator) taskv1.LabelOperator {
	switch op {
	case agentv1.LabelOperator_LABEL_OPERATOR_IN:
		return taskv1.LabelOperator_LABEL_OPERATOR_IN
	case agentv1.LabelOperator_LABEL_OPERATOR_NOT_IN:
		return taskv1.LabelOperator_LABEL_OPERATOR_NOT_IN
	case agentv1.LabelOperator_LABEL_OPERATOR_EXISTS:
		return taskv1.LabelOperator_LABEL_OPERATOR_EXISTS
	case agentv1.LabelOperator_LABEL_OPERATOR_NOT_EXISTS:
		return taskv1.LabelOperator_LABEL_OPERATOR_NOT_EXISTS
	case agentv1.LabelOperator_LABEL_OPERATOR_GT:
		return taskv1.LabelOperator_LABEL_OPERATOR_GT
	case agentv1.LabelOperator_LABEL_OPERATOR_LT:
		return taskv1.LabelOperator_LABEL_OPERATOR_LT
	default:
		return taskv1.LabelOperator_LABEL_OPERATOR_UNSPECIFIED
	}
}

// WaitForTaskCompletion waits for a task to complete on an agent
// It polls the task status from task service and waits for completion
func (am *AgentManager) WaitForTaskCompletion(ctx context.Context, taskID string, timeout time.Duration) (*StepExecutionResult, error) {
	if am.taskClient == nil {
		return nil, fmt.Errorf("task client is not initialized")
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Poll interval
	pollInterval := 2 * time.Second
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	result := &StepExecutionResult{
		StartTime: time.Now(),
		Success:   false,
		Metrics:   make(map[string]string),
	}

	for {
		select {
		case <-ctx.Done():
			// Timeout or context cancelled
			if ctx.Err() == context.DeadlineExceeded {
				// Mark task as timeout
				if err := am.cancelTask(ctx, taskID, "task execution timeout"); err != nil {
					if am.logger.Log != nil {
						am.logger.Log.Warnw("failed to cancel timeout task", "task", taskID, "error", err)
					}
				}
				return nil, fmt.Errorf("task %s execution timeout after %v", taskID, timeout)
			}
			return nil, ctx.Err()

		case <-ticker.C:
			// Poll task status
			task, err := am.getTaskStatus(ctx, taskID)
			if err != nil {
				if am.logger.Log != nil {
					am.logger.Log.Warnw("failed to get task status", "task", taskID, "error", err)
				}
				continue
			}

			if task == nil {
				continue
			}

			// Update result based on task status
			switch task.Status {
			case taskv1.TaskStatus_TASK_STATUS_SUCCESS:
				// Task completed successfully
				result.Success = true
				result.ExitCode = task.ExitCode
				if task.FinishedAt > 0 {
					result.EndTime = time.Unix(task.FinishedAt/1000, (task.FinishedAt%1000)*1000000)
				} else {
					result.EndTime = time.Now()
				}
				if task.StartedAt > 0 {
					result.StartTime = time.Unix(task.StartedAt/1000, (task.StartedAt%1000)*1000000)
				}
				if am.logger.Log != nil {
					am.logger.Log.Infow("task completed successfully", "task", taskID)
				}
				return result, nil

			case taskv1.TaskStatus_TASK_STATUS_FAILED:
				// Task failed
				result.Success = false
				result.ExitCode = task.ExitCode
				result.Error = task.ErrorMessage
				if task.FinishedAt > 0 {
					result.EndTime = time.Unix(task.FinishedAt/1000, (task.FinishedAt%1000)*1000000)
				} else {
					result.EndTime = time.Now()
				}
				if task.StartedAt > 0 {
					result.StartTime = time.Unix(task.StartedAt/1000, (task.StartedAt%1000)*1000000)
				}
				if am.logger.Log != nil {
					am.logger.Log.Errorw("task failed", "task", taskID, "error", task.ErrorMessage)
				}
				return result, fmt.Errorf("task execution failed: %s", task.ErrorMessage)

			case taskv1.TaskStatus_TASK_STATUS_CANCELLED:
				// Task cancelled
				result.Success = false
				result.ExitCode = -1
				result.Error = "task cancelled"
				result.EndTime = time.Now()
				if am.logger.Log != nil {
					am.logger.Log.Warnw("task was cancelled", "task", taskID)
				}
				return result, fmt.Errorf("task was cancelled")

			case taskv1.TaskStatus_TASK_STATUS_TIMEOUT:
				// Task timeout
				result.Success = false
				result.ExitCode = -1
				result.Error = "task execution timeout"
				result.EndTime = time.Now()
				if am.logger.Log != nil {
					am.logger.Log.Warnw("task timed out", "task", taskID)
				}
				return result, fmt.Errorf("task execution timeout")

			case taskv1.TaskStatus_TASK_STATUS_RUNNING:
				// Task is running, continue polling
				if task.StartedAt > 0 && result.StartTime.IsZero() {
					result.StartTime = time.Unix(task.StartedAt/1000, (task.StartedAt%1000)*1000000)
				}
				continue

			case taskv1.TaskStatus_TASK_STATUS_PENDING, taskv1.TaskStatus_TASK_STATUS_QUEUED:
				// Task is pending or queued, continue polling
				continue

			default:
				// Unknown status, continue polling
				continue
			}
		}
	}
}

// getTaskStatus gets task status from task service
func (am *AgentManager) getTaskStatus(ctx context.Context, taskID string) (*taskv1.TaskDetail, error) {
	req := &taskv1.GetTaskRequest{
		TaskId: taskID,
	}

	resp, err := am.taskClient.GetTask(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("get task status: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("get task failed: %s", resp.Message)
	}

	return resp.Task, nil
}

// cancelTask cancels a task
func (am *AgentManager) cancelTask(ctx context.Context, taskID, reason string) error {
	if am.taskClient == nil {
		return fmt.Errorf("task client is not initialized")
	}

	req := &taskv1.CancelTaskRequest{
		TaskId: taskID,
		Reason: reason,
	}

	resp, err := am.taskClient.CancelTask(ctx, req)
	if err != nil {
		return fmt.Errorf("cancel task: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("cancel task failed: %s", resp.Message)
	}

	return nil
}

// CancelTask cancels a running task on an agent
func (am *AgentManager) CancelTask(ctx context.Context, agentID, taskID, reason string) error {
	if am.agentClient == nil {
		return fmt.Errorf("agent client is not initialized")
	}

	req := &agentv1.CancelTaskRequest{
		AgentId: agentID,
		JobId:   taskID,
		Reason:  reason,
	}

	_, err := am.agentClient.CancelTask(ctx, req)
	if err != nil {
		return fmt.Errorf("cancel task: %w", err)
	}

	if am.logger.Log != nil {
		am.logger.Log.Infow("cancelled task on agent", "task", taskID, "agent", agentID, "reason", reason)
	}

	return nil
}

// GetAgentStatus gets the status of an agent
// Agent status is updated from heartbeat reports, so we retrieve from cache
func (am *AgentManager) GetAgentStatus(ctx context.Context, agentID string) (*AgentStatus, error) {
	if agentID == "" {
		return nil, fmt.Errorf("agent ID is required")
	}

	am.statusCacheMu.RLock()
	status, exists := am.agentStatusCache[agentID]
	am.statusCacheMu.RUnlock()

	if !exists {
		// Agent not found in cache, return offline status
		return &AgentStatus{
			AgentID:           agentID,
			Status:            "offline",
			RunningJobsCount:  0,
			MaxConcurrentJobs: 0,
			Metrics:           make(map[string]string),
			Labels:            make(map[string]string),
			LastHeartbeat:     time.Time{},
		}, nil
	}

	// Check if heartbeat is stale (more than 5 minutes old)
	if time.Since(status.LastHeartbeat) > 5*time.Minute {
		// Mark as offline if heartbeat is stale
		status.Status = "offline"
	}

	return status, nil
}

// UpdateAgentStatusFromHeartbeat updates agent status from heartbeat report
// This should be called by the agent service when processing heartbeat requests
func (am *AgentManager) UpdateAgentStatusFromHeartbeat(req *agentv1.HeartbeatRequest) {
	if req == nil || req.AgentId == "" {
		return
	}

	am.statusCacheMu.Lock()
	defer am.statusCacheMu.Unlock()

	status := &AgentStatus{
		AgentID:           req.AgentId,
		Status:            am.convertAgentStatusToString(req.Status),
		RunningJobsCount:  req.RunningJobsCount,
		MaxConcurrentJobs: req.MaxConcurrentJobs,
		Metrics:           req.Metrics,
		Labels:            req.Labels,
		LastHeartbeat:     time.Now(),
	}

	am.agentStatusCache[req.AgentId] = status

	if am.logger.Log != nil {
		am.logger.Log.Debugw("updated agent status", "agent", req.AgentId, "status", status.Status, "running_jobs", req.RunningJobsCount, "max_jobs", req.MaxConcurrentJobs)
	}
}

// convertAgentStatusToString converts agent status enum to string
func (am *AgentManager) convertAgentStatusToString(status agentv1.AgentStatus) string {
	switch status {
	case agentv1.AgentStatus_AGENT_STATUS_ONLINE:
		return "online"
	case agentv1.AgentStatus_AGENT_STATUS_OFFLINE:
		return "offline"
	case agentv1.AgentStatus_AGENT_STATUS_BUSY:
		return "busy"
	case agentv1.AgentStatus_AGENT_STATUS_IDLE:
		return "idle"
	default:
		return "unknown"
	}
}

// ListAgents lists all agents with their status
func (am *AgentManager) ListAgents() []*AgentStatus {
	am.statusCacheMu.RLock()
	defer am.statusCacheMu.RUnlock()

	agents := make([]*AgentStatus, 0, len(am.agentStatusCache))
	for _, status := range am.agentStatusCache {
		// Check if heartbeat is stale
		if time.Since(status.LastHeartbeat) > 5*time.Minute {
			status.Status = "offline"
		}
		agents = append(agents, status)
	}

	return agents
}

// RemoveAgentStatus removes agent status from cache (when agent unregisters)
func (am *AgentManager) RemoveAgentStatus(agentID string) {
	am.statusCacheMu.Lock()
	defer am.statusCacheMu.Unlock()

	delete(am.agentStatusCache, agentID)

	if am.logger.Log != nil {
		am.logger.Log.Debugw("removed agent from status cache", "agent", agentID)
	}
}

// AgentStatus represents agent status information
type AgentStatus struct {
	AgentID           string
	Status            string
	RunningJobsCount  int32
	MaxConcurrentJobs int32
	Metrics           map[string]string
	Labels            map[string]string
	LastHeartbeat     time.Time
}
