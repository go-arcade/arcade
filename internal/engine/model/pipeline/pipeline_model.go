package pipeline

import (
	"time"

	"github.com/go-arcade/arcade/internal/engine/model"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2025/01/13
 * @file: model_pipeline.go
 * @description: pipeline model
 */

// Pipeline 流水线定义表
type Pipeline struct {
	model.BaseModel
	PipelineId  string `gorm:"column:pipeline_id" json:"pipelineId"`
	Name        string `gorm:"column:name" json:"name"`
	Description string `gorm:"column:description" json:"description"`
	RepoUrl     string `gorm:"column:repo_url" json:"repoUrl"`
	Branch      string `gorm:"column:branch" json:"branch"`
	Status      int    `gorm:"column:status" json:"status"`            // 0:未知 1:等待 2:运行中 3:成功 4:失败 5:已取消 6:部分成功
	TriggerType int    `gorm:"column:trigger_type" json:"triggerType"` // 1:手动 2:Webhook 3:定时 4:API
	Cron        string `gorm:"column:cron" json:"cron"`
	Env         string `gorm:"column:env;type:json" json:"env"` // JSON格式
	TotalRuns   int    `gorm:"column:total_runs" json:"totalRuns"`
	SuccessRuns int    `gorm:"column:success_runs" json:"successRuns"`
	FailedRuns  int    `gorm:"column:failed_runs" json:"failedRuns"`
	CreatedBy   string `gorm:"column:created_by" json:"createdBy"`
	IsEnabled   int    `gorm:"column:is_enabled" json:"isEnabled"` // 0: disabled, 1: enabled
}

func (Pipeline) TableName() string {
	return "t_pipeline"
}

// PipelineRun 流水线执行记录表
type PipelineRun struct {
	model.BaseModel
	RunId         string     `gorm:"column:run_id" json:"runId"`
	PipelineId    string     `gorm:"column:pipeline_id" json:"pipelineId"`
	PipelineName  string     `gorm:"column:pipeline_name" json:"pipelineName"`
	Branch        string     `gorm:"column:branch" json:"branch"`
	CommitSha     string     `gorm:"column:commit_sha" json:"commitSha"`
	Status        int        `gorm:"column:status" json:"status"`
	TriggerType   int        `gorm:"column:trigger_type" json:"triggerType"`
	TriggeredBy   string     `gorm:"column:triggered_by" json:"triggeredBy"`
	Env           string     `gorm:"column:env;type:json" json:"env"` // JSON格式
	TotalJobs     int        `gorm:"column:total_jobs" json:"totalJobs"`
	CompletedJobs int        `gorm:"column:completed_jobs" json:"completedJobs"`
	FailedJobs    int        `gorm:"column:failed_jobs" json:"failedJobs"`
	RunningJobs   int        `gorm:"column:running_jobs" json:"runningJobs"`
	CurrentStage  int        `gorm:"column:current_stage" json:"currentStage"`
	TotalStages   int        `gorm:"column:total_stages" json:"totalStages"`
	StartTime     *time.Time `gorm:"column:start_time" json:"startTime"`
	EndTime       *time.Time `gorm:"column:end_time" json:"endTime"`
	Duration      int64      `gorm:"column:duration" json:"duration"` // 毫秒
}

func (PipelineRun) TableName() string {
	return "t_pipeline_run"
}

// PipelineStage 流水线阶段表
type PipelineStage struct {
	model.BaseModel
	StageId    string `gorm:"column:stage_id" json:"stageId"`
	PipelineId string `gorm:"column:pipeline_id" json:"pipelineId"`
	Name       string `gorm:"column:name" json:"name"`
	StageOrder int    `gorm:"column:stage_order" json:"stageOrder"`
	Parallel   int    `gorm:"column:parallel" json:"parallel"` // 0:否 1:是
}

func (PipelineStage) TableName() string {
	return "t_pipeline_stage"
}
