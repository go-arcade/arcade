package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/observabil/arcade/internal/engine/model"
	"github.com/observabil/arcade/internal/engine/service"
	"github.com/observabil/arcade/pkg/http/middleware"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2025/10/13
 * @file: router_project.go
 * @description: 项目路由（权限中间件使用示例）
 */

// RegisterProjectRoutes 注册项目相关路由
func RegisterProjectRoutes(r fiber.Router, permService *service.PermissionService) {
	projectGroup := r.Group("/projects")
	{
		// ========== 公开接口（不需要权限） ==========
		// 列出公开项目
		projectGroup.Get("/public", listPublicProjects())

		// ========== 需要基础访问权限（任意角色） ==========
		// 查看项目详情 - 要求是项目成员（任意角色即可）
		projectGroup.Get("/:projectId",
			middleware.RequireProject(permService),
			getProjectDetail())

		// 查看项目成员列表 - 要求是项目成员
		projectGroup.Get("/:projectId/members",
			middleware.RequireProject(permService),
			listProjectMembers())

		// ========== 需要访客权限（guest及以上） ==========
		// 查看构建历史
		projectGroup.Get("/:projectId/builds",
			middleware.RequireProjectGuest(permService),
			listBuilds())

		// 下载构建产物 - 要求特定权限点
		projectGroup.Get("/:projectId/builds/:buildId/artifacts",
			middleware.RequirePermission(permService, model.PermBuildArtifact),
			downloadArtifact())

		// ========== 需要报告者权限（reporter及以上） ==========
		// 创建 Issue
		projectGroup.Post("/:projectId/issues",
			middleware.RequireProjectReporter(permService),
			createIssue())

		// ========== 需要开发者权限（developer及以上） ==========
		// 触发构建 - 使用权限点检查
		projectGroup.Post("/:projectId/builds/trigger",
			middleware.RequirePermission(permService, model.PermBuildTrigger),
			triggerBuild())

		// 取消构建
		projectGroup.Post("/:projectId/builds/:buildId/cancel",
			middleware.RequirePermission(permService, model.PermBuildCancel),
			cancelBuild())

		// 创建流水线
		projectGroup.Post("/:projectId/pipelines",
			middleware.RequireProjectDeveloper(permService),
			createPipeline())

		// 触发流水线
		projectGroup.Post("/:projectId/pipelines/:pipelineId/run",
			middleware.RequireCanWrite(permService), // 使用便捷标志检查
			runPipeline())

		// ========== 需要维护者权限（maintainer及以上） ==========
		// 添加项目成员
		projectGroup.Post("/:projectId/members",
			middleware.RequireProjectMaintainer(permService),
			addProjectMember())

		// 更新成员角色
		projectGroup.Put("/:projectId/members/:userId",
			middleware.RequireCanManage(permService),
			updateMemberRole())

		// 移除项目成员
		projectGroup.Delete("/:projectId/members/:userId",
			middleware.RequireProjectMaintainer(permService),
			removeProjectMember())

		// 修改项目设置
		projectGroup.Put("/:projectId/settings",
			middleware.RequirePermission(permService, model.PermProjectSettings),
			updateProjectSettings())

		// 管理项目变量
		projectGroup.Post("/:projectId/variables",
			middleware.RequirePermission(permService, model.PermProjectVariables),
			createVariable())

		// ========== 需要所有者权限（owner） ==========
		// 删除项目
		projectGroup.Delete("/:projectId",
			middleware.RequireProjectOwner(permService),
			deleteProject())

		// 转移项目所有权
		projectGroup.Post("/:projectId/transfer",
			middleware.RequireCanDelete(permService),
			transferProject())

		// ========== 高级用法：组合多个条件 ==========
		// 部署到生产环境 - 要求 maintainer + deploy.approve 权限
		projectGroup.Post("/:projectId/deploy/production",
			middleware.PermissionMiddleware(permService, middleware.PermissionConfig{
				ResourceType:       "project",
				RequiredRole:       model.BuiltinProjectMaintainer,
				RequiredPermission: model.PermDeployApprove,
			}),
			deployToProduction())

		// 紧急回滚 - 自定义检查：优先级 >= 40 且有回滚权限
		projectGroup.Post("/:projectId/deploy/emergency-rollback",
			middleware.PermissionMiddleware(permService, middleware.PermissionConfig{
				ResourceType:       "project",
				RequiredPermission: model.PermDeployRollback,
				CheckFunc: func(p *service.ProjectPermission) bool {
					return p.Priority >= 40
				},
			}),
			emergencyRollback())
	}
}

// ========== Handler 示例 ==========

func listPublicProjects() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "public projects"})
	}
}

func getProjectDetail() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 可以从 c.Locals 获取权限信息
		perm, _ := c.Locals("permission").(*service.ProjectPermission)
		projectId, _ := c.Locals("projectId").(string)

		return c.JSON(fiber.Map{
			"projectId": projectId,
			"role":      perm.RoleName,
			"canWrite":  perm.CanWrite,
		})
	}
}

func listProjectMembers() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "project members"})
	}
}

func listBuilds() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "build list"})
	}
}

func downloadArtifact() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "download artifact"})
	}
}

func createIssue() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "create issue"})
	}
}

func triggerBuild() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "trigger build"})
	}
}

func cancelBuild() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "cancel build"})
	}
}

func createPipeline() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "create pipeline"})
	}
}

func runPipeline() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "run pipeline"})
	}
}

func addProjectMember() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "add member"})
	}
}

func updateMemberRole() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "update member role"})
	}
}

func removeProjectMember() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "remove member"})
	}
}

func updateProjectSettings() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "update settings"})
	}
}

func createVariable() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "create variable"})
	}
}

func deleteProject() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "delete project"})
	}
}

func transferProject() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "transfer project"})
	}
}

func deployToProduction() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 这里已经通过了 maintainer + deploy.approve 双重检查
		return c.JSON(fiber.Map{"message": "deploy to production"})
	}
}

func emergencyRollback() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 这里已经通过了优先级和权限点的检查
		return c.JSON(fiber.Map{"message": "emergency rollback"})
	}
}
