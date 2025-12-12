package service

import (
	"sync"
	"time"

	"github.com/go-arcade/arcade/internal/engine/model"
	pluginrepo "github.com/go-arcade/arcade/internal/engine/repo"
	"github.com/go-arcade/arcade/pkg/database"
	"github.com/go-arcade/arcade/pkg/id"
	"github.com/go-arcade/arcade/pkg/log"
)

// PluginDaemonTaskStatus 守护任务状态
type PluginDaemonTaskStatus string

const (
	DaemonTaskStatusPending   PluginDaemonTaskStatus = "pending"   // 等待中
	DaemonTaskStatusRunning   PluginDaemonTaskStatus = "running"   // 执行中
	DaemonTaskStatusCompleted PluginDaemonTaskStatus = "completed" // 已完成
	DaemonTaskStatusFailed    PluginDaemonTaskStatus = "failed"    // 失败
)

// PluginInstallDaemonTask 插件安装守护任务（内存中的结构）
type PluginInstallDaemonTask struct {
	TaskID        string                 `json:"taskId"`        // 任务ID
	PluginName    string                 `json:"pluginName"`    // 插件名称
	Version       string                 `json:"version"`       // 版本
	Status        PluginDaemonTaskStatus `json:"status"`        // 状态
	Progress      int                    `json:"progress"`      // 进度 0-100
	Message       string                 `json:"message"`       // 消息
	Error         string                 `json:"error"`         // 错误信息
	PluginID      string                 `json:"pluginId"`      // 安装成功后的插件ID
	CreateTime    time.Time              `json:"createTime"`    // 创建时间
	UpdateTime    time.Time              `json:"updateTime"`    // 更新时间
	CompletedTime *time.Time             `json:"completedTime"` // 完成时间
	Duration      int64                  `json:"duration"`      // 安装耗时（秒）
}

// DaemonTaskManager 守护任务管理器（带MongoDB持久化）
type DaemonTaskManager struct {
	daemonTaskRepo pluginrepo.IPluginTaskRepository
}

var (
	daemonTaskManager *DaemonTaskManager
	once              sync.Once
)

// GetDaemonTaskManager 获取守护任务管理器单例
func GetDaemonTaskManager() *DaemonTaskManager {
	if daemonTaskManager == nil {
		log.Warn("[DaemonTaskManager] not initialized, call InitDaemonTaskManager first")
	}
	return daemonTaskManager
}

// InitDaemonTaskManager 初始化守护任务管理器（需要在应用启动时调用）
func InitDaemonTaskManager(mongoDB database.MongoDB) *DaemonTaskManager {
	once.Do(func() {
		repo := pluginrepo.NewPluginTaskRepo(mongoDB)
		if err := repo.CreateIndexes(); err != nil {
			log.Warnw("[DaemonTaskManager] failed to create indexes", "error", err)
		}
		daemonTaskManager = &DaemonTaskManager{daemonTaskRepo: repo}
		log.Info("[DaemonTaskManager] initialized with MongoDB persistence")
	})
	return daemonTaskManager
}

// CreateDaemonTask 创建守护任务
func (tm *DaemonTaskManager) CreateDaemonTask(pluginName, version string) *PluginInstallDaemonTask {
	if tm == nil || tm.daemonTaskRepo == nil {
		log.Error("[DaemonTaskManager] not initialized")
		return nil
	}

	daemonTaskModel := &model.PluginInstallRecords{
		TaskID:     id.GetUUID(),
		PluginName: pluginName,
		Version:    version,
		Status:     string(DaemonTaskStatusPending),
		Progress:   0,
		Message:    "DaemonTask created, waiting for execution",
		Source:     "local",
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}

	if err := tm.daemonTaskRepo.CreateTask(daemonTaskModel); err != nil {
		log.Errorw("[DaemonTaskManager] failed to create daemon task", "pluginName", pluginName, "version", version, "error", err)
		return nil
	}

	log.Infow("[DaemonTaskManager] created daemon task", "taskId", daemonTaskModel.TaskID, "pluginName", pluginName, "version", version)

	return tm.modelToDaemonTask(daemonTaskModel)
}

// GetDaemonTask 获取守护任务（从MongoDB读取）
func (tm *DaemonTaskManager) GetDaemonTask(daemonTaskID string) *PluginInstallDaemonTask {
	if tm == nil || tm.daemonTaskRepo == nil {
		log.Error("[DaemonTaskManager] not initialized")
		return nil
	}

	daemonTaskModel, err := tm.daemonTaskRepo.GetTaskByID(daemonTaskID)
	if err != nil {
		log.Debugw("[DaemonTaskManager] failed to get daemon task", "taskId", daemonTaskID, "error", err)
		return nil
	}

	return tm.modelToDaemonTask(daemonTaskModel)
}

