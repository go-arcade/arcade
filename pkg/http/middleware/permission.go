package middleware

import (
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/observabil/arcade/internal/engine/model"
	"github.com/observabil/arcade/internal/engine/service"
	"github.com/observabil/arcade/pkg/ctx"
	"github.com/observabil/arcade/pkg/http"
	"github.com/observabil/arcade/pkg/http/jwt"
)

// PermissionMiddlewareV2 权限验证中间件（V2版本，支持四层权限模型）
func PermissionMiddlewareV2(appCtx *ctx.Context, permissionCode string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 获取用户ID
		userId, exists := c.Locals("user_id").(string)
		if !exists || userId == "" {
			return http.WithRepErrMsg(c, http.Unauthorized.Code, "未登录", c.Path())
		}

		// 获取资源ID和作用域
		resourceId, scope := extractResourceAndScope(c)

		// 创建权限服务
		permService := service.NewPermissionService(appCtx)

		// 检查权限
		req := &model.PermissionCheckRequest{
			UserId:         userId,
			PermissionCode: permissionCode,
			ResourceId:     resourceId,
			Scope:          scope,
		}

		resp, err := permService.CheckPermission(req)
		if err != nil {
			return http.WithRepErrMsg(c, http.InternalError.Code, "权限验证失败", c.Path())
		}

		if !resp.HasPermission {
			return http.WithRepErrMsg(c, http.Forbidden.Code, resp.Reason, c.Path())
		}

		// 将资源ID和作用域存入上下文
		c.Locals("resource_id", resourceId)
		c.Locals("scope", scope)

		return c.Next()
	}
}

// RequireSuperAdmin 需要超级管理员权限的中间件
func RequireSuperAdmin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 获取用户ID（从 JWT claims 中获取）
		userId, exists := c.Locals("user_id").(string)
		if !exists || userId == "" {
			// 尝试从 claims 中获取
			claims, ok := c.Locals("claims").(*jwt.AuthClaims)
			if !ok || claims == nil {
				return http.WithRepErrMsg(c, http.Unauthorized.Code, "未登录", c.Path())
			}
			userId = claims.UserId
		}

		// 从 Locals 获取 Context（应该在路由初始化时设置）
		appCtx, ok := c.Locals("appCtx").(*ctx.Context)
		if !ok || appCtx == nil {
			return http.WithRepErrMsg(c, http.InternalError.Code, "系统配置错误", c.Path())
		}

		permService := service.NewPermissionService(appCtx)

		// 获取用户权限
		userPerms, err := permService.GetUserPermissions(userId)
		if err != nil {
			return http.WithRepErrMsg(c, http.InternalError.Code, "权限验证失败", c.Path())
		}

		if !userPerms.IsSuperAdmin {
			return http.WithRepErrMsg(c, http.Forbidden.Code, "需要超级管理员权限", c.Path())
		}

		return c.Next()
	}
}

// RequireOrganizationAccess 需要组织访问权限的中间件
func RequireOrganizationAccess(permissionCode string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		orgId := c.Params("orgId")
		if orgId == "" {
			orgId = c.Query("orgId")
		}

		if orgId == "" {
			return http.WithRepErrMsg(c, http.BadRequest.Code, "缺少组织ID", c.Path())
		}

		userId, _ := c.Locals("user_id").(string)

		// 从 Locals 获取 Context
		appCtx, ok := c.Locals("appCtx").(*ctx.Context)
		if !ok || appCtx == nil {
			return http.WithRepErrMsg(c, http.InternalError.Code, "系统配置错误", c.Path())
		}

		permService := service.NewPermissionService(appCtx)

		// 检查权限
		req := &model.PermissionCheckRequest{
			UserId:         userId,
			PermissionCode: permissionCode,
			ResourceId:     orgId,
			Scope:          model.ScopeOrganization,
		}

		resp, err := permService.CheckPermission(req)
		if err != nil {
			return http.WithRepErrMsg(c, http.InternalError.Code, "权限验证失败", c.Path())
		}

		if !resp.HasPermission {
			return http.WithRepErrMsg(c, http.Forbidden.Code, "无权访问该组织", c.Path())
		}

		c.Locals("org_id", orgId)
		return c.Next()
	}
}

