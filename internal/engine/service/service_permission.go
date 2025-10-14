package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/observabil/arcade/internal/engine/model"
	"github.com/observabil/arcade/pkg/ctx"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2025/10/13
 * @file: service_permission.go
 * @description: 权限服务
 */

// PermissionService 权限服务
type PermissionService struct {
	ctx       *ctx.Context
	roleCache map[string]*model.Role // 角色缓存
}

// NewPermissionService 创建权限服务实例
func NewPermissionService(ctx *ctx.Context) *PermissionService {
	return &PermissionService{
		ctx:       ctx,
		roleCache: make(map[string]*model.Role),
	}
}

// GetRole 获取角色信息（带缓存）
func (ps *PermissionService) GetRole(roleId string) (*model.Role, error) {
	// 先查缓存
	if role, ok := ps.roleCache[roleId]; ok {
		return role, nil
	}

	// 从数据库查询
	var role model.Role
	err := ps.ctx.DB.Where("role_id = ? AND is_enabled = ?", roleId, model.RoleEnabled).First(&role).Error
	if err != nil {
		return nil, err
	}

	// 加入缓存
	ps.roleCache[roleId] = &role
	return &role, nil
}

// HasPermission 检查角色是否拥有指定权限点
func (ps *PermissionService) HasPermission(roleId string, permissionPoint string) (bool, error) {
	// 对于内置角色，直接查看预定义权限
	if permissions, ok := model.BuiltinRolePermissions[roleId]; ok {
		for _, perm := range permissions {
			if perm == permissionPoint {
				return true, nil
			}
		}
		return false, nil
	}

	// 对于自定义角色，从数据库加载权限
	role, err := ps.GetRole(roleId)
	if err != nil {
		return false, err
	}

	// 解析权限JSON
	var permissions []string
	if err := json.Unmarshal([]byte(role.Permissions), &permissions); err != nil {
		return false, fmt.Errorf("parse permissions failed: %w", err)
	}

	for _, perm := range permissions {
		if perm == permissionPoint {
			return true, nil
		}
	}

	return false, nil
}

// ProjectPermission 项目权限结构
type ProjectPermission struct {
	ProjectId   string
	UserId      string
	RoleId      string   // 角色ID
	RoleName    string   // 角色名称（用于显示）
	HasAccess   bool     // 是否有访问权限
	Source      string   // 权限来源: direct/team/org
	Permissions []string // 具体权限点列表
	Priority    int      // 角色优先级

	// 便捷的权限标志（从权限点计算得出）
	CanView        bool // 查看权限
	CanWrite       bool // 写入权限
	CanManage      bool // 管理权限
	CanDelete      bool // 删除权限
	CanManageTeams bool // 管理团队权限
}

// CheckProjectPermission 检查用户对项目的权限
// 权限计算逻辑：MAX(直接权限, 团队权限, 组织权限)
func (ps *PermissionService) CheckProjectPermission(ctx context.Context, userId, projectId string) (*ProjectPermission, error) {
	// 1. 获取项目信息
	var project model.Project
	if err := ps.ctx.DB.Where("project_id = ?", projectId).First(&project).Error; err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	perm := &ProjectPermission{
		ProjectId: projectId,
		UserId:    userId,
		HasAccess: false,
	}

	// 2. 项目所有者特殊处理（最高优先级）
	if project.CreatedBy == userId {
		role, _ := ps.GetRole(model.BuiltinProjectOwner)
		if role != nil {
			perm.RoleId = role.RoleId
			perm.RoleName = role.Name
			perm.Priority = role.Priority
			perm.HasAccess = true
			perm.Source = "direct"
			ps.loadRolePermissions(perm, role.RoleId)
		}
		ps.calculatePermissionFlags(perm)
		return perm, nil
	}

	// 3. 检查直接权限（项目成员）
	var directMember model.ProjectMember
	err := ps.ctx.DB.Where("project_id = ? AND user_id = ?", projectId, userId).First(&directMember).Error
	if err == nil {
		// 找到直接权限
		role, _ := ps.GetRole(directMember.RoleId)
		if role != nil {
			perm.RoleId = role.RoleId
			perm.RoleName = role.Name
			perm.Priority = role.Priority
			perm.HasAccess = true
			perm.Source = "direct"
			ps.loadRolePermissions(perm, role.RoleId)
		}
	}

	// 4. 检查团队权限
	teamRoleId, teamPriority := ps.calculateTeamPermission(ctx, userId, projectId)
	if teamRoleId != "" && teamPriority > perm.Priority {
		role, _ := ps.GetRole(teamRoleId)
		if role != nil {
			perm.RoleId = role.RoleId
			perm.RoleName = role.Name
			perm.Priority = role.Priority
			perm.HasAccess = true
			perm.Source = "team"
			ps.loadRolePermissions(perm, role.RoleId)
		}
	}

	// 5. 检查组织权限
	if project.AccessLevel == model.AccessLevelOrg {
		orgRoleId, orgPriority := ps.calculateOrgPermission(ctx, userId, project.OrgId)
		if orgRoleId != "" && orgPriority > perm.Priority {
			role, _ := ps.GetRole(orgRoleId)
			if role != nil {
				perm.RoleId = role.RoleId
				perm.RoleName = role.Name
				perm.Priority = role.Priority
				perm.HasAccess = true
				perm.Source = "org"
				ps.loadRolePermissions(perm, role.RoleId)
			}
		}
	}

	// 6. 根据权限点计算具体权限标志
	ps.calculatePermissionFlags(perm)

	return perm, nil
}

