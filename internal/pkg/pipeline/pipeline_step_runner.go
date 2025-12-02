package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// StepRunner runs a single step
type StepRunner struct {
	ctx  *ExecutionContext
	job  *Job
	step *Step
}

// NewStepRunner creates a new step runner
func NewStepRunner(ctx *ExecutionContext, job *Job, step *Step) *StepRunner {
	return &StepRunner{ctx: ctx, job: job, step: step}
}

// Run executes the step
func (r *StepRunner) Run(ctx context.Context) error {
	r.ctx.LogStep(r.job.Name, r.step.Name, "starting step")

	// Evaluate when condition with step context
	stepContext := map[string]any{
		"job": map[string]any{
			"name": r.job.Name,
		},
		"step": map[string]any{
			"name": r.step.Name,
			"uses": r.step.Uses,
		},
	}
	if ok, err := r.ctx.EvalConditionWithContext(r.step.When, stepContext); err != nil {
		return fmt.Errorf("evaluate when condition: %w", err)
	} else if !ok {
		r.ctx.LogStep(r.job.Name, r.step.Name, "skipped (when condition false)")
		return nil
	}

	// Apply step timeout if specified
	if r.step.Timeout != "" {
		timeout, err := time.ParseDuration(r.step.Timeout)
		if err != nil {
			return fmt.Errorf("invalid timeout format: %w", err)
		}
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	// Determine retry attempts
	retries := 1
	if r.job.Retry != nil && r.job.Retry.MaxAttempts > 0 {
		retries = r.job.Retry.MaxAttempts
	}

	var lastErr error
	for attempt := 1; attempt <= retries; attempt++ {
		if attempt > 1 {
			r.ctx.LogStep(r.job.Name, r.step.Name, fmt.Sprintf("retry attempt %d/%d", attempt, retries))
			if r.job.Retry != nil && r.job.Retry.Delay != "" {
				delay, err := time.ParseDuration(r.job.Retry.Delay)
				if err == nil {
					time.Sleep(delay)
				}
			}
		}

		lastErr = r.execute(ctx)
		if lastErr == nil {
			r.ctx.LogStep(r.job.Name, r.step.Name, "completed successfully")
			return nil
		}

		r.ctx.LogStep(r.job.Name, r.step.Name, fmt.Sprintf("attempt %d failed: %v", attempt, lastErr))
	}

	// Handle continue_on_error
	if lastErr != nil && r.step.ContinueOnError {
		r.ctx.LogStep(r.job.Name, r.step.Name, "failed but continueOnError=true")
		return nil
	}

	return fmt.Errorf("step failed after %d attempts: %w", retries, lastErr)
}

// execute executes the step action
func (r *StepRunner) execute(ctx context.Context) error {
	// Check if step should run on agent
	if r.step.RunOnAgent && r.ctx.AgentManager != nil {
		return r.executeOnAgent(ctx)
	}

	// Execute locally using plugin
	return r.executeLocally(ctx)
}

// executeOnAgent executes step on a remote agent
func (r *StepRunner) executeOnAgent(ctx context.Context) error {
	// Get pipeline and build ID from context
	pipelineID := r.ctx.Pipeline.Namespace // Use namespace as pipeline ID
	buildID := ""                          // TODO: Get build ID from context

	// Resolve environment variables
	env := r.ctx.ResolveStepEnv(r.job, r.step)

	// Get workspace path
	workspace := r.ctx.StepWorkspace(r.job.Name, r.step.Name)

	// Create execution request
	req := &StepExecutionRequest{
		PipelineID: pipelineID,
		BuildID:    buildID,
		JobName:    r.job.Name,
		StepName:   r.step.Name,
		Step:       r.step,
		Workspace:  workspace,
		Env:        env,
		Selector:   r.step.AgentSelector,
	}

	// Execute on agent
	result, err := r.ctx.AgentManager.ExecuteStepOnAgent(ctx, req)
	if err != nil {
		return fmt.Errorf("execute on agent: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("step execution failed on agent: exit code %d, error: %s", result.ExitCode, result.Error)
	}

	return nil
}

// executeLocally executes step locally using plugin
func (r *StepRunner) executeLocally(ctx context.Context) error {
	pluginClient, err := r.ctx.PluginManager.GetPlugin(r.step.Uses)
	if err != nil {
		return fmt.Errorf("plugin not found: %s: %w", r.step.Uses, err)
	}

	// Determine action (default to "Execute" if not specified)
	action := r.step.Action
	if action == "" {
		action = "Execute"
	}

	// Resolve environment variables
	env := r.ctx.ResolveStepEnv(r.job, r.step)

	// Resolve params with variable substitution
	resolvedParams := r.ctx.ResolveVariables(r.step.Args)

	// Prepare params JSON
	paramsJSON, err := json.Marshal(resolvedParams)
	if err != nil {
		return fmt.Errorf("marshal params: %w", err)
	}

	// Prepare opts JSON (workspace and env)
	opts := map[string]any{
		"workspace": r.ctx.StepWorkspace(r.job.Name, r.step.Name),
		"env":       env,
	}
	optsJSON, err := json.Marshal(opts)
	if err != nil {
		return fmt.Errorf("marshal opts: %w", err)
	}

	// Call plugin method
	_, err = pluginClient.CallMethod(action, paramsJSON, optsJSON)
	if err != nil {
		return fmt.Errorf("plugin execution failed: %w", err)
	}

	return nil
}
