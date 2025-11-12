package role

import (
	"github.com/go-arcade/arcade/internal/engine/model"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2025/10/13
 * @file: model_role.go
 * @description: 角色表模型（支持自定义角色）
 */

// RoleScope 角色作用域
type RoleScope string

const (
	RoleScopeProject RoleScope = "project" // 项目级角色
	RoleScopeTeam    RoleScope = "team"    // 团队级角色
	RoleScopeOrg     RoleScope = "org"     // 组织级角色
)

// Role 角色表（支持自定义角色）
type Role struct {
	model.BaseModel
	RoleId      string    `gorm:"column:role_id;not null;uniqueIndex" json:"roleId"`
	Name        string    `gorm:"column:name;not null" json:"name"`                      // 角色名称
	DisplayName string    `gorm:"column:display_name" json:"displayName"`                // 显示名称
	Description string    `gorm:"column:description" json:"description"`                 // 角色描述
	Scope       RoleScope `gorm:"column:scope;not null;type:varchar(32)" json:"scope"`   // 作用域: project/team/org
	OrgId       string    `gorm:"column:org_id;index" json:"orgId"`                      // 所属组织ID（全局角色为空）
	IsBuiltin   int       `gorm:"column:is_builtin;not null;default:0" json:"isBuiltin"` // 0: custom, 1: built-in
	IsEnabled   int       `gorm:"column:is_enabled;not null;default:1" json:"isEnabled"` // 0: disabled, 1: enabled
	Priority    int       `gorm:"column:priority;not null;default:0" json:"priority"`    // 优先级（数值越大权限越高）
	Permissions string    `gorm:"column:permissions;type:text" json:"permissions"`       // 权限点列表（JSON数组，如：["project.read","project.write"]）
	CreatedBy   string    `gorm:"column:created_by" json:"createdBy"`                    // 创建者
}

func (r *Role) TableName() string {
	return "t_role"
}

// 内置项目角色 ID
const (
	BuiltinProjectOwner      = "project_owner"      // 项目所有者
	BuiltinProjectMaintainer = "project_maintainer" // 项目维护者
	BuiltinProjectDeveloper  = "project_developer"  // 项目开发者
	BuiltinProjectReporter   = "project_reporter"   // 项目报告者
	BuiltinProjectGuest      = "project_guest"      // 项目访客
)

// 内置团队角色 ID
const (
	BuiltinTeamOwner      = "team_owner"      // 团队所有者
	BuiltinTeamMaintainer = "team_maintainer" // 团队维护者
	BuiltinTeamDeveloper  = "team_developer"  // 团队开发者
	BuiltinTeamReporter   = "team_reporter"   // 团队报告者
	BuiltinTeamGuest      = "team_guest"      // 团队访客
)

// 内置组织角色 ID
const (
	BuiltinOrgOwner  = "org_owner"  // 组织所有者
	BuiltinOrgAdmin  = "org_admin"  // 组织管理员
	BuiltinOrgMember = "org_member" // 组织成员
)

// RoleIsBuiltin role built-in status
const (
	RoleCustom  = 0 // custom role
	RoleBuiltin = 1 // built-in role
)

// RoleEnabled role enabled status
const (
	RoleDisabled = 0 // disabled
	RoleEnabled  = 1 // enabled
)

// CreateRoleRequest request for creating a role
type CreateRoleRequest struct {
	Name        string    `json:"name" binding:"required"`
	DisplayName string    `json:"displayName"`
	Description string    `json:"description"`
	Scope       RoleScope `json:"scope" binding:"required"` // project/team/org
	OrgId       string    `json:"orgId"`                    // organization ID (empty for global roles)
	Priority    int       `json:"priority"`                 // priority
	Permissions []string  `json:"permissions"`              // permission list
	CreatedBy   string    `json:"createdBy"`                // creator
}

// UpdateRoleRequest request for updating a role
type UpdateRoleRequest struct {
	DisplayName string   `json:"displayName"`
	Description string   `json:"description"`
	Priority    int      `json:"priority"`
	Permissions []string `json:"permissions"`
}

// ListRolesRequest request for listing roles
type ListRolesRequest struct {
	PageNum   int       `json:"pageNum"`
	PageSize  int       `json:"pageSize"`
	Scope     RoleScope `json:"scope"`     // filter by scope
	OrgId     string    `json:"orgId"`     // filter by org
	Name      string    `json:"name"`      // fuzzy search by name
	IsBuiltin *int      `json:"isBuiltin"` // filter by builtin status (0: custom, 1: built-in)
	IsEnabled *int      `json:"isEnabled"` // filter by enabled status (0: disabled, 1: enabled)
}

// ListRolesResponse response for listing roles
type ListRolesResponse struct {
	Roles    []Role `json:"roles"`
	Total    int64  `json:"total"`
	PageNum  int    `json:"pageNum"`
	PageSize int    `json:"pageSize"`
}

// RoleResponse response for role without timestamps
type RoleResponse struct {
	RoleId      string    `json:"roleId"`
	Name        string    `json:"name"`
	DisplayName string    `json:"displayName"`
	Description string    `json:"description"`
	Scope       RoleScope `json:"scope"`
	OrgId       string    `json:"orgId"`
	IsBuiltin   int       `json:"isBuiltin"`
	IsEnabled   int       `json:"isEnabled"`
	Priority    int       `json:"priority"`
	Permissions []string  `json:"permissions"` // parsed from JSON string
	CreatedBy   string    `json:"createdBy"`
}

// BuiltinRolePermissions 内置角色的默认权限映射
var BuiltinRolePermissions = map[string][]string{
	// 项目角色权限
	BuiltinProjectOwner: {
		"project.view", "project.edit", "project.delete", "project.settings",
		"project.member.view", "project.member.manage",
		"pipeline.view", "pipeline.create", "pipeline.edit", "pipeline.delete", "pipeline.run", "pipeline.cancel",
		"build.view", "build.trigger", "build.cancel", "build.log",
		"deploy.view", "deploy.run", "deploy.rollback", "deploy.approve",
	},
	BuiltinProjectMaintainer: {
		"project.view", "project.edit", "project.settings",
		"project.member.view",
		"pipeline.view", "pipeline.create", "pipeline.edit", "pipeline.run", "pipeline.cancel",
		"build.view", "build.trigger", "build.cancel", "build.log",
		"deploy.view", "deploy.run", "deploy.rollback",
	},
	BuiltinProjectDeveloper: {
		"project.view",
		"pipeline.view", "pipeline.run",
		"build.view", "build.trigger", "build.log",
		"deploy.view",
	},
	BuiltinProjectReporter: {
		"project.view",
		"pipeline.view",
		"build.view", "build.log",
		"deploy.view",
	},
	BuiltinProjectGuest: {
		"project.view",
		"pipeline.view",
		"build.view",
	},

	// 团队角色权限
	BuiltinTeamOwner: {
		"team.view", "team.edit", "team.delete", "team.settings",
		"team.member.view", "team.member.manage",
		"project.view", "project.create", "project.edit", "project.delete",
	},
	BuiltinTeamMaintainer: {
		"team.view", "team.edit",
		"team.member.view",
		"project.view", "project.create", "project.edit",
	},
	BuiltinTeamDeveloper: {
		"team.view",
		"project.view", "project.create",
	},
	BuiltinTeamReporter: {
		"team.view",
		"project.view",
	},
	BuiltinTeamGuest: {
		"team.view",
		"project.view",
	},

	// 组织角色权限
	BuiltinOrgOwner: {
		"organization.view", "organization.edit", "organization.delete", "organization.settings",
		"organization.member.view", "organization.member.invite", "organization.member.manage",
		"team.view", "team.create", "team.edit", "team.delete",
		"project.view", "project.create", "project.edit", "project.delete",
	},
	BuiltinOrgAdmin: {
		"organization.view", "organization.edit",
		"organization.member.view", "organization.member.invite",
		"team.view", "team.create", "team.edit",
		"project.view", "project.create", "project.edit",
	},
	BuiltinOrgMember: {
		"organization.view",
		"team.view",
		"project.view",
	},
}
