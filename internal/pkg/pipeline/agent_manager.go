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

package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	agentv1 "github.com/go-arcade/arcade/api/agent/v1"
	taskv1 "github.com/go-arcade/arcade/api/task/v1"
	"github.com/go-arcade/arcade/internal/pkg/pipeline/spec"
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
	Step       *spec.Step
	StepIndex  int // Index of step in job (0-based), used for stage calculation
	Workspace  string
	Env        map[string]string
	Selector   *spec.AgentSelector
	Context    *ExecutionContext // Execution context for variable resolution
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
func (am *AgentManager) SelectAgent(ctx context.Context, selector *spec.AgentSelector) (string, error) {
	// Get all available agents from cache
	am.statusCacheMu.RLock()
	availableAgents := make([]*AgentStatus, 0, len(am.agentStatusCache))
	for _, status := range am.agentStatusCache {
		// Filter out offline agents and agents with stale heartbeats
		if time.Since(status.LastHeartbeat) > 5*time.Minute {
			continue
		}
		// Only consider online or idle agents
		if status.Status == "online" || status.Status == "idle" {
			availableAgents = append(availableAgents, status)
		}
	}
	am.statusCacheMu.RUnlock()

	if len(availableAgents) == 0 {
		return "", fmt.Errorf("no available agents found")
	}

	// If no selector specified, return the first available agent
	if selector == nil {
		return availableAgents[0].AgentID, nil
	}

	// Filter agents by label selector
	matchedAgents := make([]*AgentStatus, 0)
	for _, agent := range availableAgents {
		if am.matchAgentLabels(agent, selector) {
			matchedAgents = append(matchedAgents, agent)
		}
	}

	if len(matchedAgents) == 0 {
		return "", fmt.Errorf("no agents match the selector criteria")
	}

	// Select agent with least running jobs (load balancing)
	selectedAgent := matchedAgents[0]
	for _, agent := range matchedAgents[1:] {
		if agent.RunningJobsCount < selectedAgent.RunningJobsCount {
			selectedAgent = agent
		}
	}

	return selectedAgent.AgentID, nil
}

// matchAgentLabels checks if an agent matches the label selector criteria
func (am *AgentManager) matchAgentLabels(agent *AgentStatus, selector *spec.AgentSelector) bool {
	agentLabels := agent.Labels

	// If selector requires labels but agent has no labels, it doesn't match
	// Note: Agent labels should be loaded from database or provided during registration
	// For now, if agent.Labels is nil/empty and selector requires labels, skip this agent
	if len(selector.MatchLabels) > 0 || len(selector.MatchExpressions) > 0 {
		if len(agentLabels) == 0 {
			// Agent labels not available in cache, cannot match
			// TODO: Query agent labels from database if needed
			if am.logger.Log != nil {
				am.logger.Log.Debugw("agent has no labels in cache, skipping label match",
					"agent", agent.AgentID)
			}
			return false
		}
	}

	// Check matchLabels (all must match)
	if len(selector.MatchLabels) > 0 {
		for key, value := range selector.MatchLabels {
			if agentValue, ok := agentLabels[key]; !ok || agentValue != value {
				return false
			}
		}
	}

	// Check matchExpressions (all must match)
	if len(selector.MatchExpressions) > 0 {
		for _, expr := range selector.MatchExpressions {
			if !am.matchExpression(agentLabels, expr) {
				return false
			}
		}
	}

	return true
}

// matchExpression checks if agent labels match a single expression
func (am *AgentManager) matchExpression(agentLabels map[string]string, expr spec.LabelExpression) bool {
	agentValue, exists := agentLabels[expr.Key]

	switch expr.Operator {
	case "In":
		if !exists {
			return false
		}
		return slices.Contains(expr.Values, agentValue)

	case "NotIn":
		if !exists {
			return true
		}
		return !slices.Contains(expr.Values, agentValue)

	case "Exists":
		return exists

	case "NotExists":
		return !exists

	case "Gt":
		if !exists {
			return false
		}
		// Simple numeric comparison (can be enhanced)
		if len(expr.Values) > 0 {
			// TODO: Implement proper numeric comparison
			// For now, just check if value exists
			return true
		}
		return false

	case "Lt":
		if !exists {
			return false
		}
		// Simple numeric comparison (can be enhanced)
		if len(expr.Values) > 0 {
			// TODO: Implement proper numeric comparison
			// For now, just check if value exists
			return true
		}
		return false

	default:
		return false
	}
}

