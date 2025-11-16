package main

import (
	"encoding/json"
	"fmt"
	"net/rpc"
	"os"
	"text/template"
	"time"

	"github.com/bytedance/sonic"
	pluginpkg "github.com/go-arcade/arcade/pkg/plugin"
	"github.com/hashicorp/go-plugin"
)

// StdoutConfig is the plugin's configuration structure (can be passed from host via Init)
type StdoutConfig struct {
	// Prefix to display when printing
	Prefix string `json:"prefix"`
	// Whether to output as JSON (message will be json.Marshal'd in Send)
	JSON bool `json:"json"`
}

// Each plugin maintains its own action and args structures

// SendArgs contains arguments for sending notifications
type SendArgs struct {
	Message json.RawMessage `json:"message"`
	Opts    json.RawMessage `json:"opts"`
}

// SendTemplateArgs contains arguments for sending template notifications
type SendTemplateArgs struct {
	Template string          `json:"template"`
	Data     json.RawMessage `json:"data"`
	Opts     json.RawMessage `json:"opts"`
}

// SendBatchArgs contains arguments for sending batch notifications
type SendBatchArgs struct {
	Messages []json.RawMessage `json:"messages"`
	Opts     json.RawMessage   `json:"opts"`
}

// Stdout implements the plugin
type Stdout struct {
	*pluginpkg.PluginBase
	name        string
	description string
	version     string
	cfg         StdoutConfig
}

// Action definitions - maintains action names and descriptions
var (
	actions = map[string]string{
		"send":          "Send a notification message",
		"send.template": "Send a notification using template",
		"send.batch":    "Send batch notifications",
	}
)

// NewStdout creates a new stdout plugin instance
func NewStdout() *Stdout {
	p := &Stdout{
		PluginBase:  pluginpkg.NewPluginBase(),
		name:        "stdout",
		description: "A simple plugin that prints messages to stdout",
		version:     "1.0.0",
	}

	// Register actions using Action Registry
	p.registerActions()
	return p
}

// registerActions registers all actions for this plugin
// Actions are maintained in the actions map above for easy management
func (p *Stdout) registerActions() {
	// Register "send" action
	if err := p.Registry().RegisterFunc("send", actions["send"], func(params json.RawMessage, opts json.RawMessage) (json.RawMessage, error) {
		var sendParams SendArgs
		if err := pluginpkg.UnmarshalParams(params, &sendParams); err != nil {
			return nil, fmt.Errorf("failed to parse send params: %w", err)
		}
		if err := p.Send(sendParams.Message, sendParams.Opts); err != nil {
			return nil, err
		}
		return sonic.Marshal(map[string]string{"status": "sent"})
	}); err != nil {
		return
	}

	// Register "send.template" action
	if err := p.Registry().RegisterFunc("send.template", actions["send.template"], func(params json.RawMessage, opts json.RawMessage) (json.RawMessage, error) {
		var tplParams SendTemplateArgs
		if err := pluginpkg.UnmarshalParams(params, &tplParams); err != nil {
			return nil, fmt.Errorf("failed to parse template params: %w", err)
		}
		if err := p.SendTemplate(tplParams.Template, tplParams.Data, tplParams.Opts); err != nil {
			return nil, err
		}
		return sonic.Marshal(map[string]string{"status": "sent"})
	}); err != nil {
		return
	}

	// Register "send.batch" action
	if err := p.Registry().RegisterFunc("send.batch", actions["send.batch"], func(params json.RawMessage, opts json.RawMessage) (json.RawMessage, error) {
		var batchParams SendBatchArgs
		if err := pluginpkg.UnmarshalParams(params, &batchParams); err != nil {
			return nil, fmt.Errorf("failed to parse batch params: %w", err)
		}
		for _, msg := range batchParams.Messages {
			if err := p.Send(msg, batchParams.Opts); err != nil {
				return nil, fmt.Errorf("failed to send batch message: %w", err)
			}
		}
		return sonic.Marshal(map[string]string{"status": "sent", "count": fmt.Sprintf("%d", len(batchParams.Messages))})
	}); err != nil {
		return
	}
}

// ===== Implement RPC Interface =====

// Name returns the plugin name
func (p *Stdout) Name() (string, error) {
	return p.name, nil
}

// Description returns the plugin description
func (p *Stdout) Description() (string, error) {
	return p.description, nil
}

// Version returns the plugin version
func (p *Stdout) Version() (string, error) {
	return p.version, nil
}

// Type returns the plugin type
func (p *Stdout) Type() (string, error) {
	return string(pluginpkg.TypeCustom), nil
}

// Init initializes the plugin
func (p *Stdout) Init(config json.RawMessage) error {
	if len(config) > 0 {
		if err := sonic.Unmarshal(config, &p.cfg); err != nil {
			// Use default config if unmarshal fails
			p.cfg = StdoutConfig{}
		}
	}
	return nil
}

// Cleanup cleans up the plugin
func (p *Stdout) Cleanup() error {
	return nil
}

