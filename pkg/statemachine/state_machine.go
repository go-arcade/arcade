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

// Event represents an event that triggers a state transition in the FSM.
// Events are optional - state transitions can also be triggered directly.
type Event string

// TransitionHook is triggered when a state transition occurs.
type TransitionHook[T comparable] func(from, to T, event Event) error

// StateHook is triggered when entering or exiting a state.
type StateHook[T comparable] func(state T) error

// TransitionValidator validates whether a state transition is allowed.
type TransitionValidator[T comparable] func(from, to T, event Event) error

// TransitionRecord records a state transition in the FSM history.
type TransitionRecord[T comparable] struct {
	From      T
	To        T
	Event     Event
	Timestamp time.Time
	Error     error
}

// StateMachine is a generic Finite State Machine implementation.
// It supports:
//   - State transitions with optional events
//   - Hooks (OnEnter, OnExit, OnTransition)
//   - Validators for transition validation
//   - Transition history tracking
//   - Graphviz DOT export for visualization
//
// The StateMachine is thread-safe and can be used concurrently.
type StateMachine[T comparable] struct {
	mu sync.RWMutex

	currentState T
	initialState T

	// Transition definitions: from state -> list of valid next states
	validTransitions map[T][]T
	// Event-based transitions: (from state, event) -> target state
	eventTransitions map[transitionKey[T]]T

	history        []TransitionRecord[T]
	maxHistorySize int

	onTransition []TransitionHook[T]
	onEnter      map[T][]StateHook[T]
	onExit       map[T][]StateHook[T]
	validators   []TransitionValidator[T]

	onError func(from, to T, event Event, err error)
}

type transitionKey[T comparable] struct {
	From  T
	Event Event
}

// New creates a new StateMachine instance.
func New[T comparable]() *StateMachine[T] {
	return &StateMachine[T]{
		validTransitions: make(map[T][]T),
		eventTransitions: make(map[transitionKey[T]]T),
		onEnter:          make(map[T][]StateHook[T]),
		onExit:           make(map[T][]StateHook[T]),
		history:          make([]TransitionRecord[T], 0),
		maxHistorySize:   100,
	}
}

// NewWithState creates a new StateMachine with an initial state.
func NewWithState[T comparable](initialState T) *StateMachine[T] {
	sm := New[T]()
	sm.currentState = initialState
	sm.initialState = initialState
	return sm
}

// Allow registers valid state transitions (compatibility method).
// This is equivalent to AddTransitions.
func (sm *StateMachine[T]) Allow(from T, to ...T) *StateMachine[T] {
	return sm.AddTransitions(from, to...)
}

// AddTransition adds a valid state transition.
// This is the basic way to define transitions without events.
func (sm *StateMachine[T]) AddTransition(from T, to T) *StateMachine[T] {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if !slices.Contains(sm.validTransitions[from], to) {
		sm.validTransitions[from] = append(sm.validTransitions[from], to)
	}
	return sm
}

// AddTransitions adds multiple valid state transitions from a source state.
func (sm *StateMachine[T]) AddTransitions(from T, to ...T) *StateMachine[T] {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	for _, target := range to {
		if !slices.Contains(sm.validTransitions[from], target) {
			sm.validTransitions[from] = append(sm.validTransitions[from], target)
		}
	}
	return sm
}

// AddEventTransition adds an event-driven state transition.
// When the specified event occurs in the from state, the FSM transitions to the to state.
func (sm *StateMachine[T]) AddEventTransition(from T, event Event, to T) *StateMachine[T] {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	key := transitionKey[T]{From: from, Event: event}
	sm.eventTransitions[key] = to
	// Also add to valid transitions
	if !slices.Contains(sm.validTransitions[from], to) {
		sm.validTransitions[from] = append(sm.validTransitions[from], to)
	}
	return sm
}

// CanTransit checks if a transition from one state to another is valid (compatibility method).
func (sm *StateMachine[T]) CanTransit(from, to T) bool {
	return sm.CanTransition(from, to)
}

// CanTransition checks if a transition from one state to another is valid.
func (sm *StateMachine[T]) CanTransition(from, to T) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return slices.Contains(sm.validTransitions[from], to)
}

