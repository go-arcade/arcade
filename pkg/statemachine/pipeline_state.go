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

type PipelineStatus string

const (
	PipelinePending  PipelineStatus = "PENDING"
	PipelineRunning  PipelineStatus = "RUNNING"
	PipelineFailed   PipelineStatus = "FAILED"
	PipelineSuccess  PipelineStatus = "SUCCESS"
	PipelineCanceled PipelineStatus = "CANCELED"
	PipelinePaused   PipelineStatus = "PAUSED"
)

// Deprecated: Use FSM instead.
// IsTerminal 判断是否为终止状态
func (ps PipelineStatus) IsTerminal() bool {
	return ps == PipelineSuccess || ps == PipelineFailed || ps == PipelineCanceled
}

// Deprecated: Use FSM instead.
// IsRunning 判断是否正在运行
func (ps PipelineStatus) IsRunning() bool {
	return ps == PipelineRunning
}

// Deprecated: Use FSM instead.
// CanResume 判断是否可以恢复
func (ps PipelineStatus) CanResume() bool {
	return ps == PipelinePaused || ps == PipelineFailed
}

// Deprecated: Use FSM instead.
// NewPipelineStateMachine 创建流水线状态机
func NewPipelineStateMachine() *StateMachine[PipelineStatus] {
	sm := NewWithState(PipelinePending)

	// 定义状态转移规则
	sm.Allow(PipelinePending, PipelineRunning).
		Allow(PipelineRunning, PipelineSuccess, PipelineFailed, PipelineCanceled, PipelinePaused).
		Allow(PipelineFailed, PipelineRunning).                  // 支持重试
		Allow(PipelinePaused, PipelineRunning, PipelineCanceled) // 支持暂停和恢复

	return sm
}

// Deprecated: Use FSM instead.
// NewPipelineStateMachineWithHooks 创建带钩子的流水线状态机
func NewPipelineStateMachineWithHooks(
	onStart func() error,
	onComplete func(status PipelineStatus) error,
	onPause func() error,
) *StateMachine[PipelineStatus] {
	sm := NewPipelineStateMachine()

	// 进入运行状态时的钩子
	if onStart != nil {
		sm.OnEnter(PipelineRunning, func(state PipelineStatus) error {
			return onStart()
		})
	}

	// 进入暂停状态时的钩子
	if onPause != nil {
		sm.OnEnter(PipelinePaused, func(state PipelineStatus) error {
			return onPause()
		})
	}

	// 进入终止状态时的钩子
	if onComplete != nil {
		sm.OnEnter(PipelineSuccess, func(state PipelineStatus) error {
			return onComplete(state)
		})
		sm.OnEnter(PipelineFailed, func(state PipelineStatus) error {
			return onComplete(state)
		})
		sm.OnEnter(PipelineCanceled, func(state PipelineStatus) error {
			return onComplete(state)
		})
	}

	return sm
}
