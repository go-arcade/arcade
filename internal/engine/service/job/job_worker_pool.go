package job

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/observabil/arcade/pkg/log"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2025/01/13
 * @file: job_worker_pool.go
 * @description: Job worker pool with goroutine pool and sync.Pool optimization
 */

// JobTask Job 任务接口
type JobTask interface {
	GetJobID() string
	GetPriority() int
	Execute(ctx context.Context) error
}

// contextPool context.Context 对象池（用于任务执行）
var contextPool = sync.Pool{
	New: func() interface{} {
		return context.Background()
	},
}

// statsUpdatePool 统计更新临时结构池
var statsUpdatePool = sync.Pool{
	New: func() interface{} {
		return &statsUpdate{}
	},
}

// statsUpdate 统计更新临时结构
type statsUpdate struct {
	completed bool
	failed    bool
	cancelled bool
	duration  time.Duration
	activeInc int
	activeDec int
}

// JobWorkerPool Job 协程池
type JobWorkerPool struct {
	mu sync.RWMutex

	// 配置
	maxWorkers    int           // 最大工作协程数
	queueSize     int           // 队列大小
	workerTimeout time.Duration // 工作超时时间

	// 任务队列
	taskQueue     chan JobTask
	priorityQueue *PriorityQueue // 优先级队列

	// 工作协程管理
	workers      []*worker
	workerCtx    context.Context
	workerCancel context.CancelFunc
	wg           sync.WaitGroup

	// 统计信息
	stats *PoolStats

	// 生命周期
	running bool
}

// worker 工作协程
type worker struct {
	id       int
	pool     *JobWorkerPool
	taskChan chan JobTask
	stopChan chan struct{}
}

// PoolStats 池统计信息
type PoolStats struct {
	mu              sync.RWMutex
	TotalSubmitted  int64
	TotalCompleted  int64
	TotalFailed     int64
	TotalCancelled  int64
	ActiveWorkers   int
	QueuedTasks     int
	AverageExecTime time.Duration
}

// workerPool worker 对象池（用于动态调整时复用）
var workerPool = sync.Pool{
	New: func() interface{} {
		return &worker{}
	},
}

// NewJobWorkerPool 创建 Job 协程池
func NewJobWorkerPool(maxWorkers, queueSize int) *JobWorkerPool {
	ctx, cancel := context.WithCancel(context.Background())

	pool := &JobWorkerPool{
		maxWorkers:    maxWorkers,
		queueSize:     queueSize,
		workerTimeout: 30 * time.Minute,
		taskQueue:     make(chan JobTask, queueSize),
		priorityQueue: NewPriorityQueue(),
		workers:       make([]*worker, 0, maxWorkers),
		workerCtx:     ctx,
		workerCancel:  cancel,
		stats: &PoolStats{
			ActiveWorkers: 0,
			QueuedTasks:   0,
		},
		running: false,
	}

	return pool
}

// Start 启动协程池
func (p *JobWorkerPool) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.running {
		return fmt.Errorf("worker pool already running")
	}

	log.Infof("starting job worker pool with %d workers", p.maxWorkers)

	// 启动工作协程（从对象池获取 worker）
	for i := 0; i < p.maxWorkers; i++ {
		w := workerPool.Get().(*worker)
		w.id = i
		w.pool = p
		w.taskChan = p.taskQueue
		w.stopChan = make(chan struct{})

		p.workers = append(p.workers, w)

		p.wg.Add(1)
		go w.run()
	}

	// 启动优先级队列调度器
	p.wg.Add(1)
	go p.priorityScheduler()

	p.running = true
	log.Infof("job worker pool started successfully")

	return nil
}

// Stop 停止协程池
func (p *JobWorkerPool) Stop() {
	p.mu.Lock()
	if !p.running {
		p.mu.Unlock()
		return
	}
	p.running = false
	p.mu.Unlock()

	log.Info("stopping job worker pool...")

	// 取消所有工作协程
	p.workerCancel()

	// 关闭任务队列
	close(p.taskQueue)

	// 等待所有工作协程完成
	p.wg.Wait()

	log.Infof("job worker pool stopped. Stats: completed=%d, failed=%d, cancelled=%d",
		p.stats.TotalCompleted, p.stats.TotalFailed, p.stats.TotalCancelled)
}

