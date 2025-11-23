package permission

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-arcade/arcade/internal/engine/model/permission"
	"github.com/go-arcade/arcade/internal/engine/model/project"
	"github.com/go-arcade/arcade/internal/engine/model/team"
	"github.com/go-arcade/arcade/internal/engine/model/user"
	"github.com/go-arcade/arcade/pkg/cache"
	"github.com/go-arcade/arcade/pkg/database"
	"gorm.io/datatypes"
)

type IPermissionRepository interface {
	CreateRoleBinding(binding *permission.UserRoleBinding) error
	DeleteRoleBinding(bindingId string) error
	GetRoleBinding(bindingId string) (*permission.UserRoleBinding, error)
	GetUserRoleBindings(userId string) ([]*permission.UserRoleBinding, error)
	GetUserRoleBindingsByScope(userId, scope string) ([]*permission.UserRoleBinding, error)
	GetUserRoleBindingByResource(userId, scope, resourceId string) (*permission.UserRoleBinding, error)
	ListRoleBindings(query *permission.RoleBindingQuery) ([]*permission.UserRoleBinding, int64, error)
	HasRoleBinding(userId, roleId, scope, resourceId string) (bool, error)
	GetUserPermissions(userId string) (*permission.UserPermissions, error)
	GetRolePermissionCodes(roleId string) ([]string, error)
	HasPermission(userId, permissionCode, resourceId, scope string) (bool, error)
	GetAccessibleResources(userId string) (*permission.AccessibleResources, error)
	GetAccessibleRoutes(userId string) ([]*permission.RouterPermission, error)
	BatchCreateRoleBindings(bindings []*permission.UserRoleBinding) error
	BatchDeleteRoleBindings(bindingIds []string) error
	DeleteUserRoleBindingsByResource(userId, scope, resourceId string) error
	DeleteRoleBindingsByResource(scope, resourceId string) error
	ClearUserPermissionsCache(userId string) error
	ClearRolePermissionsCache(roleId string) error
}

type PermissionRepo struct {
	db    database.IDatabase
	cache cache.ICache
}

func NewPermissionRepo(db database.IDatabase, cache cache.ICache) IPermissionRepository {
	return &PermissionRepo{
		db:    db,
		cache: cache,
	}
}

// ============ 用户角色绑定 ============

// CreateRoleBinding 创建用户角色绑定
func (r *PermissionRepo) CreateRoleBinding(binding *permission.UserRoleBinding) error {
	return r.db.Database().Create(binding).Error
}

// DeleteRoleBinding 删除用户角色绑定
func (r *PermissionRepo) DeleteRoleBinding(bindingId string) error {
	return r.db.Database().Where("binding_id = ?", bindingId).Delete(&permission.UserRoleBinding{}).Error
}

// GetRoleBinding 根据ID获取角色绑定
func (r *PermissionRepo) GetRoleBinding(bindingId string) (*permission.UserRoleBinding, error) {
	var binding permission.UserRoleBinding
	err := r.db.Database().Where("binding_id = ?", bindingId).First(&binding).Error
	if err != nil {
		return nil, err
	}
	return &binding, nil
}

// GetUserRoleBindings 获取用户的所有角色绑定
func (r *PermissionRepo) GetUserRoleBindings(userId string) ([]*permission.UserRoleBinding, error) {
	var bindings []*permission.UserRoleBinding
	err := r.db.Database().Where("user_id = ?", userId).Find(&bindings).Error
	return bindings, err
}

// GetUserRoleBindingsByScope 获取用户在指定作用域的角色绑定
func (r *PermissionRepo) GetUserRoleBindingsByScope(userId, scope string) ([]*permission.UserRoleBinding, error) {
	var bindings []*permission.UserRoleBinding
	err := r.db.Database().Where("user_id = ? AND scope = ?", userId, scope).Find(&bindings).Error
	return bindings, err
}

// GetUserRoleBindingByResource 获取用户在指定资源的角色绑定
func (r *PermissionRepo) GetUserRoleBindingByResource(userId, scope, resourceId string) (*permission.UserRoleBinding, error) {
	var binding permission.UserRoleBinding
	err := r.db.Database().Where("user_id = ? AND scope = ? AND resource_id = ?", userId, scope, resourceId).First(&binding).Error
	if err != nil {
		return nil, err
	}
	return &binding, nil
}

