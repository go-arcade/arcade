// Package plugin RPC call argument definitions
package plugin

import "encoding/json"

// RPC argument structures
// Uses json.RawMessage for type flexibility, allowing plugins to define custom parameter formats

// ========== Source Plugin Arguments ==========

// SourceCloneArgs contains arguments for cloning repository
type SourceCloneArgs struct {
	RepoURL string          `json:"repo_url"`
	Opts    json.RawMessage `json:"opts"`
}

// SourcePullArgs contains arguments for pulling repository
type SourcePullArgs struct {
	Opts json.RawMessage `json:"opts"`
}

// SourceCheckoutArgs contains arguments for checking out branch
type SourceCheckoutArgs struct {
	Branch string          `json:"branch"`
	Opts   json.RawMessage `json:"opts"`
}

// SourceGetCommitArgs contains arguments for getting commit info
type SourceGetCommitArgs struct {
	Opts json.RawMessage `json:"opts"`
}

// SourceGetDiffArgs contains arguments for getting diff
type SourceGetDiffArgs struct {
	Opts json.RawMessage `json:"opts"`
}

// ========== Build Plugin Arguments ==========

// BuildArgs contains arguments for build operations
type BuildArgs struct {
	ProjectConfig json.RawMessage `json:"project_config"`
	Opts          json.RawMessage `json:"opts"`
}

// BuildGetArtifactsArgs contains arguments for getting build artifacts
type BuildGetArtifactsArgs struct {
	Opts json.RawMessage `json:"opts"`
}

// BuildCleanArgs contains arguments for cleaning build
type BuildCleanArgs struct {
	Opts json.RawMessage `json:"opts"`
}

// ========== Test Plugin Arguments ==========

// TestArgs contains arguments for test operations
type TestArgs struct {
	ProjectConfig json.RawMessage `json:"project_config"`
	Opts          json.RawMessage `json:"opts"`
}

// TestGetResultsArgs contains arguments for getting test results
type TestGetResultsArgs struct {
	Opts json.RawMessage `json:"opts"`
}

// TestGetCoverageArgs contains arguments for getting test coverage
type TestGetCoverageArgs struct {
	Opts json.RawMessage `json:"opts"`
}

// ========== Deploy Plugin Arguments ==========

// DeployArgs contains arguments for deploy operations
type DeployArgs struct {
	ProjectConfig json.RawMessage `json:"project_config"`
	Opts          json.RawMessage `json:"opts"`
}

// DeployRollbackArgs contains arguments for rollback operations
type DeployRollbackArgs struct {
	Version string          `json:"version"`
	Opts    json.RawMessage `json:"opts"`
}

// DeployGetStatusArgs contains arguments for getting deployment status
type DeployGetStatusArgs struct {
	Opts json.RawMessage `json:"opts"`
}

// DeployScaleArgs contains arguments for scaling deployment
type DeployScaleArgs struct {
	Replicas int             `json:"replicas"`
	Opts     json.RawMessage `json:"opts"`
}

// ========== Security Plugin Arguments ==========

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

// SecurityGetVulnerabilitiesArgs contains arguments for getting vulnerabilities
type SecurityGetVulnerabilitiesArgs struct {
	Opts json.RawMessage `json:"opts"`
}

// SecurityCheckComplianceArgs contains arguments for checking compliance
type SecurityCheckComplianceArgs struct {
	Opts json.RawMessage `json:"opts"`
}

// ========== Notify Plugin Arguments ==========

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

// NotifySendBatchArgs contains arguments for sending batch notifications
type NotifySendBatchArgs struct {
	Messages []json.RawMessage `json:"messages"`
	Opts     json.RawMessage   `json:"opts"`
}

// ========== Approval Plugin Arguments ==========

// ApprovalCreateArgs contains arguments for creating approval request
type ApprovalCreateArgs struct {
	Request json.RawMessage `json:"request"`
	Opts    json.RawMessage `json:"opts"`
}

// ApprovalApproveArgs contains arguments for approving request
type ApprovalApproveArgs struct {
	ApprovalId string          `json:"approval_id"`
	Opts       json.RawMessage `json:"opts"`
}

// ApprovalRejectArgs contains arguments for rejecting request
type ApprovalRejectArgs struct {
	ApprovalId string          `json:"approval_id"`
	Reason     string          `json:"reason"`
	Opts       json.RawMessage `json:"opts"`
}

// ApprovalGetStatusArgs contains arguments for getting approval status
type ApprovalGetStatusArgs struct {
	ApprovalId string          `json:"approval_id"`
	Opts       json.RawMessage `json:"opts"`
}

// ========== Storage Plugin Arguments ==========

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

// StorageListArgs contains arguments for storage list operations
type StorageListArgs struct {
	Prefix string          `json:"prefix"`
	Opts   json.RawMessage `json:"opts"`
}

// StorageExistsArgs contains arguments for storage exists operations
type StorageExistsArgs struct {
	Key  string          `json:"key"`
	Opts json.RawMessage `json:"opts"`
}

// ========== Analytics Plugin Arguments ==========

// AnalyticsTrackEventArgs contains arguments for tracking events
type AnalyticsTrackEventArgs struct {
	Event json.RawMessage `json:"event"`
	Opts  json.RawMessage `json:"opts"`
}

// AnalyticsQueryArgs contains arguments for querying data
type AnalyticsQueryArgs struct {
	Query json.RawMessage `json:"query"`
	Opts  json.RawMessage `json:"opts"`
}

// AnalyticsGetMetricsArgs contains arguments for getting metrics
type AnalyticsGetMetricsArgs struct {
	Metric string          `json:"metric"`
	Opts   json.RawMessage `json:"opts"`
}

// AnalyticsGenerateReportArgs contains arguments for generating reports
type AnalyticsGenerateReportArgs struct {
	ReportType string          `json:"report_type"`
	Opts       json.RawMessage `json:"opts"`
}

// ========== Integration Plugin Arguments ==========

// IntegrationConnectArgs contains arguments for connecting to service
type IntegrationConnectArgs struct {
	Config json.RawMessage `json:"config"`
	Opts   json.RawMessage `json:"opts"`
}

// IntegrationDisconnectArgs contains arguments for disconnecting from service
type IntegrationDisconnectArgs struct {
	Opts json.RawMessage `json:"opts"`
}

// IntegrationCallArgs contains arguments for calling service method
type IntegrationCallArgs struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
	Opts   json.RawMessage `json:"opts"`
}

// IntegrationSubscribeArgs contains arguments for subscribing to events
type IntegrationSubscribeArgs struct {
	Event string          `json:"event"`
	Opts  json.RawMessage `json:"opts"`
}

// IntegrationUnsubscribeArgs contains arguments for unsubscribing from events
type IntegrationUnsubscribeArgs struct {
	Event string          `json:"event"`
	Opts  json.RawMessage `json:"opts"`
}

// ========== Custom Plugin Arguments ==========

// CustomExecuteArgs contains arguments for custom execute operations
type CustomExecuteArgs struct {
	Action string          `json:"action"`
	Params json.RawMessage `json:"params"`
	Opts   json.RawMessage `json:"opts"`
}
