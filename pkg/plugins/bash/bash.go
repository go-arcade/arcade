package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/rpc"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	pluginv1 "github.com/go-arcade/arcade/api/plugin/v1"
	pluginpkg "github.com/go-arcade/arcade/pkg/plugin"
	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

// BashConfig is the plugin's configuration structure
type BashConfig struct {
	// Shell path (default: /bin/bash)
	Shell string `json:"shell"`
	// Working directory for script execution
	WorkDir string `json:"work_dir"`
	// Default timeout in seconds (0 means no timeout)
	Timeout int `json:"timeout"`
	// Environment variables to set
	Env map[string]string `json:"env"`
	// Whether to allow dangerous operations
	AllowDangerous bool `json:"allow_dangerous"`
}

// ========== Plugin Action Arguments ==========
// Each plugin maintains its own action and args structures

// BashScriptArgs contains arguments for executing a bash script
type BashScriptArgs struct {
	Script string            `json:"script"`
	Args   []string          `json:"args,omitempty"`
	Env    map[string]string `json:"env,omitempty"`
}

// BashCommandArgs contains arguments for executing a bash command
type BashCommandArgs struct {
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

// Bash implements the custom plugin
type Bash struct {
	*pluginpkg.PluginBase
	name        string
	description string
	version     string
	cfg         BashConfig
}

// Action definitions - maintains action names and descriptions
var (
	actions = map[string]string{
		"script":  "Execute a bash script from string",
		"command": "Execute a bash command",
	}
)

// NewBash creates a new bash plugin instance
func NewBash() *Bash {
	p := &Bash{
		PluginBase:  pluginpkg.NewPluginBase(),
		name:        "bash",
		description: "A custom plugin that executes bash scripts and commands",
		version:     "1.0.0",
		cfg: BashConfig{
			Shell:   "/bin/bash",
			WorkDir: "",
			Timeout: 300, // 5 minutes default
		},
	}

	// Register actions using Action Registry
	p.registerActions()
	return p
}

// registerActions registers all actions for this plugin
// Actions are maintained in the actions map above for easy management
func (p *Bash) registerActions() {
	// Register "script" action
	if err := p.Registry().RegisterFunc("script", actions["script"], func(params json.RawMessage, opts json.RawMessage) (json.RawMessage, error) {
		return p.executeScript(params, opts)
	}); err != nil {
		return
	}

	// Register "command" action
	if err := p.Registry().RegisterFunc("command", actions["command"], func(params json.RawMessage, opts json.RawMessage) (json.RawMessage, error) {
		return p.executeCommand(params, opts)
	}); err != nil {
		return
	}
}

// ===== Implement RPC Interface =====

// Name returns the plugin name
func (p *Bash) Name() (string, error) {
	return p.name, nil
}

// Description returns the plugin description
func (p *Bash) Description() (string, error) {
	return p.description, nil
}

// Version returns the plugin version
func (p *Bash) Version() (string, error) {
	return p.version, nil
}

// Type returns the plugin type
func (p *Bash) Type() (string, error) {
	return string(pluginpkg.TypeCustom), nil
}

// Init initializes the plugin
func (p *Bash) Init(config json.RawMessage) error {
	if len(config) > 0 {
		if err := sonic.Unmarshal(config, &p.cfg); err != nil {
			return fmt.Errorf("failed to parse config: %w", err)
		}
	}

	// Validate shell path
	if p.cfg.Shell == "" {
		p.cfg.Shell = "/bin/bash"
	}

	// Check if shell exists
	if _, err := os.Stat(p.cfg.Shell); os.IsNotExist(err) {
		return fmt.Errorf("shell not found: %s", p.cfg.Shell)
	}

	// Set default working directory if not specified
	if p.cfg.WorkDir == "" {
		p.cfg.WorkDir, _ = os.Getwd()
	}

	fmt.Printf("[bash-plugin] initialized with shell: %s, work_dir: %s\n", p.cfg.Shell, p.cfg.WorkDir)
	return nil
}

// Cleanup cleans up the plugin
func (p *Bash) Cleanup() error {
	fmt.Println("[bash-plugin] cleanup completed")
	return nil
}

// Execute executes a bash script or command using Action Registry
// All actions are registered in registerActions() and routed through the registry
func (p *Bash) Execute(action string, params json.RawMessage, opts json.RawMessage) (json.RawMessage, error) {
	return p.PluginBase.Execute(action, params, opts)
}

// executeScript executes a bash script from string
func (p *Bash) executeScript(params json.RawMessage, opts json.RawMessage) (json.RawMessage, error) {
	var scriptParams BashScriptArgs

	if err := sonic.Unmarshal(params, &scriptParams); err != nil {
		return nil, fmt.Errorf("failed to parse script params: %w", err)
	}

	if scriptParams.Script == "" {
		return nil, fmt.Errorf("script content is required")
	}

	// Security check: prevent dangerous operations if not allowed
	if !p.cfg.AllowDangerous {
		dangerous := []string{"rm -rf", ":(){ :|:& };:", "mkfs", "dd if=", "> /dev/"}
		for _, pattern := range dangerous {
			if strings.Contains(scriptParams.Script, pattern) {
				return nil, fmt.Errorf("dangerous operation detected and not allowed: %s", pattern)
			}
		}
	}

	// Create a temporary script file
	tmpFile, err := os.CreateTemp("", "bash-plugin-*.sh")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			return
		}
	}(tmpFile.Name())

	// Write script content
	if _, err := tmpFile.WriteString(scriptParams.Script); err != nil {
		return nil, fmt.Errorf("failed to write script: %w", err)
	}
	err = tmpFile.Close()
	if err != nil {
		return nil, err
	}

	// Make it executable
	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		return nil, fmt.Errorf("failed to chmod script: %w", err)
	}

	// Execute the script
	return p.runCommand(p.cfg.Shell, append([]string{tmpFile.Name()}, scriptParams.Args...), scriptParams.Env)
}

