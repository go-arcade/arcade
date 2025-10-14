# 权限系统设计文档

## 概述

本系统采用基于**项目**和**团队**的细粒度权限管理，支持**自定义角色**和**权限点**组合，提供灵活的权限控制。

## 核心特性

1. **自定义角色** - 支持为组织创建自定义角色
2. **权限点组合** - 通过权限点（Permission Point）灵活组合权限
3. **多层次权限** - 支持项目、团队、组织三个层次
4. **内置角色** - 提供常用的内置角色，开箱即用
5. **权限继承** - 支持从团队和组织继承权限

## 核心概念

### 1. 组织（Organization）
- 最高层级的权限边界
- 成员角色：`owner`, `admin`, `member`
- 可以创建多个团队和项目

### 2. 团队（Team）
- 组织内的协作单位
- 成员角色：`owner`, `maintainer`, `developer`, `reporter`, `guest`
- 可以被授予对项目的访问权限

### 3. 项目（Project）
- 实际的代码仓库和CI/CD流水线
- 直接成员角色：`owner`, `maintainer`, `developer`, `reporter`, `guest`
- 访问级别：`owner`, `team`, `org`

## 角色权限矩阵

### 项目成员角色

| 角色 | 查看 | 创建分支 | 提交代码 | 触发构建 | 管理成员 | 修改设置 | 删除项目 |
|------|------|----------|----------|----------|----------|----------|----------|
| owner | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| maintainer | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ |
| developer | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ |
| reporter | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| guest | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |

### 团队成员角色

| 角色 | 查看团队 | 参与开发 | 管理成员 | 删除团队 |
|------|----------|----------|----------|----------|
| owner | ✅ | ✅ | ✅ | ✅ |
| maintainer | ✅ | ✅ | ✅ | ❌ |
| developer | ✅ | ✅ | ❌ | ❌ |
| reporter | ✅ | ❌ | ❌ | ❌ |
| guest | ✅ | ❌ | ❌ | ❌ |

### 项目团队访问权限级别

| 访问级别 | 查看 | 提交代码 | 触发构建 | 管理项目 |
|----------|------|----------|----------|----------|
| read | ✅ | ❌ | ❌ | ❌ |
| write | ✅ | ✅ | ✅ | ❌ |
| admin | ✅ | ✅ | ✅ | ✅ |

## 权限计算逻辑

用户对项目的最终权限 = **MAX**(直接权限, 团队权限, 组织权限)

### 1. 直接权限
从 `t_project_member` 表中获取，用户直接被添加到项目的角色。

### 2. 团队权限
通过以下步骤计算：
1. 查找用户所在的所有团队 (`t_team_member`)
2. 查找这些团队对项目的访问权限 (`t_project_team_access`)
3. 根据团队角色和访问级别映射为项目角色

**映射规则**：

| 团队角色 | 访问级别=read | 访问级别=write | 访问级别=admin |
|----------|---------------|----------------|----------------|
| owner | guest | developer | maintainer |
| maintainer | guest | developer | maintainer |
| developer | guest | developer | developer |
| reporter | guest | reporter | reporter |
| guest | guest | guest | guest |

### 3. 组织权限
当项目的 `access_level = 'org'` 时，组织成员自动获得基础权限：
- 组织 owner → 项目 maintainer
- 组织 admin → 项目 developer
- 组织 member → 项目 guest

## 数据库表设计

### t_project_member (项目成员表)
```sql
CREATE TABLE t_project_member (
  id INT AUTO_INCREMENT PRIMARY KEY,
  project_id VARCHAR(64) NOT NULL,
  user_id VARCHAR(64) NOT NULL,
  role VARCHAR(32) NOT NULL,
  create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
  update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_project_user (project_id, user_id)
);
```

