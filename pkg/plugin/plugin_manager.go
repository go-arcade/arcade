// Copyright 2025 Arcade Team
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	pluginv1 "github.com/go-arcade/arcade/api/plugin/v1"
	"github.com/go-arcade/arcade/pkg/log"
	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

// Manager is the plugin manager
// Responsible for managing the lifecycle of all RPC plugins, including registration, unregistration, health checks, etc.
type Manager struct {
	// Read-write lock to protect concurrent access
	mu sync.RWMutex
	// Plugin client mapping
	plugins map[string]*Client
	// go-plugin client mapping
	clients map[string]*plugin.Client
	// Manager configuration
	config *ManagerConfig
	// Plugin handler mapping
	handlers map[string]plugin.Plugin
	// Database adapter
	db DB
	// Plugin directory watcher
	watcher *Watcher
}

// ManagerConfig is the plugin manager configuration
type ManagerConfig struct {
	// Plugin directory path
	PluginDir string
	// Handshake configuration (for validating plugins)
	HandshakeConfig plugin.HandshakeConfig
	// Plugin configuration information
	PluginConfig map[string]any
	// Timeout duration
	Timeout time.Duration
	// Maximum retry count
	MaxRetries int
}

// logWriter is an io.Writer that captures logs and forwards them to LogCapture
type logWriter struct {
	logCapture *LogCapture
	stream     string
	buf        bytes.Buffer
	mu         sync.Mutex
}

// Write implements io.Writer interface
func (w *logWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	n, err = w.buf.Write(p)
	if err != nil {
		return n, err
	}

	// Process complete lines
	for {
		line, err := w.buf.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return n, err
		}
		// Remove trailing newline
		line = strings.TrimSuffix(line, "\n")
		if line != "" {
			w.logCapture.ProcessLine(line, w.stream)
		}
	}

	return n, nil
}

// NewManager creates a new plugin manager
func NewManager(config *ManagerConfig) *Manager {
	return &Manager{
		plugins:  make(map[string]*Client),
		clients:  make(map[string]*plugin.Client),
		config:   config,
		handlers: make(map[string]plugin.Plugin),
		db:       nil, // Will be set via SetDB later
	}
}

// SetDB sets the database adapter
func (m *Manager) SetDB(db DB) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.db = db
}

// RegisterPlugin registers a plugin
func (m *Manager) RegisterPlugin(name string, pluginPath string, config *RuntimePluginConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.registerPluginLocked(name, pluginPath, config)
}

