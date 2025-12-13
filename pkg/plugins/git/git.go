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
	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

// GitConfig is the plugin's configuration structure
type GitConfig struct {
	// Git executable path (default: git)
	GitPath string `json:"git_path"`
	// Default timeout in seconds (0 means no timeout)
	Timeout int `json:"timeout"`
	// Default working directory
	WorkDir string `json:"work_dir"`
	// Default user name for commits
	UserName string `json:"user_name"`
	// Default user email for commits
	UserEmail string `json:"user_email"`
	// Whether to use shallow clone by default
	Shallow bool `json:"shallow"`
	// Depth for shallow clone
	Depth int `json:"depth"`
}

// ========== Plugin Action Arguments ==========

// CloneArgs contains arguments for cloning a repository
type CloneArgs struct {
	Repo    string            `json:"repo"`    // Repository URL
	Branch  string            `json:"branch"`  // Branch to clone (optional)
	Tag     string            `json:"tag"`     // Tag to clone (optional)
	Commit  string            `json:"commit"`  // Commit SHA to clone (optional)
	Path    string            `json:"path"`    // Destination path (optional)
	Shallow *bool             `json:"shallow"` // Override shallow clone setting
	Depth   *int              `json:"depth"`   // Override depth setting
	Auth    map[string]string `json:"auth"`    // Authentication (username, password, token, ssh_key)
	Env     map[string]string `json:"env"`     // Additional environment variables
}

// CheckoutArgs contains arguments for checking out a branch/tag/commit
type CheckoutArgs struct {
	Ref    string            `json:"ref"`    // Branch, tag, or commit SHA
	Path   string            `json:"path"`   // Repository path (required)
	Force  bool              `json:"force"`  // Force checkout
	Create bool              `json:"create"` // Create branch if not exists
	Env    map[string]string `json:"env"`    // Additional environment variables
}

// PullArgs contains arguments for pulling updates
type PullArgs struct {
	Path   string            `json:"path"`   // Repository path (required)
	Branch string            `json:"branch"` // Branch to pull (optional, defaults to current)
	Remote string            `json:"remote"` // Remote name (optional, defaults to origin)
	Rebase bool              `json:"rebase"` // Use rebase instead of merge
	Env    map[string]string `json:"env"`    // Additional environment variables
}

// StatusArgs contains arguments for checking repository status
type StatusArgs struct {
	Path  string            `json:"path"`  // Repository path (required)
	Short bool              `json:"short"` // Short format output
	Env   map[string]string `json:"env"`   // Additional environment variables
}

// LogArgs contains arguments for viewing commit history
type LogArgs struct {
	Path   string            `json:"path"`   // Repository path (required)
	Ref    string            `json:"ref"`    // Branch/tag/commit to show log from (optional)
	Limit  int               `json:"limit"`  // Limit number of commits (optional)
	Since  string            `json:"since"`  // Show commits since date (optional)
	Until  string            `json:"until"`  // Show commits until date (optional)
	Author string            `json:"author"` // Filter by author (optional)
	Env    map[string]string `json:"env"`    // Additional environment variables
}

// BranchArgs contains arguments for listing branches
type BranchArgs struct {
	Path   string            `json:"path"`   // Repository path (required)
	Remote bool              `json:"remote"` // List remote branches
	All    bool              `json:"all"`    // List all branches
	Env    map[string]string `json:"env"`    // Additional environment variables
}

// TagArgs contains arguments for listing tags
type TagArgs struct {
	Path string            `json:"path"` // Repository path (required)
	Sort string            `json:"sort"` // Sort order (version, date, etc.)
	Env  map[string]string `json:"env"`  // Additional environment variables
}

// Git implements the git plugin
type Git struct {
	*pluginpkg.PluginBase
	name        string
	description string
	version     string
	cfg         GitConfig
}

// Action definitions
var (
	actions = map[string]string{
		"clone":    "Clone a git repository",
		"checkout": "Checkout a branch, tag, or commit",
		"pull":     "Pull latest changes from remote",
		"status":   "Check repository status",
		"log":      "View commit history",
		"branch":   "List branches",
		"tag":      "List tags",
	}
)

