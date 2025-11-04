package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-arcade/arcade/internal/engine/model"
	"github.com/go-arcade/arcade/internal/engine/repo"
	"github.com/go-arcade/arcade/pkg/ctx"
	"github.com/go-arcade/arcade/pkg/id"
)

type PermissionService struct {
	Ctx      *ctx.Context
	PermRepo *repo.PermissionRepo
	RoleRepo *repo.RoleRepo
	UserRepo *repo.UserRepo
}

func NewPermissionService(ctx *ctx.Context) *PermissionService {
	return &PermissionService{
		Ctx:      ctx,
		PermRepo: repo.NewPermissionRepo(ctx),
		RoleRepo: repo.NewRoleRepo(ctx),
		UserRepo: repo.NewUserRepo(ctx),
	}
}

// ============ 用户权限聚合 ============

// GetUserPermissions 获取用户完整权限（带缓存）
func (s *PermissionService) GetUserPermissions(userId string) (*model.UserPermissions, error) {
	// 尝试从缓存读取
	cacheKey := fmt.Sprintf("user:permissions:%s", userId)
	cached, err := s.Ctx.RedisSession().Get(s.Ctx.ContextIns(), cacheKey).Result()
	if err == nil && cached != "" {
		var perms model.UserPermissions
		if err := json.Unmarshal([]byte(cached), &perms); err == nil {
			return &perms, nil
		}
	}

	// 从数据库查询
	perms, err := s.PermRepo.GetUserPermissions(userId)
	if err != nil {
		return nil, err
	}

	// 写入缓存（5分钟）
	data, _ := json.Marshal(perms)
	s.Ctx.RedisSession().Set(s.Ctx.ContextIns(), cacheKey, data, 5*time.Minute)

	return perms, nil
}

// GetUserPermissionsWithRoutes 获取用户权限和可访问路由
func (s *PermissionService) GetUserPermissionsWithRoutes(userId string) (*model.UserPermissions, error) {
	perms, err := s.GetUserPermissions(userId)
	if err != nil {
		return nil, err
	}

	// 获取可访问的路由
	routes, err := s.PermRepo.GetAccessibleRoutes(userId)
	if err != nil {
		return nil, err
	}
	perms.AccessibleRoutes = routes

	// 获取可访问的资源
	resources, err := s.PermRepo.GetAccessibleResources(userId)
	if err != nil {
		return nil, err
	}
	perms.AccessibleResources = resources

	return perms, nil
}

// CheckPermission 检查用户是否拥有指定权限
func (s *PermissionService) CheckPermission(req *model.PermissionCheckRequest) (*model.PermissionCheckResponse, error) {
	hasPermission, err := s.PermRepo.HasPermission(req.UserId, req.PermissionCode, req.ResourceId, req.Scope)
	if err != nil {
		return nil, err
	}

	resp := &model.PermissionCheckResponse{
		HasPermission: hasPermission,
	}

	if !hasPermission {
		resp.Reason = "权限不足"
	}

	return resp, nil
}

// ============ 角色绑定管理 ============

// CreateRoleBinding 创建用户角色绑定
func (s *PermissionService) CreateRoleBinding(req *model.UserRoleBindingCreate) (*model.UserRoleBinding, error) {
	// 验证用户存在
	var user model.User
	if err := s.Ctx.DBSession().Where("user_id = ?", req.UserId).First(&user).Error; err != nil {
		return nil, errors.New("用户不存在")
	}

	// 验证角色存在
	if _, err := s.RoleRepo.GetRole(req.RoleId); err != nil {
		return nil, errors.New("角色不存在")
	}

	// 验证资源存在（根据scope）
	if err := s.validateResource(req.Scope, req.ResourceId); err != nil {
		return nil, err
	}

	// 检查是否已存在
	resourceId := ""
	if req.ResourceId != nil {
		resourceId = *req.ResourceId
	}
	exists, _ := s.PermRepo.HasRoleBinding(req.UserId, req.RoleId, req.Scope, resourceId)
	if exists {
		return nil, errors.New("角色绑定已存在")
	}

	// 创建绑定
	binding := &model.UserRoleBinding{
		BindingId:  id.GetUUID(),
		UserId:     req.UserId,
		RoleId:     req.RoleId,
		Scope:      req.Scope,
		ResourceId: req.ResourceId,
		GrantedBy:  req.GrantedBy,
	}

	if err := s.PermRepo.CreateRoleBinding(binding); err != nil {
		return nil, err
	}

	// 清除用户权限缓存
	s.PermRepo.ClearUserPermissionsCache(req.UserId)

	return binding, nil
}

