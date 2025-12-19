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
	"fmt"
	"sync"
)

var (
	// 全局插件注册表
	reg = &registry{
		plugins: make(map[string]Plugin),
	}
)

// registry 是插件注册表
type registry struct {
	mu      sync.RWMutex
	plugins map[string]Plugin
}

// Register 注册一个插件
// 插件应该在init函数中调用此函数进行注册
// 如果同名插件已存在，会返回错误
func Register(plugin Plugin) error {
	if plugin == nil {
		return fmt.Errorf("plugin cannot be nil")
	}

	name := plugin.Name()
	if name == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}

	reg.mu.Lock()
	defer reg.mu.Unlock()

	if _, exists := reg.plugins[name]; exists {
		return fmt.Errorf("plugin %s already registered", name)
	}

	reg.plugins[name] = plugin
	return nil
}

// MustRegister 注册一个插件，如果失败会panic
// 适用于在init函数中注册插件
func MustRegister(plugin Plugin) {
	if err := Register(plugin); err != nil {
		panic(fmt.Sprintf("failed to register plugin: %v", err))
	}
}

// Unregister 取消注册一个插件
func Unregister(name string) {
	reg.mu.Lock()
	defer reg.mu.Unlock()
	delete(reg.plugins, name)
}

// Get 获取已注册的插件
func Get(name string) (Plugin, bool) {
	reg.mu.RLock()
	defer reg.mu.RUnlock()
	plugin, ok := reg.plugins[name]
	return plugin, ok
}

// List 列出所有已注册的插件名称
func List() []string {
	reg.mu.RLock()
	defer reg.mu.RUnlock()

	names := make([]string, 0, len(reg.plugins))
	for name := range reg.plugins {
		names = append(names, name)
	}
	return names
}

// ListPlugins 列出所有已注册的插件实例
func ListPlugins() map[string]Plugin {
	reg.mu.RLock()
	defer reg.mu.RUnlock()

	plugins := make(map[string]Plugin, len(reg.plugins))
	for name, plugin := range reg.plugins {
		plugins[name] = plugin
	}
	return plugins
}

// Count 返回已注册插件的数量
func Count() int {
	reg.mu.RLock()
	defer reg.mu.RUnlock()
	return len(reg.plugins)
}

// Clear 清空所有已注册的插件（主要用于测试）
func Clear() {
	reg.mu.Lock()
	defer reg.mu.Unlock()
	reg.plugins = make(map[string]Plugin)
}