// calculateTeamPermission 计算团队权限
func (ps *PermissionService) calculateTeamPermission(ctx context.Context, userId, projectId string) (string, int) {
	// 查找用户所在的团队
	var teamMembers []model.TeamMember
	ps.ctx.DB.Where("user_id = ?", userId).Find(&teamMembers)

	var maxRoleId string
	var maxPriority int

	for _, tm := range teamMembers {
		// 查找团队对项目的访问权限
		var teamAccess model.ProjectTeamAccess
		err := ps.ctx.DB.Where("project_id = ? AND team_id = ?", projectId, tm.TeamId).First(&teamAccess).Error
		if err != nil {
			continue
		}

		// 获取团队成员角色
		teamRole, _ := ps.GetRole(tm.RoleId)
		if teamRole == nil {
			continue
		}

		// 团队成员角色 + 团队访问级别 => 项目有效角色
		effectiveRoleId := ps.mapTeamRoleToProjectRole(tm.RoleId, teamAccess.AccessLevel)
		effectiveRole, _ := ps.GetRole(effectiveRoleId)
		if effectiveRole != nil && effectiveRole.Priority > maxPriority {
			maxRoleId = effectiveRoleId
			maxPriority = effectiveRole.Priority
		}
	}

	return maxRoleId, maxPriority
}

// calculateOrgPermission 计算组织权限
func (ps *PermissionService) calculateOrgPermission(ctx context.Context, userId, orgId string) (string, int) {
	var orgMember model.OrganizationMember
	err := ps.ctx.DB.Where("org_id = ? AND user_id = ? AND status = ?",
		orgId, userId, model.OrgMemberStatusActive).First(&orgMember).Error
	if err != nil {
		return "", 0
	}

	// 组织角色映射到项目角色
	var mappedRoleId string
	switch orgMember.RoleId {
	case model.BuiltinOrgOwner:
		mappedRoleId = model.BuiltinProjectMaintainer
	case model.BuiltinOrgAdmin:
		mappedRoleId = model.BuiltinProjectDeveloper
	case model.BuiltinOrgMember:
		mappedRoleId = model.BuiltinProjectGuest
	default:
		mappedRoleId = model.BuiltinProjectGuest
	}

	role, _ := ps.GetRole(mappedRoleId)
	if role != nil {
		return role.RoleId, role.Priority
	}

	return "", 0
}

// mapTeamRoleToProjectRole 映射团队角色到项目角色
func (ps *PermissionService) mapTeamRoleToProjectRole(teamRoleId, accessLevel string) string {
	// 团队访问级别限制最大权限
	switch accessLevel {
	case model.AccessLevelRead:
		// 只读访问，最多给 guest 权限
		return model.BuiltinProjectGuest
	case model.AccessLevelWrite:
		// 读写访问
		switch teamRoleId {
		case model.BuiltinTeamOwner, model.BuiltinTeamMaintainer, model.BuiltinTeamDeveloper:
			return model.BuiltinProjectDeveloper
		default:
			return model.BuiltinProjectReporter
		}
	case model.AccessLevelAdmin:
		// 管理员访问，映射团队角色到项目角色
		switch teamRoleId {
		case model.BuiltinTeamOwner, model.BuiltinTeamMaintainer:
			return model.BuiltinProjectMaintainer
		case model.BuiltinTeamDeveloper:
			return model.BuiltinProjectDeveloper
		case model.BuiltinTeamReporter:
			return model.BuiltinProjectReporter
		default:
			return model.BuiltinProjectGuest
		}
	default:
		return model.BuiltinProjectGuest
	}
}