// NewGit creates a new git plugin instance
func NewGit() *Git {
	p := &Git{
		PluginBase:  pluginpkg.NewPluginBase(),
		name:        "git",
		description: "Git version control plugin for repository operations",
		version:     "1.0.0",
		cfg: GitConfig{
			GitPath: "git",
			Timeout: 300, // 5 minutes default
			Shallow: false,
			Depth:   1,
		},
	}

	// Register actions using Action Registry
	p.registerActions()
	return p
}

// registerActions registers all actions for this plugin
func (p *Git) registerActions() {
	// Register "clone" action
	if err := p.Registry().RegisterFunc("clone", actions["clone"], func(params json.RawMessage, opts json.RawMessage) (json.RawMessage, error) {
		return p.clone(params, opts)
	}); err != nil {
		return
	}

	// Register "checkout" action
	if err := p.Registry().RegisterFunc("checkout", actions["checkout"], func(params json.RawMessage, opts json.RawMessage) (json.RawMessage, error) {
		return p.checkout(params, opts)
	}); err != nil {
		return
	}

	// Register "pull" action
	if err := p.Registry().RegisterFunc("pull", actions["pull"], func(params json.RawMessage, opts json.RawMessage) (json.RawMessage, error) {
		return p.pull(params, opts)
	}); err != nil {
		return
	}

	// Register "status" action
	if err := p.Registry().RegisterFunc("status", actions["status"], func(params json.RawMessage, opts json.RawMessage) (json.RawMessage, error) {
		return p.status(params, opts)
	}); err != nil {
		return
	}

	// Register "log" action
	if err := p.Registry().RegisterFunc("log", actions["log"], func(params json.RawMessage, opts json.RawMessage) (json.RawMessage, error) {
		return p.log(params, opts)
	}); err != nil {
		return
	}

	// Register "branch" action
	if err := p.Registry().RegisterFunc("branch", actions["branch"], func(params json.RawMessage, opts json.RawMessage) (json.RawMessage, error) {
		return p.branch(params, opts)
	}); err != nil {
		return
	}

	// Register "tag" action
	if err := p.Registry().RegisterFunc("tag", actions["tag"], func(params json.RawMessage, opts json.RawMessage) (json.RawMessage, error) {
		return p.tag(params, opts)
	}); err != nil {
		return
	}
}

// ===== Implement RPC Interface =====

// Name returns the plugin name
func (p *Git) Name() (string, error) {
	return p.name, nil
}

// Description returns the plugin description
func (p *Git) Description() (string, error) {
	return p.description, nil
}

// Version returns the plugin version
func (p *Git) Version() (string, error) {
	return p.version, nil
}

// Type returns the plugin type
func (p *Git) Type() (string, error) {
	return pluginpkg.PluginTypeToString(pluginpkg.TypeSource), nil
}

// Init initializes the plugin
func (p *Git) Init(config json.RawMessage) error {
	if len(config) > 0 {
		if err := sonic.Unmarshal(config, &p.cfg); err != nil {
			return fmt.Errorf("failed to parse config: %w", err)
		}
	}

	// Validate git path
	if p.cfg.GitPath == "" {
		p.cfg.GitPath = "git"
	}

	// Check if git exists
	if _, err := exec.LookPath(p.cfg.GitPath); err != nil {
		return fmt.Errorf("git not found: %s", p.cfg.GitPath)
	}

	fmt.Printf("[git-plugin] initialized with git_path: %s, timeout: %d\n", p.cfg.GitPath, p.cfg.Timeout)
	return nil
}

// Cleanup cleans up the plugin
func (p *Git) Cleanup() error {
	fmt.Println("[git-plugin] cleanup completed")
	return nil
}

// Execute executes git operations using Action Registry
func (p *Git) Execute(action string, params json.RawMessage, opts json.RawMessage) (json.RawMessage, error) {
	return p.PluginBase.Execute(action, params, opts)
}

// ===== Action Handlers =====

