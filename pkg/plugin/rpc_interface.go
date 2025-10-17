// Package plugin RPC interface definitions
package plugin

import "encoding/json"

// RPC interface definitions - based on HashiCorp go-plugin
// All RPC interfaces use json.RawMessage to support dynamic type conversion

// BasePluginRPC is the base plugin RPC interface
type BasePluginRPC interface {
	Name() (string, error)
	Description() (string, error)
	Version() (string, error)
	Type() (string, error)
	Init(config json.RawMessage) error
	Cleanup() error
}

// NotifyPluginRPCInterface is the notification plugin RPC interface
type NotifyPluginRPCInterface interface {
	BasePluginRPC
	Send(message json.RawMessage, opts json.RawMessage) error
	SendTemplate(template string, data json.RawMessage, opts json.RawMessage) error
}

// CIPluginRPCInterface is the CI plugin RPC interface
type CIPluginRPCInterface interface {
	BasePluginRPC
	Build(projectConfig json.RawMessage, opts json.RawMessage) error
	Test(projectConfig json.RawMessage, opts json.RawMessage) error
	Lint(projectConfig json.RawMessage, opts json.RawMessage) error
}

// CDPluginRPCInterface is the CD plugin RPC interface
type CDPluginRPCInterface interface {
	BasePluginRPC
	Deploy(projectConfig json.RawMessage, opts json.RawMessage) error
	Rollback(projectConfig json.RawMessage, opts json.RawMessage) error
}

// SecurityPluginRPCInterface is the security plugin RPC interface
type SecurityPluginRPCInterface interface {
	BasePluginRPC
	Scan(projectConfig json.RawMessage, opts json.RawMessage) error
	Audit(projectConfig json.RawMessage, opts json.RawMessage) error
}

// StoragePluginRPCInterface is the storage plugin RPC interface
type StoragePluginRPCInterface interface {
	BasePluginRPC
	Save(key string, data json.RawMessage, opts json.RawMessage) error
	Load(key string, opts json.RawMessage) (json.RawMessage, error)
	Delete(key string, opts json.RawMessage) error
}

// ApprovalPluginRPCInterface is the approval plugin RPC interface
type ApprovalPluginRPCInterface interface {
	BasePluginRPC
	Approve(projectConfig json.RawMessage, opts json.RawMessage) error
}

// CustomPluginRPCInterface is the custom plugin RPC interface
type CustomPluginRPCInterface interface {
	BasePluginRPC
	Execute(action string, params json.RawMessage, opts json.RawMessage) (json.RawMessage, error)
}
