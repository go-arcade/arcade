package plugin

import (
	"bufio"
	"context"
	"io"
	"sync"
	"time"

	"github.com/go-arcade/arcade/pkg/log"
)

// LogCapture 插件日志捕获器
type LogCapture struct {
	pluginName string
	taskID     string
	mu         sync.RWMutex
	handlers   []LogHandler
	lineNumber int32
	ctx        context.Context
	cancel     context.CancelFunc
}

// LogHandler 日志处理函数
type LogHandler func(entry *LogEntry)

// LogEntry 日志条目
type LogEntry struct {
	Timestamp  int64  `json:"timestamp"`   // 毫秒时间戳
	LineNumber int32  `json:"line_number"` // 行号
	Level      string `json:"level"`       // 日志级别
	Content    string `json:"content"`     // 日志内容
	Stream     string `json:"stream"`      // stdout 或 stderr
	PluginName string `json:"plugin_name"` // 插件名称
	TaskID     string `json:"task_id"`     // 任务ID
}

// NewLogCapture 创建日志捕获器
func NewLogCapture(pluginName, taskID string) *LogCapture {
	ctx, cancel := context.WithCancel(context.Background())
	return &LogCapture{
		pluginName: pluginName,
		taskID:     taskID,
		handlers:   make([]LogHandler, 0),
		lineNumber: 0,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// AddHandler 添加日志处理器
func (lc *LogCapture) AddHandler(handler LogHandler) {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	lc.handlers = append(lc.handlers, handler)
}

// CaptureReader 捕获 Reader 的输出
func (lc *LogCapture) CaptureReader(reader io.Reader, stream string) {
	scanner := bufio.NewScanner(reader)
	// 设置更大的缓冲区以处理长日志行
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		select {
		case <-lc.ctx.Done():
			return
		default:
			line := scanner.Text()
			lc.processLine(line, stream)
		}
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		log.Errorf("error reading %s for plugin %s: %v", stream, lc.pluginName, err)
	}
}

// processLine 处理单行日志
func (lc *LogCapture) processLine(content, stream string) {
	lc.mu.Lock()
	lc.lineNumber++
	lineNum := lc.lineNumber
	handlers := make([]LogHandler, len(lc.handlers))
	copy(handlers, lc.handlers)
	lc.mu.Unlock()

	entry := &LogEntry{
		Timestamp:  time.Now().UnixMilli(),
		LineNumber: lineNum,
		Level:      lc.detectLogLevel(content),
		Content:    content,
		Stream:     stream,
		PluginName: lc.pluginName,
		TaskID:     lc.taskID,
	}

	// 调用所有处理器
	for _, handler := range handlers {
		if handler != nil {
			go handler(entry) // 异步处理，避免阻塞
		}
	}
}

// detectLogLevel 检测日志级别
func (lc *LogCapture) detectLogLevel(content string) string {
	// 简单的日志级别检测逻辑
	contentLower := content
	if len(content) > 100 {
		contentLower = content[:100]
	}

	switch {
	case containsKeyword(contentLower, []string{"[ERROR]", "ERROR:", "error:", "错误"}):
		return "error"
	case containsKeyword(contentLower, []string{"[WARN]", "WARN:", "WARNING:", "warn:", "警告"}):
		return "warn"
	case containsKeyword(contentLower, []string{"[DEBUG]", "DEBUG:", "debug:", "调试"}):
		return "debug"
	default:
		return "info"
	}
}

// containsKeyword 检查字符串是否包含关键字
func containsKeyword(s string, keywords []string) bool {
	for _, keyword := range keywords {
		if len(s) >= len(keyword) {
			for i := 0; i <= len(s)-len(keyword); i++ {
				if s[i:i+len(keyword)] == keyword {
					return true
				}
			}
		}
	}
	return false
}

// Stop 停止日志捕获
func (lc *LogCapture) Stop() {
	if lc.cancel != nil {
		lc.cancel()
	}
}

// GetLineNumber 获取当前行号
func (lc *LogCapture) GetLineNumber() int32 {
	lc.mu.RLock()
	defer lc.mu.RUnlock()
	return lc.lineNumber
}
