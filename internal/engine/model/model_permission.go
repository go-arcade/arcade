package model

/**
 * @author: gagral.x@gmail.com
 * @time: 2025/10/13
 * @file: model_permission.go
 * @description: 权限点定义
 */

// PermissionPoint 权限点（用于组合成角色）
const (
	// ========== 项目权限 ==========
	PermProjectView        = "project.view"        // 查看项目
	PermProjectEdit        = "project.edit"        // 编辑项目设置
	PermProjectDelete      = "project.delete"      // 删除项目
	PermProjectTransfer    = "project.transfer"    // 转移项目所有权
	PermProjectArchive     = "project.archive"     // 归档项目
	PermProjectSettings    = "project.settings"    // 修改项目设置
	PermProjectVariables   = "project.variables"   // 管理项目变量
	PermProjectWebhook     = "project.webhook"     // 管理Webhook
	PermProjectIntegration = "project.integration" // 管理集成

	// ========== 构建权限 ==========
	PermBuildView     = "build.view"     // 查看构建
	PermBuildTrigger  = "build.trigger"  // 触发构建
	PermBuildCancel   = "build.cancel"   // 取消构建
	PermBuildRetry    = "build.retry"    // 重试构建
	PermBuildDelete   = "build.delete"   // 删除构建记录
	PermBuildArtifact = "build.artifact" // 下载构建产物
	PermBuildLog      = "build.log"      // 查看构建日志

	// ========== 流水线权限 ==========
	PermPipelineView   = "pipeline.view"   // 查看流水线
	PermPipelineCreate = "pipeline.create" // 创建流水线
	PermPipelineEdit   = "pipeline.edit"   // 编辑流水线
	PermPipelineDelete = "pipeline.delete" // 删除流水线
	PermPipelineRun    = "pipeline.run"    // 运行流水线
	PermPipelineStop   = "pipeline.stop"   // 停止流水线

	// ========== 部署权限 ==========
	PermDeployView     = "deploy.view"     // 查看部署
	PermDeployCreate   = "deploy.create"   // 创建部署
	PermDeployExecute  = "deploy.execute"  // 执行部署
	PermDeployRollback = "deploy.rollback" // 回滚部署
	PermDeployApprove  = "deploy.approve"  // 审批部署

	// ========== 成员权限 ==========
	PermMemberView   = "member.view"   // 查看成员
	PermMemberInvite = "member.invite" // 邀请成员
	PermMemberEdit   = "member.edit"   // 编辑成员角色
	PermMemberRemove = "member.remove" // 移除成员

	// ========== 团队权限 ==========
	PermTeamView     = "team.view"     // 查看团队
	PermTeamCreate   = "team.create"   // 创建团队
	PermTeamEdit     = "team.edit"     // 编辑团队
	PermTeamDelete   = "team.delete"   // 删除团队
	PermTeamMember   = "team.member"   // 管理团队成员
	PermTeamProject  = "team.project"  // 管理团队项目
	PermTeamSettings = "team.settings" // 修改团队设置

	// ========== Issue/任务权限 ==========
	PermIssueView   = "issue.view"   // 查看Issue
	PermIssueCreate = "issue.create" // 创建Issue
	PermIssueEdit   = "issue.edit"   // 编辑Issue
	PermIssueClose  = "issue.close"  // 关闭Issue
	PermIssueDelete = "issue.delete" // 删除Issue

	// ========== 监控权限 ==========
	PermMonitorView      = "monitor.view"      // 查看监控
	PermMonitorMetrics   = "monitor.metrics"   // 查看指标
	PermMonitorLogs      = "monitor.logs"      // 查看日志
	PermMonitorAlert     = "monitor.alert"     // 管理告警
	PermMonitorDashboard = "monitor.dashboard" // 管理仪表板

	// ========== 安全权限 ==========
	PermSecurityScan   = "security.scan"   // 安全扫描
	PermSecurityAudit  = "security.audit"  // 安全审计
	PermSecurityPolicy = "security.policy" // 安全策略管理
)

