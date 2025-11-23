package role

import (
	"encoding/json"

	"github.com/go-arcade/arcade/internal/engine/model/role"
	"github.com/go-arcade/arcade/pkg/database"
)

type IRoleRepository interface {
	GetRole(roleId string) (*role.Role, error)
	GetRoleByName(name string, scope role.RoleScope, orgId string) (*role.Role, error)
	ListRoles(scope role.RoleScope, orgId string) ([]role.Role, error)
	ListBuiltinRoles(scope role.RoleScope) ([]role.Role, error)
	CreateRole(r *role.Role) error
	UpdateRole(r *role.Role) error
	UpdateRolePermissions(roleId string, permissions []string) error
	DeleteRole(roleId string) error
	EnableRole(roleId string, enabled bool) error
	ListRolesWithPagination(req *role.ListRolesRequest) ([]role.Role, int64, error)
	ToggleRole(roleId string) error
	RoleExists(roleId string) (bool, error)
	InitBuiltinRoles() error
}

type RoleRepo struct {
	db database.IDatabase
}

func NewRoleRepo(db database.IDatabase) IRoleRepository {
	return &RoleRepo{db: db}
}

// GetRole 获取角色
func (r *RoleRepo) GetRole(roleId string) (*role.Role, error) {
	var ro role.Role
	err := r.db.Database().Where("role_id = ?", roleId).First(&ro).Error
	return &ro, err
}

// GetRoleByName 根据名称获取角色
func (r *RoleRepo) GetRoleByName(name string, scope role.RoleScope, orgId string) (*role.Role, error) {
	var ro role.Role
	query := r.db.Database().Where("name = ? AND scope = ?", name, scope)
	if orgId != "" {
		query = query.Where("org_id = ?", orgId)
	} else {
		query = query.Where("org_id IS NULL OR org_id = ''")
	}
	err := query.First(&ro).Error
	return &ro, err
}

// ListRoles 列出角色
func (r *RoleRepo) ListRoles(scope role.RoleScope, orgId string) ([]role.Role, error) {
	var roles []role.Role
	query := r.db.Database().Select("role_id", "name", "display_name", "description", "scope", "org_id", "is_builtin", "is_enabled", "priority", "permissions", "created_by").
		Where("scope = ? AND is_enabled = ?", scope, role.RoleEnabled)
	if orgId != "" {
		// 包含全局角色和组织自定义角色
		query = query.Where("org_id IS NULL OR org_id = '' OR org_id = ?", orgId)
	} else {
		// 只返回全局角色
		query = query.Where("org_id IS NULL OR org_id = ''")
	}
	err := query.Order("priority DESC").Find(&roles).Error
	return roles, err
}

// ListBuiltinRoles 列出内置角色
func (r *RoleRepo) ListBuiltinRoles(scope role.RoleScope) ([]role.Role, error) {
	var roles []role.Role
	err := r.db.Database().
		Select("role_id", "name", "display_name", "description", "scope", "org_id", "is_builtin", "is_enabled", "priority", "permissions", "created_by").
		Where("scope = ? AND is_builtin = ?", scope, role.RoleBuiltin).
		Order("priority DESC").
		Find(&roles).Error
	return roles, err
}

// CreateRole 创建角色
func (r *RoleRepo) CreateRole(ro *role.Role) error {
	return r.db.Database().Create(ro).Error
}

// UpdateRole 更新角色
func (r *RoleRepo) UpdateRole(ro *role.Role) error {
	return r.db.Database().Save(ro).Error
}

// UpdateRolePermissions 更新角色权限
func (r *RoleRepo) UpdateRolePermissions(roleId string, permissions []string) error {
	permJson, err := json.Marshal(permissions)
	if err != nil {
		return err
	}
	return r.db.Database().Model(&role.Role{}).
		Where("role_id = ?", roleId).
		Update("permissions", string(permJson)).Error
}

// DeleteRole 删除角色（只能删除自定义角色）
func (r *RoleRepo) DeleteRole(roleId string) error {
	return r.db.Database().Where("role_id = ? AND is_builtin = ?", roleId, role.RoleCustom).
		Delete(&role.Role{}).Error
}

// EnableRole 启用/禁用角色
func (r *RoleRepo) EnableRole(roleId string, enabled bool) error {
	isEnabled := role.RoleDisabled
	if enabled {
		isEnabled = role.RoleEnabled
	}
	return r.db.Database().Model(&role.Role{}).
		Where("role_id = ?", roleId).
		Update("is_enabled", isEnabled).Error
}

