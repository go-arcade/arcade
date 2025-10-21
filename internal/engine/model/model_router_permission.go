package model

import "gorm.io/datatypes"

/**
 * @author: gagral.x@gmail.com
 * @time: 2025/01/14
 * @file: model_router_permission.go
 * @description: 路由权限映射模型
 */

// RouterPermission 路由权限映射表
type RouterPermission struct {
	BaseModel
	RouteId             string         `gorm:"column:route_id;uniqueIndex" json:"routeId"`
	Path                string         `gorm:"column:path" json:"path"`                                          // 路由路径
	Method              string         `gorm:"column:method" json:"method"`                                      // HTTP方法
	Name                string         `gorm:"column:name" json:"name"`                                          // 路由名称
	Group               string         `gorm:"column:group" json:"group"`                                        // 路由分组（如：project, org, team）
	Category            string         `gorm:"column:category" json:"category"`                                  // 路由分类（如：管理, 开发, 监控）
	RequiredPermissions datatypes.JSON `gorm:"column:required_permissions;type:json" json:"requiredPermissions"` // 所需权限列表
	Icon                string         `gorm:"column:icon" json:"icon"`                                          // 图标
	Order               int            `gorm:"column:order" json:"order"`                                        // 排序
	IsMenu              int            `gorm:"column:is_menu" json:"isMenu"`                                     // 是否显示在菜单 0:否 1:是
	IsEnabled           int            `gorm:"column:is_enabled" json:"isEnabled"`                               // 是否启用 0:否 1:是
	Description         string         `gorm:"column:description" json:"description"`                            // 描述
}

func (RouterPermission) TableName() string {
	return "t_router_permission"
}

// RoutePermissionDTO 路由权限DTO
type RoutePermissionDTO struct {
	RouteId             string   `json:"routeId"`
	Path                string   `json:"path"`
	Method              string   `json:"method"`
	Name                string   `json:"name"`
	Group               string   `json:"group"`
	Category            string   `json:"category"`
	RequiredPermissions []string `json:"requiredPermissions"`
	Icon                string   `json:"icon"`
	Order               int      `json:"order"`
	IsMenu              bool     `json:"isMenu"`
	Description         string   `json:"description"`
}

