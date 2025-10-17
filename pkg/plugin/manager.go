// Package plugin provides plugin system management based on HashiCorp go-plugin
package plugin

import (
	"encoding/json"
	"fmt"
	"net/rpc"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-plugin"
	"github.com/observabil/arcade/pkg/log"
)

// Manager is the plugin manager
// Responsible for managing the lifecycle of all RPC plugins, including registration, unregistration, health checks, etc.
type Manager struct {
	// Read-write lock to protect concurrent access
	mu sync.RWMutex
	// Plugin client mapping
	plugins map[string]*RPCPluginClient
	// go-plugin client mapping
	clients map[string]*plugin.Client
	// Manager configuration
	config *ManagerConfig
	// Plugin handler mapping
	handlers map[string]plugin.Plugin
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

// NewManager creates a new plugin manager
func NewManager(config *ManagerConfig) *Manager {
	return &Manager{
		plugins:  make(map[string]*RPCPluginClient),
		clients:  make(map[string]*plugin.Client),
		config:   config,
		handlers: make(map[string]plugin.Plugin),
	}
}

// RegisterPlugin registers a plugin
func (m *Manager) RegisterPlugin(name string, pluginPath string, config PluginConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.registerPluginLocked(name, pluginPath, config)
}

// registerPluginLocked registers a plugin (internal, assumes lock is already held)
func (m *Manager) registerPluginLocked(name string, pluginPath string, config PluginConfig) error {
	// Check if plugin already exists
	if _, exists := m.plugins[name]; exists {
		return fmt.Errorf("plugin %s already registered", name)
	}

	// Create go-plugin client
	client := plugin.NewClient(&plugin.ClientConfig{
		Cmd:             exec.Command(pluginPath),
		HandshakeConfig: m.config.HandshakeConfig,
		Plugins:         m.getPluginMap(),
		AllowedProtocols: []plugin.Protocol{
			plugin.ProtocolNetRPC,
		},
		// TODO: add logger
	})

	// Connect to plugin
	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return fmt.Errorf("connect to plugin %s: %w", name, err)
	}

	// Get plugin instance
	raw, err := rpcClient.Dispense("plugin")
	if err != nil {
		client.Kill()
		return fmt.Errorf("dispense plugin %s: %w", name, err)
	}

	// Get RPC client wrapper
	var rpcClientInstance *rpc.Client
	if wrapper, ok := raw.(*RPCPluginClientWrapper); ok {
		rpcClientInstance = wrapper.GetClient()
		log.Debugf("got RPC client from wrapper for plugin %s", name)
	} else {
		log.Errorf("dispensed plugin %s is not a RPCPluginClientWrapper: %T", name, raw)
		client.Kill()
		return fmt.Errorf("invalid plugin type for %s: expected RPCPluginClientWrapper, got %T", name, raw)
	}

	// Get plugin info from the plugin itself
	var pluginInfo PluginInfo
	if err := rpcClientInstance.Call("Plugin.GetInfo", "", &pluginInfo); err != nil {
		log.Warnf("failed to get plugin info for %s, using config values: %v", name, err)
		// Fallback to config values
		pluginInfo = PluginInfo{
			Name:    config.Name,
			Type:    config.Type,
			Version: config.Version,
		}
	} else {
		log.Infof("retrieved plugin info from %s: type=%s, version=%s", name, pluginInfo.Type, pluginInfo.Version)
	}

	// Create RPC plugin client
	rpcPluginClient := &RPCPluginClient{
		info:          pluginInfo, // Use info from plugin
		config:        config,
		pluginPath:    pluginPath,
		client:        rpcClientInstance, // Set RPC client from wrapper
		pluginClient:  client,
		instance:      raw,
		connected:     rpcClientInstance != nil,
		lastHeartbeat: time.Now().Unix(),
	}

	// Initialize plugin
	if err := m.initializePlugin(rpcPluginClient); err != nil {
		client.Kill()
		return fmt.Errorf("initialize plugin %s: %w", name, err)
	}

	// Register plugin
	m.plugins[name] = rpcPluginClient
	m.clients[name] = client

	log.Infof("plugin %s registered successfully", name)
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

	// Cleanup plugin
	if err := m.cleanupPlugin(pluginClient); err != nil {
		log.Warnf("cleanup plugin %s failed: %v", name, err)
	}

	// Close connection
	if pluginClient.pluginClient != nil {
		pluginClient.pluginClient.Kill()
	}

	// Delete plugin
	delete(m.plugins, name)
	delete(m.clients, name)

	log.Infof("plugin %s unregistered", name)
	return nil
}

// ReloadPlugin reloads a plugin by name
func (m *Manager) ReloadPlugin(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get existing plugin
	pluginClient, exists := m.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	// Save plugin path and config for re-registration
	pluginPath := pluginClient.pluginPath
	pluginConfig := pluginClient.config

	// Cleanup old plugin
	if err := m.cleanupPlugin(pluginClient); err != nil {
		log.Warnf("cleanup plugin %s failed: %v", name, err)
	}

	// Close old connection
	if pluginClient.pluginClient != nil {
		pluginClient.pluginClient.Kill()
	}

	// Remove from maps
	delete(m.plugins, name)
	delete(m.clients, name)

	// Re-register plugin with saved path and config (using internal locked version)
	if err := m.registerPluginLocked(name, pluginPath, pluginConfig); err != nil {
		return fmt.Errorf("re-register plugin %s failed: %w", name, err)
	}

	log.Infof("plugin %s reloaded successfully", name)
	return nil
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
			log.Errorf("failed to reload plugin %s: %v", name, err)
			failedPlugins = append(failedPlugins, name)
		}
	}

	if len(failedPlugins) > 0 {
		return fmt.Errorf("failed to reload %d plugin(s): %v", len(failedPlugins), failedPlugins)
	}

	log.Infof("all plugins reloaded successfully")
	return nil
}

