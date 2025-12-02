package model

import "time"

// UserRoleBinding 用户角色绑定（多层级统一管理）
type UserRoleBinding struct {
	Id         int       `json:"id" gorm:"column:id;primary_key;auto_increment"`
	BindingId  string    `json:"binding_id" gorm:"column:binding_id;unique;not null"`
	UserId     string    `json:"user_id" gorm:"column:user_id;not null;index"`
	RoleId     string    `json:"role_id" gorm:"column:role_id;not null;index"`
	Scope      string    `json:"scope" gorm:"column:scope;not null;index"` // platform/organization/team/project
	ResourceId *string   `json:"resource_id" gorm:"column:resource_id;index"`
	GrantedBy  *string   `json:"granted_by" gorm:"column:granted_by"`
	CreateTime time.Time `json:"create_time" gorm:"column:create_time;autoCreateTime"`
	UpdateTime time.Time `json:"update_time" gorm:"column:update_time;autoUpdateTime"`
}

func (UserRoleBinding) TableName() string {
	return "t_user_role_binding"
}

// UserPermissions 用户权限聚合结果
type UserPermissions struct {
	UserId              string               `json:"userId"`
	IsSuperAdmin        bool                 `json:"isSuperAdmin"`
	PlatformPermissions []string             `json:"platformPermissions"`
	OrgPermissions      map[string][]string  `json:"orgPermissions"`      // org_id -> permissions
	TeamPermissions     map[string][]string  `json:"teamPermissions"`     // team_id -> permissions
	ProjectPermissions  map[string][]string  `json:"projectPermissions"`  // project_id -> permissions
	AllPermissions      []string             `json:"allPermissions"`      // 所有权限的并集
	AccessibleRoutes    []*RouterPermission  `json:"accessibleRoutes"`    // 可访问的路由
	AccessibleResources *AccessibleResources `json:"accessibleResources"` // 可访问的资源
}

// AccessibleResources 用户可访问的资源列表
type AccessibleResources struct {
	Organizations []string `json:"organizations"` // org_id list
	Teams         []string `json:"teams"`         // team_id list
	Projects      []string `json:"projects"`      // project_id list
}

// PermissionCheckRequest 权限检查请求
type PermissionCheckRequest struct {
	UserId         string `json:"userId" binding:"required"`
	PermissionCode string `json:"permissionCode" binding:"required"`
	ResourceId     string `json:"resourceId"`
	Scope          string `json:"scope"` // platform/organization/team/project
}

// PermissionCheckResponse 权限检查响应
type PermissionCheckResponse struct {
	HasPermission bool   `json:"hasPermission"`
	Reason        string `json:"reason,omitempty"`
}

// UserRoleBindingCreate 创建用户角色绑定请求
type UserRoleBindingCreate struct {
	UserId     string  `json:"userId" binding:"required"`
	RoleId     string  `json:"roleId" binding:"required"`
	Scope      string  `json:"scope" binding:"required,oneof=platform organization team project"`
	ResourceId *string `json:"resourceId"`
	GrantedBy  *string `json:"grantedBy"`
}

// UserRoleBindingUpdate 更新用户角色绑定请求
type UserRoleBindingUpdate struct {
	RoleId string `json:"roleId" binding:"required"`
}

// RoleBindingQuery 角色绑定查询条件
type RoleBindingQuery struct {
	UserId     string
	RoleId     string
	Scope      string
	ResourceId string
	Page       int
	PageSize   int
}

// PermissionScope 权限作用域常量
const (
	ScopePlatform     = "platform"
	ScopeOrganization = "organization"
	ScopeTeam         = "team"
	ScopeProject      = "project"
)

// RouteType 路由类型常量
const (
	RouteTypeMenu   = "menu"
	RouteTypePage   = "page"
	RouteTypeButton = "button"
)

// 权限常量定义
const (
	//
	PermProjectView      = "project.view"
	PermProjectEdit      = "project.edit"
	PermProjectDelete    = "project.delete"
	PermProjectSettings  = "project.settings"
	PermProjectVariables = "project.variables"

	// 构建权限
	PermBuildView     = "build.view"
	PermBuildTrigger  = "build.trigger"
	PermBuildCancel   = "build.cancel"
	PermBuildLog      = "build.log"
	PermBuildArtifact = "build.artifact"

	// 流水线权限
	PermPipelineView   = "pipeline.view"
	PermPipelineCreate = "pipeline.create"
	PermPipelineEdit   = "pipeline.edit"
	PermPipelineDelete = "pipeline.delete"
	PermPipelineRun    = "pipeline.run"
	PermPipelineCancel = "pipeline.cancel"

	// 部署权限
	PermDeployView     = "deploy.view"
	PermDeployRun      = "deploy.run"
	PermDeployRollback = "deploy.rollback"
	PermDeployApprove  = "deploy.approve"

	// 成员权限
	PermMemberView   = "member.view"
	PermMemberInvite = "member.invite"
	PermMemberManage = "member.manage"

	// 团队权限
	PermTeamView   = "team.view"
	PermTeamCreate = "team.create"
	PermTeamEdit   = "team.edit"
	PermTeamDelete = "team.delete"

	// 组织权限
	PermOrganizationView   = "organization.view"
	PermOrganizationEdit   = "organization.edit"
	PermOrganizationDelete = "organization.delete"
)