// clone clones a git repository
func (p *Git) clone(params json.RawMessage, opts json.RawMessage) (json.RawMessage, error) {
	var cloneParams CloneArgs
	if err := sonic.Unmarshal(params, &cloneParams); err != nil {
		return nil, fmt.Errorf("failed to parse clone params: %w", err)
	}

	if cloneParams.Repo == "" {
		return nil, fmt.Errorf("repository URL is required")
	}

	// Parse opts for workspace
	var optsMap map[string]any
	if len(opts) > 0 {
		if err := sonic.Unmarshal(opts, &optsMap); err == nil {
			if workspace, ok := optsMap["workspace"].(string); ok && workspace != "" {
				if cloneParams.Path == "" {
					cloneParams.Path = workspace
				}
			}
		}
	}

	// Determine destination path
	destPath := cloneParams.Path
	if destPath == "" {
		// Extract repo name from URL
		repoName := filepath.Base(strings.TrimSuffix(cloneParams.Repo, ".git"))
		destPath = repoName
	}

	// Build git clone command
	args := []string{"clone"}

	// Shallow clone
	shallow := p.cfg.Shallow
	if cloneParams.Shallow != nil {
		shallow = *cloneParams.Shallow
	}
	if shallow {
		depth := p.cfg.Depth
		if cloneParams.Depth != nil {
			depth = *cloneParams.Depth
		}
		if depth > 0 {
			args = append(args, "--depth", fmt.Sprintf("%d", depth))
		}
		args = append(args, "--shallow-submodules")
	}

	// Branch/Tag/Commit
	if cloneParams.Branch != "" {
		args = append(args, "--branch", cloneParams.Branch)
	} else if cloneParams.Tag != "" {
		args = append(args, "--branch", cloneParams.Tag)
	}

	args = append(args, cloneParams.Repo, destPath)

	// Execute clone
	result, err := p.runGitCommand(args, nil, cloneParams.Env, "")
	if err != nil {
		return nil, fmt.Errorf("clone failed: %w", err)
	}
	if !result["success"].(bool) {
		errorMsg := result["stderr"].(string)
		if errorMsg == "" {
			if errStr, ok := result["error"].(string); ok {
				errorMsg = errStr
			} else {
				errorMsg = "clone command failed"
			}
		}
		return nil, fmt.Errorf("clone failed: %s", errorMsg)
	}

	// If specific commit was requested, checkout it
	if cloneParams.Commit != "" {
		checkoutResult, err := p.runGitCommand([]string{"checkout", cloneParams.Commit}, nil, cloneParams.Env, destPath)
		if err != nil {
			return nil, fmt.Errorf("checkout commit failed: %w", err)
		}
		result["checkout"] = checkoutResult
	}

	result["path"] = destPath
	return sonic.Marshal(result)
}

// checkout checks out a branch, tag, or commit
func (p *Git) checkout(params json.RawMessage, opts json.RawMessage) (json.RawMessage, error) {
	var checkoutParams CheckoutArgs
	if err := sonic.Unmarshal(params, &checkoutParams); err != nil {
		return nil, fmt.Errorf("failed to parse checkout params: %w", err)
	}

	if checkoutParams.Ref == "" {
		return nil, fmt.Errorf("ref (branch/tag/commit) is required")
	}

	// Parse opts for workspace
	var optsMap map[string]any
	if len(opts) > 0 {
		if err := sonic.Unmarshal(opts, &optsMap); err == nil {
			if workspace, ok := optsMap["workspace"].(string); ok && workspace != "" {
				if checkoutParams.Path == "" {
					checkoutParams.Path = workspace
				}
			}
		}
	}

	if checkoutParams.Path == "" {
		return nil, fmt.Errorf("repository path is required")
	}

	args := []string{"checkout"}
	if checkoutParams.Force {
		args = append(args, "-f")
	}
	if checkoutParams.Create {
		args = append(args, "-b", checkoutParams.Ref)
	} else {
		args = append(args, checkoutParams.Ref)
	}

	result, err := p.runGitCommand(args, nil, checkoutParams.Env, checkoutParams.Path)
	if err != nil {
		return nil, err
	}
	return sonic.Marshal(result)
}

