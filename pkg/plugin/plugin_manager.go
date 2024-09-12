package plugin

import (
	"errors"
	"plugin"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/12 22:30
 * @file: plugin_manager.go
 * @description: plugin manager
 */

type Manager struct {
	plugins map[string]Plugin
}

func NewManager() *Manager {
	return &Manager{
		plugins: make(map[string]Plugin),
	}
}

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
	pluginInstance, ok := p.(Plugin)
	if !ok {
		return errors.New("plugin does not implement the Plugin interface")
	}

	name := pluginInstance.Name()
	if _, exists := m.plugins[name]; exists {
		return errors.New("plugin already exists")
	}

	m.plugins[name] = pluginInstance

	return nil
}

func (m *Manager) AntiRegister(name string) error {
	if _, exists := m.plugins[name]; !exists {
		return errors.New("plugin does not exist")
	}

	delete(m.plugins, name)

	return nil
}

func (m *Manager) RunPlugin(name string) (string, error) {
	if _, exists := m.plugins[name]; !exists {
		return "", errors.New("plugin does not exist")
	}

	return m.plugins[name].Run(), nil
}

func (m *Manager) ListPlugins() []string {
	var list []string
	for k := range m.plugins {
		list = append(list, k)
	}
	return list
}
