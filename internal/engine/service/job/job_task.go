package job

import (
	"context"
	"fmt"
	"sync"
	"time"

	agentapi "github.com/observabil/arcade/api/agent/v1"
	"github.com/observabil/arcade/internal/engine/model"
	"github.com/observabil/arcade/internal/engine/repo"
	"github.com/observabil/arcade/pkg/log"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2025/01/13
 * @file: job_task.go
 * @description: Concrete implementation of JobTask
 */

// ConcreteJobTask 具体的 Job 任务实现
type ConcreteJobTask struct {
	job         *model.Job
	jobRepo     *repo.AgentRepo
	agentClient agentapi.AgentClient
}

// jobTaskPool JobTask 对象池
var jobTaskPool = sync.Pool{
	New: func() any {
		return &ConcreteJobTask{}
	},
}

// NewConcreteJobTask 创建 Job 任务（从对象池获取）
func NewConcreteJobTask(job *model.Job, jobRepo *repo.AgentRepo, agentClient agentapi.AgentClient) *ConcreteJobTask {
	task := jobTaskPool.Get().(*ConcreteJobTask)
	task.job = job
	task.jobRepo = jobRepo
	task.agentClient = agentClient
	return task
}

// Release 释放任务对象回对象池
func (t *ConcreteJobTask) Release() {
	// 清空字段，避免内存泄漏
	t.job = nil
	t.jobRepo = nil
	t.agentClient = nil
	jobTaskPool.Put(t)
}

// GetJobID 实现 JobTask 接口
func (t *ConcreteJobTask) GetJobID() string {
	return t.job.JobId
}

// GetPriority 实现 JobTask 接口
func (t *ConcreteJobTask) GetPriority() int {
	return t.job.Priority
}

// Execute 实现 JobTask 接口
func (t *ConcreteJobTask) Execute(ctx context.Context) error {
	jobId := t.job.JobId
	log.Infof("starting execution of job %s", jobId)

	// 确保任务执行完成后释放回对象池
	defer t.Release()

	// 更新任务状态为运行中
	if err := t.updateJobStatus(JobStatusRunning, ""); err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	// 记录开始时间
	startTime := time.Now()

	// 执行任务的主要逻辑
	// 这里可以是：
	// 1. 通过 gRPC 调用 Agent 执行任务
	// 2. 本地执行任务
	// 3. 调用其他服务
	err := t.executeJobLogic(ctx)

	// 记录结束时间和时长
	duration := time.Since(startTime)

	if err != nil {
		// 任务失败
		log.Errorf("job %s failed after %v: %v", jobId, duration, err)
		if updateErr := t.updateJobStatus(JobStatusFailed, err.Error()); updateErr != nil {
			log.Errorf("failed to update job status: %v", updateErr)
		}
		return err
	}

	// 任务成功
	log.Infof("job %s completed successfully in %v", jobId, duration)
	if err := t.updateJobStatus(JobStatusSuccess, ""); err != nil {
		log.Errorf("failed to update job status: %v", err)
	}

	return nil
}

// executeJobLogic 执行任务的核心逻辑
func (t *ConcreteJobTask) executeJobLogic(ctx context.Context) error {
	// TODO: 实现具体的任务执行逻辑
	// 示例：调用 Agent 的 gRPC 接口执行任务

	// 模拟任务执行
	select {
	case <-ctx.Done():
		return fmt.Errorf("job execution cancelled: %w", ctx.Err())
	case <-time.After(2 * time.Second):
		// 实际应该调用 Agent gRPC 接口
		log.Infof("job %s executed successfully", t.job.JobId)
		return nil
	}
}

// updateJobStatus 更新任务状态
func (t *ConcreteJobTask) updateJobStatus(status int, errorMsg string) error {
	// TODO: 调用 JobRepo 更新状态
	log.Infof("updating job %s status to %d", t.job.JobId, status)
	return nil
}

// Job 状态常量
const (
	JobStatusUnknown   = 0
	JobStatusPending   = 1
	JobStatusQueued    = 2
	JobStatusRunning   = 3
	JobStatusSuccess   = 4
	JobStatusFailed    = 5
	JobStatusCancelled = 6
	JobStatusTimeout   = 7
	JobStatusSkipped   = 8
)
