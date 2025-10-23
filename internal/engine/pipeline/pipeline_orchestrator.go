package pipeline

import (
	"github.com/go-arcade/arcade/pkg/event"
	"github.com/go-arcade/arcade/pkg/statemachine"
)

type PipelineOrchestrator struct {
	Executor *PipelineExecutor
	State    *statemachine.StateMachine[statemachine.PipelineStatus]
	EventBus *event.EventBus
}

// func (o *PipelineOrchestrator) Run(ctx context.Context, pipelineID string) error {
// 	o.State.TransitTo(statemachine.PipelinePending)
// 	stages := o.Executor.AgentClient.GetStages(pipelineID)

// 	for _, stage := range stages {
// 		if err := o.Executor.Execute(ctx, stage); err != nil {
// 			o.State.TransitTo(statemachine.PipelineFailed)
// 			return err
// 		}
// 		o.State.TransitTo(statemachine.PipelineSuccess)
// 	}
// 	return nil
// }
