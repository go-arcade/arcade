package plugin

import (
	"context"
	"errors"
	"fmt"
	"plugin"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/12 22:30
 * @file: plugin_manager.go
 * @description: plugin manager
 */

type Manager struct {
	ciPlugins       map[string]CIPlugin
	cdPlugins       map[string]CDPlugin
	securityPlugins map[string]SecurityPlugin
	config          *Config
}

var (
	name        string = "plugin"
	description string = "plugin manager"
	version     string = "1.0"
)

func NewManager() *Manager {
	return &Manager{
		ciPlugins:       make(map[string]CIPlugin),
		cdPlugins:       make(map[string]CDPlugin),
		securityPlugins: make(map[string]SecurityPlugin),
	}
}

func (m *Manager) Name() string {
	return name
}

func (m *Manager) Description() string {
	return description
}

func (m *Manager) Version() string {
	return version
}

// Register plugin
func (m *Manager) Register(path string) error {
	plug, err := plugin.Open(path)
	if err != nil {
		return err
	}

	symPlugin, err := plug.Lookup("NewPlugin")
	if err != nil {
		return err
	}

	// 断言函数签名
	newPlugin, ok := symPlugin.(func() interface{})
	if !ok {
		return errors.New("plugin does not implement the Plugin interface")
	}

	p := newPlugin()

	// 根据插件类型注册到对应的map中
	switch {
	case isCIPlugin(p):
		pluginInstance := p.(CIPlugin)
		if _, exists := m.ciPlugins[pluginInstance.Name()]; exists {
			return errors.New("CI plugin already exists")
		}
		m.ciPlugins[pluginInstance.Name()] = pluginInstance
	case isCDPlugin(p):
		pluginInstance := p.(CDPlugin)
		if _, exists := m.cdPlugins[pluginInstance.Name()]; exists {
			return errors.New("CD plugin already exists")
		}
		m.cdPlugins[pluginInstance.Name()] = pluginInstance
	case isSecurityPlugin(p):
		pluginInstance := p.(SecurityPlugin)
		if _, exists := m.securityPlugins[pluginInstance.Name()]; exists {
			return errors.New("Security plugin already exists")
		}
		m.securityPlugins[pluginInstance.Name()] = pluginInstance
	default:
		return errors.New("unknown plugin type")
	}

	return nil
}

// 类型检查辅助函数
func isCIPlugin(p interface{}) bool {
	_, ok := p.(CIPlugin)
	return ok
}

func isCDPlugin(p interface{}) bool {
	_, ok := p.(CDPlugin)
	return ok
}

func isSecurityPlugin(p interface{}) bool {
	_, ok := p.(SecurityPlugin)
	return ok
}

// AntiRegister anti register plugin
func (m *Manager) AntiRegister(name string) error {
	if _, exists := m.ciPlugins[name]; !exists {
		return fmt.Errorf("CI plugin %s does not exist", name)
	}
	if _, exists := m.cdPlugins[name]; !exists {
		return fmt.Errorf("CD plugin %s does not exist", name)
	}
	if _, exists := m.securityPlugins[name]; !exists {
		return fmt.Errorf("Security plugin %s does not exist", name)
	}

	delete(m.ciPlugins, name)
	delete(m.cdPlugins, name)
	delete(m.securityPlugins, name)

	return nil
}

// ListPlugins 列出所有插件
func (m *Manager) ListPlugins() map[PluginType][]string {
	result := make(map[PluginType][]string)

	// 列出CI插件
	for name := range m.ciPlugins {
		result[TypeCI] = append(result[TypeCI], name)
	}

	// 列出CD插件
	for name := range m.cdPlugins {
		result[TypeCD] = append(result[TypeCD], name)
	}

	// 列出安全插件
	for name := range m.securityPlugins {
		result[TypeSecurity] = append(result[TypeSecurity], name)
	}

	return result
}

// LoadPluginsFromConfig 根据配置加载插件
func (m *Manager) LoadPluginsFromConfig(configPath string) error {
	config, err := LoadConfig(configPath)
	if err != nil {
		return err
	}
	m.config = config

	for _, pluginConfig := range config.Plugins {
		if err := m.Register(pluginConfig.Path); err != nil {
			return err
		}
	}

	return nil
}

// Init 初始化所有插件
func (m *Manager) Init(ctx context.Context) error {
	// 初始化CI插件
	for _, p := range m.ciPlugins {
		if err := p.Init(ctx, m.config); err != nil {
			return err
		}
	}
	// 初始化CD插件
	for _, p := range m.cdPlugins {
		if err := p.Init(ctx, m.config); err != nil {
			return err
		}
	}
	// 初始化安全插件
	for _, p := range m.securityPlugins {
		if err := p.Init(ctx, m.config); err != nil {
			return err
		}
	}
	return nil
}

// Cleanup 清理所有插件
func (m *Manager) Cleanup() error {
	// 清理CI插件
	for _, p := range m.ciPlugins {
		if err := p.Cleanup(); err != nil {
			return err
		}
	}
	// 清理CD插件
	for _, p := range m.cdPlugins {
		if err := p.Cleanup(); err != nil {
			return err
		}
	}
	// 清理安全插件
	for _, p := range m.securityPlugins {
		if err := p.Cleanup(); err != nil {
			return err
		}
	}
	return nil
}

// GetCIPlugin 获取CI插件
func (m *Manager) GetCIPlugin(name string) (CIPlugin, error) {
	if plugin, exists := m.ciPlugins[name]; exists {
		return plugin, nil
	}
	return nil, errors.New("CI plugin does not exist")
}

// GetCDPlugin 获取CD插件
func (m *Manager) GetCDPlugin(name string) (CDPlugin, error) {
	if plugin, exists := m.cdPlugins[name]; exists {
		return plugin, nil
	}
	return nil, errors.New("CD plugin does not exist")
}

// GetSecurityPlugin 获取安全插件
func (m *Manager) GetSecurityPlugin(name string) (SecurityPlugin, error) {
	if plugin, exists := m.securityPlugins[name]; exists {
		return plugin, nil
	}
	return nil, errors.New("Security plugin does not exist")
}