// registerPluginLocked registers a plugin (internal, assumes lock is already held)
func (m *Manager) registerPluginLocked(name string, pluginPath string, config *RuntimePluginConfig) error {
	// Check if plugin already exists
	if _, exists := m.plugins[name]; exists {
		return fmt.Errorf("plugin %s already registered", name)
	}

	// Create log capture
	taskID := config.TaskID
	if taskID == "" {
		taskID = "unknown"
	}
	logCapture := NewLogCapture(name, taskID)

	// add log handlers
	for _, handler := range config.LogHandlers {
		logCapture.AddHandler(handler)
	}

	// Create custom writers for stdout/stderr capture
	stdoutWriter := &logWriter{
		logCapture: logCapture,
		stream:     "stdout",
	}
	stderrWriter := &logWriter{
		logCapture: logCapture,
		stream:     "stderr",
	}

	// Create go-plugin client with stdout/stderr capture
	cmd := exec.Command(pluginPath)
	client := plugin.NewClient(&plugin.ClientConfig{
		Cmd:             cmd,
		HandshakeConfig: m.config.HandshakeConfig,
		Plugins:         m.getPluginMap(),
		AllowedProtocols: []plugin.Protocol{
			plugin.ProtocolGRPC,
		},
		SyncStdout: stdoutWriter,
		SyncStderr: stderrWriter,
		Logger:     NewLogAdapter(&log.Logger{Log: log.GetLogger()}),
	})

	// Connect to plugin
	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return fmt.Errorf("connect to plugin %s: %w", name, err)
	}

	// Get plugin instance (gRPC connection)
	raw, err := rpcClient.Dispense("plugin")
	if err != nil {
		client.Kill()
		return fmt.Errorf("dispense plugin %s: %w", name, err)
	}

	// Get gRPC connection from wrapper
	var grpcConn *grpc.ClientConn
	if wrapper, ok := raw.(*GRPCPluginClientWrapper); ok {
		grpcConn = wrapper.GetConn()
		log.Debugw("got gRPC connection from wrapper for plugin", "plugin", name)
	} else {
		log.Errorw("dispensed plugin is not a GRPCPluginClientWrapper", "plugin", name, "type", fmt.Sprintf("%T", raw))
		client.Kill()
		return fmt.Errorf("invalid plugin type for %s: expected GRPCPluginClientWrapper, got %T", name, raw)
	}

	// Create gRPC service client
	grpcServiceClient := pluginv1.NewPluginServiceClient(grpcConn)

	// Get plugin info from the plugin itself
	ctx := context.Background()
	infoResp, err := grpcServiceClient.GetInfo(ctx, &pluginv1.GetInfoRequest{})
	var pluginInfo *PluginInfo
	if err != nil {
		log.Warnw("failed to get plugin info, using config values", "plugin", name, "error", err)
		// Fallback to config values
		pluginInfo = &PluginInfo{
			Name:    config.Name,
			Type:    config.Type,
			Version: config.Version,
		}
	} else {
		pluginInfo = infoResp.Info
		if pluginInfo == nil {
			pluginInfo = &PluginInfo{
				Name:    config.Name,
				Type:    config.Type,
				Version: config.Version,
			}
		}
		log.Infow("retrieved plugin info", "plugin", name, "version", pluginInfo.Version)
	}

	// Create gRPC plugin client
	grpcPluginClient := &Client{
		info:          pluginInfo, // Use info from plugin
		config:        config,
		pluginPath:    pluginPath,
		conn:          grpcConn,
		client:        grpcServiceClient,
		pluginClient:  client,
		connected:     grpcConn != nil,
		lastHeartbeat: time.Now().Unix(),
	}

	// Initialize plugin
	if err := m.initializePlugin(grpcPluginClient); err != nil {
		client.Kill()
		return fmt.Errorf("initialize plugin %s: %w", name, err)
	}

	// Register plugin
	m.plugins[name] = grpcPluginClient
	m.clients[name] = client

	log.Infow("plugin registered successfully", "plugin", name)
	return nil
}

// UnregisterPlugin unregisters a plugin
func (m *Manager) UnregisterPlugin(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if plugin exists
	pluginClient, exists := m.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	// Mark as disconnected to prevent new operations
	pluginClient.connected = false

	// Call plugin's cleanup method with timeout
	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cleanupCancel()

	if err := m.cleanupPluginWithContext(cleanupCtx, pluginClient); err != nil {
		// Ignore context deadline errors as they're expected during shutdown
		if err != context.DeadlineExceeded {

			log.Debugw("cleanup plugin failed", "plugin", name, "error", err)
		}
	}

	// Wait a bit for cleanup to complete
	time.Sleep(100 * time.Millisecond)

	// Close gRPC connection (this signals the plugin to shutdown)
	if pluginClient.conn != nil {
		// Close connection and ignore "connection is closing" errors as they're normal during shutdown
		err := pluginClient.conn.Close()
		if err != nil && !strings.Contains(err.Error(), "closing") && !strings.Contains(err.Error(), "Canceled") {
			log.Debugw("close connection for plugin", "plugin", name, "error", err)
		}
	}

	// Wait for plugin to gracefully shutdown
	time.Sleep(300 * time.Millisecond)

	// Kill the plugin process if it's still running
	if pluginClient.pluginClient != nil {
		// Check if plugin has already exited
		if exited := pluginClient.pluginClient.Exited(); !exited {
			pluginClient.pluginClient.Kill()
			// Wait for process to terminate
			time.Sleep(200 * time.Millisecond)
		}
	}

	// Delete plugin
	delete(m.plugins, name)
	delete(m.clients, name)

	log.Infow("plugin unregistered", "plugin", name)
	return nil
}