### t_team_member (团队成员表)
```sql
CREATE TABLE t_team_member (
  id INT AUTO_INCREMENT PRIMARY KEY,
  team_id VARCHAR(64) NOT NULL,
  user_id VARCHAR(64) NOT NULL,
  role VARCHAR(32) NOT NULL,
  create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
  update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_team_user (team_id, user_id)
);
```

### t_project_team_access (项目团队访问权限表)
```sql
CREATE TABLE t_project_team_access (
  id INT AUTO_INCREMENT PRIMARY KEY,
  project_id VARCHAR(64) NOT NULL,
  team_id VARCHAR(64) NOT NULL,
  access_level VARCHAR(32) NOT NULL,
  create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
  update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_project_team (project_id, team_id)
);
```

## API 使用示例

### HTTP 中间件使用

```go
import (
    "github.com/observabil/arcade/internal/engine/model"
    "github.com/observabil/arcade/internal/engine/service"
    "github.com/observabil/arcade/pkg/http/middleware"
)

// 初始化权限服务
permService := service.NewPermissionService(ctx)

// 1. 基础项目访问检查（任意角色即可）
router.Get("/api/v1/projects/:projectId", 
    middleware.RequireProject(permService), 
    handlerFunc)

// 2. 要求特定角色权限
router.Get("/api/v1/projects/:projectId/settings", 
    middleware.RequireProjectMaintainer(permService), 
    handlerFunc)

// 3. 要求开发者权限才能触发构建
router.Post("/api/v1/projects/:projectId/pipelines/trigger", 
    middleware.RequireProjectDeveloper(permService), 
    handlerFunc)

// 4. 要求特定权限点
router.Post("/api/v1/projects/:projectId/build", 
    middleware.RequirePermission(permService, model.PermBuildTrigger), 
    handlerFunc)

// 5. 要求写入权限（便捷函数）
router.Post("/api/v1/projects/:projectId/pipeline", 
    middleware.RequireCanWrite(permService), 
    handlerFunc)

// 6. 要求管理权限
router.Put("/api/v1/projects/:projectId/members", 
    middleware.RequireCanManage(permService), 
    handlerFunc)

// 7. 要求删除权限
router.Delete("/api/v1/projects/:projectId", 
    middleware.RequireCanDelete(permService), 
    handlerFunc)

// 8. 检查组织成员
router.Get("/api/v1/orgs/:orgId/info", 
    middleware.RequireOrgMember(permService), 
    handlerFunc)

// 9. 检查团队成员
router.Get("/api/v1/teams/:teamId/members", 
    middleware.RequireTeamMember(permService), 
    handlerFunc)

// 10. 使用统一中间件 + 自定义检查
router.Post("/api/v1/projects/:projectId/deploy", 
    middleware.PermissionMiddleware(permService, middleware.PermissionConfig{
        ResourceType: "project",
        CheckFunc: func(p *service.ProjectPermission) bool {
            return p.CanWrite && p.Priority >= 30
        },
    }), 
    handlerFunc)

// 11. 使用统一中间件 + 组合条件
router.Post("/api/v1/projects/:projectId/release", 
    middleware.PermissionMiddleware(permService, middleware.PermissionConfig{
        ResourceType:       "project",
        RequiredRole:       model.BuiltinProjectMaintainer,
        RequiredPermission: model.PermDeployApprove,
    }), 
    handlerFunc)
```

### 在 Handler 中获取权限信息

```go
func handlerFunc(c *fiber.Ctx) error {
    // 获取权限信息
    perm, ok := c.Locals("permission").(*service.ProjectPermission)
    if !ok {
        return fiber.ErrForbidden
    }

    // 使用权限信息
    if perm.CanDelete {
        // 执行删除操作
    }

    return c.JSON(fiber.Map{
        "role": perm.Role,
        "source": perm.Source,
        "canWrite": perm.CanWrite,
    })
}
```

### gRPC 拦截器使用