// 内置角色的权限集合
var BuiltinRolePermissions = map[string][]string{
	// ========== 项目角色权限 ==========
	BuiltinProjectOwner: {
		// 所有项目权限
		PermProjectView, PermProjectEdit, PermProjectDelete, PermProjectTransfer,
		PermProjectArchive, PermProjectSettings, PermProjectVariables, PermProjectWebhook, PermProjectIntegration,
		// 所有构建权限
		PermBuildView, PermBuildTrigger, PermBuildCancel, PermBuildRetry, PermBuildDelete, PermBuildArtifact, PermBuildLog,
		// 所有流水线权限
		PermPipelineView, PermPipelineCreate, PermPipelineEdit, PermPipelineDelete, PermPipelineRun, PermPipelineStop,
		// 所有部署权限
		PermDeployView, PermDeployCreate, PermDeployExecute, PermDeployRollback, PermDeployApprove,
		// 所有成员权限
		PermMemberView, PermMemberInvite, PermMemberEdit, PermMemberRemove,
		// 所有团队权限
		PermTeamView, PermTeamProject, PermTeamSettings,
		// 其他权限
		PermIssueView, PermIssueCreate, PermIssueEdit, PermIssueClose, PermIssueDelete,
		PermMonitorView, PermMonitorMetrics, PermMonitorLogs, PermMonitorAlert, PermMonitorDashboard,
		PermSecurityScan, PermSecurityAudit, PermSecurityPolicy,
	},

	BuiltinProjectMaintainer: {
		// 项目权限（不能删除、转移）
		PermProjectView, PermProjectEdit, PermProjectArchive, PermProjectSettings, PermProjectVariables, PermProjectWebhook, PermProjectIntegration,
		// 构建权限
		PermBuildView, PermBuildTrigger, PermBuildCancel, PermBuildRetry, PermBuildDelete, PermBuildArtifact, PermBuildLog,
		// 流水线权限
		PermPipelineView, PermPipelineCreate, PermPipelineEdit, PermPipelineDelete, PermPipelineRun, PermPipelineStop,
		// 部署权限
		PermDeployView, PermDeployCreate, PermDeployExecute, PermDeployRollback, PermDeployApprove,
		// 成员权限
		PermMemberView, PermMemberInvite, PermMemberEdit, PermMemberRemove,
		// 团队权限
		PermTeamView, PermTeamProject, PermTeamSettings,
		// 其他权限
		PermIssueView, PermIssueCreate, PermIssueEdit, PermIssueClose, PermIssueDelete,
		PermMonitorView, PermMonitorMetrics, PermMonitorLogs, PermMonitorAlert, PermMonitorDashboard,
		PermSecurityScan, PermSecurityAudit,
	},

	BuiltinProjectDeveloper: {
		// 基本项目权限
		PermProjectView,
		// 构建权限
		PermBuildView, PermBuildTrigger, PermBuildCancel, PermBuildRetry, PermBuildArtifact, PermBuildLog,
		// 流水线权限（不能删除）
		PermPipelineView, PermPipelineCreate, PermPipelineEdit, PermPipelineRun, PermPipelineStop,
		// 部署权限（不能批准）
		PermDeployView, PermDeployCreate, PermDeployExecute,
		// 成员权限（只读）
		PermMemberView,
		// 团队权限（只读）
		PermTeamView,
		// 其他权限
		PermIssueView, PermIssueCreate, PermIssueEdit, PermIssueClose,
		PermMonitorView, PermMonitorMetrics, PermMonitorLogs,
		PermSecurityScan,
	},

	BuiltinProjectReporter: {
		// 基本查看权限
		PermProjectView,
		PermBuildView, PermBuildArtifact, PermBuildLog,
		PermPipelineView,
		PermDeployView,
		PermMemberView,
		PermTeamView,
		// Issue 权限
		PermIssueView, PermIssueCreate, PermIssueEdit,
		PermMonitorView, PermMonitorMetrics, PermMonitorLogs,
	},

	BuiltinProjectGuest: {
		// 仅查看权限
		PermProjectView,
		PermBuildView, PermBuildLog,
		PermPipelineView,
		PermDeployView,
		PermMemberView,
		PermIssueView,
		PermMonitorView,
	},

	// ========== 团队角色权限 ==========
	BuiltinTeamOwner: {
		PermTeamView, PermTeamEdit, PermTeamDelete, PermTeamMember, PermTeamProject, PermTeamSettings,
	},

	BuiltinTeamMaintainer: {
		PermTeamView, PermTeamEdit, PermTeamMember, PermTeamProject, PermTeamSettings,
	},

	BuiltinTeamDeveloper: {
		PermTeamView, PermTeamProject,
	},

	BuiltinTeamReporter: {
		PermTeamView,
	},

	BuiltinTeamGuest: {
		PermTeamView,
	},

	// ========== 组织角色权限 ==========
	BuiltinOrgOwner: {
		// 拥有所有组织级权限
		PermTeamView, PermTeamCreate, PermTeamEdit, PermTeamDelete, PermTeamMember, PermTeamProject, PermTeamSettings,
	},

	BuiltinOrgAdmin: {
		PermTeamView, PermTeamCreate, PermTeamEdit, PermTeamMember, PermTeamProject, PermTeamSettings,
	},

	BuiltinOrgMember: {
		PermTeamView,
	},
}
