package pipeline

import (
	agentv1 "github.com/observabil/arcade/api/agent/v1"
)

type PipelineExecutor struct {
	AgentClient agentv1.AgentServiceClient
}

// func (e *PipelineExecutor) ExecuteStage(ctx context.Context, stage *Stage) error {
// 	for _, step := range stage.Steps {
// 		_, err := e.AgentClient.RunTask(ctx, &agentv1.RunTaskRequest{
// 			PluginType: step.PluginType,
// 			Config:     step.Config,
// 		})
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }
