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

package queue

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-arcade/arcade/internal/engine/model"
	"github.com/go-arcade/arcade/pkg/log"
	"gorm.io/gorm"
)

// TaskRecordManager 任务记录管理器，负责将任务状态写入 ClickHouse
type TaskRecordManager struct {
	db        *gorm.DB
	tableName string
}

// NewTaskRecordManager 创建任务记录管理器
func NewTaskRecordManager(clickHouse *gorm.DB) (*TaskRecordManager, error) {
	if clickHouse == nil {
		return nil, nil
	}

	tableName := model.TaskQueueRecord{}.CollectionName()
	manager := &TaskRecordManager{
		db:        clickHouse,
		tableName: tableName,
	}

	// 创建表（如果不存在）
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := manager.createTableIfNotExists(ctx); err != nil {
		log.Warnw("failed to create task records table", "error", err)
		// 不返回错误，允许继续运行
	}

	return manager, nil
}

// createTableIfNotExists 创建表（如果不存在）
func (m *TaskRecordManager) createTableIfNotExists(ctx context.Context) error {
	sqlDB, err := m.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get SQL DB from GORM: %w", err)
	}

	// ClickHouse 表结构
	createTableSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			task_id String,
			task_type String,
			status String,
			queue String,
			priority Int32,
			pipeline_id String,
			pipeline_run_id String,
			stage_id String,
			agent_id String,
			payload String,
			create_time DateTime,
			start_time Nullable(DateTime),
			end_time Nullable(DateTime),
			duration Nullable(Int64),
			retry_count Int32,
			current_retry Int32,
			error_message Nullable(String)
		) ENGINE = MergeTree()
		ORDER BY (task_id, create_time)
		PRIMARY KEY task_id
		SETTINGS index_granularity = 8192
	`, m.tableName)

	_, err = sqlDB.ExecContext(ctx, createTableSQL)
	return err
}

// RecordTaskEnqueued 记录任务入队
func (m *TaskRecordManager) RecordTaskEnqueued(payload *TaskPayload, queueName string) {
	if m == nil || m.db == nil {
		return
	}

	now := time.Now()
	payloadJSON, _ := json.Marshal(payload.Data)

	sqlDB, err := m.db.DB()
	if err != nil {
		log.Warnw("failed to get SQL DB from GORM", "task_id", payload.TaskID, "error", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 使用 INSERT 语句（ClickHouse 支持 INSERT ... ON DUPLICATE KEY UPDATE，但更推荐使用 ReplacingMergeTree）
	insertSQL := fmt.Sprintf(`
		INSERT INTO %s (
			task_id, task_type, status, queue, priority, pipeline_id, pipeline_run_id,
			stage_id, agent_id, payload, create_time, retry_count, current_retry
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, m.tableName)

	_, err = sqlDB.ExecContext(ctx, insertSQL,
		payload.TaskID,
		payload.TaskType,
		TaskRecordStatusPending,
		queueName,
		payload.Priority,
		payload.PipelineID,
		payload.PipelineRunID,
		payload.StageID,
		payload.AgentID,
		string(payloadJSON),
		now,
		payload.RetryCount,
		0,
	)
	if err != nil {
		log.Warnw("failed to record task enqueued", "task_id", payload.TaskID, "error", err)
	}
}

// RecordTaskStarted 记录任务开始
func (m *TaskRecordManager) RecordTaskStarted(payload *TaskPayload) {
	if m == nil || m.db == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sqlDB, err := m.db.DB()
	if err != nil {
		log.Warnw("failed to get SQL DB from GORM", "task_id", payload.TaskID, "error", err)
		return
	}

	now := time.Now()
	// ClickHouse 使用 ALTER TABLE UPDATE 或 INSERT 来更新数据
	// 由于 ClickHouse 是列式数据库，更适合使用 INSERT 覆盖旧数据
	// 这里我们使用 ALTER TABLE UPDATE（需要表引擎支持）
	updateSQL := fmt.Sprintf(`
		ALTER TABLE %s UPDATE status = ?, start_time = ? WHERE task_id = ?
	`, m.tableName)

	_, err = sqlDB.ExecContext(ctx, updateSQL, TaskRecordStatusRunning, now, payload.TaskID)
	if err != nil {
		log.Warnw("failed to record task started", "task_id", payload.TaskID, "error", err)
		// 如果 ALTER UPDATE 失败，尝试使用 INSERT 覆盖
		m.insertOrUpdateTask(ctx, payload.TaskID, TaskRecordStatusRunning, &now, nil, nil, nil)
	}
}

