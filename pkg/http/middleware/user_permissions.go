package middleware

import (
	"fmt"

	"github.com/go-arcade/arcade/internal/engine/service"
	"github.com/go-arcade/arcade/pkg/http"
	"github.com/go-arcade/arcade/pkg/http/jwt"
	"github.com/go-arcade/arcade/pkg/log"
	"github.com/gofiber/fiber/v2"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2025/01/14
 * @file: user_permissions.go
 * @description: 用户权限信息中间件 - 在请求中注入用户的完整权限信息
 */

// UserPermissionsMiddleware 用户权限信息中间件
// 作用：在请求处理前查询用户的完整权限并注入到context中
func UserPermissionsMiddleware(userPermSvc *service.UserPermissionsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 获取用户信息
		claims, ok := c.Locals("claims").(*jwt.AuthClaims)
		if !ok || claims == nil {
			// 如果没有认证信息，继续执行（可能是公开接口）
			return c.Next()
		}

		userId := claims.UserId

		// 查询用户的完整权限信息
		permissionsSummary, err := userPermSvc.GetUserPermissions(c.Context(), userId)
		if err != nil {
			log.Warnf("[UserPermissions] failed to get user permissions for %s: %v", userId, err)
			// 继续执行，但不注入权限信息
			return c.Next()
		}

		// 将权限信息注入到context中，供后续handler使用
		c.Locals("userPermissions", permissionsSummary)
		c.Locals("userAllPermissions", permissionsSummary.AllPermissions)
		c.Locals("userAccessibleRoutes", permissionsSummary.AccessibleRoutes)

		log.Debugf("[UserPermissions] user %s has %d permissions, %d routes",
			userId, len(permissionsSummary.AllPermissions), len(permissionsSummary.AccessibleRoutes))

		return c.Next()
	}
}

// GetUserPermissions 路由处理器 - 返回当前用户的权限信息
func GetUserPermissionsHandler(userPermSvc *service.UserPermissionsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 获取用户信息
		claims, ok := c.Locals("claims").(*jwt.AuthClaims)
		if !ok || claims == nil {
			return http.WithRepErrMsg(c, http.Unauthorized.Code, "user not authenticated", c.Path())
		}

		userId := claims.UserId

		// 查询用户的完整权限信息
		permissionsSummary, err := userPermSvc.GetUserPermissions(c.Context(), userId)
		if err != nil {
			log.Errorf("[UserPermissions] failed to get user permissions for %s: %v", userId, err)
			return http.WithRepErrMsg(c, http.InternalError.Code, "failed to get permissions", c.Path())
		}

		return http.WithRepJSON(c, permissionsSummary)
	}
}

// GetUserAccessibleRoutesHandler 路由处理器 - 返回当前用户可访问的路由列表
func GetUserAccessibleRoutesHandler(userPermSvc *service.UserPermissionsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 获取用户信息
		claims, ok := c.Locals("claims").(*jwt.AuthClaims)
		if !ok || claims == nil {
			return http.WithRepErrMsg(c, http.Unauthorized.Code, "user not authenticated", c.Path())
		}

		userId := claims.UserId

		// 查询用户的完整权限信息
		permissionsSummary, err := userPermSvc.GetUserPermissions(c.Context(), userId)
		if err != nil {
			log.Errorf("[UserPermissions] failed to get user permissions for %s: %v", userId, err)
			return http.WithRepErrMsg(c, http.InternalError.Code, "failed to get permissions", c.Path())
		}

		// 只返回可访问的路由
		return http.WithRepJSON(c, fiber.Map{
			"routes": permissionsSummary.AccessibleRoutes,
			"menu":   groupRoutesByCategory(permissionsSummary.AccessibleRoutes),
		})
	}
}

