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

package executor

import (
	"context"
	"testing"
)

func TestExecutorManager(t *testing.T) {
	manager := NewExecutorManager()

	// 创建测试执行器
	localExec := &mockExecutor{name: "local", canExecute: func(req *ExecutionRequest) bool {
		return req != nil && req.Step != nil && !req.Step.RunOnAgent
	}}
	agentExec := &mockExecutor{name: "agent", canExecute: func(req *ExecutionRequest) bool {
		return req != nil && req.Step != nil && req.Step.RunOnAgent
	}}

	manager.Register(localExec)
	manager.Register(agentExec)

	// 测试本地执行器选择
	localReq := &ExecutionRequest{
		Step: &StepInfo{
			Name:       "test-step",
			Uses:       "shell",
			RunOnAgent: false,
		},
	}
	executor, err := manager.SelectExecutor(localReq)
	if err != nil {
		t.Fatalf("failed to select executor: %v", err)
	}
	if executor.Name() != "local" {
		t.Errorf("expected local executor, got %s", executor.Name())
	}

	// 测试 agent 执行器选择
	agentReq := &ExecutionRequest{
		Step: &StepInfo{
			Name:       "test-step",
			Uses:       "shell",
			RunOnAgent: true,
		},
	}
	executor, err = manager.SelectExecutor(agentReq)
	if err != nil {
		t.Fatalf("failed to select executor: %v", err)
	}
	if executor.Name() != "agent" {
		t.Errorf("expected agent executor, got %s", executor.Name())
	}
}

func TestExecutionResult(t *testing.T) {
	result := NewExecutionResult("test-executor")
	if result.ExecutorName != "test-executor" {
		t.Errorf("expected executor name 'test-executor', got '%s'", result.ExecutorName)
	}

	result.Complete(true, 0, nil)
	if !result.Success {
		t.Error("expected success to be true")
	}
	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
	}
	if result.Duration == 0 {
		t.Error("expected duration to be set")
	}

	result.WithOutput("stdout", "stderr")
	if result.Output != "stdout" {
		t.Errorf("expected output 'stdout', got '%s'", result.Output)
	}
	if result.ErrorOutput != "stderr" {
		t.Errorf("expected error output 'stderr', got '%s'", result.ErrorOutput)
	}
}

// mockExecutor 用于测试的 mock 执行器
type mockExecutor struct {
	name        string
	canExecute  func(*ExecutionRequest) bool
	executeFunc func(context.Context, *ExecutionRequest) (*ExecutionResult, error)
}

func (m *mockExecutor) Name() string {
	return m.name
}

func (m *mockExecutor) CanExecute(req *ExecutionRequest) bool {
	if m.canExecute != nil {
		return m.canExecute(req)
	}
	return false
}

func (m *mockExecutor) Execute(ctx context.Context, req *ExecutionRequest) (*ExecutionResult, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, req)
	}
	result := NewExecutionResult(m.name)
	result.Complete(true, 0, nil)
	return result, nil
}