```go
import (
    "github.com/observabil/arcade/internal/engine/service"
    "github.com/observabil/arcade/internal/pkg/grpc/middleware"
)

// 初始化权限服务
permService := service.NewPermissionService(ctx)

// 在 gRPC 服务器配置中添加权限拦截器
opts := []grpc.ServerOption{
    grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
        middleware.LoggingUnaryInterceptor(log),
        middleware.AuthUnaryInterceptor(),
        middleware.PermissionUnaryInterceptor(permService),  // 添加权限拦截器
        grpc_recovery.UnaryServerInterceptor(),
    )),
    grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
        middleware.LoggingStreamInterceptor(log),
        middleware.AuthStreamInterceptor(),
        middleware.PermissionStreamInterceptor(permService),  // 添加权限拦截器
        grpc_recovery.StreamServerInterceptor(),
    )),
}
```

### gRPC 客户端发送项目ID

```go
// 在 metadata 中设置项目ID
md := metadata.Pairs("project-id", projectId)
ctx := metadata.NewOutgoingContext(context.Background(), md)

// 调用 gRPC 方法
resp, err := client.SomeMethod(ctx, req)
```

### 在 gRPC Handler 中获取权限信息

```go
func (s *MyServiceImpl) SomeMethod(ctx context.Context, req *pb.Request) (*pb.Response, error) {
    // 获取权限信息
    perm, err := middleware.GetPermissionFromContext(ctx)
    if err != nil {
        return nil, status.Error(codes.PermissionDenied, "permission not found")
    }

    // 检查权限
    if !perm.CanWrite {
        return nil, status.Error(codes.PermissionDenied, "write permission required")
    }

    // 执行业务逻辑
    return &pb.Response{}, nil
}
```

## 权限检查流程

### HTTP 请求流程
```
HTTP Request
    ↓
AuthorizationMiddleware (JWT验证)
    ↓
RequireProjectRole (权限检查)
    ↓
    1. 提取 projectId (从 URL/Query/Body)
    2. 提取 userId (从 JWT claims)
    3. 调用 PermissionService.CheckProjectPermission()
       - 检查直接权限
       - 检查团队权限
       - 检查组织权限
       - 计算最终权限
    4. 验证是否满足要求的角色
    ↓
Handler (业务逻辑)
```

### gRPC 请求流程
```
gRPC Request (with metadata: project-id)
    ↓
AuthUnaryInterceptor (认证)
    ↓
PermissionUnaryInterceptor (权限检查)
    ↓
    1. 从 metadata 提取 project-id
    2. 从 context 提取 userId
    3. 调用 PermissionService.CheckProjectPermission()
    4. 将权限信息添加到 context
    ↓
Service Handler (业务逻辑)
```

## 便捷函数

### HTTP 中间件便捷函数
```go
// 基础访问检查
middleware.RequireProject(permService)        // 要求项目访问权限（任意角色）

// 角色级别检查
middleware.RequireProjectOwner(permService)      // 要求所有者
middleware.RequireProjectMaintainer(permService) // 要求维护者
middleware.RequireProjectDeveloper(permService)  // 要求开发者
middleware.RequireProjectReporter(permService)   // 要求报告者
middleware.RequireProjectGuest(permService)      // 要求访客

// 权限点检查
middleware.RequirePermission(permService, model.PermBuildTrigger) // 要求特定权限点

// 权限标志检查
middleware.RequireCanWrite(permService)   // 要求写入权限
middleware.RequireCanManage(permService)  // 要求管理权限
middleware.RequireCanDelete(permService)  // 要求删除权限

// 组织/团队成员检查
middleware.RequireOrgMember(permService)  // 要求组织成员
middleware.RequireTeamMember(permService) // 要求团队成员

// 统一中间件（高级用法，支持组合条件）
middleware.PermissionMiddleware(permService, middleware.PermissionConfig{
    ResourceType:       "project",
    RequiredRole:       model.BuiltinProjectDeveloper,
    RequiredPermission: model.PermDeployExecute,
    CheckFunc: func(p *service.ProjectPermission) bool {
        return p.Priority >= 30
    },
})
```

