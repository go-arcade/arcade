package logger

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestBasicLogging 测试基本日志功能
func TestBasicLogging(t *testing.T) {
	cfg := Config{
		FlushInterval: 100 * time.Millisecond,
		BatchSize:     10,
		EnableFile:    true,
		LogDir:        "./testlogs",
		EnablePool:    true,
	}

	logger, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer logger.Close()

	// 写入日志
	for i := 0; i < 100; i++ {
		logger.Log("task-001", fmt.Sprintf("Log line %d", i))
	}

	time.Sleep(200 * time.Millisecond)

	stats := logger.GetStats()
	t.Logf("Stats: %+v", stats)

	if stats["total_logs"].(int64) != 100 {
		t.Errorf("Expected 100 logs, got %d", stats["total_logs"])
	}
}

// TestConcurrentLogging 测试并发日志写入
func TestConcurrentLogging(t *testing.T) {
	cfg := Config{
		FlushInterval: 100 * time.Millisecond,
		BatchSize:     50,
		EnableFile:    true,
		LogDir:        "./testlogs",
		BufferSize:    10000,
		EnablePool:    true,
	}

	logger, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer logger.Close()

	// 并发写入
	var wg sync.WaitGroup
	goroutines := 10
	logsPerGoroutine := 1000

	start := time.Now()

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			taskID := fmt.Sprintf("task-%03d", id)
			for j := 0; j < logsPerGoroutine; j++ {
				logger.Log(taskID, fmt.Sprintf("Goroutine %d log %d", id, j))
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)

	// 等待所有日志处理完成
	time.Sleep(300 * time.Millisecond)

	stats := logger.GetStats()
	t.Logf("Stats: %+v", stats)
	t.Logf("Time elapsed: %v", elapsed)
	t.Logf("Throughput: %.2f logs/sec", float64(goroutines*logsPerGoroutine)/elapsed.Seconds())

	expectedLogs := int64(goroutines * logsPerGoroutine)
	actualLogs := stats["total_logs"].(int64)

	if actualLogs != expectedLogs {
		t.Errorf("Expected %d logs, got %d", expectedLogs, actualLogs)
	}
}

// TestSubscription 测试订阅功能
func TestSubscription(t *testing.T) {
	cfg := Config{
		FlushInterval:        100 * time.Millisecond,
		BatchSize:            10,
		SubscriberBufferSize: 100,
	}

	logger, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer logger.Close()

	taskID := "task-sub-001"

	// 订阅日志
	logChan, cancel := logger.Subscribe(taskID)
	defer cancel()

	// 在另一个 goroutine 中接收日志
	var received []string
	var mu sync.Mutex
	done := make(chan bool)

	go func() {
		for entry := range logChan {
			mu.Lock()
			received = append(received, entry.Line)
			mu.Unlock()
		}
		done <- true
	}()

	// 写入日志
	expectedLogs := []string{"Log 1", "Log 2", "Log 3"}
	for _, log := range expectedLogs {
		logger.Log(taskID, log)
	}

	// 等待接收
	time.Sleep(200 * time.Millisecond)
	cancel()
	<-done

	mu.Lock()
	defer mu.Unlock()

	if len(received) != len(expectedLogs) {
		t.Errorf("Expected %d logs, got %d", len(expectedLogs), len(received))
	}

	for i, expected := range expectedLogs {
		if i >= len(received) {
			break
		}
		if received[i] != expected {
			t.Errorf("Log %d: expected %q, got %q", i, expected, received[i])
		}
	}
}

// TestMultipleTasks 测试多任务并发日志
func TestMultipleTasks(t *testing.T) {
	cfg := Config{
		FlushInterval: 50 * time.Millisecond,
		BatchSize:     10,
		EnableFile:    true,
		LogDir:        "./testlogs",
	}

	logger, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer logger.Close()

	var wg sync.WaitGroup
	tasks := 5
	logsPerTask := 20

	for i := 0; i < tasks; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			taskID := fmt.Sprintf("task-multi-%d", id)
			for j := 0; j < logsPerTask; j++ {
				logger.Log(taskID, fmt.Sprintf("Task %d log %d", id, j))
			}
		}(i)
	}

	wg.Wait()
	time.Sleep(100 * time.Millisecond)

	stats := logger.GetStats()
	expectedLogs := int64(tasks * logsPerTask)
	actualLogs := stats["total_logs"].(int64)

	if actualLogs != expectedLogs {
		t.Errorf("Expected %d logs, got %d", expectedLogs, actualLogs)
	}
}

