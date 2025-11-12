package log

import "time"

// TaskLog 任务日志(MongoDB Collection: task_logs)
type TaskLog struct {
	LogId         string    `bson:"log_id" json:"logId"`
	TaskId        string    `bson:"task_id" json:"taskId"`
	PipelineRunId string    `bson:"pipeline_run_id" json:"pipelineRunId"`
	AgentId       string    `bson:"agent_id" json:"agentId"`
	LineNumber    int       `bson:"line_number" json:"lineNumber"`
	Content       string    `bson:"content" json:"content"`
	Timestamp     time.Time `bson:"timestamp" json:"timestamp"`
	Level         string    `bson:"level" json:"level"` // INFO/WARN/ERROR/DEBUG
	CreatedAt     time.Time `bson:"created_at" json:"createdAt"`
}

// TerminalLog 终端日志/构建日志(MongoDB Collection: terminal_logs)
type TerminalLog struct {
	SessionId        string              `bson:"session_id" json:"sessionId"`
	SessionType      string              `bson:"session_type" json:"sessionType"` // build/deploy/release/debug
	Environment      string              `bson:"environment" json:"environment"`  // dev/test/staging/prod
	TaskId           string              `bson:"task_id,omitempty" json:"taskId,omitempty"`
	PipelineId       string              `bson:"pipeline_id,omitempty" json:"pipelineId,omitempty"`
	PipelineRunId    string              `bson:"pipeline_run_id,omitempty" json:"pipelineRunId,omitempty"`
	UserId           string              `bson:"user_id" json:"userId"`
	Hostname         string              `bson:"hostname" json:"hostname"`
	WorkingDirectory string              `bson:"working_directory" json:"workingDirectory"`
	Command          string              `bson:"command" json:"command"`
	ExitCode         *int                `bson:"exit_code,omitempty" json:"exitCode,omitempty"`
	Logs             []TerminalLogLine   `bson:"logs" json:"logs"`
	Metadata         TerminalLogMetadata `bson:"metadata" json:"metadata"`
	Status           string              `bson:"status" json:"status"` // running/completed/failed/timeout
	StartTime        time.Time           `bson:"start_time" json:"startTime"`
	EndTime          *time.Time          `bson:"end_time,omitempty" json:"endTime,omitempty"`
	CreatedAt        time.Time           `bson:"created_at" json:"createdAt"`
	UpdatedAt        time.Time           `bson:"updated_at" json:"updatedAt"`
}

// TerminalLogLine 终端日志行
type TerminalLogLine struct {
	Line      int       `bson:"line" json:"line"`
	Timestamp time.Time `bson:"timestamp" json:"timestamp"`
	Content   string    `bson:"content" json:"content"`
	Stream    string    `bson:"stream" json:"stream"` // stdout/stderr
}

// TerminalLogMetadata 终端日志元数据
type TerminalLogMetadata struct {
	TotalLines      int   `bson:"total_lines" json:"totalLines"`
	DurationMs      int64 `bson:"duration_ms" json:"durationMs"`
	OutputSizeBytes int64 `bson:"output_size_bytes" json:"outputSizeBytes"`
}

// BuildArtifactLog 产物构建日志(MongoDB Collection: build_artifacts_logs)
type BuildArtifactLog struct {
	ArtifactId   string    `bson:"artifact_id" json:"artifactId"`
	TaskId       string    `bson:"task_id" json:"taskId"`
	Operation    string    `bson:"operation" json:"operation"` // upload/download/delete
	FileName     string    `bson:"file_name" json:"fileName"`
	FileSize     int64     `bson:"file_size" json:"fileSize"`
	StorageType  string    `bson:"storage_type" json:"storageType"` // minio/s3/oss/gcs/cos
	StoragePath  string    `bson:"storage_path" json:"storagePath"`
	UserId       string    `bson:"user_id" json:"userId"`
	Status       string    `bson:"status" json:"status"` // success/failed
	ErrorMessage string    `bson:"error_message,omitempty" json:"errorMessage,omitempty"`
	DurationMs   int64     `bson:"duration_ms" json:"durationMs"`
	Timestamp    time.Time `bson:"timestamp" json:"timestamp"`
}
