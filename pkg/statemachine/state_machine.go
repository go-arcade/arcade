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

import (
	"fmt"
	"slices"
	"sync"
	"time"
)

// TransitionHook Triggered when the state changes
type TransitionHook[T comparable] func(from, to T) error

// StateHook Triggered when entering or exiting a state
type StateHook[T comparable] func(state T) error

// TransitionValidator 状态转移验证器
type TransitionValidator[T comparable] func(from, to T) error

// TransitionRecord 状态转移记录
type TransitionRecord[T comparable] struct {
	From      T
	To        T
	Timestamp time.Time
	Error     error
}

// StateMachine 泛型状态机（支持 OnEnter / OnExit / OnTransition）
type StateMachine[T comparable] struct {
	mu sync.RWMutex

	currentState     T
	initialState     T
	validTransitions map[T][]T
	history          []TransitionRecord[T]
	maxHistorySize   int

	onTransition []TransitionHook[T]
	onEnter      map[T][]StateHook[T]
	onExit       map[T][]StateHook[T]
	validators   []TransitionValidator[T]

	// 错误处理
	onError func(from, to T, err error)
}

// New 创建新的状态机
func New[T comparable]() *StateMachine[T] {
	return &StateMachine[T]{
		validTransitions: make(map[T][]T),
		onEnter:          make(map[T][]StateHook[T]),
		onExit:           make(map[T][]StateHook[T]),
		history:          make([]TransitionRecord[T], 0),
		maxHistorySize:   100, // 默认保留最近100条记录
	}
}

// NewWithState 创建带初始状态的状态机
func NewWithState[T comparable](initialState T) *StateMachine[T] {
	sm := New[T]()
	sm.currentState = initialState
	sm.initialState = initialState
	return sm
}

// Allow 注册合法的状态转移
func (sm *StateMachine[T]) Allow(from T, to ...T) *StateMachine[T] {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.validTransitions[from] = append(sm.validTransitions[from], to...)
	return sm
}

// CanTransit 判断是否可以状态转移
func (sm *StateMachine[T]) CanTransit(from, to T) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return slices.Contains(sm.validTransitions[from], to)
}

// Current 获取当前状态
func (sm *StateMachine[T]) Current() T {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.currentState
}

// SetCurrent 设置当前状态（不触发钩子，仅用于初始化）
func (sm *StateMachine[T]) SetCurrent(state T) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.currentState = state
	if sm.initialState == *new(T) {
		sm.initialState = state
	}
}

// Initial 获取初始状态
func (sm *StateMachine[T]) Initial() T {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.initialState
}

// Reset 重置到初始状态
func (sm *StateMachine[T]) Reset() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.currentState = sm.initialState
	sm.history = make([]TransitionRecord[T], 0)
}

// GetValidNextStates 获取当前状态可以转移到的所有状态
func (sm *StateMachine[T]) GetValidNextStates(from T) []T {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	if states, ok := sm.validTransitions[from]; ok {
		// 返回副本，避免外部修改
		result := make([]T, len(states))
		copy(result, states)
		return result
	}
	return []T{}
}

// GetAllStates 获取状态机中的所有状态
func (sm *StateMachine[T]) GetAllStates() []T {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	stateSet := make(map[T]bool)
	for from, tos := range sm.validTransitions {
		stateSet[from] = true
		for _, to := range tos {
			stateSet[to] = true
		}
	}
	states := make([]T, 0, len(stateSet))
	for state := range stateSet {
		states = append(states, state)
	}
	return states
}

// History 获取状态转移历史
func (sm *StateMachine[T]) History() []TransitionRecord[T] {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	// 返回副本
	result := make([]TransitionRecord[T], len(sm.history))
	copy(result, sm.history)
	return result
}

// SetMaxHistorySize 设置最大历史记录数
func (sm *StateMachine[T]) SetMaxHistorySize(size int) *StateMachine[T] {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.maxHistorySize = size
	if len(sm.history) > size {
		sm.history = sm.history[len(sm.history)-size:]
	}
	return sm
}

// OnTransition 注册状态转移钩子
func (sm *StateMachine[T]) OnTransition(h TransitionHook[T]) *StateMachine[T] {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.onTransition = append(sm.onTransition, h)
	return sm
}

// OnEnter 注册进入某状态的钩子
func (sm *StateMachine[T]) OnEnter(state T, h StateHook[T]) *StateMachine[T] {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.onEnter[state] = append(sm.onEnter[state], h)
	return sm
}

// OnExit 注册离开某状态的钩子
func (sm *StateMachine[T]) OnExit(state T, h StateHook[T]) *StateMachine[T] {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.onExit[state] = append(sm.onExit[state], h)
	return sm
}

