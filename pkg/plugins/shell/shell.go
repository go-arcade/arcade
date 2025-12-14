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
	"path/filepath"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	pluginv1 "github.com/go-arcade/arcade/api/plugin/v1"
	pluginpkg "github.com/go-arcade/arcade/pkg/plugin"
	"github.com/google/uuid"
	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

// ShellConfig is the plugin's configuration structure
type ShellConfig struct {
	// Shell path (default: /bin/sh)
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

// ShellScriptArgs contains arguments for executing a shell script
type ShellScriptArgs struct {
	Script string            `json:"script"`
	Args   []string          `json:"args,omitempty"`
	Env    map[string]string `json:"env,omitempty"`
}

// ShellCommandArgs contains arguments for executing a shell command
type ShellCommandArgs struct {
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

// Shell implements the custom plugin
type Shell struct {
	*pluginpkg.PluginBase
	name        string
	description string
	version     string
	cfg         ShellConfig
}

// Action definitions - maintains action names and descriptions
var (
	actions = map[string]string{
		"script":  "Execute a shell script from string",
		"command": "Execute a shell command",
	}
)

// NewShell creates a new shell plugin instance
func NewShell() *Shell {
	p := &Shell{
		PluginBase:  pluginpkg.NewPluginBase(),
		name:        "shell",
		description: "A custom plugin that executes shell scripts and commands",
		version:     "1.0.0",
		cfg: ShellConfig{
			Shell:   "/bin/sh",
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
func (p *Shell) registerActions() {
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

// Name returns the plugin name
func (p *Shell) Name() (string, error) {
	return p.name, nil
}

// Description returns the plugin description
func (p *Shell) Description() (string, error) {
	return p.description, nil
}

// Version returns the plugin version
func (p *Shell) Version() (string, error) {
	return p.version, nil
}

// Type returns the plugin type
func (p *Shell) Type() (string, error) {
	return string(pluginpkg.TypeCustom), nil
}

// Init initializes the plugin
func (p *Shell) Init(config json.RawMessage) error {
	if len(config) > 0 {
		if err := sonic.Unmarshal(config, &p.cfg); err != nil {
			return fmt.Errorf("failed to parse config: %w", err)
		}
	}

	// Validate shell path
	if p.cfg.Shell == "" {
		p.cfg.Shell = "/bin/sh"
		fmt.Fprintf(os.Stderr, "[shell-plugin] WARNING: Using default shell interpreter /bin/sh. Some bash-specific features may not work. Consider using /bin/bash for better compatibility.\n")
	}

	// Check if shell exists
	if _, err := os.Stat(p.cfg.Shell); os.IsNotExist(err) {
		return fmt.Errorf("shell not found: %s", p.cfg.Shell)
	}

	// Warn if shell is /bin/sh
	if p.cfg.Shell == "/bin/sh" {
		fmt.Fprintf(os.Stderr, "[shell-plugin] WARNING: Using /bin/sh as shell interpreter. Some bash-specific features may not work. Consider using /bin/bash for better compatibility.\n")
	}

	// Set default working directory if not specified
	if p.cfg.WorkDir == "" {
		p.cfg.WorkDir, _ = os.Getwd()
	}

	fmt.Printf("[shell-plugin] initialized with shell: %s, work_dir: %s\n", p.cfg.Shell, p.cfg.WorkDir)
	return nil
}

// Cleanup cleans up the plugin
func (p *Shell) Cleanup() error {
	fmt.Println("[shell-plugin] cleanup completed")
	return nil
}

// Execute executes a shell script or command using Action Registry
// All actions are registered in registerActions() and routed through the registry
func (p *Shell) Execute(action string, params json.RawMessage, opts json.RawMessage) (json.RawMessage, error) {
	return p.PluginBase.Execute(action, params, opts)
}

// executeScript executes a shell script from string
func (p *Shell) executeScript(params json.RawMessage, opts json.RawMessage) (json.RawMessage, error) {
	var scriptParams ShellScriptArgs

	if err := sonic.Unmarshal(params, &scriptParams); err != nil {
		return nil, fmt.Errorf("failed to parse script params: %w", err)
	}

	if scriptParams.Script == "" {
		return nil, fmt.Errorf("script content is required")
	}

	// Parse opts for runtime options (workspace, timeout, env)
	var optsMap map[string]any
	if len(opts) > 0 {
		if err := sonic.Unmarshal(opts, &optsMap); err == nil {
			// Merge additional env vars from opts
			if envOpts, ok := optsMap["env"].(map[string]any); ok {
				if scriptParams.Env == nil {
					scriptParams.Env = make(map[string]string)
				}
				for k, v := range envOpts {
					if vStr, ok := v.(string); ok {
						scriptParams.Env[k] = vStr
					}
				}
			}
		}
	}

	// Security check: prevent dangerous operations if not allowed
	if !p.cfg.AllowDangerous {
		// TODO: 使用agent的dangerous_operations配置来检查
		dangerous := []string{"rm -rf", ":(){ :|:& };:", "mkfs", "dd if=", "> /dev/"}
		for _, pattern := range dangerous {
			if strings.Contains(scriptParams.Script, pattern) {
				return nil, fmt.Errorf("dangerous operation detected and not allowed: %s", pattern)
			}
		}
	}

	// Create a temporary script file with UUID-based name
	tmpDir := os.TempDir()
	tmpFileName := filepath.Join(tmpDir, fmt.Sprintf("%s.sh", uuid.New().String()))
	tmpFile, err := os.Create(tmpFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() {
		if err := os.Remove(tmpFileName); err != nil {
			// Log cleanup error to stderr (plugin process may not have logger initialized)
			fmt.Fprintf(os.Stderr, "[shell-plugin] WARNING: failed to remove temp file %s: %v\n", tmpFileName, err)
		}
	}()

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

	// Determine working directory from opts
	workDir := p.cfg.WorkDir
	if len(opts) > 0 && optsMap != nil {
		if workspace, ok := optsMap["workspace"].(string); ok && workspace != "" {
			workDir = workspace
		}
	}

	// Execute the script with opts-aware runCommand
	return p.runCommandWithOpts(p.cfg.Shell, append([]string{tmpFile.Name()}, scriptParams.Args...), scriptParams.Env, workDir, opts)
}

// executeCommand executes a simple shell command
func (p *Shell) executeCommand(params json.RawMessage, opts json.RawMessage) (json.RawMessage, error) {
	var cmdParams ShellCommandArgs

	if err := sonic.Unmarshal(params, &cmdParams); err != nil {
		return nil, fmt.Errorf("failed to parse command params: %w", err)
	}

	if cmdParams.Command == "" {
		return nil, fmt.Errorf("command is required")
	}

	// Parse opts for runtime options (workspace, timeout, env)
	var optsMap map[string]any
	if len(opts) > 0 {
		if err := sonic.Unmarshal(opts, &optsMap); err == nil {
			// Merge additional env vars from opts
			if envOpts, ok := optsMap["env"].(map[string]any); ok {
				if cmdParams.Env == nil {
					cmdParams.Env = make(map[string]string)
				}
				for k, v := range envOpts {
					if vStr, ok := v.(string); ok {
						cmdParams.Env[k] = vStr
					}
				}
			}
		}
	}

	// Security check
	if !p.cfg.AllowDangerous {
		// TODO: 使用agent的dangerous_operations配置来检查
		dangerous := []string{"rm -rf /", ":(){ :|:& };:", "mkfs", "dd if=/dev/zero", "> /dev/"}
		for _, pattern := range dangerous {
			if strings.Contains(cmdParams.Command, pattern) {
				return nil, fmt.Errorf("dangerous operation detected and not allowed: %s", pattern)
			}
		}
	}

	// Determine working directory from opts
	workDir := p.cfg.WorkDir
	if len(opts) > 0 && optsMap != nil {
		if workspace, ok := optsMap["workspace"].(string); ok && workspace != "" {
			workDir = workspace
		}
	}

	return p.runCommandWithOpts(p.cfg.Shell, []string{"-c", cmdParams.Command}, cmdParams.Env, workDir, opts)
}

// runCommandWithOpts executes a command with the given arguments, environment, and runtime options
func (p *Shell) runCommandWithOpts(command string, args []string, env map[string]string, workDir string, opts json.RawMessage) (json.RawMessage, error) {
	ctx := context.Background()
	timeout := time.Duration(p.cfg.Timeout) * time.Second

	// Parse opts for timeout override
	if len(opts) > 0 {
		var optsMap map[string]any
		if err := sonic.Unmarshal(opts, &optsMap); err == nil {
			if timeoutOpt, ok := optsMap["timeout"].(float64); ok && timeoutOpt > 0 {
				timeout = time.Duration(timeoutOpt) * time.Second
			}
		}
	}

	// Create context with timeout if configured
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	// Create command
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = workDir

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

// ShellPluginHandler is the gRPC plugin handler
type ShellPluginHandler struct {
	plugin.Plugin
	pluginInstance *Shell
}

// Server returns the RPC server (required by plugin.Plugin interface, not used for gRPC)
func (p *ShellPluginHandler) Server(*plugin.MuxBroker) (any, error) {
	return nil, fmt.Errorf("RPC protocol not supported, use gRPC protocol instead")
}

// Client returns the RPC client (required by plugin.Plugin interface, not used for gRPC)
func (p *ShellPluginHandler) Client(*plugin.MuxBroker, *rpc.Client) (any, error) {
	return nil, fmt.Errorf("RPC protocol not supported, use gRPC protocol instead")
}

// GRPCServer returns the gRPC server
func (p *ShellPluginHandler) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
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
func (p *ShellPluginHandler) GRPCClient(context.Context, *plugin.GRPCBroker, *grpc.ClientConn) (any, error) {
	return nil, nil
}

// ===== Main Entry Point =====

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: pluginpkg.PluginHandshake,
		Plugins: map[string]plugin.Plugin{
			"plugin": &ShellPluginHandler{pluginInstance: NewShell()},
		},
		GRPCServer: func(opts []grpc.ServerOption) *grpc.Server {
			return grpc.NewServer(opts...)
		},
	})
}