// pull pulls latest changes from remote
func (p *Git) pull(params json.RawMessage, opts json.RawMessage) (json.RawMessage, error) {
	var pullParams PullArgs
	if err := sonic.Unmarshal(params, &pullParams); err != nil {
		return nil, fmt.Errorf("failed to parse pull params: %w", err)
	}

	// Parse opts for workspace
	var optsMap map[string]any
	if len(opts) > 0 {
		if err := sonic.Unmarshal(opts, &optsMap); err == nil {
			if workspace, ok := optsMap["workspace"].(string); ok && workspace != "" {
				if pullParams.Path == "" {
					pullParams.Path = workspace
				}
			}
		}
	}

	if pullParams.Path == "" {
		return nil, fmt.Errorf("repository path is required")
	}

	args := []string{"pull"}
	if pullParams.Rebase {
		args = append(args, "--rebase")
	}
	if pullParams.Remote != "" {
		args = append(args, pullParams.Remote)
	}
	if pullParams.Branch != "" {
		args = append(args, pullParams.Branch)
	}

	result, err := p.runGitCommand(args, nil, pullParams.Env, pullParams.Path)
	if err != nil {
		return nil, err
	}
	return sonic.Marshal(result)
}

// status checks repository status
func (p *Git) status(params json.RawMessage, opts json.RawMessage) (json.RawMessage, error) {
	var statusParams StatusArgs
	if err := sonic.Unmarshal(params, &statusParams); err != nil {
		return nil, fmt.Errorf("failed to parse status params: %w", err)
	}

	// Parse opts for workspace
	var optsMap map[string]any
	if len(opts) > 0 {
		if err := sonic.Unmarshal(opts, &optsMap); err == nil {
			if workspace, ok := optsMap["workspace"].(string); ok && workspace != "" {
				if statusParams.Path == "" {
					statusParams.Path = workspace
				}
			}
		}
	}

	if statusParams.Path == "" {
		return nil, fmt.Errorf("repository path is required")
	}

	args := []string{"status"}
	if statusParams.Short {
		args = append(args, "--short")
	}

	result, err := p.runGitCommand(args, nil, statusParams.Env, statusParams.Path)
	if err != nil {
		return nil, err
	}
	return sonic.Marshal(result)
}

// log views commit history
func (p *Git) log(params json.RawMessage, opts json.RawMessage) (json.RawMessage, error) {
	var logParams LogArgs
	if err := sonic.Unmarshal(params, &logParams); err != nil {
		return nil, fmt.Errorf("failed to parse log params: %w", err)
	}

	// Parse opts for workspace
	var optsMap map[string]any
	if len(opts) > 0 {
		if err := sonic.Unmarshal(opts, &optsMap); err == nil {
			if workspace, ok := optsMap["workspace"].(string); ok && workspace != "" {
				if logParams.Path == "" {
					logParams.Path = workspace
				}
			}
		}
	}

	if logParams.Path == "" {
		return nil, fmt.Errorf("repository path is required")
	}

	args := []string{"log", "--pretty=format:%H|%an|%ae|%ad|%s", "--date=iso"}
	if logParams.Limit > 0 {
		args = append(args, fmt.Sprintf("-%d", logParams.Limit))
	}
	if logParams.Since != "" {
		args = append(args, "--since", logParams.Since)
	}
	if logParams.Until != "" {
		args = append(args, "--until", logParams.Until)
	}
	if logParams.Author != "" {
		args = append(args, "--author", logParams.Author)
	}
	if logParams.Ref != "" {
		args = append(args, logParams.Ref)
	}

	result, err := p.runGitCommand(args, nil, logParams.Env, logParams.Path)
	if err != nil {
		return nil, err
	}
	return sonic.Marshal(result)
}

// branch lists branches
func (p *Git) branch(params json.RawMessage, opts json.RawMessage) (json.RawMessage, error) {
	var branchParams BranchArgs
	if err := sonic.Unmarshal(params, &branchParams); err != nil {
		return nil, fmt.Errorf("failed to parse branch params: %w", err)
	}

	// Parse opts for workspace
	var optsMap map[string]any
	if len(opts) > 0 {
		if err := sonic.Unmarshal(opts, &optsMap); err == nil {
			if workspace, ok := optsMap["workspace"].(string); ok && workspace != "" {
				if branchParams.Path == "" {
					branchParams.Path = workspace
				}
			}
		}
	}

	if branchParams.Path == "" {
		return nil, fmt.Errorf("repository path is required")
	}

	args := []string{"branch"}
	if branchParams.All {
		args = append(args, "-a")
	} else if branchParams.Remote {
		args = append(args, "-r")
	}

	result, err := p.runGitCommand(args, nil, branchParams.Env, branchParams.Path)
	if err != nil {
		return nil, err
	}
	return sonic.Marshal(result)
}

