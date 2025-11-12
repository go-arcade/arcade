package plugin

import (
	"context"
	"fmt"
	"time"

	"github.com/go-arcade/arcade/internal/engine/model/plugin"
	"github.com/go-arcade/arcade/pkg/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2025/10/16
 * @file: repo_plugin_task.go
 * @description: 插件安装任务数据访问层
 */

type IPluginTaskRepository interface {
	CreateTask(task *plugin.PluginInstallRecords) error
	GetTaskByID(taskID string) (*plugin.PluginInstallRecords, error)
	UpdateTask(task *plugin.PluginInstallRecords) error
	ListTasks(filter bson.M, skip, limit int64) ([]*plugin.PluginInstallRecords, error)
	ListAllTasks() ([]*plugin.PluginInstallRecords, error)
	ListTasksByStatus(status string) ([]*plugin.PluginInstallRecords, error)
	CountTasks(filter bson.M) (int64, error)
	DeleteTask(taskID string) error
	DeleteOldTasks(before time.Time) (int64, error)
	CreateIndexes() error
}

type PluginTaskRepo struct {
	mongo      database.MongoDB
	collection *mongo.Collection
}

func NewPluginTaskRepo(mongo database.MongoDB) IPluginTaskRepository {
	// 直接使用MongoDB接口的GetCollection方法，无需再指定数据库
	collection := mongo.GetCollection(plugin.PluginInstallRecords{}.CollectionName())
	return &PluginTaskRepo{
		mongo:      mongo,
		collection: collection,
	}
}

// CreateTask 创建任务
func (r *PluginTaskRepo) CreateTask(task *plugin.PluginInstallRecords) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	task.CreateTime = time.Now()
	task.UpdateTime = time.Now()

	_, err := r.collection.InsertOne(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to insert task: %w", err)
	}

	return nil
}

// GetTaskByID 根据任务ID获取任务
func (r *PluginTaskRepo) GetTaskByID(taskID string) (*plugin.PluginInstallRecords, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var task plugin.PluginInstallRecords
	err := r.collection.FindOne(ctx, bson.M{"task_id": taskID}).Decode(&task)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("task not found: %s", taskID)
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return &task, nil
}

// UpdateTask 更新任务
func (r *PluginTaskRepo) UpdateTask(task *plugin.PluginInstallRecords) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	task.UpdateTime = time.Now()

	update := bson.M{
		"$set": bson.M{
			"status":      task.Status,
			"progress":    task.Progress,
			"message":     task.Message,
			"error":       task.Error,
			"plugin_id":   task.PluginID,
			"update_time": task.UpdateTime,
			"duration":    task.Duration,
		},
	}

	if task.CompletedTime != nil {
		update["$set"].(bson.M)["completed_time"] = task.CompletedTime
	}

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"task_id": task.TaskID},
		update,
	)

	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("task not found: %s", task.TaskID)
	}

	return nil
}

// ListTasks 列出任务（支持分页和过滤）
func (r *PluginTaskRepo) ListTasks(filter bson.M, skip, limit int64) ([]*plugin.PluginInstallRecords, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().
		SetSort(bson.D{{Key: "create_time", Value: -1}}). // 按创建时间倒序
		SetSkip(skip).
		SetLimit(limit)

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}
	defer cursor.Close(ctx)

	var tasks []*plugin.PluginInstallRecords
	if err := cursor.All(ctx, &tasks); err != nil {
		return nil, fmt.Errorf("failed to decode tasks: %w", err)
	}

	return tasks, nil
}

// ListAllTasks 列出所有任务
func (r *PluginTaskRepo) ListAllTasks() ([]*plugin.PluginInstallRecords, error) {
	return r.ListTasks(bson.M{}, 0, 100) // 默认返回最近100条
}

// ListTasksByStatus 根据状态列出任务
func (r *PluginTaskRepo) ListTasksByStatus(status string) ([]*plugin.PluginInstallRecords, error) {
	return r.ListTasks(bson.M{"status": status}, 0, 100)
}

// CountTasks 统计任务数量
func (r *PluginTaskRepo) CountTasks(filter bson.M) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count tasks: %w", err)
	}

	return count, nil
}

// DeleteTask 删除任务
func (r *PluginTaskRepo) DeleteTask(taskID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := r.collection.DeleteOne(ctx, bson.M{"task_id": taskID})
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("task not found: %s", taskID)
	}

	return nil
}

// DeleteOldTasks 删除旧任务（完成时间早于指定时间）
func (r *PluginTaskRepo) DeleteOldTasks(before time.Time) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"completed_time": bson.M{"$lt": before},
		"status": bson.M{"$in": []string{
			plugin.TaskStatusCompleted,
			plugin.TaskStatusFailed,
		}},
	}

	result, err := r.collection.DeleteMany(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old tasks: %w", err)
	}

	return result.DeletedCount, nil
}

// CreateIndexes 创建索引
func (r *PluginTaskRepo) CreateIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
			Keys: bson.D{{Key: "plugin_name", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "create_time", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "completed_time", Value: -1}},
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	return nil
}
