package plugin

import (
	"os"

	"gopkg.in/yaml.v3"
)

// PluginType 插件类型
type PluginType string

const (
	// TypeCI CI类型插件
	TypeCI PluginType = "ci"
	// TypeCD CD类型插件
	TypeCD PluginType = "cd"
	// TypeSecurity 安全类型插件
	TypeSecurity PluginType = "security"
	// TypeNotify 通知类型插件
	TypeNotify PluginType = "notify"
	// TypeStorage 存储类型插件
	TypeStorage PluginType = "storage"
	// TypeCustom 自定义类型插件
	TypeCustom PluginType = "custom"
)

// PluginConfig 插件配置结构
type PluginConfig struct {
	Path    string     `yaml:"path"`    // 插件路径
	Name    string     `yaml:"name"`    // 插件名称
	Type    PluginType `yaml:"type"`    // 插件类型
	Version string     `yaml:"version"` // 插件版本
	Config  any        `yaml:"config"`  // 插件特定配置
}

// Config 插件管理器配置
type Config struct {
	Plugins []PluginConfig `yaml:"plugins"` // 插件列表
}

// LoadConfig 从文件加载配置
func LoadConfig(configPath string) (*Config, error) {
	var c Config
	bs, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(bs, &c); err != nil {
		return nil, err
	}
	return &c, nil
}
