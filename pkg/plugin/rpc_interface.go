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

// SourcePluginRPCInterface is the source code management plugin RPC interface
type SourcePluginRPCInterface interface {
	BasePluginRPC
	Clone(repoURL string, opts json.RawMessage) error
	Pull(opts json.RawMessage) error
	Checkout(branch string, opts json.RawMessage) error
	GetCommit(opts json.RawMessage) (json.RawMessage, error)
	GetDiff(opts json.RawMessage) (json.RawMessage, error)
}

// BuildPluginRPCInterface is the build plugin RPC interface
type BuildPluginRPCInterface interface {
	BasePluginRPC
	Build(projectConfig json.RawMessage, opts json.RawMessage) error
	GetArtifacts(opts json.RawMessage) (json.RawMessage, error)
	Clean(opts json.RawMessage) error
}

// TestPluginRPCInterface is the test plugin RPC interface
type TestPluginRPCInterface interface {
	BasePluginRPC
	Test(projectConfig json.RawMessage, opts json.RawMessage) error
	GetTestResults(opts json.RawMessage) (json.RawMessage, error)
	GetCoverage(opts json.RawMessage) (json.RawMessage, error)
}

// DeployPluginRPCInterface is the deployment plugin RPC interface
type DeployPluginRPCInterface interface {
	BasePluginRPC
	Deploy(projectConfig json.RawMessage, opts json.RawMessage) error
	Rollback(version string, opts json.RawMessage) error
	GetStatus(opts json.RawMessage) (json.RawMessage, error)
	Scale(replicas int, opts json.RawMessage) error
}

// SecurityPluginRPCInterface is the security plugin RPC interface
type SecurityPluginRPCInterface interface {
	BasePluginRPC
	Scan(projectConfig json.RawMessage, opts json.RawMessage) error
	Audit(projectConfig json.RawMessage, opts json.RawMessage) error
	GetVulnerabilities(opts json.RawMessage) (json.RawMessage, error)
	CheckCompliance(opts json.RawMessage) (json.RawMessage, error)
}

// NotifyPluginRPCInterface is the notification plugin RPC interface
type NotifyPluginRPCInterface interface {
	BasePluginRPC
	Send(message json.RawMessage, opts json.RawMessage) error
	SendTemplate(template string, data json.RawMessage, opts json.RawMessage) error
	SendBatch(messages []json.RawMessage, opts json.RawMessage) error
}

// ApprovalPluginRPCInterface is the approval plugin RPC interface
type ApprovalPluginRPCInterface interface {
	BasePluginRPC
	CreateApproval(request json.RawMessage, opts json.RawMessage) (json.RawMessage, error)
	Approve(approvalId string, opts json.RawMessage) error
	Reject(approvalId string, reason string, opts json.RawMessage) error
	GetApprovalStatus(approvalId string, opts json.RawMessage) (json.RawMessage, error)
}

// StoragePluginRPCInterface is the storage plugin RPC interface
type StoragePluginRPCInterface interface {
	BasePluginRPC
	Save(key string, data json.RawMessage, opts json.RawMessage) error
	Load(key string, opts json.RawMessage) (json.RawMessage, error)
	Delete(key string, opts json.RawMessage) error
	List(prefix string, opts json.RawMessage) (json.RawMessage, error)
	Exists(key string, opts json.RawMessage) (bool, error)
}

// AnalyticsPluginRPCInterface is the analytics plugin RPC interface
type AnalyticsPluginRPCInterface interface {
	BasePluginRPC
	TrackEvent(event json.RawMessage, opts json.RawMessage) error
	Query(query json.RawMessage, opts json.RawMessage) (json.RawMessage, error)
	GetMetrics(metric string, opts json.RawMessage) (json.RawMessage, error)
	GenerateReport(reportType string, opts json.RawMessage) (json.RawMessage, error)
}

// IntegrationPluginRPCInterface is the integration plugin RPC interface
type IntegrationPluginRPCInterface interface {
	BasePluginRPC
	Connect(config json.RawMessage, opts json.RawMessage) error
	Disconnect(opts json.RawMessage) error
	Call(method string, params json.RawMessage, opts json.RawMessage) (json.RawMessage, error)
	Subscribe(event string, opts json.RawMessage) error
	Unsubscribe(event string, opts json.RawMessage) error
}

// CustomPluginRPCInterface is the custom plugin RPC interface
type CustomPluginRPCInterface interface {
	BasePluginRPC
	Execute(action string, params json.RawMessage, opts json.RawMessage) (json.RawMessage, error)
}