// RequireTeamAccess 需要团队访问权限的中间件
func RequireTeamAccess(permissionCode string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		teamId := c.Params("teamId")
		if teamId == "" {
			teamId = c.Query("teamId")
		}

		if teamId == "" {
			return http.WithRepErrMsg(c, http.BadRequest.Code, "缺少团队ID", c.Path())
		}

		userId, _ := c.Locals("user_id").(string)

		// 从 Locals 获取 Context
		appCtx, ok := c.Locals("appCtx").(*ctx.Context)
		if !ok || appCtx == nil {
			return http.WithRepErrMsg(c, http.InternalError.Code, "系统配置错误", c.Path())
		}

		permService := service.NewPermissionService(appCtx)

		// 检查权限
		req := &model.PermissionCheckRequest{
			UserId:         userId,
			PermissionCode: permissionCode,
			ResourceId:     teamId,
			Scope:          model.ScopeTeam,
		}

		resp, err := permService.CheckPermission(req)
		if err != nil {
			return http.WithRepErrMsg(c, http.InternalError.Code, "权限验证失败", c.Path())
		}

		if !resp.HasPermission {
			return http.WithRepErrMsg(c, http.Forbidden.Code, "无权访问该团队", c.Path())
		}

		c.Locals("team_id", teamId)
		return c.Next()
	}
}

// RequireProjectAccess 需要项目访问权限的中间件
func RequireProjectAccess(permissionCode string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		projectId := c.Params("projectId")
		if projectId == "" {
			projectId = c.Query("projectId")
		}

		if projectId == "" {
			return http.WithRepErrMsg(c, http.BadRequest.Code, "缺少项目ID", c.Path())
		}

		userId, _ := c.Locals("user_id").(string)

		// 从 Locals 获取 Context
		appCtx, ok := c.Locals("appCtx").(*ctx.Context)
		if !ok || appCtx == nil {
			return http.WithRepErrMsg(c, http.InternalError.Code, "系统配置错误", c.Path())
		}

		permService := service.NewPermissionService(appCtx)

		// 检查权限
		req := &model.PermissionCheckRequest{
			UserId:         userId,
			PermissionCode: permissionCode,
			ResourceId:     projectId,
			Scope:          model.ScopeProject,
		}

		resp, err := permService.CheckPermission(req)
		if err != nil {
			return http.WithRepErrMsg(c, http.InternalError.Code, "权限验证失败", c.Path())
		}

		if !resp.HasPermission {
			return http.WithRepErrMsg(c, http.Forbidden.Code, "无权访问该项目", c.Path())
		}

		c.Locals("project_id", projectId)
		return c.Next()
	}
}

// extractResourceAndScope 从请求路径中提取资源ID和作用域
func extractResourceAndScope(c *fiber.Ctx) (string, string) {
	path := c.Path()

	// 解码URL
	path, _ = url.QueryUnescape(path)

	// 按优先级检查路径参数
	// 1. 检查 projectId (最具体)
	if projectId := c.Params("projectId"); projectId != "" {
		return projectId, model.ScopeProject
	}

	// 2. 检查 teamId
	if teamId := c.Params("teamId"); teamId != "" {
		return teamId, model.ScopeTeam
	}

	// 3. 检查 orgId
	if orgId := c.Params("orgId"); orgId != "" {
		return orgId, model.ScopeOrganization
	}

	// 4. 从查询参数检查
	if projectId := c.Query("projectId"); projectId != "" {
		return projectId, model.ScopeProject
	}
	if teamId := c.Query("teamId"); teamId != "" {
		return teamId, model.ScopeTeam
	}
	if orgId := c.Query("orgId"); orgId != "" {
		return orgId, model.ScopeOrganization
	}

	// 5. 根据路径模式判断
	if strings.Contains(path, "/platform") {
		return "", model.ScopePlatform
	}
	if strings.Contains(path, "/projects/") {
		return "", model.ScopeProject
	}
	if strings.Contains(path, "/teams/") {
		return "", model.ScopeTeam
	}
	if strings.Contains(path, "/orgs/") {
		return "", model.ScopeOrganization
	}

	// 默认为 platform 级别
	return "", model.ScopePlatform
}

// HasPermission 在Handler中检查权限的辅助函数
func HasPermission(c *fiber.Ctx, permissionCode, resourceId, scope string) bool {
	userId, exists := c.Locals("user_id").(string)
	if !exists {
		return false
	}

	// 从 Locals 获取 Context
	appCtx, ok := c.Locals("appCtx").(*ctx.Context)
	if !ok || appCtx == nil {
		return false
	}

	permService := service.NewPermissionService(appCtx)

	// 检查权限
	req := &model.PermissionCheckRequest{
		UserId:         userId,
		PermissionCode: permissionCode,
		ResourceId:     resourceId,
		Scope:          scope,
	}

	resp, err := permService.CheckPermission(req)
	if err != nil {
		return false
	}

	return resp.HasPermission
}

// GetUserPermissions 获取当前用户权限
func GetUserPermissions(c *fiber.Ctx) (*model.UserPermissions, error) {
	userId, exists := c.Locals("user_id").(string)
	if !exists {
		return nil, nil
	}

	// 从 Locals 获取 Context
	appCtx, ok := c.Locals("appCtx").(*ctx.Context)
	if !ok || appCtx == nil {
		return nil, nil
	}

	permService := service.NewPermissionService(appCtx)

	return permService.GetUserPermissionsWithRoutes(userId)
}
