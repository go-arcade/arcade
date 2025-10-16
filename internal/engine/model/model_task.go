package model

import "time"

/**
 * @author: gagral.x@gmail.com
 * @time: 2025/01/13
 * @file: model_task.go
 * @description: task model
 */

// Task 任务表
type Task struct {
	BaseModel
	TaskId        string     `gorm:"column:task_id" json:"taskId"`
	Name          string     `gorm:"column:name" json:"name"`
	PipelineId    string     `gorm:"column:pipeline_id" json:"pipelineId"`
	PipelineRunId string     `gorm:"column:pipeline_run_id" json:"pipelineRunId"`
	StageId       string     `gorm:"column:stage_id" json:"stageId"`
	Stage         int        `gorm:"column:stage" json:"stage"`
	AgentId       string     `gorm:"column:agent_id" json:"agentId"`
	Status        int        `gorm:"column:status" json:"status"`     // 1:等待 2:入队 3:运行中 4:成功 5:失败 6:已取消 7:超时 8:已跳过
	Priority      int        `gorm:"column:priority" json:"priority"` // 1:最高 5:普通 10:最低
	Image         string     `gorm:"column:image" json:"image"`
	Commands      string     `gorm:"column:commands;type:text" json:"commands"` // JSON数组
	Workspace     string     `gorm:"column:workspace" json:"workspace"`
	Env           string     `gorm:"column:env;type:json" json:"env"`         // JSON格式
	Secrets       string     `gorm:"column:secrets;type:json" json:"secrets"` // JSON格式
	Timeout       int        `gorm:"column:timeout" json:"timeout"`           // 秒
	RetryCount    int        `gorm:"column:retry_count" json:"retryCount"`
	CurrentRetry  int        `gorm:"column:current_retry" json:"currentRetry"`
	AllowFailure  int        `gorm:"column:allow_failure" json:"allowFailure"`             // 0:否 1:是
	LabelSelector string     `gorm:"column:label_selector;type:json" json:"labelSelector"` // JSON格式
	Tags          string     `gorm:"column:tags" json:"tags"`                              // 逗号分隔,已废弃
	DependsOn     string     `gorm:"column:depends_on" json:"dependsOn"`                   // 逗号分隔
	ExitCode      *int       `gorm:"column:exit_code" json:"exitCode"`
	ErrorMessage  string     `gorm:"column:error_message;type:text" json:"errorMessage"`
	StartTime     *time.Time `gorm:"column:start_time" json:"startTime"`
	EndTime       *time.Time `gorm:"column:end_time" json:"endTime"`
	Duration      int64      `gorm:"column:duration" json:"duration"` // 毫秒
	CreatedBy     string     `gorm:"column:created_by" json:"createdBy"`
}

func (Task) TableName() string {
	return "t_task"
}

// TaskArtifact 任务产物表
type TaskArtifact struct {
	BaseModel
	ArtifactId    string     `gorm:"column:artifact_id" json:"artifactId"`
	TaskId        string     `gorm:"column:task_id" json:"taskId"`
	PipelineRunId string     `gorm:"column:pipeline_run_id" json:"pipelineRunId"`
	Name          string     `gorm:"column:name" json:"name"`
	Path          string     `gorm:"column:path" json:"path"`
	Destination   string     `gorm:"column:destination" json:"destination"`
	Size          int64      `gorm:"column:size" json:"size"`                // 字节
	StorageType   string     `gorm:"column:storage_type" json:"storageType"` // minio/s3/oss/gcs/cos
	StoragePath   string     `gorm:"column:storage_path" json:"storagePath"`
	Expire        int        `gorm:"column:expire" json:"expire"` // 0:否 1:是
	ExpireDays    *int       `gorm:"column:expire_days" json:"expireDays"`
	ExpiredAt     *time.Time `gorm:"column:expired_at" json:"expiredAt"`
}

func (TaskArtifact) TableName() string {
	return "t_task_artifact"
}
