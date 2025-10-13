# Job Worker Pool 使用指南

## 概述

Job Worker Pool 是一个基于协程池的任务调度系统，用于管理和执行 CI/CD 任务。

## 架构设计

```
┌─────────────────────────────────────────────┐
│         Job Worker Pool                     │
│                                             │
│  ┌─────────────┐      ┌─────────────────┐  │
│  │ Task Queue  │─────→│  Worker Pool    │  │
│  │  (Channel)  │      │  (Goroutines)   │  │
│  └─────────────┘      └─────────────────┘  │
│         ↑                      │            │
│         │                      ↓            │
│  ┌─────────────┐      ┌─────────────────┐  │
│  │  Priority   │      │   Job Executor  │  │
│  │    Queue    │      │   (gRPC Call)   │  │
│  └─────────────┘      └─────────────────┘  │
└─────────────────────────────────────────────┘
```

## 核心组件

### 1. JobWorkerPool - 协程池管理器

**职责**:
- 管理工作协程的生命周期
- 调度任务执行
- 监控和统计

**配置参数**:
- `maxWorkers` - 最大工作协程数（默认：10）
- `queueSize` - 任务队列大小（默认：100）
- `workerTimeout` - 单个任务超时时间（默认：30分钟）

### 2. PriorityQueue - 优先级队列

**职责**:
- 当任务队列满时，暂存任务
- 按优先级排序（数字越小优先级越高）
- 支持任务取消

**优先级定义**:
- 1 - 最高优先级（紧急任务）
- 5 - 普通优先级（默认）
- 10 - 最低优先级（后台任务）

### 3. Worker - 工作协程

**职责**:
- 从队列获取任务
- 执行任务
- 处理错误和超时
- 更新统计信息

### 4. JobTask - 任务接口

```go
type JobTask interface {
    GetJobID() string
    GetPriority() int
    Execute(ctx context.Context) error
}
```

## 使用方法

### 初始化协程池

```go
// 创建工作池：10个工作协程，队列大小100
pool := NewJobWorkerPool(10, 100)

// 启动协程池
if err := pool.Start(); err != nil {
    log.Fatalf("failed to start worker pool: %v", err)
}

// 程序退出时停止
defer pool.Stop()
```

### 提交任务

#### 方式 1: 普通提交（先到先得）

```go
// 创建任务
task := NewConcreteJobTask(job, jobRepo, agentClient)

// 提交到协程池
if err := pool.Submit(task); err != nil {
    log.Errorf("failed to submit task: %v", err)
}
```

#### 方式 2: 优先级提交

```go
// 创建高优先级任务
job.Priority = 1

task := NewConcreteJobTask(job, jobRepo, agentClient)

// 使用优先级队列提交
if err := pool.SubmitWithPriority(task); err != nil {
    log.Errorf("failed to submit task: %v", err)
}
```

### 取消任务

```go
// 取消指定任务
if err := pool.CancelTask("job_123"); err != nil {
    log.Warnf("failed to cancel task: %v", err)
}
```

### 动态调整协程数

```go
// 增加到 20 个工作协程
if err := pool.Resize(20); err != nil {
    log.Errorf("failed to resize pool: %v", err)
}

// 减少到 5 个工作协程
if err := pool.Resize(5); err != nil {
    log.Errorf("failed to resize pool: %v", err)
}
```

### 获取统计信息

```go
stats := pool.GetStats()
fmt.Printf("Total Submitted: %d\n", stats.TotalSubmitted)
fmt.Printf("Total Completed: %d\n", stats.TotalCompleted)
fmt.Printf("Total Failed: %d\n", stats.TotalFailed)
fmt.Printf("Active Workers: %d\n", stats.ActiveWorkers)
fmt.Printf("Queued Tasks: %d\n", stats.QueuedTasks)
fmt.Printf("Average Execution Time: %v\n", stats.AverageExecTime)
```

## 完整示例

