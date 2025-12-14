package executor

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-arcade/arcade/pkg/log"
	"github.com/go-arcade/arcade/pkg/plugin"
)

// LocalExecutor 本地执行器
// 通过 plugin 在主程序内执行 step
type LocalExecutor struct {
	pluginManager *plugin.Manager
	logger        log.Logger
}

// NewLocalExecutor 创建本地执行器
func NewLocalExecutor(pluginManager *plugin.Manager, logger log.Logger) *LocalExecutor {
	return &LocalExecutor{
		pluginManager: pluginManager,
		logger:        logger,
	}
}

// Name 返回执行器名称
func (e *LocalExecutor) Name() string {
	return "local"
}

// CanExecute 检查是否可以执行
// 本地执行器可以执行所有不需要 RunOnAgent 的 step
func (e *LocalExecutor) CanExecute(req *ExecutionRequest) bool {
	if req == nil || req.Step == nil {
		return false
	}
	// 如果 step 指定了 RunOnAgent，则不应该使用本地执行器
	return !req.Step.RunOnAgent
}

// Execute 执行 step
func (e *LocalExecutor) Execute(ctx context.Context, req *ExecutionRequest) (*ExecutionResult, error) {
	result := NewExecutionResult(e.Name())

	if req.Step == nil {
		err := fmt.Errorf("step is nil")
		result.Complete(false, -1, err)
		return result, err
	}

	// 获取 plugin
	pluginClient, err := e.pluginManager.GetPlugin(req.Step.Uses)
	if err != nil {
		err = fmt.Errorf("plugin not found: %s: %w", req.Step.Uses, err)
		result.Complete(false, -1, err)
		return result, err
	}

	// 确定 action（默认为 "Execute"）
	action := req.Step.Action
	if action == "" {
		action = "Execute"
	}

	// 解析参数
	var resolvedParams map[string]any
	if req.Step.Args != nil {
		resolvedParams = req.Step.Args
		// TODO: 支持变量解析
	} else {
		resolvedParams = make(map[string]any)
	}

	// 准备参数 JSON
	paramsJSON, err := json.Marshal(resolvedParams)
	if err != nil {
		err = fmt.Errorf("marshal params: %w", err)
		result.Complete(false, -1, err)
		return result, err
	}

	// 准备选项 JSON（workspace 和 env）
	opts := map[string]any{
		"workspace": req.Workspace,
		"env":       req.Env,
	}
	if req.Options != nil && req.Options.Timeout > 0 {
		opts["timeout"] = req.Options.Timeout.Seconds()
	}
	optsJSON, err := json.Marshal(opts)
	if err != nil {
		err = fmt.Errorf("marshal opts: %w", err)
		result.Complete(false, -1, err)
		return result, err
	}

	// 调用 plugin 方法
	// 注意：超时已通过 opts 传递给 plugin，plugin 内部会处理超时
	pluginResult, err := pluginClient.CallMethod(action, paramsJSON, optsJSON)
	if err != nil {
		err = fmt.Errorf("plugin execution failed: %w", err)
		result.Complete(false, -1, err)
		return result, err
	}

	// 解析 plugin 返回结果
	var resultData map[string]any
	if len(pluginResult) > 0 {
		if err := json.Unmarshal(pluginResult, &resultData); err == nil {
			// 提取输出信息
			if stdout, ok := resultData["stdout"].(string); ok {
				result.Output = stdout
			}
			if stderr, ok := resultData["stderr"].(string); ok {
				result.ErrorOutput = stderr
			}
			if exitCode, ok := resultData["exit_code"].(float64); ok {
				result.ExitCode = int32(exitCode)
			}
			if success, ok := resultData["success"].(bool); ok {
				result.Success = success
			}
		}
	}

	// 如果没有从结果中提取到成功状态，默认为成功
	// 注意：此时 err 一定是 nil（如果 err != nil，函数已在前面返回）
	if !result.Success {
		result.Success = true
		result.ExitCode = 0
	}

	result.Complete(result.Success, result.ExitCode, nil)

	if e.logger.Log != nil {
		e.logger.Log.Debugw("local execution completed",
			"step", req.Step.Name,
			"success", result.Success,
			"exit_code", result.ExitCode,
			"duration", result.Duration)
	}

	return result, nil
}
