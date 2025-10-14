# 权限系统实施指南

## 🎯 已完成的工作

### 1. 插件分发系统
- ✅ Proto定义扩展（`api/agent/v1/proto/agent.proto`）
- ✅ 服务端插件下载服务（`service_plugin_download.go`）
- ✅ Agent端插件下载器示例（`examples/agent_plugin_downloader.go`）
- ✅ 完整文档（`PLUGIN_DISTRIBUTION.md`）

### 2. 权限系统重构
- ✅ 权限点独立表设计（`t_permission`）
- ✅ 角色权限关联表（`t_role_permission`）
- ✅ 路由权限映射表（`t_router_permission`）
- ✅ 权限仓库（`repo_permission.go`）
- ✅ 用户权限聚合服务（`service_user_permissions.go`）
- ✅ 权限中间件（`user_permissions.go`）

## 📋 实施步骤

### Step 1: 数据库迁移

```bash
cd /Users/gagral/go/src/arcade

# 1. 创建角色表
mysql -u root -p arcade < docs/database_migration_roles.sql

# 2. 创建权限表和关联表
mysql -u root -p arcade < docs/database_migration_permissions.sql

# 3. 创建路由权限表
mysql -u root -p arcade < docs/database_migration_router_permission.sql
```

验证：
```sql
-- 检查表是否创建成功
SHOW TABLES LIKE 't_%permission%';

-- 查看权限统计
SELECT category, COUNT(*) FROM t_permission GROUP BY category;

-- 查看角色权限统计
SELECT role_id, COUNT(*) FROM t_role_permission GROUP BY role_id;
```

### Step 2: 更新Go代码

#### 2.1 更新model包

```bash
# 新增权限模型（已创建）
# internal/engine/model/model_permission_new.go
```

#### 2.2 更新repo包

```bash
# 新增权限仓库（已创建）
# internal/engine/repo/repo_permission.go
# internal/engine/repo/repo_router_permission.go
```

#### 2.3 更新service包

服务已更新为使用关联表查询：
- `internal/engine/service/service_permission.go`
- `internal/engine/service/service_user_permissions.go`

#### 2.4 更新中间件

中间件已创建：
- `pkg/http/middleware/user_permissions.go`

### Step 3: 注册服务和路由

在 `cmd/arcade/main.go` 或 `internal/app/app.go` 中：

```go
package main

import (
    "github.com/observabil/arcade/internal/engine/service"
    "github.com/observabil/arcade/internal/engine/repo"
    "github.com/observabil/arcade/pkg/http/middleware"
)

func main() {
    // ... 其他初始化 ...
    
    // 初始化权限服务
    permService := service.NewPermissionService(ctx)
    
    // 初始化用户权限聚合服务
    userPermService := service.NewUserPermissionsService(ctx, permService)
    
    // 初始化路由权限配置（首次运行）
    routerRepo := repo.NewRouterPermissionRepo(ctx)
    if err := routerRepo.InitBuiltinRoutes(); err != nil {
        log.Errorf("failed to init builtin routes: %v", err)
    }
    
    // 设置路由
    setupRoutes(app, ctx, permService, userPermService)
}

func setupRoutes(app *fiber.App, ctx *ctx.Context, permSvc *service.PermissionService, userPermSvc *service.UserPermissionsService) {
    api := app.Group("/api/v1")
    
    // 应用全局中间件
    api.Use(middleware.Authorization(jwtManager))
    api.Use(middleware.UserPermissionsMiddleware(userPermSvc))
    
    // 用户权限相关路由
    user := api.Group("/user")
    user.Get("/permissions", middleware.GetUserPermissionsHandler(userPermSvc))
    user.Get("/routes", middleware.GetUserAccessibleRoutesHandler(userPermSvc))
    user.Get("/permissions/summary", middleware.GetUserPermissionsSummaryHandler(userPermSvc))
    
    // 项目路由（带权限检查）
    projects := api.Group("/projects")
    projects.Get("/", middleware.RequireProject(permSvc), listProjects)
    projects.Post("/", createProject) // 任何登录用户都可创建
    
    project := projects.Group("/:projectId")
    project.Get("/", middleware.RequireProject(permSvc), getProject)
    project.Put("/", middleware.RequireProjectMaintainer(permSvc), updateProject)
    project.Delete("/", middleware.RequireProjectOwner(permSvc), deleteProject)
    
    // 构建路由
    project.Get("/builds", middleware.RequirePermission(permSvc, model.PermBuildView), listBuilds)
    project.Post("/builds/trigger", middleware.RequirePermission(permSvc, model.PermBuildTrigger), triggerBuild)
}
```