// cleanupPluginWithContext cleans up a plugin with context
func (m *Manager) cleanupPluginWithContext(ctx context.Context, pluginClient *Client) error {
	if pluginClient.client == nil {
		return nil
	}

	// Call plugin's cleanup method with context
	resp, err := pluginClient.client.Cleanup(ctx, &pluginv1.CleanupRequest{})
	if err != nil {
		return fmt.Errorf("plugin cleanup failed: %w", err)
	}

	log.Infow("plugin cleanup", "plugin", pluginClient.info.GetName(), "message", resp.Message)
	return nil
}

// ReloadPlugin safely reloads a plugin by name
func (m *Manager) ReloadPlugin(name string) error {
	m.mu.Lock()
	oldClient, exists := m.plugins[name]
	if !exists {
		m.mu.Unlock()
		return fmt.Errorf("plugin %s not found", name)
	}

	// Snapshot plugin metadata
	pluginPath := oldClient.pluginPath
	pluginConfig := oldClient.config
	m.mu.Unlock()

	log.Infow("[plugin] reloading", "plugin", name, "path", pluginPath)

	// graceful cleanup
	if err := m.cleanupPlugin(oldClient); err != nil {
		log.Warnw("[plugin] cleanup failed", "plugin", name, "error", err)
	}

	// stop old client
	if oldClient.conn != nil {
		err := oldClient.conn.Close()
		if err != nil {
			return err
		}
	}
	if oldClient.pluginClient != nil {
		oldClient.pluginClient.Kill()
	}

	// wait for graceful exit
	waitStart := time.Now()
	for {
		proc := oldClient.pluginClient.Exited()
		if proc {
			break
		}
		if time.Since(waitStart) > 5*time.Second {
			log.Warnw("[plugin] wait for old plugin to exit timeout, continuing reload", "plugin", name)
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	// unregister old instance
	m.mu.Lock()
	delete(m.plugins, name)
	delete(m.clients, name)
	m.mu.Unlock()

	// re-register (with retry)
	var lastErr error
	for i := range 3 {
		err := m.RegisterPlugin(name, pluginPath, pluginConfig)
		if err == nil {
			log.Infow("[plugin] reloaded successfully", "plugin", name, "attempts", i+1)
			return nil
		}
		lastErr = err
		log.Warnw("[plugin] retry reload", "plugin", name, "attempt", i+1, "max_attempts", 3, "error", err)
		time.Sleep(time.Second)
	}

	log.Errorw("[plugin] failed to reload after 3 attempts", "plugin", name, "error", lastErr)
	return fmt.Errorf("reload plugin %s failed: %w", name, lastErr)
}

// ReloadAllPlugins reloads all registered plugins
func (m *Manager) ReloadAllPlugins() error {
	m.mu.RLock()
	pluginNames := make([]string, 0, len(m.plugins))
	for name := range m.plugins {
		pluginNames = append(pluginNames, name)
	}
	m.mu.RUnlock()

	var failedPlugins []string
	for _, name := range pluginNames {
		if err := m.ReloadPlugin(name); err != nil {
			log.Errorw("failed to reload plugin", "plugin", name, "error", err)
			failedPlugins = append(failedPlugins, name)
		}
	}

	if len(failedPlugins) > 0 {
		return fmt.Errorf("failed to reload %d plugin(s): %v", len(failedPlugins), failedPlugins)
	}

	log.Info("all plugins reloaded successfully")
	return nil
}

// GetPlugin retrieves a plugin
func (m *Manager) GetPlugin(name string) (*Client, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	pluginClient, exists := m.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", name)
	}

	return pluginClient, nil
}

// ListPlugins lists all plugins
func (m *Manager) ListPlugins() map[string]*PluginInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plugins := make(map[string]*PluginInfo)
	for name, pluginClient := range m.plugins {
		plugins[name] = pluginClient.info
	}

	return plugins
}

