package logger

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// LogEntry 日志条目
type LogEntry struct {
	Timestamp int64  `bson:"timestamp" json:"timestamp"`
	TaskID    string `bson:"task_id" json:"task_id"`
	Line      string `bson:"line" json:"line"`
}

// Config 配置
type Config struct {
	// 刷新间隔
	FlushInterval time.Duration
	// 批处理大小
	BatchSize int
	// 是否持久化到文件
	EnableFile bool
	// 日志文件目录
	LogDir string
	// 单个日志文件最大大小（字节）
	MaxFileSize int64
	// 日志文件保留数量
	MaxBackups int
	// 缓冲区大小
	BufferSize int
	// 订阅者通道大小
	SubscriberBufferSize int
	// 是否启用对象池
	EnablePool bool
	// 持久化处理器
	PersistHandler PersistHandler
}

// PersistHandler 持久化处理接口
type PersistHandler interface {
	// Persist 批量持久化日志
	Persist(ctx context.Context, entries []LogEntry) error
}

// Logger 主对象
type Logger struct {
	cfg Config

	// 日志通道
	ch chan *LogEntry

	// 订阅者管理
	subs  map[string]map[chan *LogEntry]struct{} // task_id -> subscribers
	subMu sync.RWMutex

	// 文件管理
	files  map[string]*logFile // task_id -> file
	fileMu sync.RWMutex

	// 对象池
	pool *sync.Pool

	// 生命周期管理
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// 统计信息
	stats Stats
}

// logFile 任务日志文件
type logFile struct {
	taskID string
	file   *os.File
	writer *bufio.Writer
	size   int64
	mu     sync.Mutex
}

// Stats 统计信息
type Stats struct {
	TotalLogs      atomic.Int64
	DroppedLogs    atomic.Int64
	Subscribers    atomic.Int32
	ActiveTasks    atomic.Int32
	PersistErrors  atomic.Int64
	BroadcastError atomic.Int64
}

// GetStats 获取统计信息
func (l *Logger) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"total_logs":       l.stats.TotalLogs.Load(),
		"dropped_logs":     l.stats.DroppedLogs.Load(),
		"subscribers":      l.stats.Subscribers.Load(),
		"active_tasks":     l.stats.ActiveTasks.Load(),
		"persist_errors":   l.stats.PersistErrors.Load(),
		"broadcast_errors": l.stats.BroadcastError.Load(),
	}
}

var global *Logger

// Init 初始化全局 logger
func Init(cfg Config) error {
	l, err := New(cfg)
	if err != nil {
		return err
	}
	global = l
	return nil
}

// Get 返回全局实例
func Get() *Logger {
	return global
}

// New 创建实例
func New(cfg Config) (*Logger, error) {
	// 设置默认值
	if cfg.FlushInterval == 0 {
		cfg.FlushInterval = 1 * time.Second
	}
	if cfg.BatchSize == 0 {
		cfg.BatchSize = 100
	}
	if cfg.BufferSize == 0 {
		cfg.BufferSize = 4096
	}
	if cfg.SubscriberBufferSize == 0 {
		cfg.SubscriberBufferSize = 256
	}
	if cfg.MaxFileSize == 0 {
		cfg.MaxFileSize = 100 * 1024 * 1024 // 100MB
	}
	if cfg.MaxBackups == 0 {
		cfg.MaxBackups = 10
	}
	if cfg.LogDir == "" {
		cfg.LogDir = "./logs"
	}

	// 创建日志目录
	if cfg.EnableFile {
		if err := os.MkdirAll(cfg.LogDir, 0755); err != nil {
			return nil, fmt.Errorf("create log dir: %w", err)
		}
	}

	l := &Logger{
		cfg:   cfg,
		ch:    make(chan *LogEntry, cfg.BufferSize),
		subs:  make(map[string]map[chan *LogEntry]struct{}),
		files: make(map[string]*logFile),
	}
	l.ctx, l.cancel = context.WithCancel(context.Background())

	// 初始化对象池
	if cfg.EnablePool {
		l.pool = &sync.Pool{
			New: func() interface{} {
				return &LogEntry{}
			},
		}
	}

	// 启动处理协程
	l.wg.Add(1)
	go l.run()

	return l, nil
}

// Close 关闭
func (l *Logger) Close() error {
	l.cancel()

	// 等待处理完成
	l.wg.Wait()

	// 关闭所有文件
	l.fileMu.Lock()
	for _, lf := range l.files {
		lf.close()
	}
	l.fileMu.Unlock()

	// 关闭所有订阅
	l.subMu.Lock()
	for taskID := range l.subs {
		for ch := range l.subs[taskID] {
			close(ch)
		}
	}
	l.subs = make(map[string]map[chan *LogEntry]struct{})
	l.subMu.Unlock()

	return nil
}

