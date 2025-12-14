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

type TaskStatus string

const (
	TaskPending  TaskStatus = "PENDING"
	TaskQueued   TaskStatus = "QUEUED"
	TaskRunning  TaskStatus = "RUNNING"
	TaskSuccess  TaskStatus = "SUCCESS"
	TaskFailed   TaskStatus = "FAILED"
	TaskSkipped  TaskStatus = "SKIPPED"
	TaskTimeout  TaskStatus = "TIMEOUT"
	TaskCanceled TaskStatus = "CANCELED"
)

// IsTerminal 判断是否为终止状态
func (ts TaskStatus) IsTerminal() bool {
	return ts == TaskSuccess || ts == TaskFailed || ts == TaskTimeout || ts == TaskCanceled || ts == TaskSkipped
}

// IsRunnable 判断是否为可运行状态
func (ts TaskStatus) IsRunnable() bool {
	return ts == TaskPending || ts == TaskQueued
}

// IsFailed 判断是否为失败状态
func (ts TaskStatus) IsFailed() bool {
	return ts == TaskFailed || ts == TaskTimeout
}

// NewTaskStateMachine 创建任务状态机
func NewTaskStateMachine() *StateMachine[TaskStatus] {
	sm := NewWithState(TaskPending)

	// 定义状态转移规则
	sm.Allow(TaskPending, TaskQueued, TaskSkipped).
		Allow(TaskQueued, TaskRunning, TaskCanceled, TaskSkipped).
		Allow(TaskRunning, TaskSuccess, TaskFailed, TaskTimeout, TaskCanceled).
		Allow(TaskFailed, TaskQueued) // 支持重试

	return sm
}

// NewTaskStateMachineWithHooks 创建带钩子的任务状态机
func NewTaskStateMachineWithHooks(
	onStart func() error,
	onComplete func(status TaskStatus) error,
) *StateMachine[TaskStatus] {
	sm := NewTaskStateMachine()

	// 进入运行状态时的钩子
	if onStart != nil {
		sm.OnEnter(TaskRunning, func(state TaskStatus) error {
			return onStart()
		})
	}

	// 进入终止状态时的钩子
	if onComplete != nil {
		sm.OnEnter(TaskSuccess, func(state TaskStatus) error {
			return onComplete(state)
		})
		sm.OnEnter(TaskFailed, func(state TaskStatus) error {
			return onComplete(state)
		})
		sm.OnEnter(TaskTimeout, func(state TaskStatus) error {
			return onComplete(state)
		})
		sm.OnEnter(TaskCanceled, func(state TaskStatus) error {
			return onComplete(state)
		})
		sm.OnEnter(TaskSkipped, func(state TaskStatus) error {
			return onComplete(state)
		})
	}

	return sm
}
