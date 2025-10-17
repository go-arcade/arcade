// Package plugin directory monitoring implementation
package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/observabil/arcade/pkg/log"
)

// Watcher is the plugin directory monitor
// Automatically monitors plugin directory changes, supports hot loading and unloading of plugins
type Watcher struct {
	// Plugin manager reference
	manager *Manager
	// Filesystem watcher
	watcher *fsnotify.Watcher
	// List of monitored directories
	dirs []string
	// Context
	ctx context.Context
	// Cancel function
	cancel context.CancelFunc
	// Wait group
	wg sync.WaitGroup
	// Debounce delay time
	debounceTime time.Duration
	// Mutex
	mu sync.Mutex
	// Pending operations mapping (file path -> last operation time)
	pendingOps map[string]time.Time
}

// NewWatcher creates a new plugin watcher
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
		debounceTime: 500 * time.Millisecond, // Debounce delay
		pendingOps:   make(map[string]time.Time),
	}, nil
}

// AddWatchDir adds a plugin directory to watch
func (w *Watcher) AddWatchDir(dir string) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("get absolute path for %s: %w", dir, err)
	}

	if err := w.watcher.Add(absDir); err != nil {
		return fmt.Errorf("add watch dir %s: %w", absDir, err)
	}

	w.dirs = append(w.dirs, absDir)
	log.Infof("start watch plugin dir: %s", absDir)
	return nil
}

// Start starts the monitoring
func (w *Watcher) Start() {
	w.wg.Add(1)
	go w.watchLoop()

	// Start debounce processing goroutine
	w.wg.Add(1)
	go w.debounceLoop()
}

// Stop stops the monitoring
func (w *Watcher) Stop() {
	w.cancel()
	w.watcher.Close()
	w.wg.Wait()
	log.Info("plugin watcher stopped")
}

// watchLoop monitors the event loop
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
			log.Errorf("file watch error: %v", err)
		}
	}
}

// handleEvent handles filesystem events
func (w *Watcher) handleEvent(event fsnotify.Event) {
	// Filter temporary files
	if w.shouldIgnore(event.Name) {
		return
	}

	log.Debugf("detected file event: %s %s", event.Op.String(), event.Name)

	// Handle plugin file changes
	// RPC plugins are executable binaries, check if it's likely a plugin file
	base := filepath.Base(event.Name)
	ext := filepath.Ext(base)

	// Skip common non-executable files
	if ext == ".json" || ext == ".md" || ext == ".txt" ||
		ext == ".log" || ext == ".zip" || strings.HasPrefix(base, ".") {
		return
	}

	w.schedulePluginOperation(event)
}

// schedulePluginOperation schedules plugin operations (with debounce)
func (w *Watcher) schedulePluginOperation(event fsnotify.Event) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Record operation time for debouncing
	w.pendingOps[event.Name] = time.Now()
}

// debounceLoop debounce processing loop
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

// processPendingOps processes pending operations
func (w *Watcher) processPendingOps() {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := time.Now()
	for path, opTime := range w.pendingOps {
		// If the operation time has exceeded the debounce time
		if now.Sub(opTime) >= w.debounceTime {
			w.reloadPlugin(path)
			delete(w.pendingOps, path)
		}
	}
}

// reloadPlugin reloads a plugin
func (w *Watcher) reloadPlugin(path string) {
	pluginName := w.getPluginNameFromPath(path)
	if pluginName == "" {
		log.Warnf("cannot determine plugin name from path: %s", path)
		return
	}

	// Try to unload the old plugin
	w.unloadPlugin(pluginName)

	// Load new plugin directly (RPC plugins don't have path limitations)
	w.loadPlugin(pluginName, path)
}

// unloadPlugin unloads a plugin
func (w *Watcher) unloadPlugin(pluginName string) {
	if err := w.manager.UnregisterPlugin(pluginName); err != nil {
		log.Debugf("unload plugin %s failed (maybe not loaded): %v", pluginName, err)
	} else {
		log.Infof("plugin %s unloaded", pluginName)
	}
}

// loadPlugin loads a plugin
func (w *Watcher) loadPlugin(pluginName, pluginPath string) {
	config := PluginConfig{
		Name:   pluginName,
		Type:   "", // Default type is empty, will be retrieved from plugin itself
		Config: json.RawMessage("{}"),
	}

	if err := w.manager.RegisterPlugin(pluginName, pluginPath, config); err != nil {
		log.Errorf("load plugin %s failed: %v", pluginName, err)
		return
	}

	log.Infof("plugin %s loaded successfully from %s", pluginName, pluginPath)
}

// getPluginNameFromPath extracts the plugin name from file path
func (w *Watcher) getPluginNameFromPath(path string) string {
	// Iterate through all registered plugins to find matching path
	plugins := w.manager.ListPlugins()
	for name := range plugins {
		// Match plugin name based on path
		if strings.Contains(path, name) {
			return name
		}
	}

	// If no registered plugin is found, use the filename as the plugin name
	base := filepath.Base(path)
	// For executable binaries, the filename is the plugin name (may contain version)
	// e.g., "stdout_1.0.0" -> extract plugin name before version
	parts := strings.Split(base, "_")
	if len(parts) > 0 {
		return parts[0]
	}
	return base
}

// shouldIgnore determines if the file should be ignored
func (w *Watcher) shouldIgnore(path string) bool {
	base := filepath.Base(path)

	// Ignore hidden files
	if strings.HasPrefix(base, ".") {
		return true
	}

	// Ignore temporary files
	if strings.HasSuffix(base, "~") ||
		strings.HasSuffix(base, ".tmp") ||
		strings.HasSuffix(base, ".swp") {
		return true
	}

	return false
}
