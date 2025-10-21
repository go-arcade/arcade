package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/observabil/arcade/internal/engine/model"
	"github.com/observabil/arcade/internal/engine/service"
	"github.com/observabil/arcade/pkg/ctx"
	"github.com/observabil/arcade/pkg/http"
)

// PermissionHandler 权限管理处理器
type PermissionHandler struct {
	Ctx *ctx.Context
}

func NewPermissionHandler(ctx *ctx.Context) *PermissionHandler {
	return &PermissionHandler{
		Ctx: ctx,
	}
}

// ============ 用户权限查询 ============

// GetUserPermissions 获取当前用户完整权限
func (h *PermissionHandler) GetUserPermissions(c *fiber.Ctx) error {
	userId := c.Locals("user_id").(string)

	permService := service.NewPermissionService(h.Ctx)
	perms, err := permService.GetUserPermissionsWithRoutes(userId)
	if err != nil {
		return http.WithRepErrMsg(c, http.InternalError.Code, err.Error(), c.Path())
	}

	return http.WithRepJSON(c, perms)
}

// GetUserRoutes 获取当前用户可访问的路由
func (h *PermissionHandler) GetUserRoutes(c *fiber.Ctx) error {
	userId := c.Locals("user_id").(string)

	permService := service.NewPermissionService(h.Ctx)
	perms, err := permService.GetUserPermissionsWithRoutes(userId)
	if err != nil {
		return http.WithRepErrMsg(c, http.InternalError.Code, err.Error(), c.Path())
	}

	return http.WithRepJSON(c, fiber.Map{
		"routes": perms.AccessibleRoutes,
		"menu":   h.buildMenuTree(perms.AccessibleRoutes),
	})
}

// GetAccessibleResources 获取当前用户可访问的资源
func (h *PermissionHandler) GetAccessibleResources(c *fiber.Ctx) error {
	userId := c.Locals("user_id").(string)

	permService := service.NewPermissionService(h.Ctx)
	resources, err := permService.PermRepo.GetAccessibleResources(userId)
	if err != nil {
		return http.WithRepErrMsg(c, http.InternalError.Code, err.Error(), c.Path())
	}

	return http.WithRepJSON(c, resources)
}

// ============ 权限检查 ============

// CheckPermission 检查权限
func (h *PermissionHandler) CheckPermission(c *fiber.Ctx) error {
	var req model.PermissionCheckRequest
	if err := c.BodyParser(&req); err != nil {
		return http.WithRepErrMsg(c, http.BadRequest.Code, "请求参数错误", c.Path())
	}

	// 设置当前用户ID
	req.UserId = c.Locals("user_id").(string)

	permService := service.NewPermissionService(h.Ctx)
	resp, err := permService.CheckPermission(&req)
	if err != nil {
		return http.WithRepErrMsg(c, http.InternalError.Code, err.Error(), c.Path())
	}

	return http.WithRepJSON(c, resp)
}

// ============ 角色绑定管理 ============

// CreateRoleBinding 创建角色绑定
func (h *PermissionHandler) CreateRoleBinding(c *fiber.Ctx) error {
	var req model.UserRoleBindingCreate
	if err := c.BodyParser(&req); err != nil {
		return http.WithRepErrMsg(c, http.BadRequest.Code, "请求参数错误", c.Path())
	}

	// 设置授予人
	grantedBy := c.Locals("user_id").(string)
	req.GrantedBy = &grantedBy

	permService := service.NewPermissionService(h.Ctx)
	binding, err := permService.CreateRoleBinding(&req)
	if err != nil {
		return http.WithRepErrMsg(c, http.InternalError.Code, err.Error(), c.Path())
	}

	return http.WithRepJSON(c, binding)
}

// DeleteRoleBinding 删除角色绑定
func (h *PermissionHandler) DeleteRoleBinding(c *fiber.Ctx) error {
	bindingId := c.Params("bindingId")
	if bindingId == "" {
		return http.WithRepErrMsg(c, http.BadRequest.Code, "缺少绑定ID", c.Path())
	}

	permService := service.NewPermissionService(h.Ctx)
	if err := permService.DeleteRoleBinding(bindingId); err != nil {
		return http.WithRepErrMsg(c, http.InternalError.Code, err.Error(), c.Path())
	}

	return http.WithRepJSON(c, fiber.Map{"message": "删除成功"})
}

// GetUserRoleBindings 获取用户的角色绑定
func (h *PermissionHandler) GetUserRoleBindings(c *fiber.Ctx) error {
	userId := c.Params("userId")
	if userId == "" {
		return http.WithRepErrMsg(c, http.BadRequest.Code, "缺少用户ID", c.Path())
	}

	permService := service.NewPermissionService(h.Ctx)
	bindings, err := permService.GetUserRoleBindings(userId)
	if err != nil {
		return http.WithRepErrMsg(c, http.InternalError.Code, err.Error(), c.Path())
	}

	return http.WithRepJSON(c, bindings)
}

// ListRoleBindings 列出角色绑定
func (h *PermissionHandler) ListRoleBindings(c *fiber.Ctx) error {
	query := &model.RoleBindingQuery{}

	// 解析查询参数
	if userId := c.Query("userId"); userId != "" {
		query.UserId = userId
	}
	if roleId := c.Query("roleId"); roleId != "" {
		query.RoleId = roleId
	}
	if scope := c.Query("scope"); scope != "" {
		query.Scope = scope
	}
	if resourceId := c.Query("resourceId"); resourceId != "" {
		query.ResourceId = resourceId
	}

	// 分页参数
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			query.Page = page
		}
	}
	if pageSizeStr := c.Query("pageSize"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil {
			query.PageSize = pageSize
		}
	}

	permService := service.NewPermissionService(h.Ctx)
	bindings, total, err := permService.ListRoleBindings(query)
	if err != nil {
		return http.WithRepErrMsg(c, http.InternalError.Code, err.Error(), c.Path())
	}

	return http.WithRepJSON(c, fiber.Map{
		"list":     bindings,
		"total":    total,
		"page":     query.Page,
		"pageSize": query.PageSize,
	})
}