// executeCommand executes a simple bash command
func (p *Bash) executeCommand(params json.RawMessage, opts json.RawMessage) (json.RawMessage, error) {
	var cmdParams BashCommandArgs

	if err := sonic.Unmarshal(params, &cmdParams); err != nil {
		return nil, fmt.Errorf("failed to parse command params: %w", err)
	}

	if cmdParams.Command == "" {
		return nil, fmt.Errorf("command is required")
	}

	// Security check
	if !p.cfg.AllowDangerous {
		dangerous := []string{"rm -rf /", ":(){ :|:& };:", "mkfs", "dd if=/dev/zero", "> /dev/"}
		for _, pattern := range dangerous {
			if strings.Contains(cmdParams.Command, pattern) {
				return nil, fmt.Errorf("dangerous operation detected and not allowed: %s", pattern)
			}
		}
	}

	return p.runCommand(p.cfg.Shell, []string{"-c", cmdParams.Command}, cmdParams.Env)
}

// runCommand executes a command with the given arguments and environment
func (p *Bash) runCommand(command string, args []string, env map[string]string) (json.RawMessage, error) {
	ctx := context.Background()
	timeout := time.Duration(p.cfg.Timeout) * time.Second

	// Create context with timeout if configured
	if p.cfg.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	// Create command
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = p.cfg.WorkDir

	// Set environment variables
	cmd.Env = os.Environ()
	for k, v := range p.cfg.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Record start time
	startTime := time.Now()

	// Run command
	err := cmd.Run()
	duration := time.Since(startTime)

	// Prepare result
	result := map[string]any{
		"stdout":      stdout.String(),
		"stderr":      stderr.String(),
		"duration_ms": duration.Milliseconds(),
		"success":     err == nil,
	}

	if err != nil {
		result["error"] = err.Error()
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			result["exit_code"] = exitErr.ExitCode()
		}
	} else {
		result["exit_code"] = 0
	}

	return sonic.Marshal(result)
}

// ===== gRPC Plugin Handler =====

// BashPluginHandler is the gRPC plugin handler
type BashPluginHandler struct {
	plugin.Plugin
	pluginInstance *Bash
}

// Server returns the RPC server (required by plugin.Plugin interface, not used for gRPC)
func (p *BashPluginHandler) Server(*plugin.MuxBroker) (any, error) {
	return nil, fmt.Errorf("RPC protocol not supported, use gRPC protocol instead")
}

// Client returns the RPC client (required by plugin.Plugin interface, not used for gRPC)
func (p *BashPluginHandler) Client(*plugin.MuxBroker, *rpc.Client) (any, error) {
	return nil, fmt.Errorf("RPC protocol not supported, use gRPC protocol instead")
}

// GRPCServer returns the gRPC server
func (p *BashPluginHandler) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	name, _ := p.pluginInstance.Name()
	desc, _ := p.pluginInstance.Description()
	ver, _ := p.pluginInstance.Version()
	typ, _ := p.pluginInstance.Type()

	info := &pluginpkg.PluginInfo{
		Name:        name,
		Description: desc,
		Version:     ver,
		Type:        typ,
		Author:      "Arcade Team",
		Homepage:    "https://github.com/go-arcade/arcade",
	}

	server := pluginpkg.NewServer(info, p.pluginInstance, nil)
	pluginv1.RegisterPluginServiceServer(s, server)
	return nil
}

// GRPCClient returns the gRPC client (not used in plugin side)
func (p *BashPluginHandler) GRPCClient(context.Context, *plugin.GRPCBroker, *grpc.ClientConn) (any, error) {
	return nil, nil
}

// ===== Main Entry Point =====

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: pluginpkg.PluginHandshake,
		Plugins: map[string]plugin.Plugin{
			"plugin": &BashPluginHandler{pluginInstance: NewBash()},
		},
		GRPCServer: func(opts []grpc.ServerOption) *grpc.Server {
			return grpc.NewServer(opts...)
		},
	})
}
