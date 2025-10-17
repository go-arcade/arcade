// Package plugin type definitions
package plugin

import "encoding/json"

// PluginType is the plugin type enumeration
type PluginType string

const (
	// TypeCI is the CI plugin type
	TypeCI PluginType = "ci"
	// TypeCD is the CD plugin type
	TypeCD PluginType = "cd"
	// TypeSecurity is the security plugin type
	TypeSecurity PluginType = "security"
	// TypeNotify is the notification plugin type
	TypeNotify PluginType = "notify"
	// TypeStorage is the storage plugin type
	TypeStorage PluginType = "storage"
	// TypeApproval is the approval plugin type
	TypeApproval PluginType = "approval"
	// TypeCustom is the custom plugin type
	TypeCustom PluginType = "custom"
)

// PluginConfig represents the plugin configuration
type PluginConfig struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Type        string            `json:"type"`
	Config      json.RawMessage   `json:"config"`
	Environment map[string]string `json:"environment"`
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