// convertStepToTask converts a pipeline step to an agent task
func (am *AgentManager) convertStepToTask(req *StepExecutionRequest) (*agentv1.Task, error) {
	// Resolve variables in step args if context is provided
	var resolvedArgs map[string]any
	if req.Context != nil && len(req.Step.Args) > 0 {
		resolvedArgs = req.Context.ResolveVariables(req.Step.Args)
	} else {
		resolvedArgs = req.Step.Args
	}

	// Serialize args to JSON for passing via environment variable
	var argsJSON []byte
	var err error
	if len(resolvedArgs) > 0 {
		argsJSON, err = json.Marshal(resolvedArgs)
		if err != nil {
			return nil, fmt.Errorf("marshal step args: %w", err)
		}
	}

	// Build commands from step action
	var commands []string
	if req.Context != nil {
		commands, err = am.buildCommands(req.Step, req.Context)
		if err != nil {
			return nil, fmt.Errorf("build commands: %w", err)
		}
	} else {
		// Fallback if context is not provided
		action := req.Step.Action
		if action == "" {
			action = "Execute"
		}
		commands = []string{fmt.Sprintf("plugin execute --plugin %s --action %s", req.Step.Uses, action)}
	}

	// Prepare environment variables (include plugin params)
	env := make(map[string]string)
	for k, v := range req.Env {
		env[k] = v
	}
	// Add plugin params as environment variable for agent to use
	if len(argsJSON) > 0 {
		env["PLUGIN_PARAMS"] = string(argsJSON)
	}
	// Add plugin action and name for agent reference
	env["PLUGIN_NAME"] = req.Step.Uses
	if req.Step.Action != "" {
		env["PLUGIN_ACTION"] = req.Step.Action
	} else {
		env["PLUGIN_ACTION"] = "Execute"
	}

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

	// Calculate stage number from step index
	stage := int32(req.StepIndex)

	task := &agentv1.Task{
		JobId:         fmt.Sprintf("%s-%s-%s", req.PipelineID, req.JobName, req.StepName),
		Name:          req.StepName,
		PipelineId:    req.PipelineID,
		Stage:         stage,
		Commands:      commands,
		Env:           env,
		Workspace:     req.Workspace,
		Timeout:       timeout,
		LabelSelector: labelSelector,
		Plugins:       plugins,
	}

	return task, nil
}

// buildCommands builds command list from step action and params
// This method calls plugin to resolve step into concrete commands before sending to agent
func (am *AgentManager) buildCommands(step *spec.Step, ctx *ExecutionContext) ([]string, error) {
	if ctx == nil || ctx.PluginManager == nil {
		return nil, fmt.Errorf("execution context or plugin manager is not available")
	}

	// Get plugin instance
	pluginInstance, err := ctx.PluginManager.GetPlugin(step.Uses)
	if err != nil {
		return nil, fmt.Errorf("plugin not found: %s: %w", step.Uses, err)
	}

	// Determine action (default to "Execute" if not specified)
	action := step.Action
	if action == "" {
		action = "Execute"
	}

	// Resolve params with variable substitution
	resolvedParams := ctx.ResolveVariables(step.Args)

	// Prepare params JSON
	paramsJSON, err := json.Marshal(resolvedParams)
	if err != nil {
		return nil, fmt.Errorf("marshal params: %w", err)
	}

	// Prepare opts JSON with dry-run flag to get commands instead of executing
	opts := map[string]any{
		"workspace":       ctx.StepWorkspace("", step.Name), // Will be set properly when task is created
		"dry_run":         true,                             // Request plugin to return commands instead of executing
		"build_for_agent": true,                             // Indicate this is for agent execution
	}
	optsJSON, err := json.Marshal(opts)
	if err != nil {
		return nil, fmt.Errorf("marshal opts: %w", err)
	}

	// Try to call plugin's BuildCommands action first (if supported)
	// This allows plugins to return command list without executing
	commandsJSON, err := pluginInstance.Execute("BuildCommands", paramsJSON, optsJSON)
	if err == nil {
		// Plugin supports BuildCommands, parse the returned commands
		var commands []string
		if err := json.Unmarshal(commandsJSON, &commands); err != nil {
			return nil, fmt.Errorf("unmarshal commands from plugin: %w", err)
		}
		if len(commands) > 0 {
			return commands, nil
		}
	}

	// Fallback: For plugins that don't support BuildCommands, build commands based on plugin type
	// This is a compatibility layer for existing plugins
	pluginName := pluginInstance.Name()
	switch pluginName {
	case "shell":
		return am.buildShellCommands(step, resolvedParams, ctx)
	default:
		// For other plugins, try to build command based on common patterns
		return am.buildGenericPluginCommand(step, action), nil
	}
}

