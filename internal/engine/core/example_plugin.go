package core

import (
	"context"
	"fmt"
)

// ExamplePlugin 示例插件
type ExamplePlugin struct {
	name    string
	version string
}

// NewExamplePlugin 创建示例插件
func NewExamplePlugin() *ExamplePlugin {
	return &ExamplePlugin{
		name:    "example",
		version: "1.0.0",
	}
}

// Name 返回插件名称
func (p *ExamplePlugin) Name() string {
	return p.name
}

// Version 返回插件版本
func (p *ExamplePlugin) Version() string {
	return p.version
}

// Init 初始化插件
func (p *ExamplePlugin) Init(ctx context.Context) error {
	fmt.Printf("Initializing plugin: %s v%s\n", p.name, p.version)
	return nil
}

// Execute 执行插件逻辑
func (p *ExamplePlugin) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	fmt.Printf("Executing plugin: %s with params: %v\n", p.name, params)
	return "Plugin executed successfully", nil
}

// Cleanup 清理插件资源
func (p *ExamplePlugin) Cleanup() error {
	fmt.Printf("Cleaning up plugin: %s\n", p.name)
	return nil
}