// 预定义的路由配置
var BuiltinRoutes = []RoutePermissionDTO{
	// ========== 项目管理路由 ==========
	{
		RouteId:             "route_project_list",
		Path:                "/api/v1/projects",
		Method:              "GET",
		Name:                "项目列表",
		Group:               "project",
		Category:            "项目管理",
		RequiredPermissions: []string{PermProjectView},
		Icon:                "project",
		Order:               100,
		IsMenu:              true,
	},
	{
		RouteId:             "route_project_detail",
		Path:                "/api/v1/projects/:projectId",
		Method:              "GET",
		Name:                "项目详情",
		Group:               "project",
		Category:            "项目管理",
		RequiredPermissions: []string{PermProjectView},
		Icon:                "",
		Order:               101,
		IsMenu:              false,
	},
	{
		RouteId:             "route_project_create",
		Path:                "/api/v1/projects",
		Method:              "POST",
		Name:                "创建项目",
		Group:               "project",
		Category:            "项目管理",
		RequiredPermissions: []string{}, // 任何登录用户都可以创建项目
		Icon:                "",
		Order:               102,
		IsMenu:              false,
	},
	{
		RouteId:             "route_project_update",
		Path:                "/api/v1/projects/:projectId",
		Method:              "PUT",
		Name:                "更新项目",
		Group:               "project",
		Category:            "项目管理",
		RequiredPermissions: []string{PermProjectEdit},
		Icon:                "",
		Order:               103,
		IsMenu:              false,
	},
	{
		RouteId:             "route_project_delete",
		Path:                "/api/v1/projects/:projectId",
		Method:              "DELETE",
		Name:                "删除项目",
		Group:               "project",
		Category:            "项目管理",
		RequiredPermissions: []string{PermProjectDelete},
		Icon:                "",
		Order:               104,
		IsMenu:              false,
	},
	{
		RouteId:             "route_project_settings",
		Path:                "/api/v1/projects/:projectId/settings",
		Method:              "GET",
		Name:                "项目设置",
		Group:               "project",
		Category:            "项目管理",
		RequiredPermissions: []string{PermProjectSettings},
		Icon:                "settings",
		Order:               105,
		IsMenu:              true,
	},

	// ========== 流水线路由 ==========
	{
		RouteId:             "route_pipeline_list",
		Path:                "/api/v1/projects/:projectId/pipelines",
		Method:              "GET",
		Name:                "流水线列表",
		Group:               "pipeline",
		Category:            "CI/CD",
		RequiredPermissions: []string{PermPipelineView},
		Icon:                "pipeline",
		Order:               200,
		IsMenu:              true,
	},
	{
		RouteId:             "route_pipeline_create",
		Path:                "/api/v1/projects/:projectId/pipelines",
		Method:              "POST",
		Name:                "创建流水线",
		Group:               "pipeline",
		Category:            "CI/CD",
		RequiredPermissions: []string{PermPipelineCreate},
		Icon:                "",
		Order:               201,
		IsMenu:              false,
	},
	{
		RouteId:             "route_pipeline_run",
		Path:                "/api/v1/projects/:projectId/pipelines/:pipelineId/run",
		Method:              "POST",
		Name:                "运行流水线",
		Group:               "pipeline",
		Category:            "CI/CD",
		RequiredPermissions: []string{PermPipelineRun},
		Icon:                "",
		Order:               202,
		IsMenu:              false,
	},

	// ========== 构建路由 ==========
	{
		RouteId:             "route_build_list",
		Path:                "/api/v1/projects/:projectId/builds",
		Method:              "GET",
		Name:                "构建列表",
		Group:               "build",
		Category:            "CI/CD",
		RequiredPermissions: []string{PermBuildView},
		Icon:                "build",
		Order:               300,
		IsMenu:              true,
	},
	{
		RouteId:             "route_build_trigger",
		Path:                "/api/v1/projects/:projectId/builds/trigger",
		Method:              "POST",
		Name:                "触发构建",
		Group:               "build",
		Category:            "CI/CD",
		RequiredPermissions: []string{PermBuildTrigger},
		Icon:                "",
		Order:               301,
		IsMenu:              false,
	},
	{
		RouteId:             "route_build_log",
		Path:                "/api/v1/projects/:projectId/builds/:buildId/logs",
		Method:              "GET",
		Name:                "构建日志",
		Group:               "build",
		Category:            "CI/CD",
		RequiredPermissions: []string{PermBuildLog},
		Icon:                "",
		Order:               302,
		IsMenu:              false,
	},

	// ========== 部署路由 ==========
	{
		RouteId:             "route_deploy_list",
		Path:                "/api/v1/projects/:projectId/deploys",
		Method:              "GET",
		Name:                "部署列表",
		Group:               "deploy",
		Category:            "部署",
		RequiredPermissions: []string{PermDeployView},
		Icon:                "deploy",
		Order:               400,
		IsMenu:              true,
	},
	{
		RouteId:             "route_deploy_execute",
		Path:                "/api/v1/projects/:projectId/deploys/:deployId/execute",
		Method:              "POST",
		Name:                "执行部署",
		Group:               "deploy",
		Category:            "部署",
		RequiredPermissions: []string{PermDeployRun},
		Icon:                "",
		Order:               401,
		IsMenu:              false,
	},

	// ========== 成员管理路由 ==========
	{
		RouteId:             "route_member_list",
		Path:                "/api/v1/projects/:projectId/members",
		Method:              "GET",
		Name:                "成员列表",
		Group:               "member",
		Category:            "成员管理",
		RequiredPermissions: []string{PermMemberView},
		Icon:                "team",
		Order:               500,
		IsMenu:              true,
	},
	{
		RouteId:             "route_member_invite",
		Path:                "/api/v1/projects/:projectId/members",
		Method:              "POST",
		Name:                "邀请成员",
		Group:               "member",
		Category:            "成员管理",
		RequiredPermissions: []string{PermMemberInvite},
		Icon:                "",
		Order:               501,
		IsMenu:              false,
	},

	// ========== 组织路由 ==========
	{
		RouteId:             "route_org_list",
		Path:                "/api/v1/orgs",
		Method:              "GET",
		Name:                "组织列表",
		Group:               "org",
		Category:            "组织管理",
		RequiredPermissions: []string{}, // 任何登录用户
		Icon:                "organization",
		Order:               700,
		IsMenu:              true,
	},
	{
		RouteId:             "route_org_detail",
		Path:                "/api/v1/orgs/:orgId",
		Method:              "GET",
		Name:                "组织详情",
		Group:               "org",
		Category:            "组织管理",
		RequiredPermissions: []string{}, // 组织成员即可
		Icon:                "",
		Order:               701,
		IsMenu:              false,
	},

	// ========== 团队路由 ==========
	{
		RouteId:             "route_team_list",
		Path:                "/api/v1/orgs/:orgId/teams",
		Method:              "GET",
		Name:                "团队列表",
		Group:               "team",
		Category:            "团队管理",
		RequiredPermissions: []string{PermTeamView},
		Icon:                "team",
		Order:               800,
		IsMenu:              true,
	},
	{
		RouteId:             "route_team_create",
		Path:                "/api/v1/orgs/:orgId/teams",
		Method:              "POST",
		Name:                "创建团队",
		Group:               "team",
		Category:            "团队管理",
		RequiredPermissions: []string{PermTeamCreate},
		Icon:                "",
		Order:               801,
		IsMenu:              false,
	},
}