// buildShellCommands builds shell commands from step params
func (am *AgentManager) buildShellCommands(step *spec.Step, params map[string]any, ctx *ExecutionContext) ([]string, error) {
	commands := []string{}

	// Determine shell action
	action := step.Action
	if action == "" {
		action = "command" // Default to command for shell plugin
	}

	// Build command based on action type
	switch action {
	case "script":
		// For script action, create a script file and execute it
		if script, ok := params["script"].(string); ok && script != "" {
			// Create a command that writes script to file and executes it
			scriptFile := "/tmp/script.sh"
			commands = append(commands, fmt.Sprintf("cat > %s << 'SCRIPT_EOF'\n%s\nSCRIPT_EOF", scriptFile, script))
			commands = append(commands, fmt.Sprintf("chmod +x %s", scriptFile))

			// Add script arguments if provided
			scriptArgs := ""
			if args, ok := params["args"].([]any); ok && len(args) > 0 {
				argStrs := make([]string, len(args))
				for i, arg := range args {
					argStrs[i] = fmt.Sprintf("%v", arg)
				}
				scriptArgs = " " + strings.Join(argStrs, " ")
			}

			commands = append(commands, fmt.Sprintf("sh %s%s", scriptFile, scriptArgs))
			commands = append(commands, fmt.Sprintf("rm -f %s", scriptFile))
		} else {
			return nil, fmt.Errorf("script parameter is required for shell script action")
		}

	case "command":
		// For command action, execute directly
		if command, ok := params["command"].(string); ok && command != "" {
			// Add command arguments if provided
			if args, ok := params["args"].([]any); ok && len(args) > 0 {
				argStrs := make([]string, len(args))
				for i, arg := range args {
					argStrs[i] = fmt.Sprintf("%v", arg)
				}
				commands = append(commands, fmt.Sprintf("%s %s", command, strings.Join(argStrs, " ")))
			} else {
				commands = append(commands, command)
			}
		} else {
			return nil, fmt.Errorf("command parameter is required for shell command action")
		}

	default:
		return nil, fmt.Errorf("unsupported shell action: %s", action)
	}

	return commands, nil
}

// buildGenericPluginCommand builds a generic plugin command as fallback
func (am *AgentManager) buildGenericPluginCommand(step *spec.Step, action string) []string {
	// This is a fallback for plugins that don't support BuildCommands
	// In this case, we still send plugin command to agent
	// Agent will need to have plugin runtime to execute it
	// TODO: This should be improved to fully resolve commands in main program
	return []string{fmt.Sprintf("plugin execute --plugin %s --action %s", step.Uses, action)}
}

// buildLabelSelector builds agent label selector from step selector
func (am *AgentManager) buildLabelSelector(selector *spec.AgentSelector) *agentv1.LabelSelector {
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
		AgentID:          req.AgentId,
		Status:           am.convertAgentStatusToString(req.Status),
		RunningJobsCount: req.RunningJobsCount,
		LastHeartbeat:    time.Now(),
	}

	am.agentStatusCache[req.AgentId] = status

	if am.logger.Log != nil {
		am.logger.Log.Debugw("updated agent status", "agent", req.AgentId, "status", status.Status, "running_jobs", req.RunningJobsCount)
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
	Labels            map[string]string // Agent labels for matching
	LastHeartbeat     time.Time
}
