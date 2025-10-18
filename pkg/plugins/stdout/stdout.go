package main

import (
	"encoding/json"
	"fmt"
	"net/rpc"
	"os"
	"text/template"
	"time"

	"github.com/bytedance/sonic"
	"github.com/hashicorp/go-plugin"
	pluginpkg "github.com/observabil/arcade/pkg/plugin"
)

// StdoutNotifyConfig is the plugin's configuration structure (can be passed from host via Init)
type StdoutNotifyConfig struct {
	// Prefix to display when printing
	Prefix string `json:"prefix"`
	// Whether to output as JSON (message will be json.Marshal'd in Send)
	JSON bool `json:"json"`
}

// StdoutNotify implements the notify plugin
type StdoutNotify struct {
	name        string
	description string
	version     string
	cfg         StdoutNotifyConfig
}

// NewStdoutNotify creates a new stdout notify plugin instance
func NewStdoutNotify() *StdoutNotify {
	return &StdoutNotify{
		name:        "stdout",
		description: "A simple notify plugin that prints messages to stdout",
		version:     "1.0.0",
	}
}

// ===== Implement RPC Interface =====

// Name returns the plugin name
func (p *StdoutNotify) Name() (string, error) {
	return p.name, nil
}

// Description returns the plugin description
func (p *StdoutNotify) Description() (string, error) {
	return p.description, nil
}

// Version returns the plugin version
func (p *StdoutNotify) Version() (string, error) {
	return p.version, nil
}

// Type returns the plugin type
func (p *StdoutNotify) Type() (string, error) {
	return string(pluginpkg.TypeCustom), nil
}

// Init initializes the plugin
func (p *StdoutNotify) Init(config json.RawMessage) error {
	if len(config) > 0 {
		if err := sonic.Unmarshal(config, &p.cfg); err != nil {
			// Use default config if unmarshal fails
			p.cfg = StdoutNotifyConfig{}
		}
	}
	return nil
}

// Cleanup cleans up the plugin
func (p *StdoutNotify) Cleanup() error {
	return nil
}

// Execute executes custom actions (send or template)
func (p *StdoutNotify) Execute(action string, params json.RawMessage, opts json.RawMessage) (json.RawMessage, error) {
	switch action {
	case "send":
		var sendParams struct {
			Message json.RawMessage `json:"message"`
		}
		if err := sonic.Unmarshal(params, &sendParams); err != nil {
			return nil, fmt.Errorf("failed to parse send params: %w", err)
		}
		if err := p.Send(sendParams.Message, opts); err != nil {
			return nil, err
		}
		return sonic.Marshal(map[string]string{"status": "sent"})

	case "template":
		var tplParams struct {
			Template string          `json:"template"`
			Data     json.RawMessage `json:"data"`
		}
		if err := sonic.Unmarshal(params, &tplParams); err != nil {
			return nil, fmt.Errorf("failed to parse template params: %w", err)
		}
		if err := p.SendTemplate(tplParams.Template, tplParams.Data, opts); err != nil {
			return nil, err
		}
		return sonic.Marshal(map[string]string{"status": "sent"})

	default:
		return nil, fmt.Errorf("unknown action: %s, supported actions: send, template", action)
	}
}

// Send sends a notification message
func (p *StdoutNotify) Send(message json.RawMessage, opts json.RawMessage) error {
	now := time.Now().Format(time.RFC3339)

	var line string
	if p.cfg.JSON {
		// Already in JSON format
		line = string(message)
	} else {
		// Try to parse as string first
		var msgStr string
		if err := sonic.Unmarshal(message, &msgStr); err == nil {
			line = msgStr
		} else {
			// If not a string, output as compact JSON
			line = string(message)
		}
	}

	prefix := p.cfg.Prefix
	if prefix != "" {
		prefix = "[" + prefix + "] "
	}

	_, err := fmt.Fprintf(os.Stdout, "%s%s | %s\n", prefix, now, line)
	return err
}

// SendTemplate sends a notification using a template
func (p *StdoutNotify) SendTemplate(tpl string, data json.RawMessage, opts json.RawMessage) error {
	// Parse template
	t, err := template.New("stdout_notify").Parse(tpl)
	if err != nil {
		return fmt.Errorf("stdout-notify parse template: %w", err)
	}

	// Unmarshal data
	var templateData any
	if len(data) > 0 {
		if err := sonic.Unmarshal(data, &templateData); err != nil {
			return fmt.Errorf("stdout-notify unmarshal data: %w", err)
		}
	}

	prefix := p.cfg.Prefix
	if prefix != "" {
		prefix = "[" + prefix + "] "
	}
	_, _ = fmt.Fprintf(os.Stdout, "%s%s | ", prefix, time.Now().Format(time.RFC3339))

	if err := t.Execute(os.Stdout, templateData); err != nil {
		return fmt.Errorf("stdout-notify execute template: %w", err)
	}
	_, _ = fmt.Fprintln(os.Stdout) // Newline
	return nil
}