// ListRoleBindings 列出角色绑定（支持多条件查询）
func (r *PermissionRepo) ListRoleBindings(query *permission.RoleBindingQuery) ([]*permission.UserRoleBinding, int64, error) {
	var bindings []*permission.UserRoleBinding
	var total int64

	db := r.db.Database().Model(&permission.UserRoleBinding{})

	// 条件查询
	if query.UserId != "" {
		db = db.Where("user_id = ?", query.UserId)
	}
	if query.RoleId != "" {
		db = db.Where("role_id = ?", query.RoleId)
	}
	if query.Scope != "" {
		db = db.Where("scope = ?", query.Scope)
	}
	if query.ResourceId != "" {
		db = db.Where("resource_id = ?", query.ResourceId)
	}

	// 获取总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页
	if query.Page > 0 && query.PageSize > 0 {
		offset := (query.Page - 1) * query.PageSize
		db = db.Offset(offset).Limit(query.PageSize)
	}

	err := db.Find(&bindings).Error
	return bindings, total, err
}

// HasRoleBinding 检查用户是否有指定的角色绑定
func (r *PermissionRepo) HasRoleBinding(userId, roleId, scope, resourceId string) (bool, error) {
	var count int64
	query := r.db.Database().Model(&permission.UserRoleBinding{}).Where("user_id = ? AND role_id = ? AND scope = ?", userId, roleId, scope)

	if resourceId != "" {
		query = query.Where("resource_id = ?", resourceId)
	} else {
		query = query.Where("resource_id IS NULL")
	}

	err := query.Count(&count).Error
	return count > 0, err
}

// ============ 权限聚合 ============

// GetUserPermissions 获取用户完整权限（聚合所有层级）
func (r *PermissionRepo) GetUserPermissions(userId string) (*permission.UserPermissions, error) {
	userPerms := &permission.UserPermissions{
		UserId:              userId,
		IsSuperAdmin:        false,
		PlatformPermissions: []string{},
		OrgPermissions:      make(map[string][]string),
		TeamPermissions:     make(map[string][]string),
		ProjectPermissions:  make(map[string][]string),
		AllPermissions:      []string{},
	}

	// 1. 检查是否超级管理员
	var u user.User
	if err := r.db.Database().Where("user_id = ?", userId).First(&u).Error; err != nil {
		return nil, err
	}

	if u.IsSuperAdmin == 1 {
		userPerms.IsSuperAdmin = true
		// 超管拥有所有权限，可以直接返回
		return userPerms, nil
	}

	// 2. 获取所有角色绑定
	bindings, err := r.GetUserRoleBindings(userId)
	if err != nil {
		return nil, err
	}

	// 3. 聚合各层级权限
	permissionSet := make(map[string]bool) // 用于去重

	for _, binding := range bindings {
		// 获取角色的权限列表
		permissions, err := r.GetRolePermissionCodes(binding.RoleId)
		if err != nil {
			continue
		}

		// 根据作用域分类存储
		switch binding.Scope {
		case permission.ScopePlatform:
			userPerms.PlatformPermissions = append(userPerms.PlatformPermissions, permissions...)
		case permission.ScopeOrganization:
			if binding.ResourceId != nil {
				userPerms.OrgPermissions[*binding.ResourceId] = append(
					userPerms.OrgPermissions[*binding.ResourceId], permissions...)
			}
		case permission.ScopeTeam:
			if binding.ResourceId != nil {
				userPerms.TeamPermissions[*binding.ResourceId] = append(
					userPerms.TeamPermissions[*binding.ResourceId], permissions...)
			}
		case permission.ScopeProject:
			if binding.ResourceId != nil {
				userPerms.ProjectPermissions[*binding.ResourceId] = append(
					userPerms.ProjectPermissions[*binding.ResourceId], permissions...)
			}
		}

		// 添加到总权限集合
		for _, perm := range permissions {
			permissionSet[perm] = true
		}
	}

	// 4. 转换为权限列表
	for perm := range permissionSet {
		userPerms.AllPermissions = append(userPerms.AllPermissions, perm)
	}

	return userPerms, nil
}

// GetRolePermissionCodes 获取角色的权限代码列表
func (r *PermissionRepo) GetRolePermissionCodes(roleId string) ([]string, error) {
	var codes []string

	err := r.db.Database().Table("t_role_permission rp").
		Select("p.code").
		Joins("JOIN t_permission p ON rp.permission_id = p.permission_id").
		Where("rp.role_id = ? AND p.is_enabled = ?", roleId, 1).
		Pluck("code", &codes).Error

	return codes, err
}

// ============ 权限检查 ============