// CanTransitionWithEvent checks if a transition is valid for the given event.
func (sm *StateMachine[T]) CanTransitionWithEvent(from T, event Event) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	key := transitionKey[T]{From: from, Event: event}
	_, exists := sm.eventTransitions[key]
	return exists
}

// Current returns the current state of the StateMachine.
func (sm *StateMachine[T]) Current() T {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.currentState
}

// SetCurrent sets the current state without triggering hooks.
// This is useful for initialization or recovery scenarios.
func (sm *StateMachine[T]) SetCurrent(state T) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.currentState = state
	if sm.initialState == *new(T) {
		sm.initialState = state
	}
}

// Initial returns the initial state of the StateMachine.
func (sm *StateMachine[T]) Initial() T {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.initialState
}

// Reset resets the StateMachine to its initial state and clears history.
func (sm *StateMachine[T]) Reset() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.currentState = sm.initialState
	sm.history = make([]TransitionRecord[T], 0)
}

// GetValidNextStates returns all valid next states from the given state.
func (sm *StateMachine[T]) GetValidNextStates(from T) []T {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	if states, ok := sm.validTransitions[from]; ok {
		result := make([]T, len(states))
		copy(result, states)
		return result
	}
	return []T{}
}

// GetAllStates returns all states defined in the StateMachine.
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

// History returns the transition history.
func (sm *StateMachine[T]) History() []TransitionRecord[T] {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	result := make([]TransitionRecord[T], len(sm.history))
	copy(result, sm.history)
	return result
}

// SetMaxHistorySize sets the maximum number of history records to keep.
func (sm *StateMachine[T]) SetMaxHistorySize(size int) *StateMachine[T] {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.maxHistorySize = size
	if len(sm.history) > size {
		sm.history = sm.history[len(sm.history)-size:]
	}
	return sm
}

// OnTransition registers a hook that is called during any state transition.
func (sm *StateMachine[T]) OnTransition(h TransitionHook[T]) *StateMachine[T] {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.onTransition = append(sm.onTransition, h)
	return sm
}

// OnEnter registers a hook that is called when entering a specific state.
func (sm *StateMachine[T]) OnEnter(state T, h StateHook[T]) *StateMachine[T] {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.onEnter[state] = append(sm.onEnter[state], h)
	return sm
}

// OnExit registers a hook that is called when exiting a specific state.
func (sm *StateMachine[T]) OnExit(state T, h StateHook[T]) *StateMachine[T] {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.onExit[state] = append(sm.onExit[state], h)
	return sm
}

// AddValidator adds a validator that checks if a transition is allowed.
func (sm *StateMachine[T]) AddValidator(v TransitionValidator[T]) *StateMachine[T] {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.validators = append(sm.validators, v)
	return sm
}

// OnError registers an error handler that is called when a transition fails.
func (sm *StateMachine[T]) OnError(handler func(from, to T, event Event, err error)) *StateMachine[T] {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.onError = handler
	return sm
}

// Transit performs a state transition from one state to another (compatibility method).
func (sm *StateMachine[T]) Transit(from, to T) error {
	return sm.Transition(from, to, "")
}

// Transition performs a state transition from one state to another.
// It validates the transition, runs validators, triggers hooks, and records history.
func (sm *StateMachine[T]) Transition(from, to T, event Event) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	startTime := time.Now()
	var transitionErr error

	defer func() {
		record := TransitionRecord[T]{
			From:      from,
			To:        to,
			Event:     event,
			Timestamp: startTime,
			Error:     transitionErr,
		}
		sm.history = append(sm.history, record)

		if len(sm.history) > sm.maxHistorySize {
			sm.history = sm.history[len(sm.history)-sm.maxHistorySize:]
		}

		if transitionErr != nil && sm.onError != nil {
			sm.onError(from, to, event, transitionErr)
		}
	}()

	// Validate transition
	if !slices.Contains(sm.validTransitions[from], to) {
		transitionErr = fmt.Errorf("invalid transition: %v → %v", from, to)
		return transitionErr
	}

	// Run validators
	for _, validator := range sm.validators {
		if err := validator(from, to, event); err != nil {
			transitionErr = fmt.Errorf("validation failed: %w", err)
			return transitionErr
		}
	}

	// Trigger OnExit hooks
	if hooks := sm.onExit[from]; len(hooks) > 0 {
		for _, h := range hooks {
			if err := h(from); err != nil {
				transitionErr = fmt.Errorf("exit hook failed for state %v: %w", from, err)
				return transitionErr
			}
		}
	}

	// Trigger OnTransition hooks
	for _, h := range sm.onTransition {
		if err := h(from, to, event); err != nil {
			transitionErr = fmt.Errorf("transition hook failed: %w", err)
			return transitionErr
		}
	}

	// Update current state
	sm.currentState = to

	// Trigger OnEnter hooks
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