// Execute executes custom actions using Action Registry
// All actions are registered in registerActions() and routed through the registry
func (p *Stdout) Execute(action string, params json.RawMessage, opts json.RawMessage) (json.RawMessage, error) {
	return p.PluginBase.Execute(action, params, opts)
}

// Send sends a notification message
func (p *Stdout) Send(message json.RawMessage, opts json.RawMessage) error {
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
func (p *Stdout) SendTemplate(tpl string, data json.RawMessage, opts json.RawMessage) error {
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

// StdoutPlugin is the RPC server wrapper
type StdoutPlugin struct {
	impl *Stdout
}

// Name RPC method
func (s *StdoutPlugin) Name(args string, reply *string) error {
	name, err := s.impl.Name()
	*reply = name
	return err
}

// Description RPC method
func (s *StdoutPlugin) Description(args string, reply *string) error {
	desc, err := s.impl.Description()
	*reply = desc
	return err
}

// Version RPC method
func (s *StdoutPlugin) Version(args string, reply *string) error {
	ver, err := s.impl.Version()
	*reply = ver
	return err
}

// Type RPC method
func (s *StdoutPlugin) Type(args string, reply *string) error {
	typ, err := s.impl.Type()
	*reply = typ
	return err
}

// Init RPC method
func (s *StdoutPlugin) Init(config json.RawMessage, reply *string) error {
	err := s.impl.Init(config)
	*reply = "initialized"
	return err
}

// Cleanup RPC method
func (s *StdoutPlugin) Cleanup(args string, reply *string) error {
	err := s.impl.Cleanup()
	*reply = "cleaned up"
	return err
}

// Send RPC method - uses method name + json.RawMessage
// params: json.RawMessage containing SendArgs
// opts: json.RawMessage containing optional overrides
func (s *StdoutPlugin) Send(args *pluginpkg.MethodArgs, reply *pluginpkg.MethodResult) error {
	var sendParams SendArgs
	if err := pluginpkg.UnmarshalParams(args.Params, &sendParams); err != nil {
		reply.Error = fmt.Sprintf("failed to parse params: %v", err)
		return nil
	}
	err := s.impl.Send(sendParams.Message, sendParams.Opts)
	if err != nil {
		reply.Error = err.Error()
		return nil
	}
	result, _ := sonic.Marshal(map[string]string{"status": "sent"})
	reply.Result = result
	return nil
}

// SendTemplate RPC method - uses method name + json.RawMessage
// params: json.RawMessage containing SendTemplateArgs
// opts: json.RawMessage containing optional overrides
func (s *StdoutPlugin) SendTemplate(args *pluginpkg.MethodArgs, reply *pluginpkg.MethodResult) error {
	var tplParams SendTemplateArgs
	if err := pluginpkg.UnmarshalParams(args.Params, &tplParams); err != nil {
		reply.Error = fmt.Sprintf("failed to parse params: %v", err)
		return nil
	}
	err := s.impl.SendTemplate(tplParams.Template, tplParams.Data, tplParams.Opts)
	if err != nil {
		reply.Error = err.Error()
		return nil
	}
	result, _ := sonic.Marshal(map[string]string{"status": "sent"})
	reply.Result = result
	return nil
}

// SendBatch RPC method - uses method name + json.RawMessage
// params: json.RawMessage containing SendBatchArgs
// opts: json.RawMessage containing optional overrides
func (s *StdoutPlugin) SendBatch(args *pluginpkg.MethodArgs, reply *pluginpkg.MethodResult) error {
	var batchParams SendBatchArgs
	if err := pluginpkg.UnmarshalParams(args.Params, &batchParams); err != nil {
		reply.Error = fmt.Sprintf("failed to parse params: %v", err)
		return nil
	}
	for _, msg := range batchParams.Messages {
		if err := s.impl.Send(msg, batchParams.Opts); err != nil {
			reply.Error = fmt.Sprintf("failed to send batch message: %v", err)
			return nil
		}
	}
	result, _ := sonic.Marshal(map[string]string{"status": "sent", "count": fmt.Sprintf("%d", len(batchParams.Messages))})
	reply.Result = result
	return nil
}

// Ping RPC method
func (s *StdoutPlugin) Ping(args string, reply *string) error {
	*reply = "pong"
	return nil
}

// GetInfo RPC method
func (s *StdoutPlugin) GetInfo(args string, reply *pluginpkg.PluginInfo) error {
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
		Homepage:    "https://github.com/go-arcade/arcade",
	}
	return nil
}

// GetMetrics RPC method
func (s *StdoutPlugin) GetMetrics(args string, reply *pluginpkg.PluginMetrics) error {
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
	Impl *Stdout
}

// Server returns the RPC server
func (p *StdoutNotifyPlugin) Server(*plugin.MuxBroker) (any, error) {
	return &StdoutPlugin{impl: p.Impl}, nil
}

// Client returns the RPC client (not used in plugin side)
func (p *StdoutNotifyPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (any, error) {
	return nil, nil
}

// ===== Main Entry Point =====

func main() {

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: pluginpkg.RPCHandshake,
		Plugins: map[string]plugin.Plugin{
			"plugin": &StdoutNotifyPlugin{Impl: NewStdout()},
		},
	})
}