// HasPermission 检查用户是否拥有指定权限
func (r *PermissionRepo) HasPermission(userId, permissionCode, resourceId, scope string) (bool, error) {
	userPerms, err := r.GetUserPermissions(userId)
	if err != nil {
		return false, err
	}

	// 超管直接通过
	if userPerms.IsSuperAdmin {
		return true, nil
	}

	// 检查平台级权限
	if contains(userPerms.PlatformPermissions, permissionCode) {
		return true, nil
	}

	// 根据 scope 检查对应层级权限
	switch scope {
	case permission.ScopeOrganization:
		if contains(userPerms.OrgPermissions[resourceId], permissionCode) {
			return true, nil
		}

	case permission.ScopeTeam:
		if contains(userPerms.TeamPermissions[resourceId], permissionCode) {
			return true, nil
		}
		// 检查团队所属组织的权限（继承）
		teamById, err := r.getTeamById(resourceId)
		if err == nil && teamById.OrgId != "" {
			if contains(userPerms.OrgPermissions[teamById.OrgId], permissionCode) {
				return true, nil
			}
		}

	case permission.ScopeProject:
		if contains(userPerms.ProjectPermissions[resourceId], permissionCode) {
			return true, nil
		}
		// 检查项目所属团队的权限（继承）
		projectById, err := r.getProjectById(resourceId)
		if err == nil {
			// 检查项目关联的团队
			teams, _ := r.getProjectTeams(resourceId)
			for _, teamId := range teams {
				if contains(userPerms.TeamPermissions[teamId], permissionCode) {
					return true, nil
				}
			}
			// 检查项目所属组织的权限（继承）
			if projectById.OrgId != "" {
				if contains(userPerms.OrgPermissions[projectById.OrgId], permissionCode) {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

// ============ 辅助方法 ============

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (r *PermissionRepo) getTeamById(teamId string) (*team.Team, error) {
	var t team.Team
	err := r.db.Database().Where("team_id = ?", teamId).First(&t).Error
	return &t, err
}

func (r *PermissionRepo) getProjectById(projectId string) (*project.Project, error) {
	var p project.Project
	err := r.db.Database().Where("project_id = ?", projectId).First(&p).Error
	return &p, err
}

func (r *PermissionRepo) getProjectTeams(projectId string) ([]string, error) {
	var teamIds []string
	err := r.db.Database().Table("t_project_team_relation").
		Where("project_id = ?", projectId).
		Pluck("team_id", &teamIds).Error
	return teamIds, err
}

// GetAccessibleResources 获取用户可访问的资源列表
func (r *PermissionRepo) GetAccessibleResources(userId string) (*permission.AccessibleResources, error) {
	resources := &permission.AccessibleResources{
		Organizations: []string{},
		Teams:         []string{},
		Projects:      []string{},
	}

	// 检查是否超管
	var u user.User
	if err := r.db.Database().Where("user_id = ?", userId).First(&u).Error; err != nil {
		return nil, err
	}

	if u.IsSuperAdmin == 1 {
		// 超管可访问所有资源
		r.db.Database().Table("t_organization").Pluck("org_id", &resources.Organizations)
		r.db.Database().Table("t_team").Pluck("team_id", &resources.Teams)
		r.db.Database().Table("t_project").Pluck("project_id", &resources.Projects)
		return resources, nil
	}

	// 获取角色绑定
	bindings, err := r.GetUserRoleBindings(userId)
	if err != nil {
		return nil, err
	}

	// 收集资源ID
	resourceSet := make(map[string]map[string]bool)
	resourceSet["organizations"] = make(map[string]bool)
	resourceSet["teams"] = make(map[string]bool)
	resourceSet["projects"] = make(map[string]bool)

	for _, binding := range bindings {
		if binding.ResourceId == nil {
			continue
		}

		switch binding.Scope {
		case permission.ScopeOrganization:
			resourceSet["organizations"][*binding.ResourceId] = true
		case permission.ScopeTeam:
			resourceSet["teams"][*binding.ResourceId] = true
			// 团队的组织也可访问
			if teamById, err := r.getTeamById(*binding.ResourceId); err == nil {
				resourceSet["organizations"][teamById.OrgId] = true
			}
		case permission.ScopeProject:
			resourceSet["projects"][*binding.ResourceId] = true
			// 项目的组织和团队也可访问
			if projectById, err := r.getProjectById(*binding.ResourceId); err == nil {
				resourceSet["organizations"][projectById.OrgId] = true
				teams, _ := r.getProjectTeams(*binding.ResourceId)
				for _, teamId := range teams {
					resourceSet["teams"][teamId] = true
				}
			}
		}
	}

	// 转换为列表
	for orgId := range resourceSet["organizations"] {
		resources.Organizations = append(resources.Organizations, orgId)
	}
	for teamId := range resourceSet["teams"] {
		resources.Teams = append(resources.Teams, teamId)
	}
	for projectId := range resourceSet["projects"] {
		resources.Projects = append(resources.Projects, projectId)
	}

	return resources, nil
}

// GetAccessibleRoutes 获取用户可访问的路由列表
func (r *PermissionRepo) GetAccessibleRoutes(userId string) ([]*permission.RouterPermission, error) {
	userPerms, err := r.GetUserPermissions(userId)
	if err != nil {
		return nil, err
	}

	// 超管可访问所有路由
	if userPerms.IsSuperAdmin {
		var allRoutes []*permission.RouterPermission
		r.db.Database().Where("is_enabled = ?", 1).Find(&allRoutes)
		return allRoutes, nil
	}

	// 获取所有启用的路由
	var allRoutes []*permission.RouterPermission
	if err := r.db.Database().Where("is_enabled = ?", 1).Find(&allRoutes).Error; err != nil {
		return nil, err
	}

	// 过滤用户有权限访问的路由
	var accessibleRoutes []*permission.RouterPermission
	for _, route := range allRoutes {
		// 如果路由不需要权限，直接可访问
		if len(route.RequiredPermissions) == 0 {
			accessibleRoutes = append(accessibleRoutes, route)
			continue
		}

		// 解析 RequiredPermissions (JSON 数组)
		var requiredPerms []string
		if len(route.RequiredPermissions) > 0 {
			if err := route.RequiredPermissions.UnmarshalJSON(route.RequiredPermissions); err == nil {
				// 直接使用 JSON 反序列化
				var unmarshalErr error
				requiredPerms, unmarshalErr = parseRequiredPermissions(route.RequiredPermissions)
				if unmarshalErr != nil {
					continue // 跳过无效的路由配置
				}
			}
		}

		// 检查用户是否拥有任一所需权限
		hasPermission := false
		for _, reqPerm := range requiredPerms {
			if contains(userPerms.AllPermissions, reqPerm) {
				hasPermission = true
				break
			}
		}

		if hasPermission {
			accessibleRoutes = append(accessibleRoutes, route)
		}
	}

	return accessibleRoutes, nil
}

// ============ 批量操作 ============

// BatchCreateRoleBindings 批量创建角色绑定
func (r *PermissionRepo) BatchCreateRoleBindings(bindings []*permission.UserRoleBinding) error {
	if len(bindings) == 0 {
		return errors.New("no bindings to create")
	}
	return r.db.Database().Create(&bindings).Error
}

// BatchDeleteRoleBindings 批量删除角色绑定
func (r *PermissionRepo) BatchDeleteRoleBindings(bindingIds []string) error {
	if len(bindingIds) == 0 {
		return errors.New("no bindings to delete")
	}
	return r.db.Database().Where("binding_id IN ?", bindingIds).Delete(&permission.UserRoleBinding{}).Error
}

// DeleteUserRoleBindingsByResource 删除用户在指定资源的所有角色绑定
func (r *PermissionRepo) DeleteUserRoleBindingsByResource(userId, scope, resourceId string) error {
	return r.db.Database().Where("user_id = ? AND scope = ? AND resource_id = ?", userId, scope, resourceId).
		Delete(&permission.UserRoleBinding{}).Error
}

// DeleteRoleBindingsByResource 删除指定资源的所有角色绑定
func (r *PermissionRepo) DeleteRoleBindingsByResource(scope, resourceId string) error {
	return r.db.Database().Where("scope = ? AND resource_id = ?", scope, resourceId).
		Delete(&permission.UserRoleBinding{}).Error
}

// ============ 缓存相关 ============

// ClearUserPermissionsCache 清除用户权限缓存
func (r *PermissionRepo) ClearUserPermissionsCache(userId string) error {
	if r.cache == nil {
		return nil
	}
	ctx := context.Background()
	cacheKey := fmt.Sprintf("user:permissions:%s", userId)
	return r.cache.Del(ctx, cacheKey).Err()
}

// ClearRolePermissionsCache 清除角色权限缓存
func (r *PermissionRepo) ClearRolePermissionsCache(roleId string) error {
	if r.cache == nil {
		return nil
	}
	ctx := context.Background()
	cacheKey := fmt.Sprintf("role:permissions:%s", roleId)
	return r.cache.Del(ctx, cacheKey).Err()
}

// parseRequiredPermissions 解析 JSON 格式的权限列表
func parseRequiredPermissions(jsonData datatypes.JSON) ([]string, error) {
	var perms []string
	if err := json.Unmarshal(jsonData, &perms); err != nil {
		return nil, err
	}
	return perms, nil
}
