package queue

import (
	"context"
	"time"

	"github.com/go-arcade/arcade/internal/engine/model"
	"github.com/go-arcade/arcade/pkg/database"
	"github.com/go-arcade/arcade/pkg/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TaskRecordManager 任务记录管理器，负责将任务状态写入 MongoDB
type TaskRecordManager struct {
	taskRecords *mongo.Collection
}

// NewTaskRecordManager 创建任务记录管理器
func NewTaskRecordManager(mongoDB database.MongoDB) (*TaskRecordManager, error) {
	if mongoDB == nil {
		return nil, nil
	}

	manager := &TaskRecordManager{
		taskRecords: mongoDB.GetCollection(model.TaskQueueRecord{}.CollectionName()),
	}

	// 创建索引
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "task_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "status", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "created_at", Value: -1}},
		},
	}
	_, err := manager.taskRecords.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		log.Warnw("failed to create task records indexes", "error", err)
	}

	return manager, nil
}

// RecordTaskEnqueued 记录任务入队
func (m *TaskRecordManager) RecordTaskEnqueued(payload *TaskPayload, queueName string) {
	if m == nil || m.taskRecords == nil {
		return
	}

	now := time.Now()
	record := &model.TaskQueueRecord{
		TaskID:        payload.TaskID,
		TaskType:      payload.TaskType,
		Status:        TaskRecordStatusPending,
		Queue:         queueName,
		Priority:      payload.Priority,
		PipelineID:    payload.PipelineID,
		PipelineRunID: payload.PipelineRunID,
		StageID:       payload.StageID,
		AgentID:       payload.AgentID,
		Payload:       payload.Data,
		CreateTime:    now,
		RetryCount:    payload.RetryCount,
		CurrentRetry:  0,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := m.taskRecords.InsertOne(ctx, record)
	if err != nil {
		log.Warnw("failed to record task enqueued", "task_id", payload.TaskID, "error", err)
	}
}

// RecordTaskStarted 记录任务开始
func (m *TaskRecordManager) RecordTaskStarted(payload *TaskPayload) {
	if m == nil || m.taskRecords == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"status":     TaskRecordStatusRunning,
			"start_time": &now,
		},
	}

	_, err := m.taskRecords.UpdateOne(
		ctx,
		bson.M{"task_id": payload.TaskID},
		update,
	)
	if err != nil {
		log.Warnw("failed to record task started", "task_id", payload.TaskID, "error", err)
	}
}

// RecordTaskCompleted 记录任务完成
func (m *TaskRecordManager) RecordTaskCompleted(payload *TaskPayload) {
	if m == nil || m.taskRecords == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"status":   TaskRecordStatusCompleted,
			"end_time": &now,
		},
	}

	// 先获取开始时间以计算耗时
	var record model.TaskQueueRecord
	_ = m.taskRecords.FindOne(ctx, bson.M{"task_id": payload.TaskID}).Decode(&record)
	if record.StartTime != nil {
		duration := now.Sub(*record.StartTime).Milliseconds()
		update["$set"].(bson.M)["duration"] = duration
	}

	_, err := m.taskRecords.UpdateOne(
		ctx,
		bson.M{"task_id": payload.TaskID},
		update,
	)
	if err != nil {
		log.Warnw("failed to record task completed", "task_id", payload.TaskID, "error", err)
	}
}

// RecordTaskFailed 记录任务失败
func (m *TaskRecordManager) RecordTaskFailed(payload *TaskPayload, err error) {
	if m == nil || m.taskRecords == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"status":        TaskRecordStatusFailed,
			"error_message": err.Error(),
			"end_time":      &now,
		},
	}

	// 先获取开始时间以计算耗时
	var record model.TaskQueueRecord
	_ = m.taskRecords.FindOne(ctx, bson.M{"task_id": payload.TaskID}).Decode(&record)
	if record.StartTime != nil {
		duration := now.Sub(*record.StartTime).Milliseconds()
		update["$set"].(bson.M)["duration"] = duration
	}

	_, updateErr := m.taskRecords.UpdateOne(
		ctx,
		bson.M{"task_id": payload.TaskID},
		update,
	)
	if updateErr != nil {
		log.Warnw("failed to record task failed", "task_id", payload.TaskID, "error", updateErr)
	}
}
