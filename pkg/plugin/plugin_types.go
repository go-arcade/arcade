// Package plugin type definitions and helper functions
// Types are generated from proto files, this file only contains helper functions
package plugin

import (
	"encoding/json"

	pluginv1 "github.com/go-arcade/arcade/api/plugin/v1"
)

// Type aliases for convenience (using proto-generated types)
type (
	// PluginInfo is an alias for proto-generated PluginInfo
	PluginInfo = pluginv1.PluginInfo
	// PluginMetrics is an alias for proto-generated PluginMetrics
	PluginMetrics = pluginv1.PluginMetrics
	// PluginConfig is an alias for proto-generated PluginConfig
	PluginConfig = pluginv1.PluginConfig
	// RPCError is an alias for proto-generated RPCError
	RPCError = pluginv1.RPCError
	// PluginType is an alias for proto-generated PluginType
	PluginType = pluginv1.PluginType
)

// Plugin type constants (using proto-generated enum values)
const (
	// TypeSource Source plugin type
	TypeSource PluginType = pluginv1.PluginType_PLUGIN_TYPE_SOURCE
	// TypeBuild Build plugin type
	TypeBuild PluginType = pluginv1.PluginType_PLUGIN_TYPE_BUILD
	// TypeTest Test plugin type
	TypeTest PluginType = pluginv1.PluginType_PLUGIN_TYPE_TEST
	// TypeDeploy Deploy plugin type
	TypeDeploy PluginType = pluginv1.PluginType_PLUGIN_TYPE_DEPLOY
	// TypeSecurity Security plugin type
	TypeSecurity PluginType = pluginv1.PluginType_PLUGIN_TYPE_SECURITY
	// TypeNotify Notify plugin type
	TypeNotify PluginType = pluginv1.PluginType_PLUGIN_TYPE_NOTIFY
	// TypeApproval Approval plugin type
	TypeApproval PluginType = pluginv1.PluginType_PLUGIN_TYPE_APPROVAL
	// TypeStorage Storage plugin type
	TypeStorage PluginType = pluginv1.PluginType_PLUGIN_TYPE_STORAGE
	// TypeAnalytics Analytics plugin type
	TypeAnalytics PluginType = pluginv1.PluginType_PLUGIN_TYPE_ANALYTICS
	// TypeIntegration Integration plugin type
	TypeIntegration PluginType = pluginv1.PluginType_PLUGIN_TYPE_INTEGRATION
	// TypeCustom Custom plugin type
	TypeCustom PluginType = pluginv1.PluginType_PLUGIN_TYPE_CUSTOM
)

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
		if PluginTypeToString(validType) == t {
			return true
		}
	}
	return false
}

// PluginTypeToString converts PluginType enum to string
func PluginTypeToString(pt PluginType) string {
	switch pt {
	case pluginv1.PluginType_PLUGIN_TYPE_SOURCE:
		return "source"
	case pluginv1.PluginType_PLUGIN_TYPE_BUILD:
		return "build"
	case pluginv1.PluginType_PLUGIN_TYPE_TEST:
		return "test"
	case pluginv1.PluginType_PLUGIN_TYPE_DEPLOY:
		return "deploy"
	case pluginv1.PluginType_PLUGIN_TYPE_SECURITY:
		return "security"
	case pluginv1.PluginType_PLUGIN_TYPE_NOTIFY:
		return "notify"
	case pluginv1.PluginType_PLUGIN_TYPE_APPROVAL:
		return "approval"
	case pluginv1.PluginType_PLUGIN_TYPE_STORAGE:
		return "storage"
	case pluginv1.PluginType_PLUGIN_TYPE_ANALYTICS:
		return "analytics"
	case pluginv1.PluginType_PLUGIN_TYPE_INTEGRATION:
		return "integration"
	case pluginv1.PluginType_PLUGIN_TYPE_CUSTOM:
		return "custom"
	default:
		return "unknown"
	}
}