// FindPluginByPath finds plugin name by file path
func (m *Manager) FindPluginByPath(path string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Normalize path for comparison
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}

	for name, pluginClient := range m.plugins {
		clientAbsPath, err := filepath.Abs(pluginClient.pluginPath)
		if err != nil {
			clientAbsPath = pluginClient.pluginPath
		}
		if clientAbsPath == absPath {
			return name
		}
	}

	return ""
}

// GetPluginMetrics retrieves plugin metrics
func (m *Manager) GetPluginMetrics(name string) (*PluginMetrics, error) {
	pluginClient, err := m.GetPlugin(name)
	if err != nil {
		return nil, err
	}

	// Check connection status
	if !pluginClient.connected {
		return nil, fmt.Errorf("plugin %s is not connected", name)
	}

	// Call plugin's GetMetrics method
	return pluginClient.GetMetrics()
}

// GetAllPluginMetrics retrieves metrics for all plugins
func (m *Manager) GetAllPluginMetrics() map[string]*PluginMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metrics := make(map[string]*PluginMetrics)
	for name, pluginClient := range m.plugins {
		if pluginClient.connected {
			pluginMetrics, err := pluginClient.GetMetrics()
			if err != nil {
				log.Warnw("get metrics for plugin failed", "plugin", name, "error", err)
				metrics[name] = &PluginMetrics{
					Name:      name,
					Status:    "error",
					LastError: err.Error(),
				}
			} else {
				metrics[name] = pluginMetrics
			}
		}
	}

	return metrics
}

// HealthCheck performs health checks
func (m *Manager) HealthCheck() map[string]bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	health := make(map[string]bool)
	for name, pluginClient := range m.plugins {
		// Check connection status
		if !pluginClient.connected {
			health[name] = false
			continue
		}

		// Check if gRPC client is available
		if pluginClient.client == nil {
			log.Warnw("plugin has no gRPC client", "plugin", name)
			health[name] = false
			continue
		}

		// Send heartbeat
		ctx := context.Background()
		_, err := pluginClient.client.HealthCheck(ctx, &pluginv1.HealthCheckRequest{
			PluginId:  name,
			Message:   "ping",
			Timestamp: time.Now().Unix(),
		})
		if err != nil {
			pluginClient.errorCount++
			health[name] = false
			log.Warnw("health check failed for plugin", "plugin", name, "error", err)
		} else {
			pluginClient.errorCount = 0
			pluginClient.lastHeartbeat = time.Now().Unix()
			health[name] = true
		}
	}

	return health
}

// StartHeartbeat starts heartbeat checking
func (m *Manager) StartHeartbeat(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			m.performHealthCheck()
		}
	}()
}

// performHealthCheck performs health check
func (m *Manager) performHealthCheck() {
	health := m.HealthCheck()

	for name, isHealthy := range health {
		if !isHealthy {
			log.Warnw("plugin is unhealthy", "plugin", name)
			// Auto-restart logic can be implemented here
		}
	}
}

// initializePlugin initializes a plugin
func (m *Manager) initializePlugin(pluginClient *Client) error {
	if pluginClient.client == nil {
		log.Warnw("plugin has no gRPC client, skipping initialization", "plugin", pluginClient.info.GetName())
		return nil
	}

	// Call plugin's initialization method
	ctx := context.Background()
	protoConfig := pluginClient.config.ToProto()
	resp, err := pluginClient.client.Init(ctx, &pluginv1.InitRequest{Config: protoConfig.Config})
	if err != nil {
		return fmt.Errorf("plugin init failed: %w", err)
	}

	log.Infow("plugin initialized", "plugin", pluginClient.info.GetName(), "message", resp.Message)
	return nil
}