// ===== RPC Server Implementation =====

// StdoutNotifyRPCServer is the RPC server wrapper
type StdoutNotifyRPCServer struct {
	impl *StdoutNotify
}

// Name RPC method
func (s *StdoutNotifyRPCServer) Name(args string, reply *string) error {
	name, err := s.impl.Name()
	*reply = name
	return err
}

// Description RPC method
func (s *StdoutNotifyRPCServer) Description(args string, reply *string) error {
	desc, err := s.impl.Description()
	*reply = desc
	return err
}

// Version RPC method
func (s *StdoutNotifyRPCServer) Version(args string, reply *string) error {
	ver, err := s.impl.Version()
	*reply = ver
	return err
}

// Type RPC method
func (s *StdoutNotifyRPCServer) Type(args string, reply *string) error {
	typ, err := s.impl.Type()
	*reply = typ
	return err
}

// Init RPC method
func (s *StdoutNotifyRPCServer) Init(config json.RawMessage, reply *string) error {
	err := s.impl.Init(config)
	*reply = "initialized"
	return err
}

// Cleanup RPC method
func (s *StdoutNotifyRPCServer) Cleanup(args string, reply *string) error {
	err := s.impl.Cleanup()
	*reply = "cleaned up"
	return err
}

// Send RPC method
func (s *StdoutNotifyRPCServer) Send(args *pluginpkg.NotifySendArgs, reply *string) error {
	err := s.impl.Send(args.Message, args.Opts)
	*reply = "sent"
	return err
}

// SendTemplate RPC method
func (s *StdoutNotifyRPCServer) SendTemplate(args *pluginpkg.NotifyTemplateArgs, reply *string) error {
	err := s.impl.SendTemplate(args.Template, args.Data, args.Opts)
	*reply = "sent"
	return err
}

// Execute RPC method
func (s *StdoutNotifyRPCServer) Execute(args *pluginpkg.CustomExecuteArgs, reply *json.RawMessage) error {
	result, err := s.impl.Execute(args.Action, args.Params, args.Opts)
	if err != nil {
		return err
	}
	*reply = result
	return nil
}

// Ping RPC method
func (s *StdoutNotifyRPCServer) Ping(args string, reply *string) error {
	*reply = "pong"
	return nil
}

// GetInfo RPC method
func (s *StdoutNotifyRPCServer) GetInfo(args string, reply *pluginpkg.PluginInfo) error {
	name, _ := s.impl.Name()
	desc, _ := s.impl.Description()
	ver, _ := s.impl.Version()
	typ, _ := s.impl.Type()

	*reply = pluginpkg.PluginInfo{
		Name:        name,
		Description: desc,
		Version:     ver,
		Type:        typ,
		Author:      "Arcade Team",
		Homepage:    "https://github.com/observabil/arcade",
	}
	return nil
}

// GetMetrics RPC method
func (s *StdoutNotifyRPCServer) GetMetrics(args string, reply *pluginpkg.PluginMetrics) error {
	name, _ := s.impl.Name()
	ver, _ := s.impl.Version()
	typ, _ := s.impl.Type()

	*reply = pluginpkg.PluginMetrics{
		Name:    name,
		Type:    typ,
		Version: ver,
		Status:  "running",
	}
	return nil
}

// ===== Plugin Handler =====

// StdoutNotifyPlugin is the plugin handler
type StdoutNotifyPlugin struct {
	plugin.Plugin
	Impl *StdoutNotify
}

// Server returns the RPC server
func (p *StdoutNotifyPlugin) Server(*plugin.MuxBroker) (any, error) {
	return &StdoutNotifyRPCServer{impl: p.Impl}, nil
}

// Client returns the RPC client (not used in plugin side)
func (StdoutNotifyPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (any, error) {
	return nil, nil
}

// ===== Main Entry Point =====

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: pluginpkg.RPCHandshake,
		Plugins: map[string]plugin.Plugin{
			"plugin": &StdoutNotifyPlugin{Impl: NewStdoutNotify()},
		},
	})
}
