package nova

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/go-arcade/arcade/pkg/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	// TaskRecordCollectionName is the default MongoDB collection name for task records
	TaskRecordCollectionName = "task_records"
)

// MongoTaskRecorder implements TaskRecorder interface using MongoDB
type MongoTaskRecorder struct {
	collection *mongo.Collection
}

// NewMongoTaskRecorder creates a new MongoDB task recorder
// mongoDB: MongoDB database interface
// collectionName: collection name, if empty, uses default TaskRecordCollectionName
func NewMongoTaskRecorder(mongoDB database.MongoDB, collectionName string) (*MongoTaskRecorder, error) {
	if mongoDB == nil {
		return nil, fmt.Errorf("mongoDB cannot be nil")
	}

	if collectionName == "" {
		collectionName = TaskRecordCollectionName
	}

	recorder := &MongoTaskRecorder{
		collection: mongoDB.GetCollection(collectionName),
	}

	// Create indexes
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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
			Keys: bson.D{{Key: "queue", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "priority", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "created_at", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "queued_at", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "completed_at", Value: -1}},
		},
	}

	_, err := recorder.collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return nil, fmt.Errorf("failed to create indexes: %w", err)
	}

	return recorder, nil
}

// Record records a task in MongoDB
// Uses upsert operation: updates if exists, inserts if not
func (r *MongoTaskRecorder) Record(ctx context.Context, record *TaskRecord) error {
	if record == nil {
		return fmt.Errorf("task record cannot be nil")
	}

	// Convert to MongoDB document
	doc := r.taskRecordToDoc(record)

	// Use upsert operation: update if exists, insert if not
	opts := options.Update().SetUpsert(true)
	filter := bson.M{"task_id": record.TaskID}
	update := bson.M{"$set": doc}

	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to record task: %w", err)
	}

	return nil
}

// UpdateStatus updates the task status in MongoDB
// Automatically sets timestamp fields based on the status
func (r *MongoTaskRecorder) UpdateStatus(ctx context.Context, taskID string, status TaskStatus, err error) error {
	if taskID == "" {
		return fmt.Errorf("task ID cannot be empty")
	}

	update := bson.M{
		"$set": bson.M{
			"status": status,
		},
	}

	now := time.Now()
	switch status {
	case TaskStatusProcessing:
		update["$set"].(bson.M)["started_at"] = now
	case TaskStatusCompleted:
		update["$set"].(bson.M)["completed_at"] = now
	case TaskStatusFailed:
		update["$set"].(bson.M)["failed_at"] = now
		if err != nil {
			update["$set"].(bson.M)["error"] = err.Error()
		}
	}

	filter := bson.M{"task_id": taskID}
	result, updateErr := r.collection.UpdateOne(ctx, filter, update)
	if updateErr != nil {
		return fmt.Errorf("failed to update task status: %w", updateErr)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("task not found: %s", taskID)
	}

	return nil
}

// Get retrieves a task record by task ID
func (r *MongoTaskRecorder) Get(ctx context.Context, taskID string) (*TaskRecord, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task ID cannot be empty")
	}

	filter := bson.M{"task_id": taskID}
	var doc bson.M

	err := r.collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("task record not found: %s", taskID)
		}
		return nil, fmt.Errorf("failed to get task record: %w", err)
	}

	return r.docToTaskRecord(doc)
}

// ListTaskRecords lists task records based on filter criteria
// Results are sorted by created_at in descending order
func (r *MongoTaskRecorder) ListTaskRecords(ctx context.Context, filter *TaskRecordFilter) ([]*TaskRecord, error) {
	// Build MongoDB query filter
	mongoFilter := bson.M{}

	if filter != nil {
		if len(filter.Status) > 0 {
			mongoFilter["status"] = bson.M{"$in": filter.Status}
		}
		if filter.Queue != "" {
			mongoFilter["queue"] = filter.Queue
		}
		if filter.Priority != nil {
			mongoFilter["priority"] = *filter.Priority
		}
		if filter.StartTime != nil || filter.EndTime != nil {
			timeFilter := bson.M{}
			if filter.StartTime != nil {
				timeFilter["$gte"] = *filter.StartTime
			}
			if filter.EndTime != nil {
				timeFilter["$lte"] = *filter.EndTime
			}
			mongoFilter["created_at"] = timeFilter
		}
		if len(filter.Metadata) > 0 {
			for k, v := range filter.Metadata {
				mongoFilter["metadata."+k] = v
			}
		}
	}

	// Build query options
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	if filter != nil {
		if filter.Limit > 0 {
			opts.SetLimit(int64(filter.Limit))
		}
		if filter.Offset > 0 {
			opts.SetSkip(int64(filter.Offset))
		}
	}

	// Execute query
	cursor, err := r.collection.Find(ctx, mongoFilter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list task records: %w", err)
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err = cursor.Close(ctx)
		if err != nil {
			return
		}
	}(cursor, ctx)

	var records []*TaskRecord
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}

		record, err := r.docToTaskRecord(doc)
		if err != nil {
			continue
		}

		records = append(records, record)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return records, nil
}