// cleanupPlugin cleans up a plugin
func (m *Manager) cleanupPlugin(pluginClient *Client) error {
	if pluginClient.client == nil {
		log.Warnw("plugin has no gRPC client, skipping cleanup", "plugin", pluginClient.info.GetName())
		return nil
	}

	// Call plugin's cleanup method
	ctx := context.Background()
	resp, err := pluginClient.client.Cleanup(ctx, &pluginv1.CleanupRequest{})
	if err != nil {
		return fmt.Errorf("plugin cleanup failed: %w", err)
	}

	log.Infow("plugin cleanup", "plugin", pluginClient.info.GetName(), "message", resp.Message)
	return nil
}

// getPluginMap retrieves the plugin mapping
func (m *Manager) getPluginMap() map[string]plugin.Plugin {
	if len(m.handlers) == 0 {
		// Register default plugin handler with database access
		m.handlers["plugin"] = &GRPCPlugin{
			DB: m.db,
		}
	}
	return m.handlers
}

// RegisterPluginHandler registers a plugin handler
func (m *Manager) RegisterPluginHandler(name string, handler plugin.Plugin) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers[name] = handler
}

// LoadPluginsFromDir loads all plugins from the configured plugin directory
func (m *Manager) LoadPluginsFromDir() error {
	if m.config.PluginDir == "" {
		log.Warn("plugin directory not configured, skipping auto-load")
		return nil
	}

	// Check if directory exists
	if _, err := os.Stat(m.config.PluginDir); os.IsNotExist(err) {
		log.Warnw("plugin directory does not exist", "directory", m.config.PluginDir)
		return nil
	}

	// Read directory
	entries, err := os.ReadDir(m.config.PluginDir)
	if err != nil {
		return fmt.Errorf("failed to read plugin directory: %w", err)
	}

	loadedCount := 0
	for _, entry := range entries {
		// Skip directories and hidden files
		if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		// Skip documentation files
		name := strings.ToLower(entry.Name())
		if strings.HasSuffix(name, ".json") ||
			strings.HasSuffix(name, ".md") ||
			strings.HasSuffix(name, ".txt") ||
			strings.HasSuffix(name, ".zip") {
			continue
		}

		// Extract plugin name (format: pluginName_version or just pluginName)
		pluginName := m.extractPluginName(entry.Name())
		pluginPath := filepath.Join(m.config.PluginDir, entry.Name())

		// Create default config (Type and Version will be retrieved from plugin)
		config := &RuntimePluginConfig{
			Name:   pluginName,
			Config: json.RawMessage("{}"),
			// Type and Version will be auto-detected from plugin
		}

		// Try to register the plugin
		if err := m.RegisterPlugin(pluginName, pluginPath, config); err != nil {
			log.Warnw("failed to load plugin", "plugin", pluginName, "path", pluginPath, "error", err)
			continue
		}

		loadedCount++
		log.Infow("auto-loaded plugin", "plugin", pluginName, "path", pluginPath)
	}

	log.Infow("auto-loaded plugins", "count", loadedCount, "directory", m.config.PluginDir)
	return nil
}

// extractPluginName extracts plugin name from filename
func (m *Manager) extractPluginName(filename string) string {
	// Remove extension if any
	name := filename
	if ext := filepath.Ext(filename); ext != "" {
		name = strings.TrimSuffix(name, ext)
	}

	parts := strings.Split(name, "_")
	if len(parts) > 0 {
		return parts[0]
	}
	return name
}

// Close closes the manager
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Stop watcher if it exists
	if m.watcher != nil {
		m.watcher.Stop()
		m.watcher = nil
	}

	// Cleanup all plugins
	for name, pluginClient := range m.plugins {
		if err := m.cleanupPlugin(pluginClient); err != nil {
			log.Warnw("cleanup plugin failed", "plugin", name, "error", err)
		}
	}

	// Close all clients
	for name, client := range m.clients {
		client.Kill()
		log.Infow("plugin client closed", "plugin", name)
	}

	// Clear mappings
	m.plugins = make(map[string]*Client)
	m.clients = make(map[string]*plugin.Client)

	log.Info("plugin manager closed")
	return nil
}

// Implementation of various plugin interface methods
// Specific plugin invocation logic can be implemented here as needed
