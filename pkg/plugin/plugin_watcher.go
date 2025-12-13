// Package plugin directory monitoring implementation
package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/go-arcade/arcade/pkg/log"
)

// PluginEvent represents a plugin file system event
type PluginEvent struct {
	// File path
	Path string
	// Event operation type
	Op fsnotify.Op
	// Event timestamp
	Timestamp time.Time
	// Plugin name (if known)
	PluginName string
}

// PluginFileState tracks the state of a plugin file
type PluginFileState struct {
	// File path
	Path string
	// Last event timestamp
	LastEventTime time.Time
	// Last event operation
	LastEventOp fsnotify.Op
	// Load timestamp (for ignoring events shortly after loading)
	LoadTime time.Time
	// Whether this file was recently loaded
	RecentlyLoaded bool
}

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
	// Pending operations mapping (file path -> event state)
	pendingOps map[string]*PluginFileState
	// File states (file path -> state)
	fileStates map[string]*PluginFileState
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
		pendingOps:   make(map[string]*PluginFileState),
		fileStates:   make(map[string]*PluginFileState),
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
	log.Infow("start watch plugin dir", "directory", absDir)
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
	err := w.watcher.Close()
	if err != nil {
		return
	}
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
			log.Errorw("file watch error", "error", err)
		}
	}
}

// handleEvent handles filesystem events
func (w *Watcher) handleEvent(event fsnotify.Event) {
	// Filter temporary files
	if w.shouldIgnore(event.Name) {
		return
	}

	log.Debugw("detected file event", "operation", event.Op.String(), "file", event.Name)

	// Ignore CHMOD events (permission changes don't require reload)
	if event.Op&fsnotify.Chmod == fsnotify.Chmod {
		log.Debugw("ignoring CHMOD event", "file", event.Name)
		return
	}

	// Get or create file state
	w.mu.Lock()
	state, exists := w.fileStates[event.Name]
	if !exists {
		state = &PluginFileState{
			Path:           event.Name,
			RecentlyLoaded: false,
		}
		w.fileStates[event.Name] = state
	}
	w.mu.Unlock()

	// Handle RENAME events immediately (file might be moved/deleted)
	if event.Op&fsnotify.Rename == fsnotify.Rename {
		// Check if file still exists at the original path
		if _, err := os.Stat(event.Name); os.IsNotExist(err) {
			// File was moved/deleted, find plugin and unload it
			pluginName := w.getPluginNameFromPath(event.Name)
			if pluginName != "" {
				log.Debugw("plugin file was moved/deleted (RENAME), unloading plugin", "file", event.Name, "plugin", pluginName)
				w.unloadPlugin(pluginName)
			}
			// Clean up state
			w.mu.Lock()
			delete(w.fileStates, event.Name)
			delete(w.pendingOps, event.Name)
			w.mu.Unlock()
			return
		}
		// File was renamed but still exists, treat as modification
	}

	// Check if this plugin was recently loaded (within 2 seconds)
	w.mu.Lock()
	if state.RecentlyLoaded {
		if time.Since(state.LoadTime) < 2*time.Second {
			log.Debugw("ignoring event for recently loaded plugin", "file", event.Name)
			w.mu.Unlock()
			return
		}
		// Remove recently loaded flag if enough time has passed
		state.RecentlyLoaded = false
	}
	// Update state
	state.LastEventTime = time.Now()
	state.LastEventOp = event.Op
	w.mu.Unlock()

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

	// Get or create file state
	state, exists := w.fileStates[event.Name]
	if !exists {
		state = &PluginFileState{
			Path:           event.Name,
			RecentlyLoaded: false,
		}
		w.fileStates[event.Name] = state
	}

	// Update state
	state.LastEventTime = time.Now()
	state.LastEventOp = event.Op

	// Add to pending operations for debouncing
	w.pendingOps[event.Name] = state
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
	now := time.Now()
	var toProcess []*PluginFileState
	for path, state := range w.pendingOps {
		// If the operation time has exceeded the debounce time
		if now.Sub(state.LastEventTime) >= w.debounceTime {
			toProcess = append(toProcess, state)
			delete(w.pendingOps, path)
		}
	}
	w.mu.Unlock()

	// Process events outside of lock
	for _, state := range toProcess {
		w.handlePluginChange(state.Path, state.LastEventOp)
	}
}

// handlePluginChange handles plugin file changes
func (w *Watcher) handlePluginChange(path string, op fsnotify.Op) {
	pluginName := w.getPluginNameFromPath(path)
	if pluginName == "" {
		log.Warnw("cannot determine plugin name from path", "path", path)
		return
	}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// File was deleted, just unload the plugin
		log.Infow("plugin file was deleted, unloading plugin", "file", path, "op", op.String(), "plugin", pluginName)
		w.unloadPlugin(pluginName)
		// Clean up state
		w.mu.Lock()
		delete(w.fileStates, path)
		w.mu.Unlock()
		return
	}

	// File exists or was modified, reload the plugin
	log.Debugw("handling plugin change", "path", path, "op", op.String(), "plugin", pluginName)
	w.reloadPlugin(pluginName, path)
}

// reloadPlugin reloads a plugin
func (w *Watcher) reloadPlugin(pluginName, path string) {
	// Try to unload the old plugin first
	w.unloadPlugin(pluginName)

	// Wait a bit for the plugin to fully shutdown before loading new one
	time.Sleep(100 * time.Millisecond)

	// Verify file still exists before loading
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Warnw("plugin file no longer exists, skipping load", "file", path)
		return
	}

	// Load new plugin
	w.loadPlugin(pluginName, path)
}

// unloadPlugin unloads a plugin
func (w *Watcher) unloadPlugin(pluginName string) {
	if err := w.manager.UnregisterPlugin(pluginName); err != nil {
		log.Debugw("unload plugin failed (maybe not loaded)", "plugin", pluginName, "error", err)
	} else {
		log.Infow("plugin unloaded", "plugin", pluginName)
	}
}

// loadPlugin loads a plugin
func (w *Watcher) loadPlugin(pluginName, pluginPath string) {
	config := &RuntimePluginConfig{
		Name:   pluginName,
		Type:   "", // Default type is empty, will be retrieved from plugin itself
		Config: json.RawMessage("{}"),
	}

	if err := w.manager.RegisterPlugin(pluginName, pluginPath, config); err != nil {
		log.Errorw("load plugin failed", "plugin", pluginName, "error", err)
		return
	}

	log.Infow("plugin loaded successfully", "plugin", pluginName, "path", pluginPath)
}

// getPluginNameFromPath extracts the plugin name from file path
func (w *Watcher) getPluginNameFromPath(path string) string {
	// First, try to find plugin by exact path match
	if pluginName := w.manager.FindPluginByPath(path); pluginName != "" {
		return pluginName
	}

	// If not found by exact path, try to match by filename
	base := filepath.Base(path)
	plugins := w.manager.ListPlugins()
	for name := range plugins {
		// Match plugin name based on filename
		if strings.Contains(base, name) || strings.Contains(name, base) {
			return name
		}
	}

	// If no registered plugin is found, use the filename as the plugin name
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