// GetPlugin retrieves a plugin
func (m *Manager) GetPlugin(name string) (*RPCPluginClient, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	pluginClient, exists := m.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", name)
	}

	return pluginClient, nil
}

// ListPlugins lists all plugins
func (m *Manager) ListPlugins() map[string]PluginInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plugins := make(map[string]PluginInfo)
	for name, pluginClient := range m.plugins {
		plugins[name] = pluginClient.info
	}

	return plugins
}

// GetPluginMetrics retrieves plugin metrics
func (m *Manager) GetPluginMetrics(name string) (PluginMetrics, error) {
	pluginClient, err := m.GetPlugin(name)
	if err != nil {
		return PluginMetrics{}, err
	}

	// Check connection status
	if !pluginClient.connected {
		return PluginMetrics{}, fmt.Errorf("plugin %s is not connected", name)
	}

	// Call plugin's GetMetrics method
	var metrics PluginMetrics
	err = pluginClient.client.Call("Plugin.GetMetrics", "", &metrics)
	if err != nil {
		return PluginMetrics{}, fmt.Errorf("get plugin metrics: %w", err)
	}

	return metrics, nil
}

// GetAllPluginMetrics retrieves metrics for all plugins
func (rm *Manager) GetAllPluginMetrics() map[string]PluginMetrics {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	metrics := make(map[string]PluginMetrics)
	for name, pluginClient := range rm.plugins {
		if pluginClient.connected {
			var pluginMetrics PluginMetrics
			err := pluginClient.client.Call("Plugin.GetMetrics", "", &pluginMetrics)
			if err != nil {
				log.Warnf("get metrics for plugin %s failed: %v", name, err)
				pluginMetrics = PluginMetrics{
					Name:      name,
					Status:    "error",
					LastError: err.Error(),
				}
			}
			metrics[name] = pluginMetrics
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

		// Check if RPC client is available
		if pluginClient.client == nil {
			log.Warnf("plugin %s has no RPC client", name)
			health[name] = false
			continue
		}

		// Send heartbeat
		var result string
		err := pluginClient.client.Call("Plugin.Ping", "", &result)
		if err != nil {
			pluginClient.errorCount++
			health[name] = false
			log.Warnf("health check failed for plugin %s: %v", name, err)
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
			log.Warnf("plugin %s is unhealthy", name)
			// Auto-restart logic can be implemented here
		}
	}
}

// initializePlugin initializes a plugin
func (m *Manager) initializePlugin(pluginClient *RPCPluginClient) error {
	if pluginClient.client == nil {
		log.Warnf("plugin %s has no RPC client, skipping initialization", pluginClient.info.Name)
		return nil
	}

	// Call plugin's initialization method
	var result string
	err := pluginClient.client.Call("Plugin.Init", pluginClient.config.Config, &result)
	if err != nil {
		return fmt.Errorf("plugin init failed: %w", err)
	}

	log.Infof("plugin %s initialized: %s", pluginClient.info.Name, result)
	return nil
}

// cleanupPlugin cleans up a plugin
func (m *Manager) cleanupPlugin(pluginClient *RPCPluginClient) error {
	if pluginClient.client == nil {
		log.Warnf("plugin %s has no RPC client, skipping cleanup", pluginClient.info.Name)
		return nil
	}

	// Call plugin's cleanup method
	var result string
	err := pluginClient.client.Call("Plugin.Cleanup", "", &result)
	if err != nil {
		return fmt.Errorf("plugin cleanup failed: %w", err)
	}

	log.Infof("plugin %s cleaned up: %s", pluginClient.info.Name, result)
	return nil
}

// getPluginMap retrieves the plugin mapping
func (m *Manager) getPluginMap() map[string]plugin.Plugin {
	if len(m.handlers) == 0 {
		// Register default plugin handler
		m.handlers["plugin"] = &RPCPluginHandler{}
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
		log.Warnf("plugin directory does not exist: %s", m.config.PluginDir)
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
		config := PluginConfig{
			Name:   pluginName,
			Config: json.RawMessage("{}"),
			// Type and Version will be auto-detected from plugin
		}

		// Try to register the plugin
		if err := m.RegisterPlugin(pluginName, pluginPath, config); err != nil {
			log.Warnf("failed to load plugin %s from %s: %v", pluginName, pluginPath, err)
			continue
		}

		loadedCount++
		log.Infof("auto-loaded plugin: %s from %s", pluginName, pluginPath)
	}

	log.Infof("auto-loaded %d plugin(s) from %s", loadedCount, m.config.PluginDir)
	return nil
}

// extractPluginName extracts plugin name from filename
func (m *Manager) extractPluginName(filename string) string {
	// Remove extension if any
	name := filename
	if ext := filepath.Ext(filename); ext != "" {
		name = strings.TrimSuffix(name, ext)
	}

	// Extract name before version (format: pluginName_version)
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

	// Cleanup all plugins
	for name, pluginClient := range m.plugins {
		if err := m.cleanupPlugin(pluginClient); err != nil {
			log.Warnf("cleanup plugin %s failed: %v", name, err)
		}
	}

	// Close all clients
	for name, client := range m.clients {
		client.Kill()
		log.Infof("plugin client %s closed", name)
	}

	// Clear mappings
	m.plugins = make(map[string]*RPCPluginClient)
	m.clients = make(map[string]*plugin.Client)

	log.Info("plugin manager closed")
	return nil
}

// Implementation of various plugin interface methods
// Specific plugin invocation logic can be implemented here as needed
