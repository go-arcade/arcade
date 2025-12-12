package model

import (
	"time"
)

// TaskQueueRecord queue 任务记录（MongoDB）
type TaskQueueRecord struct {
	TaskID        string         `bson:"task_id" json:"task_id"`
	TaskType      string         `bson:"task_type" json:"task_type"`
	Status        string         `bson:"status" json:"status"` // pending/running/completed/failed
	Priority      int            `bson:"priority" json:"priority"`
	Queue         string         `bson:"queue" json:"queue"`
	PipelineID    string         `bson:"pipeline_id,omitempty" json:"pipeline_id,omitempty"`
	PipelineRunID string         `bson:"pipeline_run_id,omitempty" json:"pipeline_run_id,omitempty"`
	StageID       string         `bson:"stage_id,omitempty" json:"stage_id,omitempty"`
	AgentID       string         `bson:"agent_id,omitempty" json:"agent_id,omitempty"`
	Payload       map[string]any `bson:"payload" json:"payload"` // 任务负载数据
	ErrorMessage  string         `bson:"error_message,omitempty" json:"error_message,omitempty"`
	CreateTime    time.Time      `bson:"create_time" json:"create_time"`
	StartTime     *time.Time     `bson:"start_time,omitempty" json:"start_time,omitempty"`
	EndTime       *time.Time     `bson:"end_time,omitempty" json:"end_time,omitempty"`
	Duration      int64          `bson:"duration,omitempty" json:"duration,omitempty"` // 毫秒
	RetryCount    int            `bson:"retry_count" json:"retry_count"`
	CurrentRetry  int            `bson:"current_retry" json:"current_retry"`
}

// CollectionName 返回集合名称
func (TaskQueueRecord) CollectionName() string {
	return "c_task_queue_records"
}