// DeleteRoleBinding 删除角色绑定
func (s *PermissionService) DeleteRoleBinding(bindingId string) error {
	// 获取绑定信息
	binding, err := s.PermRepo.GetRoleBinding(bindingId)
	if err != nil {
		return errors.New("绑定不存在")
	}

	// 删除绑定
	if err := s.PermRepo.DeleteRoleBinding(bindingId); err != nil {
		return err
	}

	// 清除用户权限缓存
	s.PermRepo.ClearUserPermissionsCache(binding.UserId)

	return nil
}

// GetUserRoleBindings 获取用户的所有角色绑定
func (s *PermissionService) GetUserRoleBindings(userId string) ([]*model.UserRoleBinding, error) {
	return s.PermRepo.GetUserRoleBindings(userId)
}

// ListRoleBindings 列出角色绑定（支持多条件查询）
func (s *PermissionService) ListRoleBindings(query *model.RoleBindingQuery) ([]*model.UserRoleBinding, int64, error) {
	return s.PermRepo.ListRoleBindings(query)
}

// ============ 批量操作 ============

// BatchAssignRoleToUsers 批量为用户分配角色
func (s *PermissionService) BatchAssignRoleToUsers(userIds []string, roleId, scope string, resourceId *string, grantedBy *string) error {
	var bindings []*model.UserRoleBinding

	for _, userId := range userIds {
		// 检查是否已存在
		rid := ""
		if resourceId != nil {
			rid = *resourceId
		}
		exists, _ := s.PermRepo.HasRoleBinding(userId, roleId, scope, rid)
		if exists {
			continue
		}

		binding := &model.UserRoleBinding{
			BindingId:  id.GetUUID(),
			UserId:     userId,
			RoleId:     roleId,
			Scope:      scope,
			ResourceId: resourceId,
			GrantedBy:  grantedBy,
		}
		bindings = append(bindings, binding)
	}

	if len(bindings) == 0 {
		return nil
	}

	if err := s.PermRepo.BatchCreateRoleBindings(bindings); err != nil {
		return err
	}

	// 清除所有用户的权限缓存
	for _, userId := range userIds {
		s.PermRepo.ClearUserPermissionsCache(userId)
	}

	return nil
}

// BatchRemoveUserFromResource 批量移除用户在资源的权限
func (s *PermissionService) BatchRemoveUserFromResource(userIds []string, scope, resourceId string) error {
	for _, userId := range userIds {
		if err := s.PermRepo.DeleteUserRoleBindingsByResource(userId, scope, resourceId); err != nil {
			return err
		}
		s.PermRepo.ClearUserPermissionsCache(userId)
	}
	return nil
}

// RemoveAllBindingsForResource 删除资源的所有角色绑定
func (s *PermissionService) RemoveAllBindingsForResource(scope, resourceId string) error {
	// 获取所有相关用户
	bindings, _, err := s.PermRepo.ListRoleBindings(&model.RoleBindingQuery{
		Scope:      scope,
		ResourceId: resourceId,
	})
	if err != nil {
		return err
	}

	// 删除绑定
	if err := s.PermRepo.DeleteRoleBindingsByResource(scope, resourceId); err != nil {
		return err
	}

	// 清除用户缓存
	userSet := make(map[string]bool)
	for _, binding := range bindings {
		userSet[binding.UserId] = true
	}
	for userId := range userSet {
		s.PermRepo.ClearUserPermissionsCache(userId)
	}

	return nil
}

// ============ 超级管理员管理 ============

// SetSuperAdmin 设置用户为超级管理员
func (s *PermissionService) SetSuperAdmin(userId string, isSuperAdmin bool) error {
	var user model.User
	if err := s.Ctx.DBSession().Where("user_id = ?", userId).First(&user).Error; err != nil {
		return errors.New("用户不存在")
	}

	flag := 0
	if isSuperAdmin {
		flag = 1
	}

	user.IsSuperAdmin = flag
	if err := s.UserRepo.UpdateUser(userId, &user); err != nil {
		return err
	}

	// 清除权限缓存
	s.PermRepo.ClearUserPermissionsCache(userId)

	return nil
}

// GetSuperAdminList 获取所有超级管理员列表
func (s *PermissionService) GetSuperAdminList() ([]model.User, error) {
	var admins []model.User
	err := s.Ctx.DBSession().Where("is_superadmin = ?", 1).Find(&admins).Error
	return admins, err
}

// ============ 辅助方法 ============

