# Router 中使用权限中间件指南

## 快速开始

### 1. 初始化权限服务

在 `router.go` 中初始化权限服务：

```go
package router

import (
    "github.com/observabil/arcade/internal/engine/service"
    "github.com/observabil/arcade/pkg/http/middleware"
)

type Router struct {
    Http        *httpx.Http
    Ctx         *ctx.Context
    PermService *service.PermissionService // 添加权限服务
}

func NewRouter(httpConf *httpx.Http, ctx *ctx.Context) *Router {
    return &Router{
        Http:        httpConf,
        Ctx:         ctx,
        PermService: service.NewPermissionService(ctx), // 初始化权限服务
    }
}
```

### 2. 在路由中使用

```go
func (rt *Router) routerGroup(r fiber.Router) {
    auth := middleware.AuthorizationMiddleware(...)
    
    // 项目路由（需要认证）
    rt.projectRouter(r, auth)
}

func (rt *Router) projectRouter(r fiber.Router, auth fiber.Handler) {
    // 创建项目路由组，应用认证中间件
    projectGroup := r.Group("/projects", auth)
    {
        // 在这里使用权限中间件
        projectGroup.Get("/:projectId", 
            middleware.RequireProject(rt.PermService), 
            getProjectDetail())
    }
}
```

## 使用模式

### 模式 1: 便捷函数（推荐）

适用于常见场景，代码简洁：

```go
projectGroup := r.Group("/projects", auth) // 先认证
{
    // 基础访问（任意项目成员）
    projectGroup.Get("/:projectId", 
        middleware.RequireProject(permService), 
        handler)

    // 开发者权限
    projectGroup.Post("/:projectId/builds", 
        middleware.RequireProjectDeveloper(permService), 
        handler)

    // 维护者权限
    projectGroup.Put("/:projectId/members", 
        middleware.RequireProjectMaintainer(permService), 
        handler)

    // 所有者权限
    projectGroup.Delete("/:projectId", 
        middleware.RequireProjectOwner(permService), 
        handler)
}
```

### 模式 2: 权限点检查

适用于需要精确控制特定操作权限：

```go
projectGroup := r.Group("/projects", auth)
{
    // 触发构建 - 检查 build.trigger 权限点
    projectGroup.Post("/:projectId/builds/trigger", 
        middleware.RequirePermission(permService, model.PermBuildTrigger), 
        handler)

    // 审批部署 - 检查 deploy.approve 权限点
    projectGroup.Post("/:projectId/deploys/:deployId/approve", 
        middleware.RequirePermission(permService, model.PermDeployApprove), 
        handler)

    // 管理变量 - 检查 project.variables 权限点
    projectGroup.Post("/:projectId/variables", 
        middleware.RequirePermission(permService, model.PermProjectVariables), 
        handler)
}
```

### 模式 3: 权限标志检查

适用于通用的权限能力检查：

```go
projectGroup := r.Group("/projects", auth)
{
    // 需要写入权限的操作
    projectGroup.Post("/:projectId/pipeline", 
        middleware.RequireCanWrite(permService), 
        handler)

    // 需要管理权限的操作
    projectGroup.Put("/:projectId/settings", 
        middleware.RequireCanManage(permService), 
        handler)

    // 需要删除权限的操作
    projectGroup.Delete("/:projectId", 
        middleware.RequireCanDelete(permService), 
        handler)
}
```

### 模式 4: 统一中间件（高级）

适用于需要组合多个条件或自定义逻辑：

```go
projectGroup := r.Group("/projects", auth)
{
    // 组合条件：要求 maintainer + deploy.approve 权限
    projectGroup.Post("/:projectId/deploy/production", 
        middleware.PermissionMiddleware(permService, middleware.PermissionConfig{
            ResourceType:       "project",
            RequiredRole:       model.BuiltinProjectMaintainer,
            RequiredPermission: model.PermDeployApprove,
        }), 
        deployToProduction())

    // 自定义检查：优先级 >= 40 且有回滚权限
    projectGroup.Post("/:projectId/deploy/rollback", 
        middleware.PermissionMiddleware(permService, middleware.PermissionConfig{
            ResourceType: "project",
            CheckFunc: func(p *service.ProjectPermission) bool {
                return p.Priority >= 40 && 
                       p.HasPermissionPoint(model.PermDeployRollback)
            },
        }), 
        emergencyRollback())

    // 可选权限检查：如果没有 projectId 就跳过
    projectGroup.Get("/search", 
        middleware.PermissionMiddleware(permService, middleware.PermissionConfig{
            ResourceType: "project",
            Optional:     true, // 可选的
        }), 
        searchProjects())
}
```

## 完整示例

### 项目路由完整示例

