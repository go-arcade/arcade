// Package plugin unified interface definitions
package plugin

import (
	"encoding/json"

	streamv1 "github.com/go-arcade/arcade/api/stream/v1"
)

// IPlugin is the unified plugin interface
// Unifies all plugin types into a common Execute(action, payload) execution model
//
// Design principles:
//   - All plugin operations are distinguished by action strings
//   - Plugins use Action Registry to manage and route actions dynamically
//   - Host side only needs to call Execute method, no need to care about plugin types
//   - Action Registry provides extensibility, multi-language support, and maintainability
//
// Standard Action naming conventions:
//   - Source Plugin: "clone", "pull", "checkout", "commit.get", "commit.diff"
//   - Build Plugin: "build", "artifacts.get", "clean"
//   - Deploy Plugin: "deploy", "rollback", "status", "scale"
//   - Notify Plugin: "send", "send.template", "send.batch"
//   - Security Plugin: "scan", "audit", "vulnerabilities.get", "compliance.check"
//   - Test Plugin: "test", "results.get", "coverage.get"
//   - Storage Plugin: "save", "load", "delete", "list", "exists"
//   - Analytics Plugin: "track", "query", "metrics.get", "report.generate"
//   - Integration Plugin: "connect", "disconnect", "call", "subscribe", "unsubscribe"
//   - Approval Plugin: "approval.create", "approval.approve", "approval.reject", "approval.status"
//   - Config Actions: "config.query", "config.query.key", "config.list"
//   - Custom Plugin: "script", "command", "file" and other custom actions
type IPlugin interface {
	// Name returns the plugin name
	Name() (string, error)
	// Description returns the plugin description
	Description() (string, error)
	// Version returns the plugin version
	Version() (string, error)
	// Type returns the plugin type as string (source, build, deploy, etc.)
	Type() (string, error)
	// Init initializes the plugin with configuration
	Init(config json.RawMessage) error
	// Execute executes an action with given parameters
	// This is the unified entry point for all plugin operations.
	// action: action name (e.g., "clone", "build", "send.template", "config.query")
	// params: action-specific parameters (JSON)
	// opts: optional overrides (JSON, e.g., timeout, workdir, env)
	// Returns: action result (JSON) and error
	Execute(action string, params json.RawMessage, opts json.RawMessage) (json.RawMessage, error)
	// Cleanup cleans up the plugin resources
	Cleanup() error
}

// LogStreamWriter is used for streaming logs
type LogStreamWriter struct {
	PipelineID string
	StepID     string
	Client     streamv1.StreamServiceClient
}
