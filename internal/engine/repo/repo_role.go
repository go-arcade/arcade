package repo

import (
	"encoding/json"

	"github.com/observabil/arcade/internal/engine/model"
	"github.com/observabil/arcade/pkg/ctx"
)

type RoleRepo struct {
	ctx *ctx.Context
}

func NewRoleRepo(ctx *ctx.Context) *RoleRepo {
	return &RoleRepo{ctx: ctx}
}

// GetRole 获取角色
func (r *RoleRepo) GetRole(roleId string) (*model.Role, error) {
	var role model.Role
	err := r.ctx.DB.Where("role_id = ?", roleId).First(&role).Error
	return &role, err
}

// GetRoleByName 根据名称获取角色
func (r *RoleRepo) GetRoleByName(name string, scope model.RoleScope, orgId string) (*model.Role, error) {
	var role model.Role
	query := r.ctx.DB.Where("name = ? AND scope = ?", name, scope)
	if orgId != "" {
		query = query.Where("org_id = ?", orgId)
	} else {
		query = query.Where("org_id IS NULL OR org_id = ''")
	}
	err := query.First(&role).Error
	return &role, err
}

// ListRoles 列出角色
func (r *RoleRepo) ListRoles(scope model.RoleScope, orgId string) ([]model.Role, error) {
	var roles []model.Role
	query := r.ctx.DB.Where("scope = ? AND is_enabled = ?", scope, model.RoleEnabled)
	if orgId != "" {
		// 包含全局角色和组织自定义角色
		query = query.Where("org_id IS NULL OR org_id = '' OR org_id = ?", orgId)
	} else {
		// 只返回全局角色
		query = query.Where("org_id IS NULL OR org_id = ''")
	}
	err := query.Order("priority DESC, create_time ASC").Find(&roles).Error
	return roles, err
}

// ListBuiltinRoles 列出内置角色
func (r *RoleRepo) ListBuiltinRoles(scope model.RoleScope) ([]model.Role, error) {
	var roles []model.Role
	err := r.ctx.DB.Where("scope = ? AND is_builtin = ?", scope, model.RoleBuiltin).
		Order("priority DESC").Find(&roles).Error
	return roles, err
}

// CreateRole 创建角色
func (r *RoleRepo) CreateRole(role *model.Role) error {
	return r.ctx.DB.Create(role).Error
}

// UpdateRole 更新角色
func (r *RoleRepo) UpdateRole(role *model.Role) error {
	return r.ctx.DB.Save(role).Error
}

// UpdateRolePermissions 更新角色权限
func (r *RoleRepo) UpdateRolePermissions(roleId string, permissions []string) error {
	permJson, err := json.Marshal(permissions)
	if err != nil {
		return err
	}
	return r.ctx.DB.Model(&model.Role{}).
		Where("role_id = ?", roleId).
		Update("permissions", string(permJson)).Error
}

// DeleteRole 删除角色（只能删除自定义角色）
func (r *RoleRepo) DeleteRole(roleId string) error {
	return r.ctx.DB.Where("role_id = ? AND is_builtin = ?", roleId, model.RoleCustom).
		Delete(&model.Role{}).Error
}

// EnableRole 启用/禁用角色
func (r *RoleRepo) EnableRole(roleId string, enabled bool) error {
	isEnabled := model.RoleDisabled
	if enabled {
		isEnabled = model.RoleEnabled
	}
	return r.ctx.DB.Model(&model.Role{}).
		Where("role_id = ?", roleId).
		Update("is_enabled", isEnabled).Error
}

// InitBuiltinRoles 初始化内置角色（首次启动时调用）
func (r *RoleRepo) InitBuiltinRoles() error {
	// 检查是否已初始化
	var count int64
	r.ctx.DB.Model(&model.Role{}).Where("is_builtin = ?", model.RoleBuiltin).Count(&count)
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
		role := &model.Role{
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
		if err := r.ctx.DB.Create(role).Error; err != nil {
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
		role := &model.Role{
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
		if err := r.ctx.DB.Create(role).Error; err != nil {
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
		role := &model.Role{
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
		if err := r.ctx.DB.Create(role).Error; err != nil {
			return err
		}
	}

	return nil
}