```go
func (rt *Router) projectRouter(r fiber.Router, auth fiber.Handler) {
    // 创建项目路由组
    projectGroup := r.Group("/projects", auth)
    
    // 初始化权限服务
    permService := rt.PermService
    
    {
        // ===== 查看相关（guest 及以上）=====
        projectGroup.Get("/:projectId", 
            middleware.RequireProjectGuest(permService), 
            rt.getProject)
        
        projectGroup.Get("/:projectId/pipelines", 
            middleware.RequireProjectGuest(permService), 
            rt.listPipelines)
        
        // ===== 操作相关（developer 及以上）=====
        projectGroup.Post("/:projectId/builds", 
            middleware.RequireProjectDeveloper(permService), 
            rt.triggerBuild)
        
        projectGroup.Post("/:projectId/pipelines/:pipelineId/run", 
            middleware.RequireProjectDeveloper(permService), 
            rt.runPipeline)
        
        // ===== 管理相关（maintainer 及以上）=====
        projectGroup.Put("/:projectId/settings", 
            middleware.RequireProjectMaintainer(permService), 
            rt.updateSettings)
        
        projectGroup.Post("/:projectId/members", 
            middleware.RequireProjectMaintainer(permService), 
            rt.addMember)
        
        // ===== 删除相关（owner）=====
        projectGroup.Delete("/:projectId", 
            middleware.RequireProjectOwner(permService), 
            rt.deleteProject)
    }
}
```

### 组织路由示例

```go
func (rt *Router) orgRouter(r fiber.Router, auth fiber.Handler) {
    orgGroup := r.Group("/orgs", auth)
    permService := rt.PermService
    
    {
        // 查看组织信息 - 要求是组织成员
        orgGroup.Get("/:orgId", 
            middleware.RequireOrgMember(permService), 
            rt.getOrg)
        
        // 列出组织成员
        orgGroup.Get("/:orgId/members", 
            middleware.RequireOrgMember(permService), 
            rt.listOrgMembers)
        
        // 创建团队 - 要求组织管理员
        orgGroup.Post("/:orgId/teams", 
            middleware.PermissionMiddleware(permService, middleware.PermissionConfig{
                ResourceType:       "org",
                RequiredPermission: model.PermTeamCreate,
            }), 
            rt.createTeam)
    }
}
```

### 团队路由示例

```go
func (rt *Router) teamRouter(r fiber.Router, auth fiber.Handler) {
    teamGroup := r.Group("/teams", auth)
    permService := rt.PermService
    
    {
        // 查看团队 - 要求是团队成员
        teamGroup.Get("/:teamId", 
            middleware.RequireTeamMember(permService), 
            rt.getTeam)
        
        // 管理团队成员 - 要求特定权限
        teamGroup.Post("/:teamId/members", 
            middleware.PermissionMiddleware(permService, middleware.PermissionConfig{
                ResourceType:       "team",
                RequiredPermission: model.PermTeamMember,
            }), 
            rt.addTeamMember)
    }
}
```

## Handler 中获取权限信息

```go
func (rt *Router) getProject(c *fiber.Ctx) error {
    // 1. 获取资源ID（已由中间件提取）
    projectId, _ := c.Locals("projectId").(string)
    
    // 2. 获取权限信息（已由中间件计算）
    perm, ok := c.Locals("permission").(*service.ProjectPermission)
    if !ok {
        return fiber.ErrForbidden
    }
    
    // 3. 使用权限信息
    response := fiber.Map{
        "projectId": projectId,
        "userRole":  perm.RoleName,
        "priority":  perm.Priority,
        "source":    perm.Source,
        "permissions": fiber.Map{
            "canView":        perm.CanView,
            "canWrite":       perm.CanWrite,
            "canManage":      perm.CanManage,
            "canDelete":      perm.CanDelete,
            "canManageTeams": perm.CanManageTeams,
        },
    }
    
    // 4. 根据权限返回不同的数据
    if perm.CanManage {
        // 如果有管理权限，返回敏感信息
        response["webhookSecret"] = "xxx"
        response["apiTokens"] = []string{"token1", "token2"}
    }
    
    return c.JSON(response)
}
```

## 中间件链使用

可以组合多个中间件：

```go
projectGroup.Post("/:projectId/release", 
    auth,                                                  // 1. 先认证
    middleware.RequireProjectMaintainer(permService),     // 2. 再检查权限
    validateReleaseRequest(),                             // 3. 验证请求
    handler)                                               // 4. 业务处理
```

## 实际集成到 router.go

修改 `router.go` 文件：