// Submit 提交任务到协程池
func (p *JobWorkerPool) Submit(task JobTask) error {
	p.mu.RLock()
	if !p.running {
		p.mu.RUnlock()
		return fmt.Errorf("worker pool not running")
	}
	p.mu.RUnlock()

	// 增加提交计数
	p.stats.mu.Lock()
	p.stats.TotalSubmitted++
	p.stats.mu.Unlock()

	// 尝试直接放入任务队列
	select {
	case p.taskQueue <- task:
		log.Infof("task %s submitted to queue", task.GetJobID())
		return nil
	default:
		// 队列已满，放入优先级队列
		p.priorityQueue.Push(task)
		log.Infof("task %s added to priority queue (queue full)", task.GetJobID())
		return nil
	}
}

// SubmitWithPriority 提交带优先级的任务
func (p *JobWorkerPool) SubmitWithPriority(task JobTask) error {
	p.mu.RLock()
	if !p.running {
		p.mu.RUnlock()
		return fmt.Errorf("worker pool not running")
	}
	p.mu.RUnlock()

	// 增加提交计数
	p.stats.mu.Lock()
	p.stats.TotalSubmitted++
	p.stats.mu.Unlock()

	// 放入优先级队列
	p.priorityQueue.Push(task)
	log.Infof("task %s added to priority queue (priority=%d)", task.GetJobID(), task.GetPriority())

	return nil
}

// CancelTask 取消指定任务
func (p *JobWorkerPool) CancelTask(jobId string) error {
	// 从优先级队列中移除
	if p.priorityQueue.Remove(jobId) {
		p.stats.mu.Lock()
		p.stats.TotalCancelled++
		p.stats.mu.Unlock()
		log.Infof("task %s cancelled from priority queue", jobId)
		return nil
	}

	// TODO: 支持取消正在执行的任务（需要通过 context）
	log.Warnf("task %s not found in queue or already executing", jobId)
	return fmt.Errorf("task %s not found or already executing", jobId)
}

// GetStats 获取池统计信息
func (p *JobWorkerPool) GetStats() PoolStats {
	p.stats.mu.RLock()
	defer p.stats.mu.RUnlock()

	// 更新实时队列信息
	p.stats.QueuedTasks = len(p.taskQueue) + p.priorityQueue.Len()

	return PoolStats{
		TotalSubmitted:  p.stats.TotalSubmitted,
		TotalCompleted:  p.stats.TotalCompleted,
		TotalFailed:     p.stats.TotalFailed,
		TotalCancelled:  p.stats.TotalCancelled,
		ActiveWorkers:   p.stats.ActiveWorkers,
		QueuedTasks:     p.stats.QueuedTasks,
		AverageExecTime: p.stats.AverageExecTime,
	}
}

// Resize 动态调整工作协程数量
func (p *JobWorkerPool) Resize(newSize int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.running {
		return fmt.Errorf("worker pool not running")
	}

	currentSize := len(p.workers)
	if newSize == currentSize {
		return nil
	}

	if newSize > currentSize {
		// 增加工作协程（从对象池获取）
		for i := currentSize; i < newSize; i++ {
			w := workerPool.Get().(*worker)
			w.id = i
			w.pool = p
			w.taskChan = p.taskQueue
			w.stopChan = make(chan struct{})

			p.workers = append(p.workers, w)

			p.wg.Add(1)
			go w.run()
		}
		log.Infof("worker pool resized from %d to %d workers", currentSize, newSize)
	} else {
		// 减少工作协程（归还对象池）
		for i := currentSize - 1; i >= newSize; i-- {
			w := p.workers[i]
			close(w.stopChan)

			// 清空 worker 并归还对象池
			w.id = 0
			w.pool = nil
			w.taskChan = nil
			w.stopChan = nil
			workerPool.Put(w)

			p.workers = p.workers[:i]
		}
		log.Infof("worker pool resized from %d to %d workers", currentSize, newSize)
	}

	p.maxWorkers = newSize
	return nil
}

