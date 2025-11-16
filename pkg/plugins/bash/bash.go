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
	pluginpkg "github.com/go-arcade/arcade/pkg/plugin"
	"github.com/hashicorp/go-plugin"
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

	if err := pluginpkg.UnmarshalParams(params, &scriptParams); err != nil {
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

	if err := pluginpkg.UnmarshalParams(params, &cmdParams); err != nil {
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

// ===== RPC Server Implementation =====

// BashPlugin is the RPC server wrapper
type BashPlugin struct {
	impl *Bash
}

// Name RPC method
func (s *BashPlugin) Name(args string, reply *string) error {
	name, err := s.impl.Name()
	*reply = name
	return err
}

// Description RPC method
func (s *BashPlugin) Description(args string, reply *string) error {
	desc, err := s.impl.Description()
	*reply = desc
	return err
}

// Version RPC method
func (s *BashPlugin) Version(args string, reply *string) error {
	ver, err := s.impl.Version()
	*reply = ver
	return err
}

// Type RPC method
func (s *BashPlugin) Type(args string, reply *string) error {
	typ, err := s.impl.Type()
	*reply = typ
	return err
}

// Init RPC method
func (s *BashPlugin) Init(config json.RawMessage, reply *string) error {
	err := s.impl.Init(config)
	*reply = "initialized"
	return err
}

// Cleanup RPC method
func (s *BashPlugin) Cleanup(args string, reply *string) error {
	err := s.impl.Cleanup()
	*reply = "cleaned up"
	return err
}

// Script RPC method - uses method name + json.RawMessage
// params: json.RawMessage containing BashScriptArgs
// opts: json.RawMessage containing optional overrides
func (s *BashPlugin) Script(args *pluginpkg.MethodArgs, reply *pluginpkg.MethodResult) error {
	result, err := s.impl.Execute("script", args.Params, args.Opts)
	if err != nil {
		reply.Error = err.Error()
		return nil
	}
	reply.Result = result
	return nil
}

// Command RPC method - uses method name + json.RawMessage
// params: json.RawMessage containing BashCommandArgs
// opts: json.RawMessage containing optional overrides
func (s *BashPlugin) Command(args *pluginpkg.MethodArgs, reply *pluginpkg.MethodResult) error {
	result, err := s.impl.Execute("command", args.Params, args.Opts)
	if err != nil {
		reply.Error = err.Error()
		return nil
	}
	reply.Result = result
	return nil
}

// Ping RPC method
func (s *BashPlugin) Ping(args string, reply *string) error {
	*reply = "pong"
	return nil
}

// GetInfo RPC method
func (s *BashPlugin) GetInfo(args string, reply *pluginpkg.PluginInfo) error {
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
func (s *BashPlugin) GetMetrics(args string, reply *pluginpkg.PluginMetrics) error {
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

// BashPluginHandler is the plugin handler
type BashPluginHandler struct {
	plugin.Plugin
	Impl *Bash
}

// Server returns the RPC server
func (p *BashPluginHandler) Server(*plugin.MuxBroker) (any, error) {
	return &BashPlugin{impl: p.Impl}, nil
}

// Client returns the RPC client (not used in plugin side)
func (p *BashPluginHandler) Client(b *plugin.MuxBroker, c *rpc.Client) (any, error) {
	return nil, nil
}

// ===== Main Entry Point =====

func main() {

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: pluginpkg.RPCHandshake,
		Plugins: map[string]plugin.Plugin{
			"plugin": &BashPluginHandler{Impl: NewBash()},
		},
	})
}
