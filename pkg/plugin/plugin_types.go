// Package plugin type definitions
package plugin

import "encoding/json"

// PluginType is the plugin type enumeration
type PluginType string

const (
	// TypeSource Source plugin type
	TypeSource PluginType = "source"
	// TypeBuild Build plugin type
	TypeBuild PluginType = "build"
	// TypeTest Test plugin type
	TypeTest PluginType = "test"
	// TypeDeploy Deploy plugin type
	TypeDeploy PluginType = "deploy"
	// TypeSecurity Security plugin type
	TypeSecurity PluginType = "security"
	// TypeNotify Notify plugin type
	TypeNotify PluginType = "notify"
	// TypeApproval Approval plugin type
	TypeApproval PluginType = "approval"
	// TypeStorage Storage plugin type
	TypeStorage PluginType = "storage"
	// TypeAnalytics Analytics plugin type
	TypeAnalytics PluginType = "analytics"
	// TypeIntegration Integration plugin type
	TypeIntegration PluginType = "integration"
	// TypeCustom Custom plugin type
	TypeCustom PluginType = "custom"
)

// PluginConfig represents the plugin configuration
type PluginConfig struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Type        string            `json:"type"`
	Config      json.RawMessage   `json:"config"`
	Environment map[string]string `json:"environment"`
	TaskID      string            `json:"task_id"` // Task ID for log correlation
	LogHandlers []LogHandler      `json:"-"`       // Log handlers (not serialized)
}

// PluginInfo contains plugin information
type PluginInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Author      string `json:"author"`
	Homepage    string `json:"homepage,omitempty"`
}

// PluginMetrics contains plugin runtime metrics
type PluginMetrics struct {
	Name          string                 `json:"name"`
	Type          string                 `json:"type"`
	Version       string                 `json:"version"`
	Status        string                 `json:"status"`
	Uptime        int64                  `json:"uptime"`
	CallCount     int64                  `json:"call_count"`
	ErrorCount    int64                  `json:"error_count"`
	LastError     string                 `json:"last_error,omitempty"`
	LastCallTime  int64                  `json:"last_call_time"`
	CustomMetrics map[string]interface{} `json:"custom_metrics,omitempty"`
}

// RPCError represents an RPC error structure
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
}

// Error implements the error interface
func (e *RPCError) Error() string {
	return e.Message
}

// AllPluginTypes returns all supported plugin types
func AllPluginTypes() []PluginType {
	return []PluginType{
		TypeSource,
		TypeBuild,
		TypeTest,
		TypeDeploy,
		TypeSecurity,
		TypeNotify,
		TypeApproval,
		TypeStorage,
		TypeAnalytics,
		TypeIntegration,
		TypeCustom,
	}
}

// IsValidPluginType checks if a plugin type is valid
func IsValidPluginType(t string) bool {
	for _, validType := range AllPluginTypes() {
		if string(validType) == t {
			return true
		}
	}
	return false
}

// GetPluginTypeDescription returns a description for a plugin type
func GetPluginTypeDescription(t PluginType) string {
	descriptions := map[PluginType]string{
		TypeSource:      "Source code management plugin for repository operations (clone, pull, checkout, etc.)",
		TypeBuild:       "Build plugin for compiling and building projects (compile, package, generate artifacts, etc.)",
		TypeTest:        "Test plugin for running tests and generating reports (unit tests, integration tests, coverage, etc.)",
		TypeDeploy:      "Deployment plugin for application deployment and management (deploy, rollback, scaling, etc.)",
		TypeSecurity:    "Security plugin for security scanning and auditing (vulnerability scanning, compliance checks, etc.)",
		TypeNotify:      "Notification plugin for sending various notifications (email, webhook, instant messaging, etc.)",
		TypeApproval:    "Approval plugin for approval workflow management (create approval, approve, reject, etc.)",
		TypeStorage:     "Storage plugin for data storage and management (save, load, delete, list, etc.)",
		TypeAnalytics:   "Analytics plugin for data analysis and reporting (event tracking, queries, metrics, reports, etc.)",
		TypeIntegration: "Integration plugin for third-party service integration (connect, call, subscribe, etc.)",
		TypeCustom:      "Custom plugin for special-purpose customized functionality",
	}
	return descriptions[t]
}

// String returns the string representation of PluginType
func (pt PluginType) String() string {
	return string(pt)
}

// Validate validates the PluginType
func (pt PluginType) Validate() error {
	if !IsValidPluginType(string(pt)) {
		return &RPCError{
			Code:    400,
			Message: "invalid plugin type",
			Data:    string(pt),
		}
	}
	return nil
}