// tag lists tags
func (p *Git) tag(params json.RawMessage, opts json.RawMessage) (json.RawMessage, error) {
	var tagParams TagArgs
	if err := sonic.Unmarshal(params, &tagParams); err != nil {
		return nil, fmt.Errorf("failed to parse tag params: %w", err)
	}

	// Parse opts for workspace
	var optsMap map[string]any
	if len(opts) > 0 {
		if err := sonic.Unmarshal(opts, &optsMap); err == nil {
			if workspace, ok := optsMap["workspace"].(string); ok && workspace != "" {
				if tagParams.Path == "" {
					tagParams.Path = workspace
				}
			}
		}
	}

	if tagParams.Path == "" {
		return nil, fmt.Errorf("repository path is required")
	}

	args := []string{"tag"}
	if tagParams.Sort != "" {
		args = append(args, "--sort", tagParams.Sort)
	}

	result, err := p.runGitCommand(args, nil, tagParams.Env, tagParams.Path)
	if err != nil {
		return nil, err
	}
	return sonic.Marshal(result)
}

// runGitCommand executes a git command
func (p *Git) runGitCommand(args []string, auth map[string]string, env map[string]string, workDir string) (map[string]any, error) {
	ctx := context.Background()
	timeout := time.Duration(p.cfg.Timeout) * time.Second

	if p.cfg.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, p.cfg.GitPath, args...)
	if workDir != "" {
		cmd.Dir = workDir
	} else if p.cfg.WorkDir != "" {
		cmd.Dir = p.cfg.WorkDir
	}

	// Set environment variables
	cmd.Env = os.Environ()

	// Configure git user if provided
	if p.cfg.UserName != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("GIT_AUTHOR_NAME=%s", p.cfg.UserName))
		cmd.Env = append(cmd.Env, fmt.Sprintf("GIT_COMMITTER_NAME=%s", p.cfg.UserName))
	}
	if p.cfg.UserEmail != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("GIT_AUTHOR_EMAIL=%s", p.cfg.UserEmail))
		cmd.Env = append(cmd.Env, fmt.Sprintf("GIT_COMMITTER_EMAIL=%s", p.cfg.UserEmail))
	}

	// Add authentication environment variables
	if auth != nil {
		if username, ok := auth["username"]; ok {
			cmd.Env = append(cmd.Env, fmt.Sprintf("GIT_USERNAME=%s", username))
		}
		if password, ok := auth["password"]; ok {
			cmd.Env = append(cmd.Env, fmt.Sprintf("GIT_PASSWORD=%s", password))
		}
		if token, ok := auth["token"]; ok {
			cmd.Env = append(cmd.Env, fmt.Sprintf("GIT_TOKEN=%s", token))
		}
	}

	// Add custom environment variables
	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	startTime := time.Now()
	err := cmd.Run()
	duration := time.Since(startTime)

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

	return result, nil
}

// ===== gRPC Plugin Handler =====

// GitPluginHandler is the gRPC plugin handler
type GitPluginHandler struct {
	plugin.Plugin
	pluginInstance *Git
}

// Server returns the RPC server (required by plugin.Plugin interface, not used for gRPC)
func (p *GitPluginHandler) Server(*plugin.MuxBroker) (any, error) {
	return nil, fmt.Errorf("RPC protocol not supported, use gRPC protocol instead")
}

// Client returns the RPC client (required by plugin.Plugin interface, not used for gRPC)
func (p *GitPluginHandler) Client(*plugin.MuxBroker, *rpc.Client) (any, error) {
	return nil, fmt.Errorf("RPC protocol not supported, use gRPC protocol instead")
}

// GRPCServer returns the gRPC server
func (p *GitPluginHandler) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
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
func (p *GitPluginHandler) GRPCClient(context.Context, *plugin.GRPCBroker, *grpc.ClientConn) (any, error) {
	return nil, nil
}

// ===== Main Entry Point =====

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: pluginpkg.PluginHandshake,
		Plugins: map[string]plugin.Plugin{
			"plugin": &GitPluginHandler{pluginInstance: NewGit()},
		},
		GRPCServer: func(opts []grpc.ServerOption) *grpc.Server {
			return grpc.NewServer(opts...)
		},
	})
}
