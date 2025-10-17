// Package plugin RPC call argument definitions
package plugin

import "encoding/json"

// RPC argument structures
// Uses json.RawMessage for type flexibility, allowing plugins to define custom parameter formats

// NotifySendArgs contains arguments for sending notifications
type NotifySendArgs struct {
	Message json.RawMessage `json:"message"`
	Opts    json.RawMessage `json:"opts"`
}

// NotifyTemplateArgs contains arguments for sending template notifications
type NotifyTemplateArgs struct {
	Template string          `json:"template"`
	Data     json.RawMessage `json:"data"`
	Opts     json.RawMessage `json:"opts"`
}

// CIBuildArgs contains arguments for CI build operations
type CIBuildArgs struct {
	ProjectConfig json.RawMessage `json:"project_config"`
	Opts          json.RawMessage `json:"opts"`
}

// CITestArgs contains arguments for CI test operations
type CITestArgs struct {
	ProjectConfig json.RawMessage `json:"project_config"`
	Opts          json.RawMessage `json:"opts"`
}

// CILintArgs contains arguments for CI lint operations
type CILintArgs struct {
	ProjectConfig json.RawMessage `json:"project_config"`
	Opts          json.RawMessage `json:"opts"`
}

// CDDeployArgs contains arguments for CD deploy operations
type CDDeployArgs struct {
	ProjectConfig json.RawMessage `json:"project_config"`
	Opts          json.RawMessage `json:"opts"`
}

// CDRollbackArgs contains arguments for CD rollback operations
type CDRollbackArgs struct {
	ProjectConfig json.RawMessage `json:"project_config"`
	Opts          json.RawMessage `json:"opts"`
}

// SecurityScanArgs contains arguments for security scan operations
type SecurityScanArgs struct {
	ProjectConfig json.RawMessage `json:"project_config"`
	Opts          json.RawMessage `json:"opts"`
}

// SecurityAuditArgs contains arguments for security audit operations
type SecurityAuditArgs struct {
	ProjectConfig json.RawMessage `json:"project_config"`
	Opts          json.RawMessage `json:"opts"`
}

// StorageSaveArgs contains arguments for storage save operations
type StorageSaveArgs struct {
	Key  string          `json:"key"`
	Data json.RawMessage `json:"data"`
	Opts json.RawMessage `json:"opts"`
}

// StorageLoadArgs contains arguments for storage load operations
type StorageLoadArgs struct {
	Key  string          `json:"key"`
	Opts json.RawMessage `json:"opts"`
}

// StorageDeleteArgs contains arguments for storage delete operations
type StorageDeleteArgs struct {
	Key  string          `json:"key"`
	Opts json.RawMessage `json:"opts"`
}

// ApprovalApproveArgs contains arguments for approval operations
type ApprovalApproveArgs struct {
	ProjectConfig json.RawMessage `json:"project_config"`
	Opts          json.RawMessage `json:"opts"`
}

// CustomExecuteArgs contains arguments for custom execute operations
type CustomExecuteArgs struct {
	Action string          `json:"action"`
	Params json.RawMessage `json:"params"`
	Opts   json.RawMessage `json:"opts"`
}