## 自定义角色

### 创建自定义角色

```go
import "github.com/observabil/arcade/internal/engine/model"

// 创建一个构建管理员角色
role := &model.Role{
    RoleId:      "custom_build_admin",
    Name:        "build_admin",
    DisplayName: "构建管理员",
    Description: "只能触发和管理构建",
    Scope:       model.RoleScopeProject,
    OrgId:       "org_123", // 组织自定义角色
    IsBuiltin:   model.RoleCustom,
    IsEnabled:   model.RoleEnabled,
    Priority:    25, // 介于 reporter(20) 和 developer(30) 之间
    Permissions: `["project.view","build.view","build.trigger","build.cancel","build.retry","build.artifact","build.log","pipeline.view","pipeline.run"]`,
    CreatedBy:   "user_123",
}

roleRepo.CreateRole(role)
```

### 自定义角色示例

#### 1. 构建管理员（Build Admin）
```json
{
  "role_id": "custom_build_admin",
  "name": "build_admin",
  "priority": 25,
  "permissions": [
    "project.view",
    "build.view",
    "build.trigger",
    "build.cancel",
    "build.retry",
    "build.artifact",
    "build.log",
    "pipeline.view",
    "pipeline.run"
  ]
}
```
**用途**：专门负责构建管理的角色，可以触发和管理构建

#### 2. 部署管理员（Deploy Admin）
```json
{
  "role_id": "custom_deploy_admin",
  "name": "deploy_admin",
  "priority": 35,
  "permissions": [
    "project.view",
    "build.view",
    "build.artifact",
    "pipeline.view",
    "pipeline.run",
    "deploy.view",
    "deploy.create",
    "deploy.execute",
    "deploy.rollback",
    "deploy.approve"
  ]
}
```
**用途**：专门负责部署的角色，有完整部署权限

#### 3. 监控管理员（Monitor Admin）
```json
{
  "role_id": "custom_monitor_admin",
  "name": "monitor_admin",
  "priority": 15,
  "permissions": [
    "project.view",
    "build.view",
    "build.log",
    "monitor.view",
    "monitor.metrics",
    "monitor.logs",
    "monitor.alert",
    "monitor.dashboard"
  ]
}
```
**用途**：专门负责监控告警的角色

#### 4. 安全审计员（Security Auditor）
```json
{
  "role_id": "custom_security_auditor",
  "name": "security_auditor",
  "priority": 28,
  "permissions": [
    "project.view",
    "build.view",
    "security.scan",
    "security.audit",
    "security.policy"
  ]
}
```
**用途**：专门负责安全审计的角色

### 权限点检查

除了角色级别检查，还可以直接检查权限点：

```go
// HTTP 中间件：检查特定权限点
router.Post("/api/v1/projects/:projectId/deploy", 
    middleware.RequirePermission(permService, model.PermDeployExecute), 
    handlerFunc)

// 在代码中检查权限点
err := permService.RequirePermissionPoint(ctx, userId, projectId, model.PermDeployApprove)
if err != nil {
    return errors.New("you don't have deploy approval permission")
}

// 检查是否有权限点
hasPermission, err := permService.HasPermission(roleId, model.PermDeployApprove)
```

## 迁移指南

### 1. 运行数据库迁移
```bash
mysql -u root -p arcade < docs/database_migration_roles.sql
```

### 2. 初始化内置角色（代码中自动执行）
```go
roleRepo := repo.NewRoleRepo(ctx)
if err := roleRepo.InitBuiltinRoles(); err != nil {
    log.Errorf("failed to init builtin roles: %v", err)
}
```

### 3. 为现有项目添加成员
```sql
-- 将项目创建者添加为 owner 角色
INSERT INTO t_project_member (project_id, user_id, role_id)
SELECT project_id, created_by, 'project_owner'
FROM t_project
WHERE created_by IS NOT NULL;
```