### Step 4: 重新生成Proto代码（插件分发）

```bash
cd api
buf generate
```

这会重新生成：
- `api/agent/v1/agent.pb.go`
- `api/agent/v1/agent_grpc.pb.go`

### Step 5: 实现Agent端gRPC服务

在 `internal/engine/service/agent/service_agent_pb.go` 中添加：

```go
// DownloadPlugin 实现插件下载接口
func (s *AgentServicePB) DownloadPlugin(ctx context.Context, req *v1.DownloadPluginRequest) (*v1.DownloadPluginResponse, error) {
    log.Infof("agent %s requesting plugin: %s (v%s)", req.AgentId, req.PluginId, req.Version)

    // 调用插件下载服务
    resp, err := s.pluginDownloadService.GetPluginForDownload(req.PluginId, req.Version)
    if err != nil {
        log.Errorf("failed to get plugin for download: %v", err)
        return &v1.DownloadPluginResponse{
            Success: false,
            Message: fmt.Sprintf("failed to get plugin: %v", err),
        }, nil
    }

    return resp, nil
}

// ListAvailablePlugins 列出可用插件
func (s *AgentServicePB) ListAvailablePlugins(ctx context.Context, req *v1.ListAvailablePluginsRequest) (*v1.ListAvailablePluginsResponse, error) {
    plugins, err := s.pluginDownloadService.ListAvailablePlugins(req.PluginType)
    if err != nil {
        return &v1.ListAvailablePluginsResponse{
            Success: false,
            Message: fmt.Sprintf("failed to list plugins: %v", err),
        }, nil
    }

    return &v1.ListAvailablePluginsResponse{
        Success: true,
        Message: "success",
        Plugins: plugins,
    }, nil
}
```

### Step 6: 测试

```bash
# 启动服务
make run

# 测试用户权限API
curl -H "Authorization: Bearer <token>" \
     http://localhost:8080/api/v1/user/permissions

# 测试路由列表
curl -H "Authorization: Bearer <token>" \
     http://localhost:8080/api/v1/user/routes
```

## 🔍 验证

### 数据库验证

```sql
-- 1. 验证权限点总数
SELECT COUNT(*) AS '权限点总数' FROM t_permission WHERE is_enabled = 1;

-- 2. 验证角色权限关联
SELECT 
    r.name AS '角色',
    COUNT(rp.permission_id) AS '权限数量'
FROM t_role r
LEFT JOIN t_role_permission rp ON r.role_id = rp.role_id
WHERE r.is_builtin = 1
GROUP BY r.role_id, r.name;

-- 3. 验证路由配置
SELECT COUNT(*) AS '路由总数', 
       SUM(CASE WHEN is_menu = 1 THEN 1 ELSE 0 END) AS '菜单数量'
FROM t_router_permission;
```

### API验证

```bash
# 测试用户权限
curl http://localhost:8080/api/v1/user/permissions \
  -H "Authorization: Bearer YOUR_TOKEN"

# 预期返回：
# {
#   "userId": "...",
#   "organizations": [...],
#   "teams": [...],
#   "projects": [...],
#   "allPermissions": ["project.view", "build.trigger", ...],
#   "accessibleRoutes": [...]
# }
```

## 📝 文档索引

### 插件分发
- `docs/PLUGIN_DISTRIBUTION.md` - 完整设计文档
- `docs/PLUGIN_DISTRIBUTION_QUICKSTART.md` - 快速入门
- `docs/examples/agent_plugin_downloader.go` - Agent端示例代码

### 权限系统
- `docs/PERMISSION_SYSTEM.md` - 原权限系统文档
- `docs/PERMISSION_SYSTEM_REFACTORED.md` - 重构后设计
- `docs/USER_PERMISSIONS_API.md` - API使用指南

### 数据库
- `docs/database_migration_roles.sql` - 角色表
- `docs/database_migration_permissions.sql` - 权限表和关联表
- `docs/database_migration_router_permission.sql` - 路由权限表

## 🚀 下一步

### 必须完成
1. ✅ 运行数据库迁移
2. ✅ 重新生成Proto代码
3. ⏳ 注册API路由
4. ⏳ 实现Agent端插件下载（需要Agent项目）

### 可选增强
- Redis缓存权限查询
- WebSocket推送权限变更
- 权限变更审计日志
- 前端权限指令和路由守卫
- 性能监控和优化

## 💡 使用建议

1. **开发环境**：先在本地验证所有功能
2. **测试环境**：部署测试，验证权限隔离
3. **生产环境**：灰度发布，监控性能指标
4. **文档维护**：及时更新权限点和路由配置

完成！🎉