// priorityScheduler 优先级队列调度器
func (p *JobWorkerPool) priorityScheduler() {
	defer p.wg.Done()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-p.workerCtx.Done():
			return
		case <-ticker.C:
			// 从优先级队列取出任务放入工作队列
			if p.priorityQueue.Len() > 0 && len(p.taskQueue) < cap(p.taskQueue) {
				task := p.priorityQueue.Pop()
				if task != nil {
					select {
					case p.taskQueue <- task:
						log.Debugf("task %s moved from priority queue to work queue", task.GetJobID())
					default:
						// 队列满了，放回优先级队列
						p.priorityQueue.Push(task)
					}
				}
			}
		}
	}
}

// worker.run 工作协程执行逻辑
func (w *worker) run() {
	defer w.pool.wg.Done()

	log.Infof("worker %d started", w.id)

	for {
		select {
		case <-w.pool.workerCtx.Done():
			log.Infof("worker %d stopped (context cancelled)", w.id)
			return
		case <-w.stopChan:
			log.Infof("worker %d stopped (stop signal)", w.id)
			return
		case task, ok := <-w.taskChan:
			if !ok {
				log.Infof("worker %d stopped (channel closed)", w.id)
				return
			}

			// 执行任务
			w.executeTask(task)
		}
	}
}

// executeTask 执行单个任务（使用 sync.Pool 优化）
func (w *worker) executeTask(task JobTask) {
	jobId := task.GetJobID()
	startTime := time.Now()

	// 从对象池获取统计更新对象
	update := statsUpdatePool.Get().(*statsUpdate)
	defer func() {
		// 重置并归还对象池
		*update = statsUpdate{}
		statsUpdatePool.Put(update)
	}()

	// 更新活跃工作协程数
	w.pool.stats.mu.Lock()
	w.pool.stats.ActiveWorkers++
	w.pool.stats.mu.Unlock()

	defer func() {
		// 恢复 panic
		if r := recover(); r != nil {
			log.Errorf("worker %d panic while executing task %s: %v", w.id, jobId, r)
			update.failed = true
			update.activeDec = 1
			w.applyStatsUpdate(update, time.Since(startTime))
		}
	}()

	log.Infof("worker %d executing task %s (priority=%d)", w.id, jobId, task.GetPriority())

	// 创建带超时的上下文（复用基础 context）
	baseCtx := contextPool.Get().(context.Context)
	ctx, cancel := context.WithTimeout(baseCtx, w.pool.workerTimeout)
	defer func() {
		cancel()
		contextPool.Put(context.Background())
	}()

	// 执行任务
	err := task.Execute(ctx)
	duration := time.Since(startTime)

	// 准备统计更新
	update.activeDec = 1
	update.duration = duration

	if err != nil {
		// 任务失败
		update.failed = true
		log.Errorf("worker %d task %s failed after %v: %v", w.id, jobId, duration, err)
	} else {
		// 任务成功
		update.completed = true
		log.Infof("worker %d task %s completed in %v", w.id, jobId, duration)
	}

	// 应用统计更新
	w.applyStatsUpdate(update, duration)
}

// applyStatsUpdate 应用统计更新（集中处理，减少锁竞争）
func (w *worker) applyStatsUpdate(update *statsUpdate, duration time.Duration) {
	w.pool.stats.mu.Lock()
	defer w.pool.stats.mu.Unlock()

	if update.completed {
		w.pool.stats.TotalCompleted++
	}
	if update.failed {
		w.pool.stats.TotalFailed++
	}
	if update.cancelled {
		w.pool.stats.TotalCancelled++
	}

	w.pool.stats.ActiveWorkers -= update.activeDec

	// 更新平均执行时间
	totalTasks := w.pool.stats.TotalCompleted + w.pool.stats.TotalFailed
	if totalTasks > 0 {
		avgTime := w.pool.stats.AverageExecTime
		w.pool.stats.AverageExecTime = (avgTime*time.Duration(totalTasks-1) + duration) / time.Duration(totalTasks)
	}
}