// loadRolePermissions 加载角色的权限点列表
func (ps *PermissionService) loadRolePermissions(perm *ProjectPermission, roleId string) {
	// 对于内置角色，使用预定义权限
	if permissions, ok := model.BuiltinRolePermissions[roleId]; ok {
		perm.Permissions = permissions
		return
	}

	// 对于自定义角色，从数据库加载
	role, err := ps.GetRole(roleId)
	if err != nil {
		perm.Permissions = []string{}
		return
	}

	var permissions []string
	if err := json.Unmarshal([]byte(role.Permissions), &permissions); err != nil {
		perm.Permissions = []string{}
		return
	}

	perm.Permissions = permissions
}

// calculatePermissionFlags 根据权限点计算具体权限标志
func (ps *PermissionService) calculatePermissionFlags(perm *ProjectPermission) {
	// 根据权限点列表计算标志
	perm.CanView = ps.hasPermissionPoint(perm.Permissions, model.PermProjectView)

	perm.CanWrite = ps.hasPermissionPoint(perm.Permissions, model.PermBuildTrigger) ||
		ps.hasPermissionPoint(perm.Permissions, model.PermPipelineCreate)

	perm.CanManage = ps.hasPermissionPoint(perm.Permissions, model.PermProjectSettings) ||
		ps.hasPermissionPoint(perm.Permissions, model.PermMemberEdit)

	perm.CanDelete = ps.hasPermissionPoint(perm.Permissions, model.PermProjectDelete)

	perm.CanManageTeams = ps.hasPermissionPoint(perm.Permissions, model.PermTeamProject)
}

// hasPermissionPoint 检查权限点列表中是否包含指定权限
func (ps *PermissionService) hasPermissionPoint(permissions []string, point string) bool {
	for _, p := range permissions {
		if p == point {
			return true
		}
	}
	return false
}

// RequireProjectPermission 要求特定项目权限（按角色ID）
func (ps *PermissionService) RequireProjectPermission(ctx context.Context, userId, projectId string, requiredRoleId string) error {
	perm, err := ps.CheckProjectPermission(ctx, userId, projectId)
	if err != nil {
		return err
	}

	if !perm.HasAccess {
		return fmt.Errorf("access denied: user %s has no access to project %s", userId, projectId)
	}

	// 获取要求的角色优先级
	requiredRole, _ := ps.GetRole(requiredRoleId)
	if requiredRole == nil {
		return fmt.Errorf("invalid required role: %s", requiredRoleId)
	}

	if perm.Priority < requiredRole.Priority {
		return fmt.Errorf("insufficient permission: required %s, got %s", requiredRole.Name, perm.RoleName)
	}

	return nil
}

// RequirePermissionPoint 要求特定权限点
func (ps *PermissionService) RequirePermissionPoint(ctx context.Context, userId, projectId string, permissionPoint string) error {
	perm, err := ps.CheckProjectPermission(ctx, userId, projectId)
	if err != nil {
		return err
	}

	if !perm.HasAccess {
		return fmt.Errorf("access denied: user %s has no access to project %s", userId, projectId)
	}

	if !ps.hasPermissionPoint(perm.Permissions, permissionPoint) {
		return fmt.Errorf("permission denied: %s required", permissionPoint)
	}

	return nil
}

// CheckOrganizationPermission 检查用户是否为组织成员
func (ps *PermissionService) CheckOrganizationPermission(ctx context.Context, userId, orgId string) (string, error) {
	var orgMember model.OrganizationMember
	err := ps.ctx.DB.Where("org_id = ? AND user_id = ? AND status = ?",
		orgId, userId, model.OrgMemberStatusActive).First(&orgMember).Error
	if err != nil {
		return "", fmt.Errorf("user is not a member of organization")
	}

	return orgMember.RoleId, nil
}

// CheckTeamPermission 检查用户是否为团队成员
func (ps *PermissionService) CheckTeamPermission(ctx context.Context, userId, teamId string) (string, error) {
	var teamMember model.TeamMember
	err := ps.ctx.DB.Where("team_id = ? AND user_id = ?", teamId, userId).First(&teamMember).Error
	if err != nil {
		return "", fmt.Errorf("user is not a member of team")
	}

	return teamMember.RoleId, nil
}
