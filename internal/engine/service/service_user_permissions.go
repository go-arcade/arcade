package service

import (
	"context"
	"fmt"
	"slices"
	"sync"

	"github.com/go-arcade/arcade/internal/engine/model"
	"github.com/go-arcade/arcade/internal/engine/model/entity"
	"github.com/go-arcade/arcade/internal/engine/repo"
	"github.com/go-arcade/arcade/pkg/ctx"
	"github.com/go-arcade/arcade/pkg/database"
	"github.com/go-arcade/arcade/pkg/log"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2025/01/14
 * @file: service_user_permissions.go
 * @description: 用户权限聚合服务 - 查询用户在所有层级的权限并返回可访问路由
 */

// UserPermissionsService 用户权限聚合服务
type UserPermissionsService struct {
	ctx        *ctx.Context
	db         database.DB
	permSvc    *PermissionService
	routerRepo repo.IRouterPermissionRepository
}

// NewUserPermissionsService 创建用户权限聚合服务
func NewUserPermissionsService(ctx *ctx.Context, db database.DB, permSvc *PermissionService, routerRepo repo.IRouterPermissionRepository) *UserPermissionsService {
	return &UserPermissionsService{
		ctx:        ctx,
		db:         db,
		permSvc:    permSvc,
		routerRepo: routerRepo,
	}
}

// UserPermissionSummary 用户权限汇总
type UserPermissionSummary struct {
	UserId              string                     `json:"userId"`
	Organizations       []OrganizationPermission   `json:"organizations"`       // 用户所属组织
	Teams               []TeamPermission           `json:"teams"`               // 用户所属团队
	Projects            []ProjectPermissionSummary `json:"projects"`            // 用户可访问的项目
	AllPermissions      []string                   `json:"allPermissions"`      // 所有权限点合集
	AccessibleRoutes    []AccessibleRoute          `json:"accessibleRoutes"`    // 可访问的路由
	AccessibleResources map[string][]string        `json:"accessibleResources"` // 可访问的资源
}

// OrganizationPermission 组织权限信息
type OrganizationPermission struct {
	OrgId       string   `json:"orgId"`
	OrgName     string   `json:"orgName"`
	RoleId      string   `json:"roleId"`
	RoleName    string   `json:"roleName"`
	Permissions []string `json:"permissions"`
	Status      string   `json:"status"` // active/pending
}

// TeamPermission 团队权限信息
type TeamPermission struct {
	TeamId      string   `json:"teamId"`
	TeamName    string   `json:"teamName"`
	OrgId       string   `json:"orgId"`
	RoleId      string   `json:"roleId"`
	RoleName    string   `json:"roleName"`
	Permissions []string `json:"permissions"`
}

// ProjectPermissionSummary 项目权限摘要
type ProjectPermissionSummary struct {
	ProjectId   string   `json:"projectId"`
	ProjectName string   `json:"projectName"`
	OrgId       string   `json:"orgId"`
	RoleId      string   `json:"roleId"`
	RoleName    string   `json:"roleName"`
	Source      string   `json:"source"` // direct/team/org
	Permissions []string `json:"permissions"`
	Priority    int      `json:"priority"`
}

// AccessibleRoute 可访问的路由
type AccessibleRoute struct {
	Path        string   `json:"path"`
	Method      string   `json:"method"`
	Name        string   `json:"name"`
	Group       string   `json:"group"`       // 路由分组
	Category    string   `json:"category"`    // 路由分类
	Permissions []string `json:"permissions"` // 需要的权限
	Icon        string   `json:"icon"`        // 图标
	Order       int      `json:"order"`       // 排序
	IsMenu      bool     `json:"isMenu"`      // 是否显示在菜单
}

// GetUserPermissions 获取用户的所有权限汇总
func (s *UserPermissionsService) GetUserPermissions(ctx context.Context, userId string) (*UserPermissionSummary, error) {
	log.Infof("[UserPermissions] getting permissions for user: %s", userId)

	summary := &UserPermissionSummary{
		UserId:              userId,
		Organizations:       []OrganizationPermission{},
		Teams:               []TeamPermission{},
		Projects:            []ProjectPermissionSummary{},
		AllPermissions:      []string{},
		AccessibleRoutes:    []AccessibleRoute{},
		AccessibleResources: make(map[string][]string),
	}

	// 使用 WaitGroup 并发查询
	var wg sync.WaitGroup
	var mu sync.Mutex
	errors := []error{}

	// 1. 查询组织权限
	wg.Add(1)
	go func() {
		defer wg.Done()
		orgs, err := s.getUserOrganizations(ctx, userId)
		if err != nil {
			mu.Lock()
			errors = append(errors, fmt.Errorf("get organizations: %w", err))
			mu.Unlock()
			return
		}
		mu.Lock()
		summary.Organizations = orgs
		mu.Unlock()
	}()

	// 2. 查询团队权限
	wg.Add(1)
	go func() {
		defer wg.Done()
		teams, err := s.getUserTeams(ctx, userId)
		if err != nil {
			mu.Lock()
			errors = append(errors, fmt.Errorf("get teams: %w", err))
			mu.Unlock()
			return
		}
		mu.Lock()
		summary.Teams = teams
		mu.Unlock()
	}()

	// 3. 查询项目权限
	wg.Add(1)
	go func() {
		defer wg.Done()
		projects, err := s.getUserProjects(ctx, userId)
		if err != nil {
			mu.Lock()
			errors = append(errors, fmt.Errorf("get projects: %w", err))
			mu.Unlock()
			return
		}
		mu.Lock()
		summary.Projects = projects
		mu.Unlock()
	}()

	wg.Wait()

	if len(errors) > 0 {
		log.Warnf("[UserPermissions] errors occurred: %v", errors)
	}

	// 4. 汇总所有权限点
	s.aggregatePermissions(summary)

	// 5. 根据权限计算可访问路由
	routes, err := s.calculateAccessibleRoutes(ctx, summary.AllPermissions)
	if err != nil {
		log.Warnf("[UserPermissions] failed to calculate accessible routes: %v", err)
	} else {
		summary.AccessibleRoutes = routes
	}

	// 6. 汇总可访问资源
	s.aggregateAccessibleResources(summary)

	log.Infof("[UserPermissions] user %s has %d orgs, %d teams, %d projects, %d permissions, %d routes",
		userId, len(summary.Organizations), len(summary.Teams), len(summary.Projects),
		len(summary.AllPermissions), len(summary.AccessibleRoutes))

	return summary, nil
}

// getUserOrganizations 获取用户所属组织
func (s *UserPermissionsService) getUserOrganizations(ctx context.Context, userId string) ([]OrganizationPermission, error) {
	var orgMembers []model.OrganizationMember
	err := s.db.DB().Where("user_id = ?", userId).Find(&orgMembers).Error
	if err != nil {
		return nil, err
	}

	result := make([]OrganizationPermission, 0, len(orgMembers))
	for _, om := range orgMembers {
		// 获取组织信息
		var org model.Organization
		if err := s.db.DB().Where("org_id = ?", om.OrgId).First(&org).Error; err != nil {
			log.Warnf("[UserPermissions] failed to get org %s: %v", om.OrgId, err)
			continue
		}

		// 获取角色信息
		role, _ := s.permSvc.GetRole(om.RoleId)
		roleName := om.RoleId
		if role != nil {
			roleName = role.Name
		}

		// 获取权限列表
		permissions := s.getPermissionsForRole(om.RoleId)

		// 转换状态
		status := "inactive"
		switch om.Status {
		case model.OrgMemberStatusActive:
			status = "active"
		case model.OrgMemberStatusPending:
			status = "pending"
		}

		result = append(result, OrganizationPermission{
			OrgId:       om.OrgId,
			OrgName:     org.Name,
			RoleId:      om.RoleId,
			RoleName:    roleName,
			Permissions: permissions,
			Status:      status,
		})
	}

	return result, nil
}

// getUserTeams 获取用户所属团队
func (s *UserPermissionsService) getUserTeams(ctx context.Context, userId string) ([]TeamPermission, error) {
	var teamMembers []model.TeamMember
	err := s.db.DB().Where("user_id = ?", userId).Find(&teamMembers).Error
	if err != nil {
		return nil, err
	}

	result := make([]TeamPermission, 0, len(teamMembers))
	for _, tm := range teamMembers {
		// 获取团队信息
		var team entity.Team
		if err := s.db.DB().Where("team_id = ?", tm.TeamId).First(&team).Error; err != nil {
			log.Warnf("[UserPermissions] failed to get team %s: %v", tm.TeamId, err)
			continue
		}

		// 获取角色信息
		role, _ := s.permSvc.GetRole(tm.RoleId)
		roleName := tm.RoleId
		if role != nil {
			roleName = role.Name
		}

		// 获取权限列表
		permissions := s.getPermissionsForRole(tm.RoleId)

		result = append(result, TeamPermission{
			TeamId:      tm.TeamId,
			TeamName:    team.Name,
			OrgId:       team.OrgId,
			RoleId:      tm.RoleId,
			RoleName:    roleName,
			Permissions: permissions,
		})
	}

	return result, nil
}

// getUserProjects 获取用户可访问的项目
func (s *UserPermissionsService) getUserProjects(ctx context.Context, userId string) ([]ProjectPermissionSummary, error) {
	// 1. 获取用户直接参与的项目
	var projectMembers []model.ProjectMember
	err := s.db.DB().Where("user_id = ?", userId).Find(&projectMembers).Error
	if err != nil {
		return nil, err
	}

	projectMap := make(map[string]*ProjectPermissionSummary)

	// 处理直接项目成员关系
	for _, pm := range projectMembers {
		var project model.Project
		if err := s.db.DB().Where("project_id = ?", pm.ProjectId).First(&project).Error; err != nil {
			continue
		}

		role, _ := s.permSvc.GetRole(pm.RoleId)
		roleName := pm.RoleId
		priority := 0
		if role != nil {
			roleName = role.Name
			priority = role.Priority
		}

		permissions := s.getPermissionsForRole(pm.RoleId)

		projectMap[pm.ProjectId] = &ProjectPermissionSummary{
			ProjectId:   pm.ProjectId,
			ProjectName: project.Name,
			OrgId:       project.OrgId,
			RoleId:      pm.RoleId,
			RoleName:    roleName,
			Source:      "direct",
			Permissions: permissions,
			Priority:    priority,
		}
	}

	// 2. 通过团队访问的项目
	var teamMembers []model.TeamMember
	s.db.DB().Where("user_id = ?", userId).Find(&teamMembers)

	for _, tm := range teamMembers {
		var teamAccesses []model.ProjectTeamAccess
		s.db.DB().Where("team_id = ?", tm.TeamId).Find(&teamAccesses)

		for _, ta := range teamAccesses {
			// 如果已经有直接权限，比较优先级
			if existing, exists := projectMap[ta.ProjectId]; exists {
				// 计算团队权限对应的项目角色
				teamRole := s.permSvc.mapTeamRoleToProjectRole(tm.RoleId, ta.AccessLevel)
				teamRoleObj, _ := s.permSvc.GetRole(teamRole)
				if teamRoleObj != nil && teamRoleObj.Priority > existing.Priority {
					// 团队权限更高，更新
					existing.RoleId = teamRole
					existing.RoleName = teamRoleObj.Name
					existing.Source = "team"
					existing.Permissions = s.getPermissionsForRole(teamRole)
					existing.Priority = teamRoleObj.Priority
				}
				continue
			}

			// 新的项目访问
			var project model.Project
			if err := s.db.DB().Where("project_id = ?", ta.ProjectId).First(&project).Error; err != nil {
				continue
			}

			teamRole := s.permSvc.mapTeamRoleToProjectRole(tm.RoleId, ta.AccessLevel)
			teamRoleObj, _ := s.permSvc.GetRole(teamRole)
			roleName := teamRole
			priority := 0
			if teamRoleObj != nil {
				roleName = teamRoleObj.Name
				priority = teamRoleObj.Priority
			}

			permissions := s.getPermissionsForRole(teamRole)

			projectMap[ta.ProjectId] = &ProjectPermissionSummary{
				ProjectId:   ta.ProjectId,
				ProjectName: project.Name,
				OrgId:       project.OrgId,
				RoleId:      teamRole,
				RoleName:    roleName,
				Source:      "team",
				Permissions: permissions,
				Priority:    priority,
			}
		}
	}

	// 3. 通过组织访问的项目（access_level = org）
	var orgMembers []model.OrganizationMember
	s.db.DB().Where("user_id = ? AND status = ?", userId, model.OrgMemberStatusActive).Find(&orgMembers)

	for _, om := range orgMembers {
		var orgProjects []model.Project
		s.db.DB().Where("org_id = ? AND access_level = ?", om.OrgId, model.AccessLevelOrg).Find(&orgProjects)

		for _, project := range orgProjects {
			if _, exists := projectMap[project.ProjectId]; exists {
				// 已经有更高优先级的权限
				continue
			}

			// 组织角色映射到项目角色
			var mappedRoleId string
			switch om.RoleId {
			case model.BuiltinOrgOwner:
				mappedRoleId = model.BuiltinProjectMaintainer
			case model.BuiltinOrgAdmin:
				mappedRoleId = model.BuiltinProjectDeveloper
			case model.BuiltinOrgMember:
				mappedRoleId = model.BuiltinProjectGuest
			default:
				mappedRoleId = model.BuiltinProjectGuest
			}

			role, _ := s.permSvc.GetRole(mappedRoleId)
			roleName := mappedRoleId
			priority := 0
			if role != nil {
				roleName = role.Name
				priority = role.Priority
			}

			permissions := s.getPermissionsForRole(mappedRoleId)

			projectMap[project.ProjectId] = &ProjectPermissionSummary{
				ProjectId:   project.ProjectId,
				ProjectName: project.Name,
				OrgId:       project.OrgId,
				RoleId:      mappedRoleId,
				RoleName:    roleName,
				Source:      "org",
				Permissions: permissions,
				Priority:    priority,
			}
		}
	}

	// 转换为数组
	result := make([]ProjectPermissionSummary, 0, len(projectMap))
	for _, p := range projectMap {
		result = append(result, *p)
	}

	return result, nil
}

// getPermissionsForRole 获取角色的权限列表
func (s *UserPermissionsService) getPermissionsForRole(roleId string) []string {
	// 从关联表查询角色权限
	var permissionCodes []string

	err := s.db.DB().Table("t_role_permission rp").
		Select("p.code").
		Joins("JOIN t_permission p ON rp.permission_id = p.permission_id").
		Where("rp.role_id = ? AND p.is_enabled = ?", roleId, 1).
		Pluck("code", &permissionCodes).Error

	if err != nil {
		log.Warnf("[UserPermissions] failed to get permissions for role %s: %v", roleId, err)
		return []string{}
	}

	return permissionCodes
}

// aggregatePermissions 汇总所有权限点
func (s *UserPermissionsService) aggregatePermissions(summary *UserPermissionSummary) {
	permSet := make(map[string]bool)

	// 从组织收集权限
	for _, org := range summary.Organizations {
		for _, perm := range org.Permissions {
			permSet[perm] = true
		}
	}

	// 从团队收集权限
	for _, team := range summary.Teams {
		for _, perm := range team.Permissions {
			permSet[perm] = true
		}
	}

	// 从项目收集权限
	for _, proj := range summary.Projects {
		for _, perm := range proj.Permissions {
			permSet[perm] = true
		}
	}

	// 转换为数组
	allPerms := make([]string, 0, len(permSet))
	for perm := range permSet {
		allPerms = append(allPerms, perm)
	}

	summary.AllPermissions = allPerms
}

// calculateAccessibleRoutes 计算可访问的路由
func (s *UserPermissionsService) calculateAccessibleRoutes(ctx context.Context, permissions []string) ([]AccessibleRoute, error) {
	// 获取所有路由权限配置
	routePerms, err := s.routerRepo.GetAllRoutePermissions()
	if err != nil {
		return nil, err
	}

	permSet := make(map[string]bool)
	for _, perm := range permissions {
		permSet[perm] = true
	}

	accessibleRoutes := []AccessibleRoute{}

	for _, rp := range routePerms {
		// 检查是否有足够权限访问此路由
		hasAccess := false

		// 如果路由不需要权限，默认可访问
		if len(rp.RequiredPermissions) == 0 {
			hasAccess = true
		} else {
			// 检查是否拥有所需的任意一个权限（OR逻辑）
			for _, reqPerm := range rp.RequiredPermissions {
				if permSet[reqPerm] {
					hasAccess = true
					break
				}
			}
		}

		if hasAccess {
			accessibleRoutes = append(accessibleRoutes, AccessibleRoute{
				Path:        rp.Path,
				Method:      rp.Method,
				Name:        rp.Name,
				Group:       rp.Group,
				Category:    rp.Category,
				Permissions: rp.RequiredPermissions,
				Icon:        rp.Icon,
				Order:       rp.Order,
				IsMenu:      rp.IsMenu,
			})
		}
	}

	return accessibleRoutes, nil
}

// aggregateAccessibleResources 汇总可访问的资源
func (s *UserPermissionsService) aggregateAccessibleResources(summary *UserPermissionSummary) {
	// 组织资源
	orgIds := make([]string, 0, len(summary.Organizations))
	for _, org := range summary.Organizations {
		orgIds = append(orgIds, org.OrgId)
	}
	summary.AccessibleResources["organizations"] = orgIds

	// 团队资源
	teamIds := make([]string, 0, len(summary.Teams))
	for _, team := range summary.Teams {
		teamIds = append(teamIds, team.TeamId)
	}
	summary.AccessibleResources["teams"] = teamIds

	// 项目资源
	projectIds := make([]string, 0, len(summary.Projects))
	for _, proj := range summary.Projects {
		projectIds = append(projectIds, proj.ProjectId)
	}
	summary.AccessibleResources["projects"] = projectIds
}

// HasPermission 检查用户是否拥有指定权限
func (s *UserPermissionsService) HasPermission(ctx context.Context, userId string, permission string) (bool, error) {
	summary, err := s.GetUserPermissions(ctx, userId)
	if err != nil {
		return false, err
	}

	if slices.Contains(summary.AllPermissions, permission) {
		return true, nil
	}

	return false, nil
}

// HasAnyPermission 检查用户是否拥有任意一个指定权限
func (s *UserPermissionsService) HasAnyPermission(ctx context.Context, userId string, permissions []string) (bool, error) {
	summary, err := s.GetUserPermissions(ctx, userId)
	if err != nil {
		return false, err
	}

	permSet := make(map[string]bool)
	for _, perm := range summary.AllPermissions {
		permSet[perm] = true
	}

	for _, requiredPerm := range permissions {
		if permSet[requiredPerm] {
			return true, nil
		}
	}

	return false, nil
}

// HasAllPermissions 检查用户是否拥有所有指定权限
func (s *UserPermissionsService) HasAllPermissions(ctx context.Context, userId string, permissions []string) (bool, error) {
	summary, err := s.GetUserPermissions(ctx, userId)
	if err != nil {
		return false, err
	}

	permSet := make(map[string]bool)
	for _, perm := range summary.AllPermissions {
		permSet[perm] = true
	}

	for _, requiredPerm := range permissions {
		if !permSet[requiredPerm] {
			return false, nil
		}
	}

	return true, nil
}