// PriorityQueue 优先级队列（最小堆，使用 sync.Pool 优化）
type PriorityQueue struct {
	mu    sync.RWMutex
	tasks []JobTask
	index map[string]int // jobId -> index in tasks
}

// taskSlicePool 任务切片对象池
var taskSlicePool = sync.Pool{
	New: func() interface{} {
		slice := make([]JobTask, 0, 100) // 预分配容量
		return &slice
	},
}

// indexMapPool 索引映射对象池
var indexMapPool = sync.Pool{
	New: func() interface{} {
		m := make(map[string]int, 100) // 预分配容量
		return &m
	},
}

// NewPriorityQueue 创建优先级队列（使用对象池）
func NewPriorityQueue() *PriorityQueue {
	return &PriorityQueue{
		tasks: make([]JobTask, 0, 100),
		index: make(map[string]int, 100),
	}
}

// Push 添加任务
func (pq *PriorityQueue) Push(task JobTask) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	pq.tasks = append(pq.tasks, task)
	pq.index[task.GetJobID()] = len(pq.tasks) - 1
	pq.up(len(pq.tasks) - 1)
}

// Pop 取出优先级最高的任务
func (pq *PriorityQueue) Pop() JobTask {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	if len(pq.tasks) == 0 {
		return nil
	}

	task := pq.tasks[0]
	lastIdx := len(pq.tasks) - 1

	pq.tasks[0] = pq.tasks[lastIdx]
	pq.index[pq.tasks[0].GetJobID()] = 0
	pq.tasks = pq.tasks[:lastIdx]
	delete(pq.index, task.GetJobID())

	if len(pq.tasks) > 0 {
		pq.down(0)
	}

	return task
}

// Remove 移除指定任务
func (pq *PriorityQueue) Remove(jobId string) bool {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	idx, exists := pq.index[jobId]
	if !exists {
		return false
	}

	lastIdx := len(pq.tasks) - 1
	if idx != lastIdx {
		pq.tasks[idx] = pq.tasks[lastIdx]
		pq.index[pq.tasks[idx].GetJobID()] = idx
	}

	pq.tasks = pq.tasks[:lastIdx]
	delete(pq.index, jobId)

	if idx < len(pq.tasks) {
		pq.up(idx)
		pq.down(idx)
	}

	return true
}

// Len 返回队列长度
func (pq *PriorityQueue) Len() int {
	pq.mu.RLock()
	defer pq.mu.RUnlock()
	return len(pq.tasks)
}

// up 向上调整堆
func (pq *PriorityQueue) up(i int) {
	for {
		parent := (i - 1) / 2
		if parent == i || pq.tasks[parent].GetPriority() <= pq.tasks[i].GetPriority() {
			break
		}
		pq.swap(parent, i)
		i = parent
	}
}

// down 向下调整堆
func (pq *PriorityQueue) down(i int) {
	for {
		left := 2*i + 1
		if left >= len(pq.tasks) {
			break
		}

		j := left
		if right := left + 1; right < len(pq.tasks) && pq.tasks[right].GetPriority() < pq.tasks[left].GetPriority() {
			j = right
		}

		if pq.tasks[i].GetPriority() <= pq.tasks[j].GetPriority() {
			break
		}

		pq.swap(i, j)
		i = j
	}
}

// swap 交换两个元素
func (pq *PriorityQueue) swap(i, j int) {
	pq.tasks[i], pq.tasks[j] = pq.tasks[j], pq.tasks[i]
	pq.index[pq.tasks[i].GetJobID()] = i
	pq.index[pq.tasks[j].GetJobID()] = j
}