// UpdateDaemonTask 更新守护任务
func (tm *DaemonTaskManager) UpdateDaemonTask(daemonTaskID string, status PluginDaemonTaskStatus, progress int, message string) {
	if tm == nil || tm.daemonTaskRepo == nil {
		log.Error("[DaemonTaskManager] not initialized")
		return
	}

	// 先获取守护任务
	daemonTaskModel, err := tm.daemonTaskRepo.GetTaskByID(daemonTaskID)
	if err != nil {
		log.Errorw("[DaemonTaskManager] failed to get daemon task for update", "taskId", daemonTaskID, "error", err)
		return
	}

	// 更新字段
	daemonTaskModel.Status = string(status)
	daemonTaskModel.Progress = progress
	daemonTaskModel.Message = message
	daemonTaskModel.UpdateTime = time.Now()

	if status == DaemonTaskStatusCompleted || status == DaemonTaskStatusFailed {
		now := time.Now()
		daemonTaskModel.CompletedTime = &now
		// 计算耗时（秒）
		daemonTaskModel.Duration = int64(now.Sub(daemonTaskModel.CreateTime).Seconds())
	}

	// 保存到MongoDB
	if err := tm.daemonTaskRepo.UpdateTask(daemonTaskModel); err != nil {
		log.Errorw("[DaemonTaskManager] failed to update daemon task", "taskId", daemonTaskID, "error", err)
	}
}

// UpdateDaemonTaskError 更新守护任务错误（持久化到MongoDB）
func (tm *DaemonTaskManager) UpdateDaemonTaskError(daemonTaskID string, err error) {
	if tm == nil || tm.daemonTaskRepo == nil {
		log.Error("[DaemonTaskManager] not initialized")
		return
	}

	daemonTaskModel, getErr := tm.daemonTaskRepo.GetTaskByID(daemonTaskID)
	if getErr != nil {
		log.Errorw("[DaemonTaskManager] failed to get daemon task for error update", "taskId", daemonTaskID, "error", getErr)
		return
	}

	daemonTaskModel.Status = string(DaemonTaskStatusFailed)
	daemonTaskModel.Error = err.Error()
	daemonTaskModel.Message = "install failed"
	daemonTaskModel.UpdateTime = time.Now()
	now := time.Now()
	daemonTaskModel.CompletedTime = &now
	// 计算耗时（秒）
	daemonTaskModel.Duration = int64(now.Sub(daemonTaskModel.CreateTime).Seconds())

	if updateErr := tm.daemonTaskRepo.UpdateTask(daemonTaskModel); updateErr != nil {
		log.Errorw("[DaemonTaskManager] failed to update daemon task error", "taskId", daemonTaskID, "error", updateErr)
	}
}

// UpdateDaemonTaskSuccess 更新守护任务成功
func (tm *DaemonTaskManager) UpdateDaemonTaskSuccess(daemonTaskID string, pluginID string) {
	if tm == nil || tm.daemonTaskRepo == nil {
		log.Error("[DaemonTaskManager] not initialized")
		return
	}

	daemonTaskModel, err := tm.daemonTaskRepo.GetTaskByID(daemonTaskID)
	if err != nil {
		log.Errorw("[DaemonTaskManager] failed to get daemon task for success update", "taskId", daemonTaskID, "error", err)
		return
	}

	daemonTaskModel.Status = string(DaemonTaskStatusCompleted)
	daemonTaskModel.Progress = 100
	daemonTaskModel.PluginID = pluginID
	daemonTaskModel.Message = "install success"
	daemonTaskModel.UpdateTime = time.Now()
	now := time.Now()
	daemonTaskModel.CompletedTime = &now
	// 计算耗时（秒）
	daemonTaskModel.Duration = int64(now.Sub(daemonTaskModel.CreateTime).Seconds())

	if err := tm.daemonTaskRepo.UpdateTask(daemonTaskModel); err != nil {
		log.Errorw("[DaemonTaskManager] failed to update daemon task success", "taskId", daemonTaskID, "error", err)
	}
}

// ListDaemonTasks 列出所有守护任务
func (tm *DaemonTaskManager) ListDaemonTasks() []*PluginInstallDaemonTask {
	if tm == nil || tm.daemonTaskRepo == nil {
		log.Error("[DaemonTaskManager] not initialized")
		return []*PluginInstallDaemonTask{}
	}

	daemonTaskModels, err := tm.daemonTaskRepo.ListAllTasks()
	if err != nil {
		log.Errorw("[DaemonTaskManager] failed to list daemon tasks", "error", err)
		return []*PluginInstallDaemonTask{}
	}

	daemonTasks := make([]*PluginInstallDaemonTask, 0, len(daemonTaskModels))
	for _, daemonTaskModel := range daemonTaskModels {
		daemonTasks = append(daemonTasks, tm.modelToDaemonTask(daemonTaskModel))
	}

	return daemonTasks
}

// CleanupOldDaemonTasks 清理旧守护任务（保留最近24小时的任务）
func (tm *DaemonTaskManager) CleanupOldDaemonTasks() {
	if tm == nil || tm.daemonTaskRepo == nil {
		log.Error("[DaemonTaskManager] not initialized")
		return
	}

	cutoff := time.Now().Add(-24 * time.Hour)
	count, err := tm.daemonTaskRepo.DeleteOldTasks(cutoff)
	if err != nil {
		log.Errorw("[DaemonTaskManager] failed to cleanup old daemon tasks", "error", err)
		return
	}

	if count > 0 {
		log.Infow("[DaemonTaskManager] cleaned up old daemon tasks", "count", count)
	}
}

// modelToDaemonTask 将MongoDB模型转换为内存守护任务结构
func (tm *DaemonTaskManager) modelToDaemonTask(m *model.PluginInstallRecords) *PluginInstallDaemonTask {
	return &PluginInstallDaemonTask{
		TaskID:        m.TaskID,
		PluginName:    m.PluginName,
		Version:       m.Version,
		Status:        PluginDaemonTaskStatus(m.Status),
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
