package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/template"
	"time"

	"github.com/observabil/arcade/pkg/plugin"
)

// StdoutNotifyConfig 插件自身的配置结构（可从宿主 Init 传入）
type StdoutNotifyConfig struct {
	// 前缀（打印时展示）
	Prefix string `yaml:"prefix" json:"prefix"`
	// 是否输出 JSON（Send 时 message 将被 json.Marshal）
	JSON bool `yaml:"json" json:"json"`
}

type StdoutNotify struct {
	name        string
	description string
	version     string
	cfg         StdoutNotifyConfig
}

func NewStdoutNotify() *StdoutNotify {
	return &StdoutNotify{
		name:        "stdout",
		description: "A simple notify plugin that prints messages to stdout",
		version:     "1.0.0",
	}
}

// ===== 实现 BasePlugin =====
func (p *StdoutNotify) Name() string        { return p.name }
func (p *StdoutNotify) Description() string { return p.description }
func (p *StdoutNotify) Version() string     { return p.version }
func (p *StdoutNotify) Type() plugin.PluginType {
	return plugin.TypeNotify
}

func (p *StdoutNotify) Init(_ context.Context, config any) error {
	// 宿主在调用 Init 时，可以把插件段配置（any）透传进来
	// 这里尽量宽松地做类型断言：map[string]any / 已经解码好的 StdoutNotifyConfig 都兼容
	switch c := config.(type) {
	case map[string]any:
		// 尝试从 map 解到 cfg
		b, _ := json.Marshal(c)
		_ = json.Unmarshal(b, &p.cfg)
	case *StdoutNotifyConfig:
		p.cfg = *c
	case StdoutNotifyConfig:
		p.cfg = c
	case *plugin.Config:
		// 如果你把整个宿主 Config 传了进来，这里可在 Plugins 里按 name 找到自己的 Config
		for _, pc := range c.Plugins {
			if pc.Name == p.name || pc.Type == plugin.TypeNotify {
				// 尝试把 pc.Config 解到自己 cfg
				b, _ := json.Marshal(pc.Config)
				_ = json.Unmarshal(b, &p.cfg)
			}
		}
	default:
		// 忽略未知类型，使用默认配置
	}
	return nil
}

func (p *StdoutNotify) Cleanup() error { return nil }

// ===== 实现 NotifyPlugin =====
func (p *StdoutNotify) Send(_ context.Context, message any, _ ...plugin.Option) error {
	now := time.Now().Format(time.RFC3339)

	var line string
	if p.cfg.JSON {
		b, err := json.Marshal(message)
		if err != nil {
			return fmt.Errorf("stdout-notify marshal message: %w", err)
		}
		line = string(b)
	} else {
		switch v := message.(type) {
		case string:
			line = v
		default:
			// 非字符串就做个紧凑 JSON
			b, _ := json.Marshal(v)
			line = string(b)
		}
	}

	prefix := p.cfg.Prefix
	if prefix != "" {
		prefix = "[" + prefix + "] "
	}

	_, err := fmt.Fprintf(os.Stdout, "%s%s%s\n", prefix, now, " | "+line)
	return err
}

func (p *StdoutNotify) SendTemplate(_ context.Context, tpl string, data any, _ ...plugin.Option) error {
	// 用 text/template 渲染到 stdout
	t, err := template.New("stdout_notify").Parse(tpl)
	if err != nil {
		return fmt.Errorf("stdout-notify parse template: %w", err)
	}

	prefix := p.cfg.Prefix
	if prefix != "" {
		prefix = "[" + prefix + "] "
	}
	_, _ = fmt.Fprintf(os.Stdout, "%s%s | ", prefix, time.Now().Format(time.RFC3339))

	if err := t.Execute(os.Stdout, data); err != nil {
		return fmt.Errorf("stdout-notify execute template: %w", err)
	}
	_, _ = fmt.Fprintln(os.Stdout) // 换行
	return nil
}

// 插件入口点
var Plugin = NewStdoutNotify()
