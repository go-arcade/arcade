package task

import (
	"context"
	"fmt"
	"sync"
	"time"

	agentv1 "github.com/go-arcade/arcade/api/agent/v1"
	taskmodel "github.com/go-arcade/arcade/internal/engine/model/task"
	agentrepo "github.com/go-arcade/arcade/internal/engine/repo/agent"
	"github.com/go-arcade/arcade/pkg/log"
)


// ConcreteTask 具体的 Task 任务实现
type ConcreteTask struct {
	task        *taskmodel.Task
	taskRepo    *agentrepo.IAgentRepository
	agentClient agentv1.AgentServiceClient
}

// taskPool Task 对象池
var taskPool = sync.Pool{
	New: func() any {
		return &ConcreteTask{}
	},
}

// NewConcreteTask 创建 Task 任务（从对象池获取）
func NewConcreteTask(task *taskmodel.Task, taskRepo *agentrepo.IAgentRepository, agentClient agentv1.AgentServiceClient) *ConcreteTask {
	t := taskPool.Get().(*ConcreteTask)
	t.task = task
	t.taskRepo = taskRepo
	t.agentClient = agentClient
	return t
}

// Release 释放任务对象回对象池
func (t *ConcreteTask) Release() {
	// 清空字段，避免内存泄漏
	t.task = nil
	t.taskRepo = nil
	t.agentClient = nil
	taskPool.Put(t)
}

// GetTaskID 实现 Task 接口
func (t *ConcreteTask) GetTaskID() string {
	return t.task.TaskId
}

// GetPriority 实现 Task 接口
func (t *ConcreteTask) GetPriority() int {
	return t.task.Priority
}

// Execute 实现 Task 接口
func (t *ConcreteTask) Execute(ctx context.Context) error {
	taskId := t.task.TaskId
	log.Infof("starting execution of task %s", taskId)

	// 确保任务执行完成后释放回对象池
	defer t.Release()

	// 更新任务状态为运行中
	if err := t.updateTaskStatus(TaskStatusRunning, ""); err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	// 记录开始时间
	startTime := time.Now()

	// 执行任务的主要逻辑
	// 这里可以是：
	// 1. 通过 gRPC 调用 Agent 执行任务
	// 2. 本地执行任务
	// 3. 调用其他服务
	err := t.executeTaskLogic(ctx)

	// 记录结束时间和时长
	duration := time.Since(startTime)

	if err != nil {
		// 任务失败
		log.Errorf("task %s failed after %v: %v", taskId, duration, err)
		if updateErr := t.updateTaskStatus(TaskStatusFailed, err.Error()); updateErr != nil {
			log.Errorf("failed to update task status: %v", updateErr)
		}
		return err
	}

	// 任务成功
	log.Infof("task %s completed successfully in %v", taskId, duration)
	if err := t.updateTaskStatus(TaskStatusSuccess, ""); err != nil {
		log.Errorf("failed to update task status: %v", err)
	}

	return nil
}

// executeTaskLogic 执行任务的核心逻辑
func (t *ConcreteTask) executeTaskLogic(ctx context.Context) error {
	// TODO: 实现具体的任务执行逻辑
	// 示例：调用 Agent 的 gRPC 接口执行任务

	// 模拟任务执行
	select {
	case <-ctx.Done():
		return fmt.Errorf("task execution cancelled: %w", ctx.Err())
	case <-time.After(2 * time.Second):
		// 实际应该调用 Agent gRPC 接口
		log.Infof("task %s executed successfully", t.task.TaskId)
		return nil
	}
}

// updateTaskStatus 更新任务状态
func (t *ConcreteTask) updateTaskStatus(status int, errorMsg string) error {
	// TODO: 调用 TaskRepo 更新状态
	log.Infof("updating task %s status to %d", t.task.TaskId, status)
	return nil
}

// Task 状态常量
const (
	TaskStatusUnknown   = 0
	TaskStatusPending   = 1
	TaskStatusQueued    = 2
	TaskStatusRunning   = 3
	TaskStatusSuccess   = 4
	TaskStatusFailed    = 5
	TaskStatusCancelled = 6
	TaskStatusTimeout   = 7
	TaskStatusSkipped   = 8
)
