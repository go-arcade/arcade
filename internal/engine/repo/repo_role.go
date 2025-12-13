package repo

import (
	"encoding/json"

	"github.com/go-arcade/arcade/internal/engine/model"
	"github.com/go-arcade/arcade/pkg/database"
)

type IRoleRepository interface {
	GetRole(roleId string) (*model.Role, error)
	GetRoleByName(name string, scope model.RoleScope, orgId string) (*model.Role, error)
	ListRoles(scope model.RoleScope, orgId string) ([]model.Role, error)
	ListBuiltinRoles(scope model.RoleScope) ([]model.Role, error)
	CreateRole(r *model.Role) error
	UpdateRole(r *model.Role) error
	UpdateRolePermissions(roleId string, permissions []string) error
	DeleteRole(roleId string) error
	EnableRole(roleId string, enabled bool) error
	ListRolesWithPagination(req *model.ListRolesRequest) ([]model.Role, int64, error)
	ToggleRole(roleId string) error
	RoleExists(roleId string) (bool, error)
	InitBuiltinRoles() error
}

type RoleRepo struct {
	database.IDatabase
}

func NewRoleRepo(db database.IDatabase) IRoleRepository {
	return &RoleRepo{IDatabase: db}
}

// GetRole 获取角色
func (r *RoleRepo) GetRole(roleId string) (*model.Role, error) {
	var ro model.Role
	err := r.Database().Select("id", "role_id", "name", "display_name", "description", "scope", "org_id", "is_builtin", "is_enabled", "priority", "permissions", "created_by", "created_at", "updated_at").
		Where("role_id = ?", roleId).First(&ro).Error
	return &ro, err
}

// GetRoleByName 根据名称获取角色
func (r *RoleRepo) GetRoleByName(name string, scope model.RoleScope, orgId string) (*model.Role, error) {
	var ro model.Role
	query := r.Database().Select("id", "role_id", "name", "display_name", "description", "scope", "org_id", "is_builtin", "is_enabled", "priority", "permissions", "created_by", "created_at", "updated_at").
		Where("name = ? AND scope = ?", name, scope)
	if orgId != "" {
		query = query.Where("org_id = ?", orgId)
	} else {
		query = query.Where("org_id IS NULL OR org_id = ''")
	}
	err := query.First(&ro).Error
	return &ro, err
}

// ListRoles 列出角色
func (r *RoleRepo) ListRoles(scope model.RoleScope, orgId string) ([]model.Role, error) {
	var roles []model.Role
	query := r.Database().Select("role_id", "name", "display_name", "description", "scope", "org_id", "is_builtin", "is_enabled", "priority", "permissions", "created_by").
		Where("scope = ? AND is_enabled = ?", scope, model.RoleEnabled)
	if orgId != "" {
		// 包含全局角色和组织自定义角色
		query = query.Where("org_id IS NULL OR org_id = '' OR org_id = ?", orgId)
	} else {
		// 只返回全局角色
		query = query.Where("org_id IS NULL OR org_id = ''")
	}
	err := query.Select("id", "role_id", "name", "display_name", "description", "scope", "org_id", "is_builtin", "is_enabled", "priority", "permissions", "created_by", "created_at", "updated_at").
		Order("priority DESC").Find(&roles).Error
	return roles, err
}

// ListBuiltinRoles 列出内置角色
func (r *RoleRepo) ListBuiltinRoles(scope model.RoleScope) ([]model.Role, error) {
	var roles []model.Role
	err := r.Database().
		Select("id", "role_id", "name", "display_name", "description", "scope", "org_id", "is_builtin", "is_enabled", "priority", "permissions", "created_by", "created_at", "updated_at").
		Where("scope = ? AND is_builtin = ?", scope, model.RoleBuiltin).
		Order("priority DESC").
		Find(&roles).Error
	return roles, err
}

// CreateRole 创建角色
func (r *RoleRepo) CreateRole(ro *model.Role) error {
	return r.Database().Create(ro).Error
}

// UpdateRole 更新角色
func (r *RoleRepo) UpdateRole(ro *model.Role) error {
	return r.Database().Save(ro).Error
}

// UpdateRolePermissions 更新角色权限
func (r *RoleRepo) UpdateRolePermissions(roleId string, permissions []string) error {
	permJson, err := json.Marshal(permissions)
	if err != nil {
		return err
	}
	return r.Database().Model(&model.Role{}).
		Where("role_id = ?", roleId).
		Update("permissions", string(permJson)).Error
}

