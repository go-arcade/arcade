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

func TestTaskStatus_IsTerminal(t *testing.T) {
	tests := []struct {
		status   TaskStatus
		expected bool
	}{
		{TaskPending, false},
		{TaskQueued, false},
		{TaskRunning, false},
		{TaskSuccess, true},
		{TaskFailed, true},
		{TaskTimeout, true},
		{TaskCanceled, true},
		{TaskSkipped, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if got := tt.status.IsTerminal(); got != tt.expected {
				t.Errorf("IsTerminal() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestTaskStatus_IsRunnable(t *testing.T) {
	tests := []struct {
		status   TaskStatus
		expected bool
	}{
		{TaskPending, true},
		{TaskQueued, true},
		{TaskRunning, false},
		{TaskSuccess, false},
		{TaskFailed, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if got := tt.status.IsRunnable(); got != tt.expected {
				t.Errorf("IsRunnable() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestTaskStatus_IsFailed(t *testing.T) {
	tests := []struct {
		status   TaskStatus
		expected bool
	}{
		{TaskPending, false},
		{TaskRunning, false},
		{TaskSuccess, false},
		{TaskFailed, true},
		{TaskTimeout, true},
		{TaskCanceled, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if got := tt.status.IsFailed(); got != tt.expected {
				t.Errorf("IsFailed() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewTaskStateMachine(t *testing.T) {
	sm := NewTaskStateMachine()

	// 验证初始状态
	if sm.Current() != TaskPending {
		t.Errorf("expected initial state to be %v, got %v", TaskPending, sm.Current())
	}

	// 测试合法转移：PENDING -> QUEUED -> RUNNING -> SUCCESS
	if err := sm.TransitTo(TaskQueued); err != nil {
		t.Errorf("PENDING -> QUEUED should be valid: %v", err)
	}

	if err := sm.TransitTo(TaskRunning); err != nil {
		t.Errorf("QUEUED -> RUNNING should be valid: %v", err)
	}

	if err := sm.TransitTo(TaskSuccess); err != nil {
		t.Errorf("RUNNING -> SUCCESS should be valid: %v", err)
	}
}

func TestNewTaskStateMachine_Skip(t *testing.T) {
	sm := NewTaskStateMachine()

	// 测试跳过：PENDING -> SKIPPED
	if err := sm.TransitTo(TaskSkipped); err != nil {
		t.Errorf("PENDING -> SKIPPED should be valid: %v", err)
	}

	if sm.Current() != TaskSkipped {
		t.Errorf("expected state to be %v, got %v", TaskSkipped, sm.Current())
	}
}

func TestNewTaskStateMachine_Cancel(t *testing.T) {
	sm := NewTaskStateMachine()

	// 测试取消：PENDING -> QUEUED -> CANCELED
	sm.TransitTo(TaskQueued)

	if err := sm.TransitTo(TaskCanceled); err != nil {
		t.Errorf("QUEUED -> CANCELED should be valid: %v", err)
	}

	if sm.Current() != TaskCanceled {
		t.Errorf("expected state to be %v, got %v", TaskCanceled, sm.Current())
	}
}

func TestNewTaskStateMachine_Retry(t *testing.T) {
	sm := NewTaskStateMachine()

	// 测试重试：PENDING -> QUEUED -> RUNNING -> FAILED -> QUEUED
	sm.TransitTo(TaskQueued)
	sm.TransitTo(TaskRunning)
	sm.TransitTo(TaskFailed)

	if err := sm.TransitTo(TaskQueued); err != nil {
		t.Errorf("FAILED -> QUEUED (retry) should be valid: %v", err)
	}

	if sm.Current() != TaskQueued {
		t.Errorf("expected state to be %v after retry, got %v", TaskQueued, sm.Current())
	}
}

func TestNewTaskStateMachine_Timeout(t *testing.T) {
	sm := NewTaskStateMachine()

	// 测试超时：PENDING -> QUEUED -> RUNNING -> TIMEOUT
	sm.TransitTo(TaskQueued)
	sm.TransitTo(TaskRunning)

	if err := sm.TransitTo(TaskTimeout); err != nil {
		t.Errorf("RUNNING -> TIMEOUT should be valid: %v", err)
	}

	if sm.Current() != TaskTimeout {
		t.Errorf("expected state to be %v, got %v", TaskTimeout, sm.Current())
	}
}

func TestNewTaskStateMachineWithHooks(t *testing.T) {
	var started, completed bool
	var completionStatus TaskStatus

	sm := NewTaskStateMachineWithHooks(
		func() error {
			started = true
			return nil
		},
		func(status TaskStatus) error {
			completed = true
			completionStatus = status
			return nil
		},
	)

	// 执行状态转移
	sm.TransitTo(TaskQueued)
	sm.TransitTo(TaskRunning)

	if !started {
		t.Error("expected onStart hook to be called")
	}

	sm.TransitTo(TaskSuccess)

	if !completed {
		t.Error("expected onComplete hook to be called")
	}

	if completionStatus != TaskSuccess {
		t.Errorf("expected completion status to be %v, got %v", TaskSuccess, completionStatus)
	}
}

func TestNewTaskStateMachine_InvalidTransitions(t *testing.T) {
	sm := NewTaskStateMachine()

	// 测试非法转移
	invalidTransitions := []struct {
		from TaskStatus
		to   TaskStatus
	}{
		{TaskPending, TaskRunning},  // 必须先入队
		{TaskPending, TaskSuccess},  // 不能直接成功
		{TaskSuccess, TaskFailed},   // 终止状态不能转移
		{TaskCanceled, TaskRunning}, // 取消后不能运行
		{TaskTimeout, TaskSuccess},  // 超时不能转为成功
	}

	for _, tt := range invalidTransitions {
		t.Run(string(tt.from)+"->"+string(tt.to), func(t *testing.T) {
			sm = NewTaskStateMachine()
			sm.SetCurrent(tt.from)

			if err := sm.TransitTo(tt.to); err == nil {
				t.Errorf("%v -> %v should be invalid", tt.from, tt.to)
			}
		})
	}
}