// ListRolesWithPagination lists roles with pagination and filters
func (r *RoleRepo) ListRolesWithPagination(req *role.ListRolesRequest) ([]role.Role, int64, error) {
	var roles []role.Role
	var total int64

	query := r.db.Database().Model(&role.Role{})

	// apply filters
	if req.Scope != "" {
		query = query.Where("scope = ?", req.Scope)
	}
	if req.OrgId != "" {
		// include global roles and org-specific roles
		query = query.Where("org_id IS NULL OR org_id = '' OR org_id = ?", req.OrgId)
	}
	if req.IsBuiltin != nil {
		query = query.Where("is_builtin = ?", *req.IsBuiltin)
	}
	if req.IsEnabled != nil {
		query = query.Where("is_enabled = ?", *req.IsEnabled)
	}
	if req.Name != "" {
		query = query.Where("name LIKE ?", "%"+req.Name+"%")
	}

	// get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// apply pagination and sorting (specify fields, exclude created_at and updated_at)
	offset := (req.PageNum - 1) * req.PageSize
	err := query.Select("role_id", "name", "display_name", "description", "scope", "org_id", "is_builtin", "is_enabled", "priority", "permissions", "created_by").
		Order("priority DESC").
		Offset(offset).
		Limit(req.PageSize).
		Find(&roles).Error

	return roles, total, err
}

// ToggleRole toggles the enabled status of a role
func (r *RoleRepo) ToggleRole(roleId string) error {
	return r.db.Database().Model(&role.Role{}).
		Where("role_id = ?", roleId).
		Update("is_enabled", r.db.Database().Raw("1 - is_enabled")).Error
}

// RoleExists checks if a role exists
func (r *RoleRepo) RoleExists(roleId string) (bool, error) {
	var count int64
	err := r.db.Database().Model(&role.Role{}).Where("role_id = ?", roleId).Count(&count).Error
	return count > 0, err
}

// InitBuiltinRoles 初始化内置角色（首次启动时调用）
func (r *RoleRepo) InitBuiltinRoles() error {
	// 检查是否已初始化
	var count int64
	r.db.Database().Model(&role.Role{}).Where("is_builtin = ?", role.RoleBuiltin).Count(&count)
	if count > 0 {
		// 已初始化，跳过
		return nil
	}

	// 创建内置项目角色
	projectRoles := []struct {
		RoleId      string
		Name        string
		DisplayName string
		Priority    int
	}{
		{role.BuiltinProjectOwner, "owner", "所有者", 50},
		{role.BuiltinProjectMaintainer, "maintainer", "维护者", 40},
		{role.BuiltinProjectDeveloper, "developer", "开发者", 30},
		{role.BuiltinProjectReporter, "reporter", "报告者", 20},
		{role.BuiltinProjectGuest, "guest", "访客", 10},
	}

	for _, pr := range projectRoles {
		permJson, _ := json.Marshal(role.BuiltinRolePermissions[pr.RoleId])
		role2 := &role.Role{
			RoleId:      pr.RoleId,
			Name:        pr.Name,
			DisplayName: pr.DisplayName,
			Description: "内置" + pr.DisplayName + "角色",
			Scope:       role.RoleScopeProject,
			IsBuiltin:   role.RoleBuiltin,
			IsEnabled:   role.RoleEnabled,
			Priority:    pr.Priority,
			Permissions: string(permJson),
		}
		if err := r.db.Database().Create(role2).Error; err != nil {
			return err
		}
	}

	// 创建内置团队角色
	teamRoles := []struct {
		RoleId      string
		Name        string
		DisplayName string
		Priority    int
	}{
		{role.BuiltinTeamOwner, "owner", "所有者", 50},
		{role.BuiltinTeamMaintainer, "maintainer", "维护者", 40},
		{role.BuiltinTeamDeveloper, "developer", "开发者", 30},
		{role.BuiltinTeamReporter, "reporter", "报告者", 20},
		{role.BuiltinTeamGuest, "guest", "访客", 10},
	}

	for _, tr := range teamRoles {
		permJson, _ := json.Marshal(role.BuiltinRolePermissions[tr.RoleId])
		role2 := &role.Role{
			RoleId:      tr.RoleId,
			Name:        tr.Name,
			DisplayName: tr.DisplayName,
			Description: "内置" + tr.DisplayName + "角色",
			Scope:       role.RoleScopeTeam,
			IsBuiltin:   role.RoleBuiltin,
			IsEnabled:   role.RoleEnabled,
			Priority:    tr.Priority,
			Permissions: string(permJson),
		}
		if err := r.db.Database().Create(role2).Error; err != nil {
			return err
		}
	}

	// 创建内置组织角色
	orgRoles := []struct {
		RoleId      string
		Name        string
		DisplayName string
		Priority    int
	}{
		{role.BuiltinOrgOwner, "owner", "所有者", 50},
		{role.BuiltinOrgAdmin, "admin", "管理员", 40},
		{role.BuiltinOrgMember, "member", "成员", 10},
	}

	for _, or := range orgRoles {
		permJson, _ := json.Marshal(role.BuiltinRolePermissions[or.RoleId])
		role2 := &role.Role{
			RoleId:      or.RoleId,
			Name:        or.Name,
			DisplayName: or.DisplayName,
			Description: "内置" + or.DisplayName + "角色",
			Scope:       role.RoleScopeOrg,
			IsBuiltin:   role.RoleBuiltin,
			IsEnabled:   role.RoleEnabled,
			Priority:    or.Priority,
			Permissions: string(permJson),
		}
		if err := r.db.Database().Create(role2).Error; err != nil {
			return err
		}
	}

	return nil
}