// ============ 批量操作 ============

// BatchAssignRole 批量分配角色
func (h *PermissionHandler) BatchAssignRole(c *fiber.Ctx) error {
	var req struct {
		UserIds    []string `json:"userIds" validate:"required"`
		RoleId     string   `json:"roleId" validate:"required"`
		Scope      string   `json:"scope" validate:"required"`
		ResourceId *string  `json:"resourceId"`
	}

	if err := c.BodyParser(&req); err != nil {
		return http.WithRepErrMsg(c, http.BadRequest.Code, "请求参数错误", c.Path())
	}

	grantedBy := c.Locals("user_id").(string)

	permService := service.NewPermissionService(h.Ctx)
	if err := permService.BatchAssignRoleToUsers(req.UserIds, req.RoleId, req.Scope, req.ResourceId, &grantedBy); err != nil {
		return http.WithRepErrMsg(c, http.InternalError.Code, err.Error(), c.Path())
	}

	return http.WithRepJSON(c, fiber.Map{"message": "批量分配成功"})
}

// BatchRemoveRole 批量移除角色
func (h *PermissionHandler) BatchRemoveRole(c *fiber.Ctx) error {
	var req struct {
		UserIds    []string `json:"userIds" validate:"required"`
		Scope      string   `json:"scope" validate:"required"`
		ResourceId string   `json:"resourceId" validate:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		return http.WithRepErrMsg(c, http.BadRequest.Code, "请求参数错误", c.Path())
	}

	permService := service.NewPermissionService(h.Ctx)
	if err := permService.BatchRemoveUserFromResource(req.UserIds, req.Scope, req.ResourceId); err != nil {
		return http.WithRepErrMsg(c, http.InternalError.Code, err.Error(), c.Path())
	}

	return http.WithRepJSON(c, fiber.Map{"message": "批量移除成功"})
}

// ============ 超级管理员管理 ============

// SetSuperAdmin 设置超级管理员
func (h *PermissionHandler) SetSuperAdmin(c *fiber.Ctx) error {
	userId := c.Params("userId")
	if userId == "" {
		return http.WithRepErrMsg(c, http.BadRequest.Code, "缺少用户ID", c.Path())
	}

	var req struct {
		IsSuperAdmin bool `json:"isSuperAdmin"`
	}

	if err := c.BodyParser(&req); err != nil {
		return http.WithRepErrMsg(c, http.BadRequest.Code, "请求参数错误", c.Path())
	}

	permService := service.NewPermissionService(h.Ctx)
	if err := permService.SetSuperAdmin(userId, req.IsSuperAdmin); err != nil {
		return http.WithRepErrMsg(c, http.InternalError.Code, err.Error(), c.Path())
	}

	return http.WithRepJSON(c, fiber.Map{"message": "设置成功"})
}

// GetSuperAdminList 获取超级管理员列表
func (h *PermissionHandler) GetSuperAdminList(c *fiber.Ctx) error {
	permService := service.NewPermissionService(h.Ctx)
	admins, err := permService.GetSuperAdminList()
	if err != nil {
		return http.WithRepErrMsg(c, http.InternalError.Code, err.Error(), c.Path())
	}

	return http.WithRepJSON(c, admins)
}

// ============ 资源访问 ============

// GetUserOrganizations 获取用户可访问的组织
func (h *PermissionHandler) GetUserOrganizations(c *fiber.Ctx) error {
	userId := c.Locals("user_id").(string)

	permService := service.NewPermissionService(h.Ctx)
	orgs, err := permService.GetUserAccessibleOrganizations(userId)
	if err != nil {
		return http.WithRepErrMsg(c, http.InternalError.Code, err.Error(), c.Path())
	}

	return http.WithRepJSON(c, orgs)
}

// GetUserTeams 获取用户可访问的团队
func (h *PermissionHandler) GetUserTeams(c *fiber.Ctx) error {
	userId := c.Locals("user_id").(string)
	orgId := c.Query("orgId")

	permService := service.NewPermissionService(h.Ctx)
	teams, err := permService.GetUserAccessibleTeams(userId, orgId)
	if err != nil {
		return http.WithRepErrMsg(c, http.InternalError.Code, err.Error(), c.Path())
	}

	return http.WithRepJSON(c, teams)
}

// GetUserProjects 获取用户可访问的项目
func (h *PermissionHandler) GetUserProjects(c *fiber.Ctx) error {
	userId := c.Locals("user_id").(string)
	orgId := c.Query("orgId")

	permService := service.NewPermissionService(h.Ctx)
	projects, err := permService.GetUserAccessibleProjects(userId, orgId)
	if err != nil {
		return http.WithRepErrMsg(c, http.InternalError.Code, err.Error(), c.Path())
	}

	return http.WithRepJSON(c, projects)
}

// ============ 辅助方法 ============

// buildMenuTree 构建菜单树
func (h *PermissionHandler) buildMenuTree(routes []*model.RouterPermission) map[string]interface{} {
	menu := make(map[string]interface{})

	// 按分类分组
	menuItems := make(map[string][]*model.RouterPermission)
	for _, route := range routes {
		// 只包含菜单路由
		if route.IsMenu == 1 {
			// 按分类分组
			category := route.Category
			if category == "" {
				category = "其他"
			}
			menuItems[category] = append(menuItems[category], route)
		}
	}

	// 构建菜单结构
	for category, items := range menuItems {
		menu[category] = items
	}

	return menu
}
