package plugin

import (
	"sync"
	"time"

	"github.com/go-arcade/arcade/internal/engine/model/plugin"
	pluginrepo "github.com/go-arcade/arcade/internal/engine/repo/plugin"
	"github.com/go-arcade/arcade/pkg/database"
	"github.com/go-arcade/arcade/pkg/id"
	"github.com/go-arcade/arcade/pkg/log"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2025/10/16
 * @file: service_plugin_task.go
 * @description: 插件安装任务管理（MongoDB持久化）
 */

// PluginTaskStatus 任务状态
type PluginTaskStatus string

const (
	TaskStatusPending   PluginTaskStatus = "pending"   // 等待中
	TaskStatusRunning   PluginTaskStatus = "running"   // 执行中
	TaskStatusCompleted PluginTaskStatus = "completed" // 已完成
	TaskStatusFailed    PluginTaskStatus = "failed"    // 失败
)

// PluginInstallTask 插件安装任务（内存中的结构）
type PluginInstallTask struct {
	TaskID        string           `json:"taskId"`        // 任务ID
	PluginName    string           `json:"pluginName"`    // 插件名称
	Version       string           `json:"version"`       // 版本
	Status        PluginTaskStatus `json:"status"`        // 状态
	Progress      int              `json:"progress"`      // 进度 0-100
	Message       string           `json:"message"`       // 消息
	Error         string           `json:"error"`         // 错误信息
	PluginID      string           `json:"pluginId"`      // 安装成功后的插件ID
	CreateTime    time.Time        `json:"createTime"`    // 创建时间
	UpdateTime    time.Time        `json:"updateTime"`    // 更新时间
	CompletedTime *time.Time       `json:"completedTime"` // 完成时间
	Duration      int64            `json:"duration"`      // 安装耗时（秒）
}

// TaskManager 任务管理器（带MongoDB持久化）
type TaskManager struct {
	taskRepo pluginrepo.IPluginTaskRepository
}

var (
	taskManager *TaskManager
	once        sync.Once
)

// GetTaskManager 获取任务管理器单例
func GetTaskManager() *TaskManager {
	if taskManager == nil {
		log.Warn("[TaskManager] not initialized, call InitTaskManager first")
	}
	return taskManager
}

// InitTaskManager 初始化任务管理器（需要在应用启动时调用）
func InitTaskManager(mongoDB database.MongoDB) *TaskManager {
	once.Do(func() {
		repo := pluginrepo.NewPluginTaskRepo(mongoDB)
		if err := repo.CreateIndexes(); err != nil {
			log.Warnf("[TaskManager] failed to create indexes: %v", err)
		}
		taskManager = &TaskManager{taskRepo: repo}
		log.Info("[TaskManager] initialized with MongoDB persistence")
	})
	return taskManager
}

// CreateTask 创建任务
func (tm *TaskManager) CreateTask(pluginName, version string) *PluginInstallTask {
	if tm == nil || tm.taskRepo == nil {
		log.Error("[TaskManager] not initialized")
		return nil
	}

	taskModel := &plugin.PluginInstallRecords{
		TaskID:     id.GetUUID(),
		PluginName: pluginName,
		Version:    version,
		Status:     string(TaskStatusPending),
		Progress:   0,
		Message:    "Task created, waiting for execution",
		Source:     "local",
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}

	if err := tm.taskRepo.CreateTask(taskModel); err != nil {
		log.Errorf("[TaskManager] failed to create task: %v", err)
		return nil
	}

	log.Infof("[TaskManager] created task: %s for plugin %s v%s", taskModel.TaskID, pluginName, version)

	return tm.modelToTask(taskModel)
}

// GetTask 获取任务（从MongoDB读取）
func (tm *TaskManager) GetTask(taskID string) *PluginInstallTask {
	if tm == nil || tm.taskRepo == nil {
		log.Error("[TaskManager] not initialized")
		return nil
	}

	taskModel, err := tm.taskRepo.GetTaskByID(taskID)
	if err != nil {
		log.Debugf("[TaskManager] failed to get task %s: %v", taskID, err)
		return nil
	}

	return tm.modelToTask(taskModel)
}

// UpdateTask 更新任务
func (tm *TaskManager) UpdateTask(taskID string, status PluginTaskStatus, progress int, message string) {
	if tm == nil || tm.taskRepo == nil {
		log.Error("[TaskManager] not initialized")
		return
	}

	// 先获取任务
	taskModel, err := tm.taskRepo.GetTaskByID(taskID)
	if err != nil {
		log.Errorf("[TaskManager] failed to get task %s for update: %v", taskID, err)
		return
	}

	// 更新字段
	taskModel.Status = string(status)
	taskModel.Progress = progress
	taskModel.Message = message
	taskModel.UpdateTime = time.Now()

	if status == TaskStatusCompleted || status == TaskStatusFailed {
		now := time.Now()
		taskModel.CompletedTime = &now
		// 计算耗时（秒）
		taskModel.Duration = int64(now.Sub(taskModel.CreateTime).Seconds())
	}

	// 保存到MongoDB
	if err := tm.taskRepo.UpdateTask(taskModel); err != nil {
		log.Errorf("[TaskManager] failed to update task %s: %v", taskID, err)
	}
}

// UpdateTaskError 更新任务错误（持久化到MongoDB）
func (tm *TaskManager) UpdateTaskError(taskID string, err error) {
	if tm == nil || tm.taskRepo == nil {
		log.Error("[TaskManager] not initialized")
		return
	}

	taskModel, getErr := tm.taskRepo.GetTaskByID(taskID)
	if getErr != nil {
		log.Errorf("[TaskManager] failed to get task %s for error update: %v", taskID, getErr)
		return
	}

	taskModel.Status = string(TaskStatusFailed)
	taskModel.Error = err.Error()
	taskModel.Message = "install failed"
	taskModel.UpdateTime = time.Now()
	now := time.Now()
	taskModel.CompletedTime = &now
	// 计算耗时（秒）
	taskModel.Duration = int64(now.Sub(taskModel.CreateTime).Seconds())

	if updateErr := tm.taskRepo.UpdateTask(taskModel); updateErr != nil {
		log.Errorf("[TaskManager] failed to update task error %s: %v", taskID, updateErr)
	}
}

// UpdateTaskSuccess 更新任务成功
func (tm *TaskManager) UpdateTaskSuccess(taskID string, pluginID string) {
	if tm == nil || tm.taskRepo == nil {
		log.Error("[TaskManager] not initialized")
		return
	}

	taskModel, err := tm.taskRepo.GetTaskByID(taskID)
	if err != nil {
		log.Errorf("[TaskManager] failed to get task %s for success update: %v", taskID, err)
		return
	}

	taskModel.Status = string(TaskStatusCompleted)
	taskModel.Progress = 100
	taskModel.PluginID = pluginID
	taskModel.Message = "install success"
	taskModel.UpdateTime = time.Now()
	now := time.Now()
	taskModel.CompletedTime = &now
	// 计算耗时（秒）
	taskModel.Duration = int64(now.Sub(taskModel.CreateTime).Seconds())

	if err := tm.taskRepo.UpdateTask(taskModel); err != nil {
		log.Errorf("[TaskManager] failed to update task success %s: %v", taskID, err)
	}
}

// ListTasks 列出所有任务
func (tm *TaskManager) ListTasks() []*PluginInstallTask {
	if tm == nil || tm.taskRepo == nil {
		log.Error("[TaskManager] not initialized")
		return []*PluginInstallTask{}
	}

	taskModels, err := tm.taskRepo.ListAllTasks()
	if err != nil {
		log.Errorf("[TaskManager] failed to list tasks: %v", err)
		return []*PluginInstallTask{}
	}

	tasks := make([]*PluginInstallTask, 0, len(taskModels))
	for _, taskModel := range taskModels {
		tasks = append(tasks, tm.modelToTask(taskModel))
	}

	return tasks
}

// CleanupOldTasks 清理旧任务（保留最近24小时的任务）
func (tm *TaskManager) CleanupOldTasks() {
	if tm == nil || tm.taskRepo == nil {
		log.Error("[TaskManager] not initialized")
		return
	}

	cutoff := time.Now().Add(-24 * time.Hour)
	count, err := tm.taskRepo.DeleteOldTasks(cutoff)
	if err != nil {
		log.Errorf("[TaskManager] failed to cleanup old tasks: %v", err)
		return
	}

	if count > 0 {
		log.Infof("[TaskManager] cleaned up %d old tasks", count)
	}
}

// modelToTask 将MongoDB模型转换为内存任务结构
func (tm *TaskManager) modelToTask(m *plugin.PluginInstallRecords) *PluginInstallTask {
	return &PluginInstallTask{
		TaskID:        m.TaskID,
		PluginName:    m.PluginName,
		Version:       m.Version,
		Status:        PluginTaskStatus(m.Status),
		Progress:      m.Progress,
		Message:       m.Message,
		Error:         m.Error,
		PluginID:      m.PluginID,
		CreateTime:    m.CreateTime,
		UpdateTime:    m.UpdateTime,
		CompletedTime: m.CompletedTime,
		Duration:      m.Duration,
	}
}