// RecordTaskCompleted 记录任务完成
func (m *TaskRecordManager) RecordTaskCompleted(payload *TaskPayload) {
	if m == nil || m.db == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sqlDB, err := m.db.DB()
	if err != nil {
		log.Warnw("failed to get SQL DB from GORM", "task_id", payload.TaskID, "error", err)
		return
	}

	now := time.Now()

	// 先获取开始时间以计算耗时
	var startTime sql.NullTime
	var duration sql.NullInt64
	selectSQL := fmt.Sprintf(`SELECT start_time FROM %s WHERE task_id = ?`, m.tableName)
	row := sqlDB.QueryRowContext(ctx, selectSQL, payload.TaskID)
	if err := row.Scan(&startTime); err == nil && startTime.Valid {
		durationValue := now.Sub(startTime.Time).Milliseconds()
		duration = sql.NullInt64{Int64: durationValue, Valid: true}
	}

	updateSQL := fmt.Sprintf(`
		ALTER TABLE %s UPDATE status = ?, end_time = ?, duration = ? WHERE task_id = ?
	`, m.tableName)

	_, err = sqlDB.ExecContext(ctx, updateSQL, TaskRecordStatusCompleted, now, duration, payload.TaskID)
	if err != nil {
		log.Warnw("failed to record task completed", "task_id", payload.TaskID, "error", err)
		// 如果 ALTER UPDATE 失败，尝试使用 INSERT 覆盖
		m.insertOrUpdateTask(ctx, payload.TaskID, TaskRecordStatusCompleted, nil, &now, &duration, nil)
	}
}

// RecordTaskFailed 记录任务失败
func (m *TaskRecordManager) RecordTaskFailed(payload *TaskPayload, err error) {
	if m == nil || m.db == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sqlDB, dbErr := m.db.DB()
	if dbErr != nil {
		log.Warnw("failed to get SQL DB from GORM", "task_id", payload.TaskID, "error", dbErr)
		return
	}

	now := time.Now()
	errorMsg := ""
	if err != nil {
		errorMsg = err.Error()
	}

	// 先获取开始时间以计算耗时
	var startTime sql.NullTime
	var duration sql.NullInt64
	selectSQL := fmt.Sprintf(`SELECT start_time FROM %s WHERE task_id = ?`, m.tableName)
	row := sqlDB.QueryRowContext(ctx, selectSQL, payload.TaskID)
	if scanErr := row.Scan(&startTime); scanErr == nil && startTime.Valid {
		durationValue := now.Sub(startTime.Time).Milliseconds()
		duration = sql.NullInt64{Int64: durationValue, Valid: true}
	}

	updateSQL := fmt.Sprintf(`
		ALTER TABLE %s UPDATE status = ?, error_message = ?, end_time = ?, duration = ? WHERE task_id = ?
	`, m.tableName)

	_, updateErr := sqlDB.ExecContext(ctx, updateSQL, TaskRecordStatusFailed, errorMsg, now, duration, payload.TaskID)
	if updateErr != nil {
		log.Warnw("failed to record task failed", "task_id", payload.TaskID, "error", updateErr)
		// 如果 ALTER UPDATE 失败，尝试使用 INSERT 覆盖
		m.insertOrUpdateTask(ctx, payload.TaskID, TaskRecordStatusFailed, nil, &now, &duration, &errorMsg)
	}
}

// insertOrUpdateTask 插入或更新任务记录（备用方法）
func (m *TaskRecordManager) insertOrUpdateTask(ctx context.Context, taskID, status string, startTime, endTime *time.Time, duration *sql.NullInt64, errorMsg *string) {
	sqlDB, err := m.db.DB()
	if err != nil {
		return
	}

	// 先查询现有记录
	var existingRecord struct {
		TaskType      string
		Queue         string
		Priority      int32
		PipelineID    string
		PipelineRunID string
		StageID       string
		AgentID       string
		Payload       string
		CreateTime    time.Time
		RetryCount    int32
		CurrentRetry  int32
	}

	selectSQL := fmt.Sprintf(`SELECT task_type, queue, priority, pipeline_id, pipeline_run_id, stage_id, agent_id, payload, create_time, retry_count, current_retry FROM %s WHERE task_id = ?`, m.tableName)
	row := sqlDB.QueryRowContext(ctx, selectSQL, taskID)
	if err := row.Scan(&existingRecord.TaskType, &existingRecord.Queue, &existingRecord.Priority,
		&existingRecord.PipelineID, &existingRecord.PipelineRunID, &existingRecord.StageID,
		&existingRecord.AgentID, &existingRecord.Payload, &existingRecord.CreateTime,
		&existingRecord.RetryCount, &existingRecord.CurrentRetry); err != nil {
		// 记录不存在，无法更新
		return
	}

	// 插入新记录（覆盖旧记录）
	insertSQL := fmt.Sprintf(`
		INSERT INTO %s (
			task_id, task_type, status, queue, priority, pipeline_id, pipeline_run_id,
			stage_id, agent_id, payload, create_time, start_time, end_time, duration,
			retry_count, current_retry, error_message
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, m.tableName)

	var durationValue interface{}
	if duration != nil && duration.Valid {
		durationValue = duration.Int64
	} else {
		durationValue = nil
	}

	var errorMsgValue interface{}
	if errorMsg != nil {
		errorMsgValue = *errorMsg
	} else {
		errorMsgValue = nil
	}

	_, err = sqlDB.ExecContext(ctx, insertSQL,
		taskID,
		existingRecord.TaskType,
		status,
		existingRecord.Queue,
		existingRecord.Priority,
		existingRecord.PipelineID,
		existingRecord.PipelineRunID,
		existingRecord.StageID,
		existingRecord.AgentID,
		existingRecord.Payload,
		existingRecord.CreateTime,
		startTime,
		endTime,
		durationValue,
		existingRecord.RetryCount,
		existingRecord.CurrentRetry,
		errorMsgValue,
	)
	if err != nil {
		log.Warnw("failed to insert/update task record", "task_id", taskID, "error", err)
	}
}