// StringToPluginType converts string to PluginType enum
func StringToPluginType(s string) PluginType {
	switch s {
	case "source":
		return pluginv1.PluginType_PLUGIN_TYPE_SOURCE
	case "build":
		return pluginv1.PluginType_PLUGIN_TYPE_BUILD
	case "test":
		return pluginv1.PluginType_PLUGIN_TYPE_TEST
	case "deploy":
		return pluginv1.PluginType_PLUGIN_TYPE_DEPLOY
	case "security":
		return pluginv1.PluginType_PLUGIN_TYPE_SECURITY
	case "notify":
		return pluginv1.PluginType_PLUGIN_TYPE_NOTIFY
	case "approval":
		return pluginv1.PluginType_PLUGIN_TYPE_APPROVAL
	case "storage":
		return pluginv1.PluginType_PLUGIN_TYPE_STORAGE
	case "analytics":
		return pluginv1.PluginType_PLUGIN_TYPE_ANALYTICS
	case "integration":
		return pluginv1.PluginType_PLUGIN_TYPE_INTEGRATION
	case "custom":
		return pluginv1.PluginType_PLUGIN_TYPE_CUSTOM
	default:
		return pluginv1.PluginType_PLUGIN_TYPE_UNSPECIFIED
	}
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

// PluginTypeString returns the string representation of PluginType
func PluginTypeString(pt PluginType) string {
	return PluginTypeToString(pt)
}

// ValidatePluginType validates the PluginType
func ValidatePluginType(pt PluginType) error {
	if pt == pluginv1.PluginType_PLUGIN_TYPE_UNSPECIFIED {
		return RPCErrorToError(&RPCError{
			Code:    400,
			Message: "invalid plugin type",
			Data:    PluginTypeToString(pt),
		})
	}
	return nil
}

// RPCErrorToError converts RPCError to error
func RPCErrorToError(rpcErr *RPCError) error {
	if rpcErr == nil {
		return nil
	}
	return &rpcErrorWrapper{rpcErr: rpcErr}
}

// rpcErrorWrapper wraps RPCError to implement error interface
type rpcErrorWrapper struct {
	rpcErr *RPCError
}

func (e *rpcErrorWrapper) Error() string {
	if e.rpcErr == nil {
		return ""
	}
	return e.rpcErr.GetMessage()
}

func (e *rpcErrorWrapper) Unwrap() error {
	// Return nil if rpcErr is nil, otherwise return the RPCError itself
	// Since RPCError doesn't implement error, we return nil
	// Callers can access the RPCError directly via the wrapper
	return nil
}

// GetRPCError returns the underlying RPCError
func (e *rpcErrorWrapper) GetRPCError() *RPCError {
	return e.rpcErr
}

// RuntimePluginConfig represents the plugin runtime configuration
// This is used internally and includes non-serializable fields like LogHandlers
type RuntimePluginConfig struct {
	Name        string
	Version     string
	Type        string
	Config      json.RawMessage
	Environment map[string]string
	TaskID      string
	LogHandlers []LogHandler // Log handlers (not serialized)
}

// ToProto converts RuntimePluginConfig to proto PluginConfig
func (c *RuntimePluginConfig) ToProto() *pluginv1.PluginConfig {
	return &pluginv1.PluginConfig{
		Name:        c.Name,
		Version:     c.Version,
		Type:        c.Type,
		Config:      c.Config,
		Environment: c.Environment,
		TaskId:      c.TaskID,
	}
}

// FromProto creates RuntimePluginConfig from proto PluginConfig
func (c *RuntimePluginConfig) FromProto(pc *pluginv1.PluginConfig) {
	if pc == nil {
		return
	}
	c.Name = pc.Name
	c.Version = pc.Version
	c.Type = pc.Type
	c.Config = pc.Config
	c.Environment = pc.Environment
	c.TaskID = pc.TaskId
}