## 性能优化建议

1. **缓存权限结果** - 对频繁访问的项目权限进行缓存
2. **批量查询** - 列表接口批量查询权限，避免 N+1 问题
3. **索引优化** - 确保所有外键字段都有索引
4. **权限预计算** - 对复杂权限可以预计算并存储

## 安全考虑

1. **最小权限原则** - 默认拒绝，显式授权
2. **权限审计** - 记录所有权限变更操作
3. **定期检查** - 定期清理无效的权限关系
4. **敏感操作** - 删除项目等敏感操作需要二次确认

## 示例场景

### 场景1：用户通过团队访问项目
1. 用户 Alice 是 Team A 的 developer
2. Team A 对 Project X 有 write 访问权限
3. Alice 访问 Project X 时：
   - 团队角色: developer
   - 访问级别: write
   - 最终项目角色: developer
   - 权限: 可以查看、提交代码、触发构建

### 场景2：用户同时拥有多个权限来源
1. 用户 Bob 直接被添加为 Project Y 的 reporter
2. Bob 也是 Team B 的 maintainer
3. Team B 对 Project Y 有 admin 访问权限
4. Bob 访问 Project Y 时：
   - 直接权限: reporter
   - 团队权限: maintainer (maintainer + admin)
   - 最终权限: maintainer (取最大值)
   - 权限: 可以管理项目但不能删除

### 场景3：组织级访问
1. Project Z 的 access_level 设置为 'org'
2. 用户 Carol 是组织的 member
3. Carol 访问 Project Z 时：
   - 组织权限: guest
   - 最终权限: guest
   - 权限: 只能查看，不能修改

## 自定义角色 API

### 创建自定义角色
```http
POST /api/v1/orgs/:orgId/roles
Content-Type: application/json

{
  "name": "release_manager",
  "display_name": "发布管理员",
  "description": "负责版本发布的角色",
  "scope": "project",
  "priority": 35,
  "permissions": [
    "project.view",
    "build.view",
    "build.trigger",
    "pipeline.view",
    "pipeline.create",
    "pipeline.run",
    "deploy.execute",
    "deploy.approve"
  ]
}
```

### 列出可用角色
```http
GET /api/v1/orgs/:orgId/roles?scope=project

Response:
{
  "roles": [
    {
      "role_id": "project_owner",
      "name": "owner",
      "display_name": "所有者",
      "is_builtin": true,
      "priority": 50,
      "permissions": [...]
    },
    {
      "role_id": "custom_release_manager",
      "name": "release_manager",
      "display_name": "发布管理员",
      "is_builtin": false,
      "priority": 35,
      "permissions": [...]
    }
  ]
}
```

### 为用户分配自定义角色
```http
POST /api/v1/projects/:projectId/members
Content-Type: application/json

{
  "user_id": "user_123",
  "role_id": "custom_release_manager"
}
```

## 权限点完整列表

所有可用的权限点定义在 `model/model_permission.go` 文件中，包括：

| 分类 | 权限点数量 | 说明 |
|------|-----------|------|
| 项目权限 | 9 | 项目级别的操作 |
| 构建权限 | 7 | CI/CD 构建操作 |
| 流水线权限 | 6 | 流水线管理 |
| 部署权限 | 5 | 部署相关操作 |
| 成员权限 | 4 | 成员管理 |
| 团队权限 | 7 | 团队管理 |
| Issue权限 | 5 | 问题追踪 |
| 监控权限 | 5 | 监控告警 |
| 安全权限 | 3 | 安全审计 |

## 扩展性

系统设计支持未来扩展：
- ✅ 自定义角色和权限点
- ✅ 组织级自定义角色
- ✅ 基于优先级的权限合并
- 可以添加更细粒度的权限（如分支保护规则）
- 可以支持项目组（Project Group）概念
- 可以支持权限继承和委托机制
- 可以支持临时权限和过期机制