```go
func NewRouter(httpConf *httpx.Http, ctx *ctx.Context) *Router {
    return &Router{
        Http:        httpConf,
        Ctx:         ctx,
        PermService: service.NewPermissionService(ctx), // 添加这行
    }
}

func (rt *Router) routerGroup(r fiber.Router) {
    auth := middleware.AuthorizationMiddleware(...)
    
    // 注册路由
    rt.userRouter(r, auth)
    rt.authRouter(r, auth)
    rt.projectRouter(r, auth)  // 添加项目路由
    rt.orgRouter(r, auth)      // 添加组织路由
    rt.teamRouter(r, auth)     // 添加团队路由
}

// 项目路由
func (rt *Router) projectRouter(r fiber.Router, auth fiber.Handler) {
    projectGroup := r.Group("/projects", auth)
    {
        projectGroup.Get("/:projectId", 
            middleware.RequireProject(rt.PermService), 
            rt.getProject)
        
        projectGroup.Post("/:projectId/builds", 
            middleware.RequireProjectDeveloper(rt.PermService), 
            rt.triggerBuild)
        
        // ... 更多路由
    }
}
```

## 注意事项

1. **中间件顺序** - 认证中间件必须在权限中间件之前
2. **资源ID** - projectId/orgId/teamId 会从 URL参数/查询参数/请求体 自动提取
3. **性能** - 权限服务内置角色缓存，避免重复查询
4. **错误处理** - 权限检查失败会自动返回 403 Forbidden
5. **可选检查** - 设置 `Optional: true` 可以在资源ID不存在时跳过检查

## 推荐的路由结构

```
/api/v1
├── /auth           # 认证相关（无需权限）
├── /users          # 用户相关（认证即可）
├── /orgs           # 组织相关（组织成员权限）
│   └── /:orgId
│       ├── /members     # 组织成员管理
│       ├── /teams       # 组织团队管理
│       └── /projects    # 组织项目列表
├── /teams          # 团队相关（团队成员权限）
│   └── /:teamId
│       ├── /members     # 团队成员管理
│       └── /projects    # 团队项目管理
└── /projects       # 项目相关（项目权限）
    └── /:projectId
        ├── /builds      # 构建管理（developer+）
        ├── /pipelines   # 流水线（developer+）
        ├── /deploys     # 部署（根据权限点）
        ├── /members     # 成员管理（maintainer+）
        └── /settings    # 设置（maintainer+）
```

## 常见场景

### 场景 1: 公开接口 + 可选权限

列表接口可能包含公开项目和私有项目：

```go
projectGroup.Get("", 
    middleware.PermissionMiddleware(permService, middleware.PermissionConfig{
        ResourceType: "project",
        Optional:     true, // 可选的，没有 projectId 也能访问
    }), 
    listProjects)
```

### 场景 2: 敏感操作需要多重检查

生产环境部署需要高权限：

```go
projectGroup.Post("/:projectId/deploy/production", 
    middleware.PermissionMiddleware(permService, middleware.PermissionConfig{
        ResourceType:       "project",
        RequiredRole:       model.BuiltinProjectMaintainer, // 至少是维护者
        RequiredPermission: model.PermDeployApprove,        // 且有审批权限
    }), 
    deployToProduction)
```

### 场景 3: 自定义角色检查

检查用户是否有自定义的"发布管理员"角色：

```go
projectGroup.Post("/:projectId/releases", 
    middleware.PermissionMiddleware(permService, middleware.PermissionConfig{
        ResourceType: "project",
        CheckFunc: func(p *service.ProjectPermission) bool {
            // 检查是否有特定自定义角色或足够的权限
            return p.RoleId == "custom_release_manager" || p.Priority >= 40
        },
    }), 
    createRelease)
```

### 场景 4: 根据不同资源类型检查

```go
// 项目资源
r.Get("/projects/:projectId", 
    middleware.PermissionMiddleware(permService, middleware.PermissionConfig{
        ResourceType: "project",
    }), handler)

// 组织资源
r.Get("/orgs/:orgId", 
    middleware.PermissionMiddleware(permService, middleware.PermissionConfig{
        ResourceType: "org",
    }), handler)

// 团队资源
r.Get("/teams/:teamId", 
    middleware.PermissionMiddleware(permService, middleware.PermissionConfig{
        ResourceType: "team",
    }), handler)
```

## 调试技巧

在 handler 中打印权限信息：

```go
func handler(c *fiber.Ctx) error {
    perm, _ := c.Locals("permission").(*service.ProjectPermission)
    
    log.Infof("User %s accessing project %s with role %s (priority: %d, source: %s)", 
        perm.UserId, perm.ProjectId, perm.RoleName, perm.Priority, perm.Source)
    
    log.Infof("Permissions: %v", perm.Permissions)
    log.Infof("CanWrite: %v, CanManage: %v, CanDelete: %v", 
        perm.CanWrite, perm.CanManage, perm.CanDelete)
    
    // 业务逻辑...
}
```