```go
package main

import (
    "context"
    "time"
    "github.com/observabil/arcade/internal/engine/service/job"
)

func main() {
    // 1. 创建协程池
    pool := job.NewJobWorkerPool(10, 100)
    
    // 2. 启动协程池
    if err := pool.Start(); err != nil {
        panic(err)
    }
    defer pool.Stop()
    
    // 3. 提交多个任务
    for i := 0; i < 50; i++ {
        task := &job.ConcreteJobTask{
            // 初始化任务...
        }
        
        if err := pool.Submit(task); err != nil {
            log.Errorf("failed to submit task %d: %v", i, err)
        }
    }
    
    // 4. 等待一段时间
    time.Sleep(30 * time.Second)
    
    // 5. 查看统计信息
    stats := pool.GetStats()
    log.Infof("Pool stats: %+v", stats)
}
```

## 任务执行流程

```
1. 提交任务
   ↓
2. 检查队列是否有空位
   ├─ 有空位 → 直接进入任务队列
   └─ 无空位 → 进入优先级队列
   ↓
3. 优先级调度器（定时检查）
   ├─ 从优先级队列取出高优先级任务
   └─ 放入任务队列
   ↓
4. Worker 从任务队列获取任务
   ↓
5. 执行任务
   ├─ 更新状态为 Running
   ├─ 调用 Execute()
   ├─ 记录执行时间
   └─ 处理 panic
   ↓
6. 更新统计信息
   ├─ 成功 → TotalCompleted++
   ├─ 失败 → TotalFailed++
   └─ 更新平均执行时间
   ↓
7. Worker 继续处理下一个任务
```

## 配置建议

### 小规模环境（< 100 并发任务）

```go
pool := NewJobWorkerPool(
    5,    // 5个工作协程
    50,   // 队列大小50
)
```

### 中等规模环境（100-500 并发任务）

```go
pool := NewJobWorkerPool(
    20,   // 20个工作协程
    200,  // 队列大小200
)
```

### 大规模环境（> 500 并发任务）

```go
pool := NewJobWorkerPool(
    50,   // 50个工作协程
    500,  // 队列大小500
)
```

## 监控指标

### 关键指标

| 指标 | 说明 | 告警阈值 |
|------|------|----------|
| QueuedTasks | 队列中等待的任务数 | > 80% 队列容量 |
| ActiveWorkers | 活跃的工作协程数 | > 90% maxWorkers |
| TotalFailed / TotalCompleted | 失败率 | > 5% |
| AverageExecTime | 平均执行时间 | 持续上升 |

### 监控接口

```http
GET /api/v1/jobs/pool/stats

Response:
{
  "total_submitted": 1000,
  "total_completed": 950,
  "total_failed": 30,
  "total_cancelled": 20,
  "active_workers": 8,
  "queued_tasks": 15,
  "average_exec_time": "5m30s"
}
```

## 性能优化

### 1. 合理设置协程数

```go
// 根据 CPU 核心数动态设置
numCPU := runtime.NumCPU()
maxWorkers := numCPU * 2  // 一般为 CPU 核心数的 2-4 倍
```

### 2. 队列大小设置

```go
// 队列大小应该能容纳短时间内的突发任务
queueSize := maxWorkers * 10  // 一般为工作协程数的 10 倍
```

### 3. 任务分批提交

```go
// 避免一次性提交大量任务
const batchSize = 50

for i := 0; i < len(tasks); i += batchSize {
    end := i + batchSize
    if end > len(tasks) {
        end = len(tasks)
    }
    
    for _, task := range tasks[i:end] {
        pool.Submit(task)
    }
    
    time.Sleep(100 * time.Millisecond) // 批次间隔
}
```

### 4. 使用优先级队列

```go
// 关键任务使用高优先级
criticalJob.Priority = 1
pool.SubmitWithPriority(criticalTask)

// 普通任务使用默认优先级
normalJob.Priority = 5
pool.Submit(normalTask)

// 后台任务使用低优先级
backgroundJob.Priority = 10
pool.SubmitWithPriority(backgroundTask)
```

## 故障处理

### Worker Panic 恢复

协程池自动捕获和恢复 worker panic：