// AddValidator 添加状态转移验证器
func (sm *StateMachine[T]) AddValidator(v TransitionValidator[T]) *StateMachine[T] {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.validators = append(sm.validators, v)
	return sm
}

// OnError 注册错误处理器
func (sm *StateMachine[T]) OnError(handler func(from, to T, err error)) *StateMachine[T] {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.onError = handler
	return sm
}

// Transit 执行状态转移（自动触发钩子）
func (sm *StateMachine[T]) Transit(from, to T) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// 记录开始时间
	startTime := time.Now()
	var transitionErr error

	defer func() {
		// 记录状态转移历史
		record := TransitionRecord[T]{
			From:      from,
			To:        to,
			Timestamp: startTime,
			Error:     transitionErr,
		}
		sm.history = append(sm.history, record)

		// 限制历史记录大小
		if len(sm.history) > sm.maxHistorySize {
			sm.history = sm.history[len(sm.history)-sm.maxHistorySize:]
		}

		// 如果有错误，调用错误处理器
		if transitionErr != nil && sm.onError != nil {
			sm.onError(from, to, transitionErr)
		}
	}()

	// 验证转移合法性
	if !slices.Contains(sm.validTransitions[from], to) {
		transitionErr = fmt.Errorf("invalid transition: %v → %v", from, to)
		return transitionErr
	}

	// 运行验证器
	for _, validator := range sm.validators {
		if err := validator(from, to); err != nil {
			transitionErr = fmt.Errorf("validation failed: %w", err)
			return transitionErr
		}
	}

	// 触发 OnExit 钩子
	if hooks := sm.onExit[from]; len(hooks) > 0 {
		for _, h := range hooks {
			if err := h(from); err != nil {
				transitionErr = fmt.Errorf("exit hook failed for state %v: %w", from, err)
				return transitionErr
			}
		}
	}

	// 触发 OnTransition 钩子
	for _, h := range sm.onTransition {
		if err := h(from, to); err != nil {
			transitionErr = fmt.Errorf("transition hook failed: %w", err)
			return transitionErr
		}
	}

	// 更新当前状态
	sm.currentState = to

	// 触发 OnEnter 钩子
	if hooks := sm.onEnter[to]; len(hooks) > 0 {
		for _, h := range hooks {
			if err := h(to); err != nil {
				transitionErr = fmt.Errorf("enter hook failed for state %v: %w", to, err)
				return transitionErr
			}
		}
	}

	return nil
}

// TransitTo 从当前状态转移到目标状态
func (sm *StateMachine[T]) TransitTo(to T) error {
	from := sm.Current()
	return sm.Transit(from, to)
}

// MustTransit 强制状态转移（panic 版）
func (sm *StateMachine[T]) MustTransit(from, to T) {
	if err := sm.Transit(from, to); err != nil {
		panic(err)
	}
}

// MustTransitTo 从当前状态强制转移到目标状态（panic 版）
func (sm *StateMachine[T]) MustTransitTo(to T) {
	if err := sm.TransitTo(to); err != nil {
		panic(err)
	}
}

// Is 检查当前状态是否为指定状态
func (sm *StateMachine[T]) Is(state T) bool {
	return sm.Current() == state
}

// IsOneOf 检查当前状态是否为指定状态之一
func (sm *StateMachine[T]) IsOneOf(states ...T) bool {
	current := sm.Current()
	return slices.Contains(states, current)
}

// CanTransitTo 检查是否可以从当前状态转移到目标状态
func (sm *StateMachine[T]) CanTransitTo(to T) bool {
	return sm.CanTransit(sm.Current(), to)
}

// ToDot 导出状态机为 Graphviz DOT 格式
func (sm *StateMachine[T]) ToDot(name string) string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	dot := fmt.Sprintf("digraph %s {\n", name)
	dot += "  rankdir=LR;\n"
	dot += "  node [shape=circle];\n"

	// 标记初始状态
	if sm.initialState != *new(T) {
		dot += "  start [shape=point];\n"
		dot += fmt.Sprintf("  start -> \"%v\";\n", sm.initialState)
	}

	// 标记当前状态
	if sm.currentState != *new(T) {
		dot += fmt.Sprintf("  \"%v\" [style=filled, fillcolor=lightblue];\n", sm.currentState)
	}

	// 添加转移边
	for from, tos := range sm.validTransitions {
		for _, to := range tos {
			dot += fmt.Sprintf("  \"%v\" -> \"%v\";\n", from, to)
		}
	}

	dot += "}\n"
	return dot
}
