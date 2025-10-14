package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/observabil/arcade/internal/engine/model"
	"github.com/observabil/arcade/internal/engine/service"
	"github.com/observabil/arcade/pkg/http"
	"github.com/observabil/arcade/pkg/http/jwt"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2025/10/13
 * @file: permission.go
 * @description: 统一权限验证中间件
 */

// PermissionConfig 权限配置
type PermissionConfig struct {
	// 需要的资源类型
	ResourceType string // project/org/team

	// 项目权限配置
	RequiredRole       string                                // 要求的角色ID（如 model.BuiltinProjectDeveloper）
	RequiredPermission string                                // 要求的权限点（如 model.PermBuildTrigger）
	CheckFunc          func(*service.ProjectPermission) bool // 自定义检查函数

	// 是否可选（如果资源ID不存在，是否跳过检查）
	Optional bool
}

// PermissionMiddleware 统一权限验证中间件
func PermissionMiddleware(permService *service.PermissionService, config PermissionConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. 获取用户信息
		claims, ok := c.Locals("claims").(*jwt.AuthClaims)
		if !ok || claims == nil {
			http.WithRepErrMsg(c, http.Unauthorized.Code, "user not authenticated", c.Path())
			return fiber.ErrUnauthorized
		}

		// 2. 根据资源类型进行不同的权限检查
		switch config.ResourceType {
		case "project":
			return checkProjectPermission(c, permService, claims.UserId, config)
		case "org":
			return checkOrgPermission(c, permService, claims.UserId, config)
		case "team":
			return checkTeamPermission(c, permService, claims.UserId, config)
		default:
			// 默认为项目权限检查
			return checkProjectPermission(c, permService, claims.UserId, config)
		}
	}
}

// checkProjectPermission 检查项目权限
func checkProjectPermission(c *fiber.Ctx, permService *service.PermissionService, userId string, config PermissionConfig) error {
	// 获取 projectId
	projectId := getResourceId(c, "projectId")
	if projectId == "" {
		if config.Optional {
			return c.Next()
		}
		http.WithRepErrMsg(c, http.BadRequest.Code, "projectId is required", c.Path())
		return fiber.ErrBadRequest
	}

	// 获取权限信息
	perm, err := permService.CheckProjectPermission(c.Context(), userId, projectId)
	if err != nil {
		http.WithRepErrMsg(c, http.Forbidden.Code, err.Error(), c.Path())
		return fiber.ErrForbidden
	}

	if !perm.HasAccess {
		http.WithRepErrMsg(c, http.Forbidden.Code, "access denied", c.Path())
		return fiber.ErrForbidden
	}

	// 检查角色要求
	if config.RequiredRole != "" {
		if err := permService.RequireProjectPermission(c.Context(), userId, projectId, config.RequiredRole); err != nil {
			http.WithRepErrMsg(c, http.Forbidden.Code, err.Error(), c.Path())
			return fiber.ErrForbidden
		}
	}

	// 检查权限点要求
	if config.RequiredPermission != "" {
		if err := permService.RequirePermissionPoint(c.Context(), userId, projectId, config.RequiredPermission); err != nil {
			http.WithRepErrMsg(c, http.Forbidden.Code, err.Error(), c.Path())
			return fiber.ErrForbidden
		}
	}

	// 自定义检查函数
	if config.CheckFunc != nil && !config.CheckFunc(perm) {
		http.WithRepErrMsg(c, http.Forbidden.Code, "insufficient permission", c.Path())
		return fiber.ErrForbidden
	}

	// 将权限信息保存到 context
	c.Locals("permission", perm)
	c.Locals("projectId", projectId)

	return c.Next()
}

// checkOrgPermission 检查组织权限
func checkOrgPermission(c *fiber.Ctx, permService *service.PermissionService, userId string, config PermissionConfig) error {
	orgId := getResourceId(c, "orgId")
	if orgId == "" {
		if config.Optional {
			return c.Next()
		}
		http.WithRepErrMsg(c, http.BadRequest.Code, "orgId is required", c.Path())
		return fiber.ErrBadRequest
	}

	roleId, err := permService.CheckOrganizationPermission(c.Context(), userId, orgId)
	if err != nil {
		http.WithRepErrMsg(c, http.Forbidden.Code, err.Error(), c.Path())
		return fiber.ErrForbidden
	}

	c.Locals("orgRole", roleId)
	c.Locals("orgId", orgId)

	return c.Next()
}