// GetUserPermissionsSummaryHandler 路由处理器 - 返回用户权限摘要（简化版）
func GetUserPermissionsSummaryHandler(userPermSvc *service.UserPermissionsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 获取用户信息
		claims, ok := c.Locals("claims").(*jwt.AuthClaims)
		if !ok || claims == nil {
			return http.WithRepErrMsg(c, http.Unauthorized.Code, "user not authenticated", c.Path())
		}

		userId := claims.UserId

		// 查询用户的完整权限信息
		permissionsSummary, err := userPermSvc.GetUserPermissions(c.Context(), userId)
		if err != nil {
			log.Errorf("[UserPermissions] failed to get user permissions for %s: %v", userId, err)
			return http.WithRepErrMsg(c, http.InternalError.Code, "failed to get permissions", c.Path())
		}

		// 返回简化的摘要信息
		summary := fiber.Map{
			"userId":               userId,
			"organizationCount":    len(permissionsSummary.Organizations),
			"teamCount":            len(permissionsSummary.Teams),
			"projectCount":         len(permissionsSummary.Projects),
			"permissionCount":      len(permissionsSummary.AllPermissions),
			"accessibleRouteCount": len(permissionsSummary.AccessibleRoutes),
			"accessibleResources":  permissionsSummary.AccessibleResources,
		}

		return http.WithRepJSON(c, summary)
	}
}

// RequireAnyPermission 中间件 - 要求用户拥有任意一个指定权限
func RequireAnyPermission(userPermSvc *service.UserPermissionsService, permissions []string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 获取用户信息
		claims, ok := c.Locals("claims").(*jwt.AuthClaims)
		if !ok || claims == nil {
			return http.WithRepErrMsg(c, http.Unauthorized.Code, "user not authenticated", c.Path())
		}

		userId := claims.UserId

		// 检查权限
		hasPermission, err := userPermSvc.HasAnyPermission(c.Context(), userId, permissions)
		if err != nil {
			log.Errorf("[UserPermissions] failed to check permissions for %s: %v", userId, err)
			return http.WithRepErrMsg(c, http.InternalError.Code, "failed to check permissions", c.Path())
		}

		if !hasPermission {
			return http.WithRepErrMsg(c, http.Forbidden.Code,
				"insufficient permissions: require any of "+fmt.Sprint(permissions), c.Path())
		}

		return c.Next()
	}
}

// RequireAllPermissions 中间件 - 要求用户拥有所有指定权限
func RequireAllPermissions(userPermSvc *service.UserPermissionsService, permissions []string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 获取用户信息
		claims, ok := c.Locals("claims").(*jwt.AuthClaims)
		if !ok || claims == nil {
			return http.WithRepErrMsg(c, http.Unauthorized.Code, "user not authenticated", c.Path())
		}

		userId := claims.UserId

		// 检查权限
		hasPermission, err := userPermSvc.HasAllPermissions(c.Context(), userId, permissions)
		if err != nil {
			log.Errorf("[UserPermissions] failed to check permissions for %s: %v", userId, err)
			return http.WithRepErrMsg(c, http.InternalError.Code, "failed to check permissions", c.Path())
		}

		if !hasPermission {
			return http.WithRepErrMsg(c, http.Forbidden.Code,
				"insufficient permissions: require all of "+fmt.Sprint(permissions), c.Path())
		}

		return c.Next()
	}
}

// groupRoutesByCategory 将路由按分类分组（用于生成菜单）
func groupRoutesByCategory(routes []service.AccessibleRoute) map[string][]service.AccessibleRoute {
	grouped := make(map[string][]service.AccessibleRoute)

	for _, route := range routes {
		if !route.IsMenu {
			continue // 只包含菜单路由
		}

		category := route.Category
		if category == "" {
			category = "其他"
		}

		grouped[category] = append(grouped[category], route)
	}

	return grouped
}

// 辅助函数：从context中获取用户权限信息
func GetUserPermissionsFromContext(c *fiber.Ctx) (*service.UserPermissionSummary, bool) {
	perms, ok := c.Locals("userPermissions").(*service.UserPermissionSummary)
	return perms, ok
}

// 辅助函数：检查用户是否有指定权限（从context中获取）
func HasPermissionInContext(c *fiber.Ctx, permission string) bool {
	perms, ok := GetUserPermissionsFromContext(c)
	if !ok {
		return false
	}

	for _, p := range perms.AllPermissions {
		if p == permission {
			return true
		}
	}

	return false
}