```go
defer func() {
    if r := recover(); r != nil {
        log.Errorf("worker panic: %v", r)
        // 自动恢复，worker 继续运行
    }
}()
```

### 任务超时处理

```go
// 每个任务都有超时控制
ctx, cancel := context.WithTimeout(context.Background(), workerTimeout)
defer cancel()

err := task.Execute(ctx)
if err == context.DeadlineExceeded {
    // 任务超时
    updateJobStatus(JobStatusTimeout, "execution timeout")
}
```

### 优雅关闭

```go
// 停止接收新任务
pool.Stop()

// 等待所有任务完成（最多等待 30 秒）
done := make(chan struct{})
go func() {
    pool.wg.Wait()
    close(done)
}()

select {
case <-done:
    log.Info("all tasks completed")
case <-time.After(30 * time.Second):
    log.Warn("force shutdown after timeout")
}
```

## 最佳实践

### 1. 任务幂等性

确保任务可以安全重试：

```go
func (t *ConcreteJobTask) Execute(ctx context.Context) error {
    // 检查任务是否已完成
    if t.isCompleted() {
        return nil
    }
    
    // 执行任务...
}
```

### 2. 错误分类

区分可重试和不可重试的错误：

```go
if err != nil {
    if isRetryable(err) {
        // 可重试错误，放回队列
        return fmt.Errorf("retryable error: %w", err)
    } else {
        // 不可重试错误，直接失败
        return fmt.Errorf("fatal error: %w", err)
    }
}
```

### 3. 资源清理

确保任务执行后清理资源：

```go
func (t *ConcreteJobTask) Execute(ctx context.Context) error {
    // 获取资源
    resource, err := acquireResource()
    if err != nil {
        return err
    }
    defer resource.Release() // 确保释放
    
    // 执行任务...
    return nil
}
```

### 4. 进度上报

定期上报任务进度：

```go
func (t *ConcreteJobTask) Execute(ctx context.Context) error {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    
    go func() {
        for range ticker.C {
            t.reportProgress()
        }
    }()
    
    // 执行任务...
    return nil
}
```

## 集成到系统

### Wire 依赖注入

```go
// internal/engine/service/job/provider.go
package job

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
    ProvideJobWorkerPool,
)

func ProvideJobWorkerPool() *JobWorkerPool {
    pool := NewJobWorkerPool(20, 200)
    if err := pool.Start(); err != nil {
        panic(fmt.Sprintf("failed to start worker pool: %v", err))
    }
    return pool
}
```

### HTTP API 示例

```go
// POST /api/v1/jobs
func (rt *Router) createJob(c *fiber.Ctx) error {
    var jobReq CreateJobRequest
    if err := c.BodyParser(&jobReq); err != nil {
        return err
    }
    
    // 创建 Job
    job := &model.Job{
        JobId: id.GetUUID(),
        Name: jobReq.Name,
        Priority: jobReq.Priority,
        // ...
    }
    
    // 保存到数据库
    if err := jobRepo.Create(job); err != nil {
        return err
    }
    
    // 创建任务
    task := NewConcreteJobTask(job, jobRepo, agentClient)
    
    // 提交到协程池
    if err := workerPool.Submit(task); err != nil {
        return err
    }
    
    return c.JSON(fiber.Map{
        "job_id": job.JobId,
        "status": "queued",
    })
}

// GET /api/v1/jobs/pool/stats
func (rt *Router) getPoolStats(c *fiber.Ctx) error {
    stats := workerPool.GetStats()
    return c.JSON(stats)
}

// POST /api/v1/jobs/:jobId/cancel
func (rt *Router) cancelJob(c *fiber.Ctx) error {
    jobId := c.Params("jobId")
    
    if err := workerPool.CancelTask(jobId); err != nil {
        return err
    }
    
    return c.JSON(fiber.Map{
        "message": "job cancelled",
    })
}
```

## 性能测试

### 基准测试

```go
func BenchmarkWorkerPool(b *testing.B) {
    pool := NewJobWorkerPool(10, 100)
    pool.Start()
    defer pool.Stop()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        task := &MockTask{jobId: fmt.Sprintf("job_%d", i)}
        pool.Submit(task)
    }
}
```