// checkTeamPermission 检查团队权限
func checkTeamPermission(c *fiber.Ctx, permService *service.PermissionService, userId string, config PermissionConfig) error {
	teamId := getResourceId(c, "teamId")
	if teamId == "" {
		if config.Optional {
			return c.Next()
		}
		http.WithRepErrMsg(c, http.BadRequest.Code, "teamId is required", c.Path())
		return fiber.ErrBadRequest
	}

	roleId, err := permService.CheckTeamPermission(c.Context(), userId, teamId)
	if err != nil {
		http.WithRepErrMsg(c, http.Forbidden.Code, err.Error(), c.Path())
		return fiber.ErrForbidden
	}

	c.Locals("teamRole", roleId)
	c.Locals("teamId", teamId)

	return c.Next()
}

// getResourceId 从多个来源获取资源ID（优先级：URL参数 > 查询参数 > 请求体）
func getResourceId(c *fiber.Ctx, paramName string) string {
	// 1. 从路径参数获取
	if id := c.Params(paramName); id != "" {
		return id
	}

	// 2. 从查询参数获取
	if id := c.Query(paramName); id != "" {
		return id
	}

	// 3. 从请求体获取
	var body map[string]interface{}
	if err := c.BodyParser(&body); err == nil {
		if id, ok := body[paramName].(string); ok && id != "" {
			return id
		}
	}

	return ""
}

// ========== 便捷的中间件构造函数 ==========

// RequireProject 要求项目访问权限（基础检查，任意角色）
func RequireProject(permService *service.PermissionService) fiber.Handler {
	return PermissionMiddleware(permService, PermissionConfig{
		ResourceType: "project",
	})
}

// RequireProjectOwner 要求项目所有者权限
func RequireProjectOwner(permService *service.PermissionService) fiber.Handler {
	return PermissionMiddleware(permService, PermissionConfig{
		ResourceType: "project",
		RequiredRole: model.BuiltinProjectOwner,
	})
}

// RequireProjectMaintainer 要求项目维护者权限
func RequireProjectMaintainer(permService *service.PermissionService) fiber.Handler {
	return PermissionMiddleware(permService, PermissionConfig{
		ResourceType: "project",
		RequiredRole: model.BuiltinProjectMaintainer,
	})
}

// RequireProjectDeveloper 要求项目开发者权限
func RequireProjectDeveloper(permService *service.PermissionService) fiber.Handler {
	return PermissionMiddleware(permService, PermissionConfig{
		ResourceType: "project",
		RequiredRole: model.BuiltinProjectDeveloper,
	})
}

// RequireProjectReporter 要求项目报告者权限
func RequireProjectReporter(permService *service.PermissionService) fiber.Handler {
	return PermissionMiddleware(permService, PermissionConfig{
		ResourceType: "project",
		RequiredRole: model.BuiltinProjectReporter,
	})
}

// RequireProjectGuest 要求项目访客权限（基本访问）
func RequireProjectGuest(permService *service.PermissionService) fiber.Handler {
	return PermissionMiddleware(permService, PermissionConfig{
		ResourceType: "project",
		RequiredRole: model.BuiltinProjectGuest,
	})
}

// RequirePermission 要求特定权限点
func RequirePermission(permService *service.PermissionService, permissionPoint string) fiber.Handler {
	return PermissionMiddleware(permService, PermissionConfig{
		ResourceType:       "project",
		RequiredPermission: permissionPoint,
	})
}

// RequireCanWrite 要求写入权限
func RequireCanWrite(permService *service.PermissionService) fiber.Handler {
	return PermissionMiddleware(permService, PermissionConfig{
		ResourceType: "project",
		CheckFunc: func(p *service.ProjectPermission) bool {
			return p.CanWrite
		},
	})
}

// RequireCanManage 要求管理权限
func RequireCanManage(permService *service.PermissionService) fiber.Handler {
	return PermissionMiddleware(permService, PermissionConfig{
		ResourceType: "project",
		CheckFunc: func(p *service.ProjectPermission) bool {
			return p.CanManage
		},
	})
}

// RequireCanDelete 要求删除权限
func RequireCanDelete(permService *service.PermissionService) fiber.Handler {
	return PermissionMiddleware(permService, PermissionConfig{
		ResourceType: "project",
		CheckFunc: func(p *service.ProjectPermission) bool {
			return p.CanDelete
		},
	})
}

// RequireOrgMember 要求组织成员权限
func RequireOrgMember(permService *service.PermissionService) fiber.Handler {
	return PermissionMiddleware(permService, PermissionConfig{
		ResourceType: "org",
	})
}

// RequireTeamMember 要求团队成员权限
func RequireTeamMember(permService *service.PermissionService) fiber.Handler {
	return PermissionMiddleware(permService, PermissionConfig{
		ResourceType: "team",
	})
}
