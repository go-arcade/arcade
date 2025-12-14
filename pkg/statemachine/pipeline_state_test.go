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

package statemachine

import "testing"

func TestPipelineStatus_IsTerminal(t *testing.T) {
	tests := []struct {
		status   PipelineStatus
		expected bool
	}{
		{PipelinePending, false},
		{PipelineRunning, false},
		{PipelinePaused, false},
		{PipelineSuccess, true},
		{PipelineFailed, true},
		{PipelineCanceled, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if got := tt.status.IsTerminal(); got != tt.expected {
				t.Errorf("IsTerminal() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPipelineStatus_IsRunning(t *testing.T) {
	tests := []struct {
		status   PipelineStatus
		expected bool
	}{
		{PipelinePending, false},
		{PipelineRunning, true},
		{PipelinePaused, false},
		{PipelineSuccess, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if got := tt.status.IsRunning(); got != tt.expected {
				t.Errorf("IsRunning() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPipelineStatus_CanResume(t *testing.T) {
	tests := []struct {
		status   PipelineStatus
		expected bool
	}{
		{PipelinePending, false},
		{PipelineRunning, false},
		{PipelinePaused, true},
		{PipelineFailed, true},
		{PipelineSuccess, false},
		{PipelineCanceled, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if got := tt.status.CanResume(); got != tt.expected {
				t.Errorf("CanResume() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewPipelineStateMachine(t *testing.T) {
	sm := NewPipelineStateMachine()

	// 验证初始状态
	if sm.Current() != PipelinePending {
		t.Errorf("expected initial state to be %v, got %v", PipelinePending, sm.Current())
	}

	// 测试正常流程：PENDING -> RUNNING -> SUCCESS
	if err := sm.TransitTo(PipelineRunning); err != nil {
		t.Errorf("PENDING -> RUNNING should be valid: %v", err)
	}

	if err := sm.TransitTo(PipelineSuccess); err != nil {
		t.Errorf("RUNNING -> SUCCESS should be valid: %v", err)
	}
}

func TestNewPipelineStateMachine_Pause(t *testing.T) {
	sm := NewPipelineStateMachine()

	// 测试暂停：PENDING -> RUNNING -> PAUSED
	sm.TransitTo(PipelineRunning)

	if err := sm.TransitTo(PipelinePaused); err != nil {
		t.Errorf("RUNNING -> PAUSED should be valid: %v", err)
	}

	if sm.Current() != PipelinePaused {
		t.Errorf("expected state to be %v, got %v", PipelinePaused, sm.Current())
	}
}

func TestNewPipelineStateMachine_Resume(t *testing.T) {
	sm := NewPipelineStateMachine()

	// 测试恢复：PENDING -> RUNNING -> PAUSED -> RUNNING
	sm.TransitTo(PipelineRunning)
	sm.TransitTo(PipelinePaused)

	if err := sm.TransitTo(PipelineRunning); err != nil {
		t.Errorf("PAUSED -> RUNNING should be valid: %v", err)
	}

	if sm.Current() != PipelineRunning {
		t.Errorf("expected state to be %v, got %v", PipelineRunning, sm.Current())
	}
}

func TestNewPipelineStateMachine_Retry(t *testing.T) {
	sm := NewPipelineStateMachine()

	// 测试重试：PENDING -> RUNNING -> FAILED -> RUNNING
	sm.TransitTo(PipelineRunning)
	sm.TransitTo(PipelineFailed)

	if err := sm.TransitTo(PipelineRunning); err != nil {
		t.Errorf("FAILED -> RUNNING (retry) should be valid: %v", err)
	}

	if sm.Current() != PipelineRunning {
		t.Errorf("expected state to be %v after retry, got %v", PipelineRunning, sm.Current())
	}
}

func TestNewPipelineStateMachine_Cancel(t *testing.T) {
	sm := NewPipelineStateMachine()

	// 测试取消：PENDING -> RUNNING -> CANCELED
	sm.TransitTo(PipelineRunning)

	if err := sm.TransitTo(PipelineCanceled); err != nil {
		t.Errorf("RUNNING -> CANCELED should be valid: %v", err)
	}

	if sm.Current() != PipelineCanceled {
		t.Errorf("expected state to be %v, got %v", PipelineCanceled, sm.Current())
	}
}

func TestNewPipelineStateMachine_CancelFromPaused(t *testing.T) {
	sm := NewPipelineStateMachine()

	// 测试从暂停状态取消：PENDING -> RUNNING -> PAUSED -> CANCELED
	sm.TransitTo(PipelineRunning)
	sm.TransitTo(PipelinePaused)

	if err := sm.TransitTo(PipelineCanceled); err != nil {
		t.Errorf("PAUSED -> CANCELED should be valid: %v", err)
	}

	if sm.Current() != PipelineCanceled {
		t.Errorf("expected state to be %v, got %v", PipelineCanceled, sm.Current())
	}
}

func TestNewPipelineStateMachineWithHooks(t *testing.T) {
	var started, paused, completed bool
	var completionStatus PipelineStatus

	sm := NewPipelineStateMachineWithHooks(
		func() error {
			started = true
			return nil
		},
		func(status PipelineStatus) error {
			completed = true
			completionStatus = status
			return nil
		},
		func() error {
			paused = true
			return nil
		},
	)

	// 执行状态转移
	sm.TransitTo(PipelineRunning)

	if !started {
		t.Error("expected onStart hook to be called")
	}

	sm.TransitTo(PipelinePaused)

	if !paused {
		t.Error("expected onPause hook to be called")
	}

	sm.TransitTo(PipelineRunning)
	sm.TransitTo(PipelineSuccess)

	if !completed {
		t.Error("expected onComplete hook to be called")
	}

	if completionStatus != PipelineSuccess {
		t.Errorf("expected completion status to be %v, got %v", PipelineSuccess, completionStatus)
	}
}

func TestNewPipelineStateMachine_InvalidTransitions(t *testing.T) {
	// 测试非法转移
	invalidTransitions := []struct {
		from PipelineStatus
		to   PipelineStatus
	}{
		{PipelinePending, PipelineSuccess},  // 不能直接成功
		{PipelinePending, PipelineFailed},   // 不能直接失败
		{PipelinePending, PipelinePaused},   // 不能在未运行时暂停
		{PipelineSuccess, PipelineRunning},  // 成功后不能重新运行
		{PipelineCanceled, PipelineRunning}, // 取消后不能运行
	}

	for _, tt := range invalidTransitions {
		t.Run(string(tt.from)+"->"+string(tt.to), func(t *testing.T) {
			sm := NewPipelineStateMachine()
			sm.SetCurrent(tt.from)

			if err := sm.TransitTo(tt.to); err == nil {
				t.Errorf("%v -> %v should be invalid", tt.from, tt.to)
			}
		})
	}
}

func TestPipelineStateMachine_CompleteWorkflow(t *testing.T) {
	sm := NewPipelineStateMachine()

	// 模拟完整的工作流程
	workflow := []PipelineStatus{
		PipelineRunning, // 开始运行
		PipelinePaused,  // 暂停
		PipelineRunning, // 恢复
		PipelineFailed,  // 失败
		PipelineRunning, // 重试
		PipelineSuccess, // 成功
	}

	for i, state := range workflow {
		if err := sm.TransitTo(state); err != nil {
			t.Errorf("step %d: failed to transit to %v: %v", i, state, err)
		}

		if sm.Current() != state {
			t.Errorf("step %d: expected state to be %v, got %v", i, state, sm.Current())
		}
	}

	// 验证历史记录
	history := sm.History()
	if len(history) != len(workflow) {
		t.Errorf("expected %d history records, got %d", len(workflow), len(history))
	}
}