### 压力测试

```bash
# 提交 10000 个任务
for i in {1..10000}; do
    curl -X POST http://localhost:8080/api/v1/jobs \
      -H "Content-Type: application/json" \
      -d '{"name": "test-job-'$i'", "priority": 5}'
done

# 查看统计
curl http://localhost:8080/api/v1/jobs/pool/stats
```

## 监控和告警

### Prometheus 指标

```go
var (
    jobsSubmitted = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "arcade_jobs_submitted_total",
        Help: "Total number of jobs submitted",
    })
    
    jobsCompleted = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "arcade_jobs_completed_total",
        Help: "Total number of jobs completed",
    })
    
    jobsFailed = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "arcade_jobs_failed_total",
        Help: "Total number of jobs failed",
    })
    
    activeWorkers = prometheus.NewGauge(prometheus.GaugeOpts{
        Name: "arcade_active_workers",
        Help: "Number of active workers",
    })
    
    queuedTasks = prometheus.NewGauge(prometheus.GaugeOpts{
        Name: "arcade_queued_tasks",
        Help: "Number of tasks in queue",
    })
)
```

### 告警规则

```yaml
groups:
  - name: job_worker_pool
    rules:
      # 队列堆积告警
      - alert: JobQueueBacklog
        expr: arcade_queued_tasks > 80
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Job queue backlog detected"
          
      # 失败率告警
      - alert: HighJobFailureRate
        expr: rate(arcade_jobs_failed_total[5m]) / rate(arcade_jobs_completed_total[5m]) > 0.1
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High job failure rate detected"
          
      # Worker 利用率告警
      - alert: HighWorkerUtilization
        expr: arcade_active_workers / arcade_max_workers > 0.9
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "High worker utilization, consider scaling up"
```

## 常见问题

### Q: 任务卡住不执行怎么办？

A: 检查以下几点：
1. 协程池是否已启动（`pool.Start()`）
2. 队列是否已满
3. 所有 worker 是否都在执行长时间任务
4. 查看日志是否有错误

### Q: 如何提高吞吐量？

A: 
1. 增加工作协程数（`Resize()`）
2. 增加队列大小
3. 优化任务执行逻辑
4. 使用批量操作

### Q: 任务执行顺序不对？

A:
1. 使用优先级队列（`SubmitWithPriority()`）
2. 设置合适的优先级值
3. 注意：同优先级任务按提交顺序执行

### Q: 如何确保任务不丢失？

A:
1. 任务提交前先保存到数据库
2. 状态更新采用事务
3. 启用任务持久化
4. 实现任务重试机制

## 扩展功能

### 分布式任务调度

```go
// 结合 Redis 实现分布式任务队列
type DistributedJobPool struct {
    localPool  *JobWorkerPool
    redisQueue *RedisQueue
}

func (d *DistributedJobPool) Submit(task JobTask) error {
    // 先存入 Redis
    if err := d.redisQueue.Push(task); err != nil {
        return err
    }
    
    // 再提交到本地池
    return d.localPool.Submit(task)
}
```

### 任务重试

```go
type RetryableTask struct {
    JobTask
    maxRetries int
    currentRetry int
}

func (t *RetryableTask) Execute(ctx context.Context) error {
    for t.currentRetry < t.maxRetries {
        err := t.JobTask.Execute(ctx)
        if err == nil {
            return nil
        }
        
        t.currentRetry++
        if t.currentRetry < t.maxRetries {
            log.Warnf("task %s failed, retrying (%d/%d)", 
                t.GetJobID(), t.currentRetry, t.maxRetries)
            time.Sleep(time.Duration(t.currentRetry) * time.Second)
        }
    }
    
    return fmt.Errorf("task failed after %d retries", t.maxRetries)
}
```

## 参考资料

- [Go Concurrency Patterns](https://go.dev/blog/pipelines)
- [Worker Pool Pattern](https://gobyexample.com/worker-pools)
- [Priority Queue in Go](https://pkg.go.dev/container/heap)

