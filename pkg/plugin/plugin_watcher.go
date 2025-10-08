package plugin

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/observabil/arcade/pkg/log"
)

// Watcher 插件目录监控器
type Watcher struct {
	manager      *Manager
	watcher      *fsnotify.Watcher
	dirs         []string
	configPath   string
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	debounceTime time.Duration
	mu           sync.Mutex
	pendingOps   map[string]time.Time // 文件路径 -> 最后操作时间
}

// NewWatcher 创建新的插件监控器
func NewWatcher(manager *Manager) (*Watcher, error) {
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("create fsnotify watcher: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Watcher{
		manager:      manager,
		watcher:      fw,
		dirs:         []string{},
		ctx:          ctx,
		cancel:       cancel,
		debounceTime: 500 * time.Millisecond, // 防抖延迟
		pendingOps:   make(map[string]time.Time),
	}, nil
}

// AddWatchDir 添加要监控的插件目录
func (w *Watcher) AddWatchDir(dir string) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("get absolute path for %s: %w", dir, err)
	}

	if err := w.watcher.Add(absDir); err != nil {
		return fmt.Errorf("add watch dir %s: %w", absDir, err)
	}

	w.dirs = append(w.dirs, absDir)
	log.Infof("开始监控插件目录: %s", absDir)
	return nil
}

// WatchConfig 监控配置文件变化
func (w *Watcher) WatchConfig(configPath string) error {
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return fmt.Errorf("get absolute path for config %s: %w", configPath, err)
	}

	// 监控配置文件所在目录
	configDir := filepath.Dir(absPath)
	if err := w.watcher.Add(configDir); err != nil {
		return fmt.Errorf("add watch config dir %s: %w", configDir, err)
	}

	w.configPath = absPath
	log.Infof("开始监控配置文件: %s", absPath)
	return nil
}

// Start 启动监控
func (w *Watcher) Start() {
	w.wg.Add(1)
	go w.watchLoop()

	// 启动防抖处理协程
	w.wg.Add(1)
	go w.debounceLoop()
}

// Stop 停止监控
func (w *Watcher) Stop() {
	w.cancel()
	w.watcher.Close()
	w.wg.Wait()
	log.Info("插件监控器已停止")
}

// watchLoop 监控事件循环
func (w *Watcher) watchLoop() {
	defer w.wg.Done()

	for {
		select {
		case <-w.ctx.Done():
			return

		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			w.handleEvent(event)

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			log.Errorf("文件监控错误: %v", err)
		}
	}
}

// handleEvent 处理文件系统事件
func (w *Watcher) handleEvent(event fsnotify.Event) {
	// 过滤临时文件和非 .so 文件
	if w.shouldIgnore(event.Name) {
		return
	}

	log.Debugf("检测到文件事件: %s %s", event.Op.String(), event.Name)

	// 配置文件变化
	if w.configPath != "" && event.Name == w.configPath {
		if event.Op&fsnotify.Write == fsnotify.Write {
			w.scheduleConfigReload()
		}
		return
	}

	// 插件文件变化
	if !strings.HasSuffix(event.Name, ".so") {
		return
	}

	w.schedulePluginOperation(event)
}

// schedulePluginOperation 调度插件操作（防抖）
func (w *Watcher) schedulePluginOperation(event fsnotify.Event) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 记录操作时间，用于防抖
	w.pendingOps[event.Name] = time.Now()
}

// scheduleConfigReload 调度配置重载（防抖）
func (w *Watcher) scheduleConfigReload() {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.pendingOps["__config__"] = time.Now()
}

// debounceLoop 防抖处理循环
func (w *Watcher) debounceLoop() {
	defer w.wg.Done()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return

		case <-ticker.C:
			w.processPendingOps()
		}
	}
}

// processPendingOps 处理待处理的操作
func (w *Watcher) processPendingOps() {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := time.Now()
	for path, opTime := range w.pendingOps {
		// 如果操作时间已经超过防抖时间
		if now.Sub(opTime) >= w.debounceTime {
			if path == "__config__" {
				w.reloadConfig()
			} else {
				w.reloadPlugin(path)
			}
			delete(w.pendingOps, path)
		}
	}
}

// reloadPlugin 重新加载插件
func (w *Watcher) reloadPlugin(path string) {
	// 先尝试卸载旧插件
	pluginName := w.getPluginNameFromPath(path)
	if pluginName != "" {
		if err := w.manager.AntiRegister(pluginName); err != nil {
			log.Debugf("卸载插件 %s 失败（可能未加载）: %v", pluginName, err)
		} else {
			log.Infof("已卸载插件: %s", pluginName)
		}
	}

	// 加载新插件
	if err := w.manager.Register(path); err != nil {
		log.Errorf("加载插件 %s 失败: %v", path, err)
		return
	}

	// 初始化插件
	if err := w.manager.Init(w.ctx); err != nil {
		log.Errorf("初始化插件 %s 失败: %v", path, err)
		return
	}

	log.Infof("✓ 成功加载插件: %s", path)
}

// reloadConfig 重新加载配置文件
func (w *Watcher) reloadConfig() {
	if w.configPath == "" {
		return
	}

	log.Info("检测到配置文件变化，正在重新加载...")

	// 重新加载配置
	if err := w.manager.LoadPluginsFromConfig(w.configPath); err != nil {
		log.Errorf("重新加载配置失败: %v", err)
		return
	}

	// 初始化新插件
	if err := w.manager.Init(w.ctx); err != nil {
		log.Errorf("初始化插件失败: %v", err)
		return
	}

	log.Info("✓ 配置文件重新加载成功")
}

// getPluginNameFromPath 从路径获取插件名称
func (w *Watcher) getPluginNameFromPath(path string) string {
	// 遍历所有已注册插件，查找匹配的路径
	plugins := w.manager.ListPlugins()
	for _, p := range plugins {
		if p.Path == path {
			return p.Name
		}
	}

	// 如果找不到，使用文件名作为插件名
	base := filepath.Base(path)
	return strings.TrimSuffix(base, ".so")
}

// shouldIgnore 判断是否应该忽略该文件
func (w *Watcher) shouldIgnore(path string) bool {
	base := filepath.Base(path)

	// 忽略隐藏文件
	if strings.HasPrefix(base, ".") {
		return true
	}

	// 忽略临时文件
	if strings.HasSuffix(base, "~") ||
		strings.HasSuffix(base, ".tmp") ||
		strings.HasSuffix(base, ".swp") {
		return true
	}

	return false
}
