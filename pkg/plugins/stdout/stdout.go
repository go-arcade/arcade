package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/rpc"
	"os"
	"text/template"
	"time"

	"github.com/bytedance/sonic"
	pluginv1 "github.com/go-arcade/arcade/api/plugin/v1"
	pluginpkg "github.com/go-arcade/arcade/pkg/plugin"
	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
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

// ===== gRPC Plugin Handler =====

// StdoutNotifyPlugin is the gRPC plugin handler
type StdoutNotifyPlugin struct {
	plugin.Plugin
	Impl *Stdout
}

// Server returns the RPC server (required by plugin.Plugin interface, not used for gRPC)
func (p *StdoutNotifyPlugin) Server(*plugin.MuxBroker) (any, error) {
	return nil, fmt.Errorf("RPC protocol not supported, use gRPC protocol instead")
}

// Client returns the RPC client (required by plugin.Plugin interface, not used for gRPC)
func (p *StdoutNotifyPlugin) Client(*plugin.MuxBroker, *rpc.Client) (any, error) {
	return nil, fmt.Errorf("RPC protocol not supported, use gRPC protocol instead")
}

// GRPCServer returns the gRPC server
func (p *StdoutNotifyPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	name, _ := p.Impl.Name()
	desc, _ := p.Impl.Description()
	ver, _ := p.Impl.Version()
	typ, _ := p.Impl.Type()

	info := &pluginpkg.PluginInfo{
		Name:        name,
		Description: desc,
		Version:     ver,
		Type:        typ,
		Author:      "Arcade Team",
		Homepage:    "https://github.com/go-arcade/arcade",
	}

	server := pluginpkg.NewServer(info, p.Impl, nil)
	pluginv1.RegisterPluginServiceServer(s, server)
	return nil
}

// GRPCClient returns the gRPC client (not used in plugin side)
func (p *StdoutNotifyPlugin) GRPCClient(context.Context, *plugin.GRPCBroker, *grpc.ClientConn) (any, error) {
	return nil, nil
}

// ===== Main Entry Point =====

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: pluginpkg.PluginHandshake,
		Plugins: map[string]plugin.Plugin{
			"plugin": &StdoutNotifyPlugin{Impl: NewStdout()},
		},
		GRPCServer: func(opts []grpc.ServerOption) *grpc.Server {
			return grpc.NewServer(opts...)
		},
	})
}