// TestCloseTask 测试关闭任务日志文件
func TestCloseTask(t *testing.T) {
	cfg := Config{
		FlushInterval: 50 * time.Millisecond,
		BatchSize:     10,
		EnableFile:    true,
		LogDir:        "./testlogs",
	}

	logger, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer logger.Close()

	taskID := "task-close-001"

	// 写入日志
	for i := 0; i < 10; i++ {
		logger.Log(taskID, fmt.Sprintf("Log %d", i))
	}

	// 等待写入完成
	time.Sleep(100 * time.Millisecond)

	// 关闭任务文件
	if err := logger.CloseTask(taskID); err != nil {
		t.Errorf("Failed to close task: %v", err)
	}

	// 验证文件已从 map 中移除
	logger.fileMu.RLock()
	_, exists := logger.files[taskID]
	logger.fileMu.RUnlock()

	if exists {
		t.Error("Task file should be removed from map")
	}
}

// MockPersistHandler 模拟持久化处理器
type MockPersistHandler struct {
	mu      sync.Mutex
	entries []LogEntry
}

func (m *MockPersistHandler) Persist(ctx context.Context, entries []LogEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries = append(m.entries, entries...)
	return nil
}

// TestCustomPersistHandler 测试自定义持久化处理器
func TestCustomPersistHandler(t *testing.T) {
	handler := &MockPersistHandler{}

	cfg := Config{
		FlushInterval:  50 * time.Millisecond,
		BatchSize:      5,
		PersistHandler: handler,
	}

	logger, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer logger.Close()

	// 写入日志
	taskID := "task-persist"
	for i := 0; i < 10; i++ {
		logger.Log(taskID, fmt.Sprintf("Log %d", i))
	}

	// 等待持久化完成
	time.Sleep(200 * time.Millisecond)

	handler.mu.Lock()
	count := len(handler.entries)
	handler.mu.Unlock()

	if count != 10 {
		t.Errorf("Expected 10 persisted logs, got %d", count)
	}
}

// BenchmarkLogging 性能基准测试
func BenchmarkLogging(b *testing.B) {
	cfg := Config{
		FlushInterval: 1 * time.Second,
		BatchSize:     100,
		BufferSize:    10000,
		EnablePool:    true,
	}

	logger, err := New(cfg)
	if err != nil {
		b.Fatal(err)
	}
	defer logger.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		taskID := "bench-task"
		i := 0
		for pb.Next() {
			logger.Log(taskID, fmt.Sprintf("Benchmark log %d", i))
			i++
		}
	})
}

// BenchmarkLoggingWithPool 使用对象池的性能基准测试
func BenchmarkLoggingWithPool(b *testing.B) {
	cfg := Config{
		FlushInterval: 1 * time.Second,
		BatchSize:     100,
		BufferSize:    10000,
		EnablePool:    true,
	}

	logger, err := New(cfg)
	if err != nil {
		b.Fatal(err)
	}
	defer logger.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Log("bench-task", fmt.Sprintf("Benchmark log %d", i))
	}
}

// BenchmarkLoggingWithoutPool 不使用对象池的性能基准测试
func BenchmarkLoggingWithoutPool(b *testing.B) {
	cfg := Config{
		FlushInterval: 1 * time.Second,
		BatchSize:     100,
		BufferSize:    10000,
		EnablePool:    false,
	}

	logger, err := New(cfg)
	if err != nil {
		b.Fatal(err)
	}
	defer logger.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Log("bench-task", fmt.Sprintf("Benchmark log %d", i))
	}
}
