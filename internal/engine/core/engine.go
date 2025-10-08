package core

import (
	"context"
	"sync"
)

// Plugin 定义插件接口
type Plugin interface {
	// Name 返回插件名称
	Name() string
	// Version 返回插件版本
	Version() string
	// Init 初始化插件
	Init(ctx context.Context) error
	// Execute 执行插件逻辑
	Execute(ctx context.Context, params map[string]interface{}) (interface{}, error)
	// Cleanup 清理插件资源
	Cleanup() error
}

// Engine 构建引擎核心
type Engine struct {
	plugins    map[string]Plugin
	mu         sync.RWMutex
	ctx        context.Context
	cancelFunc context.CancelFunc
}

// NewEngine 创建新的构建引擎实例
func NewEngine() *Engine {
	ctx, cancel := context.WithCancel(context.Background())
	return &Engine{
		plugins:    make(map[string]Plugin),
		ctx:        ctx,
		cancelFunc: cancel,
	}
}

// RegisterPlugin 注册插件
func (e *Engine) RegisterPlugin(plugin Plugin) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.plugins[plugin.Name()]; exists {
		return ErrPluginAlreadyExists
	}

	if err := plugin.Init(e.ctx); err != nil {
		return err
	}

	e.plugins[plugin.Name()] = plugin
	return nil
}

// UnregisterPlugin 注销插件
func (e *Engine) UnregisterPlugin(name string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	plugin, exists := e.plugins[name]
	if !exists {
		return ErrPluginNotFound
	}

	if err := plugin.Cleanup(); err != nil {
		return err
	}

	delete(e.plugins, name)
	return nil
}

// ExecutePlugin 执行指定插件
func (e *Engine) ExecutePlugin(name string, params map[string]interface{}) (interface{}, error) {
	e.mu.RLock()
	plugin, exists := e.plugins[name]
	e.mu.RUnlock()

	if !exists {
		return nil, ErrPluginNotFound
	}

	return plugin.Execute(e.ctx, params)
}

// GetPlugin 获取插件实例
func (e *Engine) GetPlugin(name string) (Plugin, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	plugin, exists := e.plugins[name]
	if !exists {
		return nil, ErrPluginNotFound
	}

	return plugin, nil
}

// ListPlugins 列出所有已注册的插件
func (e *Engine) ListPlugins() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	plugins := make([]string, 0, len(e.plugins))
	for name := range e.plugins {
		plugins = append(plugins, name)
	}
	return plugins
}

// Shutdown 关闭引擎
func (e *Engine) Shutdown() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, plugin := range e.plugins {
		if err := plugin.Cleanup(); err != nil {
			return err
		}
	}

	e.cancelFunc()
	return nil
}