// DeleteRole 删除角色（只能删除自定义角色）
func (r *RoleRepo) DeleteRole(roleId string) error {
	return r.Database().Where("role_id = ? AND is_builtin = ?", roleId, model.RoleCustom).
		Delete(&model.Role{}).Error
}

// EnableRole 启用/禁用角色
func (r *RoleRepo) EnableRole(roleId string, enabled bool) error {
	isEnabled := model.RoleDisabled
	if enabled {
		isEnabled = model.RoleEnabled
	}
	return r.Database().Model(&model.Role{}).
		Where("role_id = ?", roleId).
		Update("is_enabled", isEnabled).Error
}

// ListRolesWithPagination lists roles with pagination and filters
func (r *RoleRepo) ListRolesWithPagination(req *model.ListRolesRequest) ([]model.Role, int64, error) {
	var roles []model.Role
	var total int64

	query := r.Database().Model(&model.Role{})

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
	return r.Database().Model(&model.Role{}).
		Where("role_id = ?", roleId).
		Update("is_enabled", r.Database().Raw("1 - is_enabled")).Error
}

// RoleExists checks if a role exists
func (r *RoleRepo) RoleExists(roleId string) (bool, error) {
	var count int64
	err := r.Database().Model(&model.Role{}).Where("role_id = ?", roleId).Count(&count).Error
	return count > 0, err
}

// InitBuiltinRoles 初始化内置角色（首次启动时调用）
func (r *RoleRepo) InitBuiltinRoles() error {
	// 检查是否已初始化
	var count int64
	r.Database().Model(&model.Role{}).Where("is_builtin = ?", model.RoleBuiltin).Count(&count)
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
		{model.BuiltinProjectOwner, "owner", "所有者", 50},
		{model.BuiltinProjectMaintainer, "maintainer", "维护者", 40},
		{model.BuiltinProjectDeveloper, "developer", "开发者", 30},
		{model.BuiltinProjectReporter, "reporter", "报告者", 20},
		{model.BuiltinProjectGuest, "guest", "访客", 10},
	}

	for _, pr := range projectRoles {
		permJson, _ := json.Marshal(model.BuiltinRolePermissions[pr.RoleId])
		role2 := &model.Role{
			RoleId:      pr.RoleId,
			Name:        pr.Name,
			DisplayName: pr.DisplayName,
			Description: "内置" + pr.DisplayName + "角色",
			Scope:       model.RoleScopeProject,
			IsBuiltin:   model.RoleBuiltin,
			IsEnabled:   model.RoleEnabled,
			Priority:    pr.Priority,
			Permissions: string(permJson),
		}
		if err := r.Database().Create(role2).Error; err != nil {
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
		{model.BuiltinTeamOwner, "owner", "所有者", 50},
		{model.BuiltinTeamMaintainer, "maintainer", "维护者", 40},
		{model.BuiltinTeamDeveloper, "developer", "开发者", 30},
		{model.BuiltinTeamReporter, "reporter", "报告者", 20},
		{model.BuiltinTeamGuest, "guest", "访客", 10},
	}

	for _, tr := range teamRoles {
		permJson, _ := json.Marshal(model.BuiltinRolePermissions[tr.RoleId])
		role2 := &model.Role{
			RoleId:      tr.RoleId,
			Name:        tr.Name,
			DisplayName: tr.DisplayName,
			Description: "内置" + tr.DisplayName + "角色",
			Scope:       model.RoleScopeTeam,
			IsBuiltin:   model.RoleBuiltin,
			IsEnabled:   model.RoleEnabled,
			Priority:    tr.Priority,
			Permissions: string(permJson),
		}
		if err := r.Database().Create(role2).Error; err != nil {
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
		{model.BuiltinOrgOwner, "owner", "所有者", 50},
		{model.BuiltinOrgAdmin, "admin", "管理员", 40},
		{model.BuiltinOrgMember, "member", "成员", 10},
	}

	for _, or := range orgRoles {
		permJson, _ := json.Marshal(model.BuiltinRolePermissions[or.RoleId])
		role2 := &model.Role{
			RoleId:      or.RoleId,
			Name:        or.Name,
			DisplayName: or.DisplayName,
			Description: "内置" + or.DisplayName + "角色",
			Scope:       model.RoleScopeOrg,
			IsBuiltin:   model.RoleBuiltin,
			IsEnabled:   model.RoleEnabled,
			Priority:    or.Priority,
			Permissions: string(permJson),
		}
		if err := r.Database().Create(role2).Error; err != nil {
			return err
		}
	}

	return nil
}