func (l *Logger) run() {
	defer l.wg.Done()

	batch := make([]*LogEntry, 0, l.cfg.BatchSize)
	ticker := time.NewTicker(l.cfg.FlushInterval)
	defer ticker.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}

		// 持久化
		if l.cfg.PersistHandler != nil {
			entries := make([]LogEntry, len(batch))
			for i, e := range batch {
				entries[i] = *e
			}
			if err := l.cfg.PersistHandler.Persist(l.ctx, entries); err != nil {
				l.stats.PersistErrors.Add(1)
			}
		}

		// 写入文件
		if l.cfg.EnableFile {
			l.writeBatchToFile(batch)
		}

		// 释放对象到池
		if l.cfg.EnablePool {
			for _, e := range batch {
				l.pool.Put(e)
			}
		}

		batch = batch[:0]
	}

	for {
		select {
		case e := <-l.ch:
			// 广播给订阅者
			l.broadcast(e)

			// 添加到批次
			batch = append(batch, e)

			// 如果达到批次大小，立即刷新
			if len(batch) >= l.cfg.BatchSize {
				flush()
			}

		case <-ticker.C:
			flush()

		case <-l.ctx.Done():
			// 处理剩余日志
			for len(l.ch) > 0 {
				e := <-l.ch
				l.broadcast(e)
				batch = append(batch, e)
			}
			flush()
			return
		}
	}
}

// Log 输出日志
func (l *Logger) Log(taskID, line string) {
	var e *LogEntry
	if l.cfg.EnablePool {
		e = l.pool.Get().(*LogEntry)
	} else {
		e = &LogEntry{}
	}

	e.Timestamp = time.Now().UnixMilli()
	e.TaskID = taskID
	e.Line = strings.TrimRight(line, "\n")

	l.stats.TotalLogs.Add(1)

	select {
	case l.ch <- e:
	default:
		// 缓冲区满，丢弃最旧的日志
		l.stats.DroppedLogs.Add(1)
		select {
		case <-l.ch:
		default:
		}
		l.ch <- e
	}
}

// Subscribe 订阅任务日志流（用于 SSE/WebSocket）
func (l *Logger) Subscribe(taskID string) (chan *LogEntry, func()) {
	ch := make(chan *LogEntry, l.cfg.SubscriberBufferSize)

	l.subMu.Lock()
	if _, ok := l.subs[taskID]; !ok {
		l.subs[taskID] = make(map[chan *LogEntry]struct{})
		l.stats.ActiveTasks.Add(1)
	}
	l.subs[taskID][ch] = struct{}{}
	l.stats.Subscribers.Add(1)
	l.subMu.Unlock()

	var once sync.Once
	cancel := func() {
		once.Do(func() {
			l.subMu.Lock()
			delete(l.subs[taskID], ch)
			if len(l.subs[taskID]) == 0 {
				delete(l.subs, taskID)
				l.stats.ActiveTasks.Add(-1)
			}
			l.stats.Subscribers.Add(-1)
			l.subMu.Unlock()
			close(ch)
		})
	}
	return ch, cancel
}

func (l *Logger) broadcast(e *LogEntry) {
	l.subMu.RLock()
	subs, ok := l.subs[e.TaskID]
	l.subMu.RUnlock()

	if !ok {
		return
	}

	// 复制一份用于广播
	entry := &LogEntry{
		Timestamp: e.Timestamp,
		TaskID:    e.TaskID,
		Line:      e.Line,
	}

	for ch := range subs {
		select {
		case ch <- entry:
		default:
			// 订阅者消费太慢，跳过
			l.stats.BroadcastError.Add(1)
		}
	}
}

func (l *Logger) writeBatchToFile(batch []*LogEntry) {
	// 按 taskID 分组
	groups := make(map[string][]*LogEntry)
	for _, e := range batch {
		groups[e.TaskID] = append(groups[e.TaskID], e)
	}

	// 写入各自的文件
	for taskID, entries := range groups {
		lf, err := l.getOrCreateLogFile(taskID)
		if err != nil {
			l.stats.PersistErrors.Add(1)
			continue
		}

		for _, e := range entries {
			line := fmt.Sprintf("[%s] %s\n",
				time.UnixMilli(e.Timestamp).Format("2006-01-02 15:04:05.000"),
				e.Line)
			if err := lf.write(line); err != nil {
				l.stats.PersistErrors.Add(1)
			}
		}
	}
}

func (l *Logger) getOrCreateLogFile(taskID string) (*logFile, error) {
	l.fileMu.RLock()
	lf, ok := l.files[taskID]
	l.fileMu.RUnlock()

	if ok {
		return lf, nil
	}

	l.fileMu.Lock()
	defer l.fileMu.Unlock()

	// 双重检查
	if lf, ok := l.files[taskID]; ok {
		return lf, nil
	}

	// 创建新文件
	filename := filepath.Join(l.cfg.LogDir, fmt.Sprintf("%s.log", taskID))
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}

	lf = &logFile{
		taskID: taskID,
		file:   file,
		writer: bufio.NewWriterSize(file, 32*1024), // 32KB buffer
		size:   stat.Size(),
	}

	l.files[taskID] = lf
	return lf, nil
}

func (lf *logFile) write(line string) error {
	lf.mu.Lock()
	defer lf.mu.Unlock()

	n, err := lf.writer.WriteString(line)
	if err != nil {
		return err
	}

	lf.size += int64(n)
	return lf.writer.Flush()
}

func (lf *logFile) close() error {
	lf.mu.Lock()
	defer lf.mu.Unlock()

	if err := lf.writer.Flush(); err != nil {
		lf.file.Close()
		return err
	}
	return lf.file.Close()
}

// CloseTask 关闭特定任务的日志文件
func (l *Logger) CloseTask(taskID string) error {
	l.fileMu.Lock()
	defer l.fileMu.Unlock()

	lf, ok := l.files[taskID]
	if !ok {
		return nil
	}

	delete(l.files, taskID)
	return lf.close()
}

// Log 全局便捷函数
func Log(taskID, line string) {
	if global != nil {
		global.Log(taskID, line)
	}
}