// Delete deletes a task record by task ID
func (r *MongoTaskRecorder) Delete(ctx context.Context, taskID string) error {
	if taskID == "" {
		return fmt.Errorf("task ID cannot be empty")
	}

	filter := bson.M{"task_id": taskID}
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete task record: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("task record not found: %s", taskID)
	}

	return nil
}

// taskRecordToDoc converts TaskRecord to MongoDB document
func (r *MongoTaskRecorder) taskRecordToDoc(record *TaskRecord) bson.M {
	doc := bson.M{
		"task_id":     record.TaskID,
		"status":      record.Status,
		"queue":       record.Queue,
		"priority":    record.Priority,
		"created_at":  record.CreatedAt,
		"retry_count": record.RetryCount,
	}

	if record.Task != nil {
		doc["task"] = bson.M{
			"type":    record.Task.Type,
			"payload": record.Task.Payload,
		}
	}

	if record.QueuedAt != nil {
		doc["queued_at"] = *record.QueuedAt
	}
	if record.ProcessAt != nil {
		doc["process_at"] = *record.ProcessAt
	}
	if record.StartedAt != nil {
		doc["started_at"] = *record.StartedAt
	}
	if record.CompletedAt != nil {
		doc["completed_at"] = *record.CompletedAt
	}
	if record.FailedAt != nil {
		doc["failed_at"] = *record.FailedAt
	}
	if record.Error != "" {
		doc["error"] = record.Error
	}
	if len(record.Metadata) > 0 {
		doc["metadata"] = record.Metadata
	}

	return doc
}

// docToTaskRecord converts MongoDB document to TaskRecord
func (r *MongoTaskRecorder) docToTaskRecord(doc bson.M) (*TaskRecord, error) {
	record := &TaskRecord{}

	if taskID, ok := doc["task_id"].(string); ok {
		record.TaskID = taskID
	}

	if status, ok := doc["status"].(string); ok {
		record.Status = TaskStatus(status)
	}

	if queue, ok := doc["queue"].(string); ok {
		record.Queue = queue
	}

	if priority, ok := doc["priority"].(int32); ok {
		record.Priority = Priority(priority)
	} else if priority, ok := doc["priority"].(int64); ok {
		record.Priority = Priority(priority)
	} else if priority, ok := doc["priority"].(int); ok {
		record.Priority = Priority(priority)
	}

	if createdAt, ok := doc["created_at"].(primitive.DateTime); ok {
		record.CreatedAt = createdAt.Time()
	} else if createdAt, ok := doc["created_at"].(time.Time); ok {
		record.CreatedAt = createdAt
	}

	if retryCount, ok := doc["retry_count"].(int32); ok {
		record.RetryCount = int(retryCount)
	} else if retryCount, ok := doc["retry_count"].(int64); ok {
		record.RetryCount = int(retryCount)
	} else if retryCount, ok := doc["retry_count"].(int); ok {
		record.RetryCount = retryCount
	}

	// Parse task
	if taskDoc, ok := doc["task"].(bson.M); ok {
		record.Task = &Task{}
		if taskType, ok := taskDoc["type"].(string); ok {
			record.Task.Type = taskType
		}
		if payload, ok := taskDoc["payload"].(primitive.Binary); ok {
			record.Task.Payload = payload.Data
		} else if payload, ok := taskDoc["payload"].([]byte); ok {
			record.Task.Payload = payload
		}
	}

	// Parse time fields
	if queuedAt, ok := doc["queued_at"].(primitive.DateTime); ok {
		t := queuedAt.Time()
		record.QueuedAt = &t
	} else if queuedAt, ok := doc["queued_at"].(time.Time); ok {
		record.QueuedAt = &queuedAt
	}

	if processAt, ok := doc["process_at"].(primitive.DateTime); ok {
		t := processAt.Time()
		record.ProcessAt = &t
	} else if processAt, ok := doc["process_at"].(time.Time); ok {
		record.ProcessAt = &processAt
	}

	if startedAt, ok := doc["started_at"].(primitive.DateTime); ok {
		t := startedAt.Time()
		record.StartedAt = &t
	} else if startedAt, ok := doc["started_at"].(time.Time); ok {
		record.StartedAt = &startedAt
	}

	if completedAt, ok := doc["completed_at"].(primitive.DateTime); ok {
		t := completedAt.Time()
		record.CompletedAt = &t
	} else if completedAt, ok := doc["completed_at"].(time.Time); ok {
		record.CompletedAt = &completedAt
	}

	if failedAt, ok := doc["failed_at"].(primitive.DateTime); ok {
		t := failedAt.Time()
		record.FailedAt = &t
	} else if failedAt, ok := doc["failed_at"].(time.Time); ok {
		record.FailedAt = &failedAt
	}

	if err, ok := doc["error"].(string); ok {
		record.Error = err
	}

	if metadata, ok := doc["metadata"].(bson.M); ok {
		record.Metadata = make(map[string]any)
		maps.Copy(record.Metadata, metadata)
	}

	return record, nil
}
