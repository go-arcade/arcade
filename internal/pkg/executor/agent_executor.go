package executor

import (
	"context"
	"fmt"
	"time"

	"github.com/go-arcade/arcade/pkg/log"
)

// AgentExecutor agent 执行器
// 通过 agent manager 在远程 agent 上执行 step
type AgentExecutor struct {
	agentManager AgentManager
	logger       log.Logger
}

// AgentManager 定义 agent manager 接口
// 避免循环依赖，使用接口而不是具体类型
type AgentManager interface {
	ExecuteStepOnAgent(ctx context.Context, req *AgentExecutionRequest) (*AgentExecutionResult, error)
}

// AgentExecutionRequest agent 执行请求
type AgentExecutionRequest struct {
	PipelineID string
	BuildID    string
	JobName    string
	StepName   string
	Step       *AgentStep
	StepIndex  int
	Workspace  string
	Env        map[string]string
	Selector   *AgentSelector
}

// AgentStep agent step 信息
type AgentStep struct {
	Name            string
	Uses            string
	Action          string
	Args            map[string]any
	Env             map[string]string
	Timeout         string
	RunOnAgent      bool
	AgentSelector   *AgentSelector
	ContinueOnError bool
}

// AgentExecutionResult agent 执行结果
type AgentExecutionResult struct {
	Success   bool
	ExitCode  int32
	Error     string
	StartTime time.Time
	EndTime   time.Time
	Metrics   map[string]string
}

// NewAgentExecutor 创建 agent 执行器
func NewAgentExecutor(agentManager AgentManager, logger log.Logger) *AgentExecutor {
	return &AgentExecutor{
		agentManager: agentManager,
		logger:       logger,
	}
}

// Name 返回执行器名称
func (e *AgentExecutor) Name() string {
	return "agent"
}

// CanExecute 检查是否可以执行
// Agent 执行器可以执行所有指定了 RunOnAgent 的 step
func (e *AgentExecutor) CanExecute(req *ExecutionRequest) bool {
	if req == nil || req.Step == nil {
		return false
	}
	// 需要 RunOnAgent 为 true，并且 agent manager 可用
	return req.Step.RunOnAgent && e.agentManager != nil
}

// Execute 执行 step
func (e *AgentExecutor) Execute(ctx context.Context, req *ExecutionRequest) (*ExecutionResult, error) {
	result := NewExecutionResult(e.Name())

	if req.Step == nil {
		err := fmt.Errorf("step is nil")
		result.Complete(false, -1, err)
		return result, err
	}

	if e.agentManager == nil {
		err := fmt.Errorf("agent manager is not available")
		result.Complete(false, -1, err)
		return result, err
	}

	// 转换 step 信息
	agentStep := &AgentStep{
		Name:            req.Step.Name,
		Uses:            req.Step.Uses,
		Action:          req.Step.Action,
		Args:            req.Step.Args,
		Env:             req.Step.Env,
		Timeout:         req.Step.Timeout,
		RunOnAgent:      req.Step.RunOnAgent,
		AgentSelector:   req.Step.AgentSelector,
		ContinueOnError: req.Options != nil && req.Options.ContinueOnError,
	}

	// 构建 agent 执行请求
	agentReq := &AgentExecutionRequest{
		PipelineID: req.Pipeline.Namespace,
		BuildID:    "", // TODO: 从 context 获取
		JobName:    req.Job.Name,
		StepName:   req.Step.Name,
		Step:       agentStep,
		StepIndex:  0, // TODO: 从 context 获取
		Workspace:  req.Workspace,
		Env:        req.Env,
		Selector:   req.Step.AgentSelector,
	}

	// 应用超时
	if req.Options != nil && req.Options.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, req.Options.Timeout)
		defer cancel()
	}

	// 执行
	agentResult, err := e.agentManager.ExecuteStepOnAgent(ctx, agentReq)
	if err != nil {
		err = fmt.Errorf("agent execution failed: %w", err)
		result.Complete(false, -1, err)
		return result, err
	}

	// 转换结果
	result.Success = agentResult.Success
	result.ExitCode = agentResult.ExitCode
	result.Error = agentResult.Error
	result.StartTime = agentResult.StartTime
	result.EndTime = agentResult.EndTime
	result.Duration = agentResult.EndTime.Sub(agentResult.StartTime)

	// 设置元数据
	if agentResult.Metrics != nil {
		for k, v := range agentResult.Metrics {
			result.WithMetadata(k, v)
		}
	}

	result.Complete(result.Success, result.ExitCode, nil)

	if e.logger.Log != nil {
		e.logger.Log.Debugw("agent execution completed",
			"step", req.Step.Name,
			"success", result.Success,
			"exit_code", result.ExitCode,
			"duration", result.Duration)
	}

	return result, nil
}
