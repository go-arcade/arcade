package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/rpc"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/hashicorp/go-plugin"
	pluginpkg "github.com/observabil/arcade/pkg/plugin"
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

// BashPlugin implements the custom plugin
type BashPlugin struct {
	name        string
	description string
	version     string
	cfg         BashConfig
}

// NewBashPlugin creates a new bash plugin instance
func NewBashPlugin() *BashPlugin {
	return &BashPlugin{
		name:        "bash",
		description: "A custom plugin that executes bash scripts and commands",
		version:     "1.0.0",
		cfg: BashConfig{
			Shell:   "/bin/bash",
			WorkDir: "",
			Timeout: 300, // 5 minutes default
		},
	}
}

// ===== Implement RPC Interface =====

// Name returns the plugin name
func (p *BashPlugin) Name() (string, error) {
	return p.name, nil
}

// Description returns the plugin description
func (p *BashPlugin) Description() (string, error) {
	return p.description, nil
}

// Version returns the plugin version
func (p *BashPlugin) Version() (string, error) {
	return p.version, nil
}

// Type returns the plugin type
func (p *BashPlugin) Type() (string, error) {
	return string(pluginpkg.TypeCustom), nil
}

// Init initializes the plugin
func (p *BashPlugin) Init(config json.RawMessage) error {
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
func (p *BashPlugin) Cleanup() error {
	fmt.Println("[bash-plugin] cleanup completed")
	return nil
}

// Execute executes a bash script or command
func (p *BashPlugin) Execute(action string, params json.RawMessage, opts json.RawMessage) (json.RawMessage, error) {
	switch action {
	case "script":
		return p.executeScript(params, opts)
	case "command":
		return p.executeCommand(params, opts)
	case "file":
		return p.executeFile(params, opts)
	default:
		return nil, fmt.Errorf("unknown action: %s, supported actions: script, command, file", action)
	}
}

// executeScript executes a bash script from string
func (p *BashPlugin) executeScript(params json.RawMessage, opts json.RawMessage) (json.RawMessage, error) {
	var scriptParams struct {
		Script string            `json:"script"`
		Args   []string          `json:"args"`
		Env    map[string]string `json:"env"`
	}

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
	defer os.Remove(tmpFile.Name())

	// Write script content
	if _, err := tmpFile.WriteString(scriptParams.Script); err != nil {
		return nil, fmt.Errorf("failed to write script: %w", err)
	}
	tmpFile.Close()

	// Make it executable
	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		return nil, fmt.Errorf("failed to chmod script: %w", err)
	}

	// Execute the script
	return p.runCommand(p.cfg.Shell, append([]string{tmpFile.Name()}, scriptParams.Args...), scriptParams.Env)
}

// executeCommand executes a simple bash command
func (p *BashPlugin) executeCommand(params json.RawMessage, opts json.RawMessage) (json.RawMessage, error) {
	var cmdParams struct {
		Command string            `json:"command"`
		Env     map[string]string `json:"env"`
	}

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

// executeFile executes a bash script from a file
func (p *BashPlugin) executeFile(params json.RawMessage, opts json.RawMessage) (json.RawMessage, error) {
	var fileParams struct {
		FilePath string            `json:"file_path"`
		Args     []string          `json:"args"`
		Env      map[string]string `json:"env"`
	}

	if err := sonic.Unmarshal(params, &fileParams); err != nil {
		return nil, fmt.Errorf("failed to parse file params: %w", err)
	}

	if fileParams.FilePath == "" {
		return nil, fmt.Errorf("file_path is required")
	}

	// Check if file exists
	absPath := fileParams.FilePath
	if !filepath.IsAbs(absPath) {
		absPath = filepath.Join(p.cfg.WorkDir, fileParams.FilePath)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("script file not found: %s", absPath)
	}

	return p.runCommand(p.cfg.Shell, append([]string{absPath}, fileParams.Args...), fileParams.Env)
}

// runCommand executes a command with the given arguments and environment
func (p *BashPlugin) runCommand(command string, args []string, env map[string]string) (json.RawMessage, error) {
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
		if exitErr, ok := err.(*exec.ExitError); ok {
			result["exit_code"] = exitErr.ExitCode()
		}
	} else {
		result["exit_code"] = 0
	}

	return sonic.Marshal(result)
}

// ===== RPC Server Implementation =====

// BashPluginRPCServer is the RPC server wrapper
type BashPluginRPCServer struct {
	impl *BashPlugin
}

// Name RPC method
func (s *BashPluginRPCServer) Name(args string, reply *string) error {
	name, err := s.impl.Name()
	*reply = name
	return err
}

// Description RPC method
func (s *BashPluginRPCServer) Description(args string, reply *string) error {
	desc, err := s.impl.Description()
	*reply = desc
	return err
}

// Version RPC method
func (s *BashPluginRPCServer) Version(args string, reply *string) error {
	ver, err := s.impl.Version()
	*reply = ver
	return err
}

// Type RPC method
func (s *BashPluginRPCServer) Type(args string, reply *string) error {
	typ, err := s.impl.Type()
	*reply = typ
	return err
}

// Init RPC method
func (s *BashPluginRPCServer) Init(config json.RawMessage, reply *string) error {
	err := s.impl.Init(config)
	*reply = "initialized"
	return err
}

// Cleanup RPC method
func (s *BashPluginRPCServer) Cleanup(args string, reply *string) error {
	err := s.impl.Cleanup()
	*reply = "cleaned up"
	return err
}

// Execute RPC method
func (s *BashPluginRPCServer) Execute(args *pluginpkg.CustomExecuteArgs, reply *json.RawMessage) error {
	result, err := s.impl.Execute(args.Action, args.Params, args.Opts)
	if err != nil {
		return err
	}
	*reply = result
	return nil
}

// Ping RPC method
func (s *BashPluginRPCServer) Ping(args string, reply *string) error {
	*reply = "pong"
	return nil
}

// GetInfo RPC method
func (s *BashPluginRPCServer) GetInfo(args string, reply *pluginpkg.PluginInfo) error {
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
func (s *BashPluginRPCServer) GetMetrics(args string, reply *pluginpkg.PluginMetrics) error {
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
	Impl *BashPlugin
}

// Server returns the RPC server
func (p *BashPluginHandler) Server(*plugin.MuxBroker) (any, error) {
	return &BashPluginRPCServer{impl: p.Impl}, nil
}

// Client returns the RPC client (not used in plugin side)
func (BashPluginHandler) Client(b *plugin.MuxBroker, c *rpc.Client) (any, error) {
	return nil, nil
}

// ===== Main Entry Point =====

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: pluginpkg.RPCHandshake,
		Plugins: map[string]plugin.Plugin{
			"plugin": &BashPluginHandler{Impl: NewBashPlugin()},
		},
	})
}