// validateResource 验证资源是否存在
func (s *PermissionService) validateResource(scope string, resourceId *string) error {
	// Platform 级别不需要资源ID
	if scope == model.ScopePlatform {
		if resourceId != nil && *resourceId != "" {
			return errors.New("Platform级别不应指定资源ID")
		}
		return nil
	}

	// 其他级别必须有资源ID
	if resourceId == nil || *resourceId == "" {
		return fmt.Errorf("%s级别必须指定资源ID", scope)
	}

	// 验证资源存在
	switch scope {
	case model.ScopeOrganization:
		// 检查组织是否存在
		var count int64
		if err := s.Ctx.DBSession().Table("t_organization").Where("org_id = ?", *resourceId).Count(&count).Error; err != nil || count == 0 {
			return errors.New("组织不存在")
		}
	case model.ScopeTeam:
		// 检查团队是否存在
		var count int64
		if err := s.Ctx.DBSession().Table("t_team").Where("team_id = ?", *resourceId).Count(&count).Error; err != nil || count == 0 {
			return errors.New("团队不存在")
		}
	case model.ScopeProject:
		// 检查项目是否存在
		var count int64
		if err := s.Ctx.DBSession().Table("t_project").Where("project_id = ?", *resourceId).Count(&count).Error; err != nil || count == 0 {
			return errors.New("项目不存在")
		}
	default:
		return fmt.Errorf("无效的scope: %s", scope)
	}

	return nil
}

// GetUserAccessibleOrganizations 获取用户可访问的组织列表
func (s *PermissionService) GetUserAccessibleOrganizations(userId string) ([]map[string]interface{}, error) {
	resources, err := s.PermRepo.GetAccessibleResources(userId)
	if err != nil {
		return nil, err
	}

	if len(resources.Organizations) == 0 {
		return []map[string]interface{}{}, nil
	}

	var orgs []map[string]interface{}
	err = s.Ctx.DBSession().Table("t_organization").Where("org_id IN ?", resources.Organizations).Find(&orgs).Error
	return orgs, err
}

// GetUserAccessibleTeams 获取用户可访问的团队列表
func (s *PermissionService) GetUserAccessibleTeams(userId string, orgId string) ([]map[string]interface{}, error) {
	resources, err := s.PermRepo.GetAccessibleResources(userId)
	if err != nil {
		return nil, err
	}

	if len(resources.Teams) == 0 {
		return []map[string]interface{}{}, nil
	}

	query := s.Ctx.DBSession().Table("t_team").Where("team_id IN ?", resources.Teams)
	if orgId != "" {
		query = query.Where("org_id = ?", orgId)
	}

	var teams []map[string]interface{}
	err = query.Find(&teams).Error
	return teams, err
}

// GetUserAccessibleProjects 获取用户可访问的项目列表
func (s *PermissionService) GetUserAccessibleProjects(userId string, orgId string) ([]map[string]interface{}, error) {
	resources, err := s.PermRepo.GetAccessibleResources(userId)
	if err != nil {
		return nil, err
	}

	if len(resources.Projects) == 0 {
		return []map[string]interface{}{}, nil
	}

	query := s.Ctx.DBSession().Table("t_project").Where("project_id IN ?", resources.Projects)
	if orgId != "" {
		query = query.Where("org_id = ?", orgId)
	}

	var projects []map[string]interface{}
	err = query.Find(&projects).Error
	return projects, err
}

// ClearAllUserPermissionsCache 清除所有用户权限缓存（用于角色权限变更时）
func (s *PermissionService) ClearAllUserPermissionsCache() error {
	// 获取所有用户
	var users []*model.User
	if err := s.Ctx.DBSession().Select("user_id").Find(&users).Error; err != nil {
		return err
	}

	// 清除缓存
	for _, user := range users {
		s.PermRepo.ClearUserPermissionsCache(user.UserId)
	}

	return nil
}

// GetRole 获取角色信息
func (s *PermissionService) GetRole(roleId string) (*model.Role, error) {
	return s.RoleRepo.GetRole(roleId)
}

// mapTeamRoleToProjectRole 将团队角色映射到项目角色
func (s *PermissionService) mapTeamRoleToProjectRole(teamRoleId string, accessLevel string) string {
	// 根据团队角色和访问级别映射到项目角色
	switch teamRoleId {
	case model.BuiltinTeamOwner:
		return model.BuiltinProjectOwner
	case model.BuiltinTeamMaintainer:
		if accessLevel == "write" {
			return model.BuiltinProjectMaintainer
		}
		return model.BuiltinProjectDeveloper
	case model.BuiltinTeamDeveloper:
		return model.BuiltinProjectDeveloper
	case model.BuiltinTeamReporter:
		return model.BuiltinProjectReporter
	case model.BuiltinTeamGuest:
		return model.BuiltinProjectGuest
	default:
		return model.BuiltinProjectGuest
	}
}