// TransitTo performs a transition from the current state to the target state.
func (sm *StateMachine[T]) TransitTo(to T) error {
	return sm.TransitionTo(to)
}

// TransitionTo performs a transition from the current state to the target state.
func (sm *StateMachine[T]) TransitionTo(to T) error {
	return sm.Transition(sm.Current(), to, "")
}

// TriggerEvent triggers a state transition based on an event.
// It looks up the event transition table to find the target state.
func (sm *StateMachine[T]) TriggerEvent(event Event) error {
	sm.mu.RLock()
	current := sm.currentState
	sm.mu.RUnlock()

	sm.mu.RLock()
	key := transitionKey[T]{From: current, Event: event}
	to, exists := sm.eventTransitions[key]
	sm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no transition defined for event %v in state %v", event, current)
	}

	return sm.Transition(current, to, event)
}

// MustTransit performs a transition and panics on error (compatibility method).
func (sm *StateMachine[T]) MustTransit(from, to T) {
	sm.MustTransition(from, to, "")
}

// MustTransition performs a transition and panics on error.
func (sm *StateMachine[T]) MustTransition(from, to T, event Event) {
	if err := sm.Transition(from, to, event); err != nil {
		panic(err)
	}
}

// MustTransitTo performs a transition from current state and panics on error (compatibility method).
func (sm *StateMachine[T]) MustTransitTo(to T) {
	sm.MustTransitionTo(to)
}

// MustTransitionTo performs a transition from current state and panics on error.
func (sm *StateMachine[T]) MustTransitionTo(to T) {
	if err := sm.TransitionTo(to); err != nil {
		panic(err)
	}
}

// MustTriggerEvent triggers an event and panics on error.
func (sm *StateMachine[T]) MustTriggerEvent(event Event) {
	if err := sm.TriggerEvent(event); err != nil {
		panic(err)
	}
}

// Is checks if the current state matches the given state.
func (sm *StateMachine[T]) Is(state T) bool {
	return sm.Current() == state
}

// IsOneOf checks if the current state is one of the given states.
func (sm *StateMachine[T]) IsOneOf(states ...T) bool {
	current := sm.Current()
	return slices.Contains(states, current)
}

// CanTransitTo checks if a transition to the target state is valid from the current state (compatibility method).
func (sm *StateMachine[T]) CanTransitTo(to T) bool {
	return sm.CanTransitionTo(to)
}

// CanTransitionTo checks if a transition to the target state is valid from the current state.
func (sm *StateMachine[T]) CanTransitionTo(to T) bool {
	return sm.CanTransition(sm.Current(), to)
}

// ToDot exports the StateMachine as a Graphviz DOT format string.
func (sm *StateMachine[T]) ToDot(name string) string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	dot := fmt.Sprintf("digraph %s {\n", name)
	dot += "  rankdir=LR;\n"
	dot += "  node [shape=circle];\n"

	if sm.initialState != *new(T) {
		dot += "  start [shape=point];\n"
		dot += fmt.Sprintf("  start -> \"%v\";\n", sm.initialState)
	}

	if sm.currentState != *new(T) {
		dot += fmt.Sprintf("  \"%v\" [style=filled, fillcolor=lightblue];\n", sm.currentState)
	}

	for from, tos := range sm.validTransitions {
		for _, to := range tos {
			label := ""
			// Check if there's an event for this transition
			for key, target := range sm.eventTransitions {
				if key.From == from && target == to {
					if label != "" {
						label += ", "
					}
					label += string(key.Event)
				}
			}
			if label != "" {
				dot += fmt.Sprintf("  \"%v\" -> \"%v\" [label=\"%s\"];\n", from, to, label)
			} else {
				dot += fmt.Sprintf("  \"%v\" -> \"%v\";\n", from, to)
			}
		}
	}

	dot += "}\n"
	return dot
}
