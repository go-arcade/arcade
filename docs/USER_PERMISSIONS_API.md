# 用户权限聚合系统使用指南

## 概述

用户权限聚合系统提供了一个统一的方式来查询和管理用户在**组织-团队-项目**三个层级的权限，并基于权限返回用户可访问的路由列表。

## 核心功能

1. **权限聚合** - 自动聚合用户在组织、团队、项目中的所有权限
2. **路由过滤** - 根据用户权限返回可访问的路由列表
3. **资源汇总** - 列出用户可访问的所有资源（组织、团队、项目）
4. **动态菜单** - 基于权限动态生成前端菜单
5. **权限检查** - 提供便捷的权限检查中间件

## 系统架构

```
┌─────────────────────────────────────────────────────┐
│              HTTP请求 (带JWT Token)                  │
└───────────────────┬─────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────────┐
│          AuthorizationMiddleware (JWT验证)           │
└───────────────────┬─────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────────┐
│       UserPermissionsMiddleware (权限聚合)           │
│  1. 查询组织权限                                      │
│  2. 查询团队权限                                      │
│  3. 查询项目权限                                      │
│  4. 汇总所有权限点                                    │
│  5. 计算可访问路由                                    │
│  6. 注入到context                                    │
└───────────────────┬─────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────────┐
│            Handler (业务逻辑)                         │
│  - 可从context获取权限信息                            │
│  - 可使用便捷函数检查权限                              │
└─────────────────────────────────────────────────────┘
```

## API端点

### 1. 获取用户完整权限信息

**接口**: `GET /api/v1/user/permissions`

**描述**: 返回当前用户的完整权限信息，包括组织、团队、项目权限及可访问路由

**请求头**:
```http
Authorization: Bearer <JWT_TOKEN>
```

**响应示例**:
```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "userId": "user_123",
    "organizations": [
      {
        "orgId": "org_001",
        "orgName": "示例组织",
        "roleId": "org_admin",
        "roleName": "管理员",
        "permissions": [
          "team.view",
          "team.create",
          "team.edit"
        ],
        "status": "active"
      }
    ],
    "teams": [
      {
        "teamId": "team_001",
        "teamName": "开发团队",
        "orgId": "org_001",
        "roleId": "team_developer",
        "roleName": "开发者",
        "permissions": [
          "team.view",
          "team.project"
        ]
      }
    ],
    "projects": [
      {
        "projectId": "proj_001",
        "projectName": "核心项目",
        "orgId": "org_001",
        "roleId": "project_developer",
        "roleName": "开发者",
        "source": "direct",
        "permissions": [
          "project.view",
          "build.trigger",
          "pipeline.run"
        ],
        "priority": 30
      }
    ],
    "allPermissions": [
      "project.view",
      "build.view",
      "build.trigger",
      "pipeline.view",
      "pipeline.run",
      "team.view"
    ],
    "accessibleRoutes": [
      {
        "path": "/api/v1/projects",
        "method": "GET",
        "name": "项目列表",
        "group": "project",
        "category": "项目管理",
        "permissions": ["project.view"],
        "icon": "project",
        "order": 100,
        "isMenu": true
      }
    ],
    "accessibleResources": {
      "organizations": ["org_001"],
      "teams": ["team_001", "team_002"],
      "projects": ["proj_001", "proj_002", "proj_003"]
    }
  }
}
```

### 2. 获取用户可访问路由

**接口**: `GET /api/v1/user/routes`

**描述**: 返回用户可访问的路由列表（按分类分组，用于生成前端菜单）

**响应示例**:
```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "routes": [
      {
        "path": "/api/v1/projects",
        "method": "GET",
        "name": "项目列表",
        "group": "project",
        "category": "项目管理",
        "icon": "project",
        "order": 100,
        "isMenu": true
      },
      {
        "path": "/api/v1/projects/:projectId/pipelines",
        "method": "GET",
        "name": "流水线列表",
        "group": "pipeline",
        "category": "CI/CD",
        "icon": "pipeline",
        "order": 200,
        "isMenu": true
      }
    ],
    "menu": {
      "项目管理": [
        {
          "path": "/api/v1/projects",
          "name": "项目列表",
          "icon": "project"
        }
      ],
      "CI/CD": [
        {
          "path": "/api/v1/projects/:projectId/pipelines",
          "name": "流水线列表",
          "icon": "pipeline"
        }
      ]
    }
  }
}
```

### 3. 获取用户权限摘要

**接口**: `GET /api/v1/user/permissions/summary`

**描述**: 返回用户权限的简化摘要信息

**响应示例**:
```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "userId": "user_123",
    "organizationCount": 2,
    "teamCount": 5,
    "projectCount": 15,
    "permissionCount": 28,
    "accessibleRouteCount": 42,
    "accessibleResources": {
      "organizations": ["org_001", "org_002"],
      "teams": ["team_001", "team_002", "team_003"],
      "projects": ["proj_001", "proj_002", "proj_003"]
    }
  }
}
```

## 集成到路由系统

### 1. 初始化服务

```go
// cmd/arcade/main.go 或 internal/app/app.go

import (
    "github.com/observabil/arcade/internal/engine/service"
    "github.com/observabil/arcade/internal/engine/repo"
    "github.com/observabil/arcade/pkg/http/middleware"
)

func initServices(ctx *ctx.Context) {
    // 初始化权限服务
    permService := service.NewPermissionService(ctx)
    
    // 初始化用户权限聚合服务
    userPermService := service.NewUserPermissionsService(ctx, permService)
    
    // 初始化路由权限配置
    routerRepo := repo.NewRouterPermissionRepo(ctx)
    if err := routerRepo.InitBuiltinRoutes(); err != nil {
        log.Errorf("failed to init builtin routes: %v", err)
    }
    
    return userPermService
}
```

### 2. 注册全局中间件

```go
// internal/engine/router/router.go

func SetupRouter(app *fiber.App, ctx *ctx.Context) {
    // JWT认证中间件
    app.Use(middleware.Authorization(jwtManager))
    
    // 用户权限聚合中间件（在认证后、业务路由前）
    app.Use(middleware.UserPermissionsMiddleware(userPermService))
    
    // 后续的业务路由
    // ...
}
```

### 3. 注册API端点

```go
// internal/engine/router/router_user.go

func RegisterUserRoutes(router fiber.Router, userPermService *service.UserPermissionsService) {
    user := router.Group("/user")
    
    // 获取用户完整权限信息
    user.Get("/permissions", middleware.GetUserPermissionsHandler(userPermService))
    
    // 获取用户可访问路由
    user.Get("/routes", middleware.GetUserAccessibleRoutesHandler(userPermService))
    
    // 获取用户权限摘要
    user.Get("/permissions/summary", middleware.GetUserPermissionsSummaryHandler(userPermService))
}
```

### 4. 在Handler中使用权限信息

```go
// 方式1：从context获取完整权限信息
func MyHandler(c *fiber.Ctx) error {
    perms, ok := middleware.GetUserPermissionsFromContext(c)
    if !ok {
        return fiber.ErrForbidden
    }
    
    // 使用权限信息
    log.Infof("User has %d projects", len(perms.Projects))
    
    // 检查特定权限
    if middleware.HasPermissionInContext(c, model.PermBuildTrigger) {
        // 用户有触发构建的权限
    }
    
    return c.JSON(perms)
}

// 方式2：使用便捷中间件
router.Post("/api/v1/projects/:projectId/build",
    middleware.RequireAnyPermission(userPermService, []string{
        model.PermBuildTrigger,
        model.PermPipelineRun,
    }),
    handleBuildTrigger,
)

// 方式3：要求所有权限
router.Post("/api/v1/projects/:projectId/release",
    middleware.RequireAllPermissions(userPermService, []string{
        model.PermDeployExecute,
        model.PermDeployApprove,
    }),
    handleRelease,
)
```

## 数据库表结构

### t_router_permission (路由权限映射表)

```sql
CREATE TABLE t_router_permission (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    route_id VARCHAR(128) NOT NULL UNIQUE COMMENT '路由ID',
    path VARCHAR(255) NOT NULL COMMENT '路由路径',
    method VARCHAR(16) NOT NULL COMMENT 'HTTP方法',
    name VARCHAR(128) NOT NULL COMMENT '路由名称',
    `group` VARCHAR(64) COMMENT '路由分组',
    category VARCHAR(64) COMMENT '路由分类',
    required_permissions JSON COMMENT '所需权限列表',
    icon VARCHAR(64) COMMENT '图标',
    `order` INT DEFAULT 0 COMMENT '排序',
    is_menu TINYINT DEFAULT 0 COMMENT '是否显示在菜单 0:否 1:是',
    is_enabled TINYINT DEFAULT 1 COMMENT '是否启用',
    description VARCHAR(512) COMMENT '描述',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_path_method (path, method),
    INDEX idx_group (` group`),
    INDEX idx_category (category),
    INDEX idx_is_menu (is_menu)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='路由权限映射表';
```

## 前端集成示例

### 1. 登录后获取用户权限

```typescript
// frontend/src/api/user.ts
import axios from 'axios';

export interface UserPermissions {
  userId: string;
  organizations: Organization[];
  teams: Team[];
  projects: Project[];
  allPermissions: string[];
  accessibleRoutes: Route[];
  accessibleResources: {
    organizations: string[];
    teams: string[];
    projects: string[];
  };
}

export interface Route {
  path: string;
  method: string;
  name: string;
  group: string;
  category: string;
  permissions: string[];
  icon: string;
  order: number;
  isMenu: boolean;
}

// 获取用户权限
export async function getUserPermissions(): Promise<UserPermissions> {
  const response = await axios.get('/api/v1/user/permissions');
  return response.data.data;
}

// 获取用户可访问路由
export async function getUserRoutes() {
  const response = await axios.get('/api/v1/user/routes');
  return response.data.data;
}
```

### 2. 动态生成路由和菜单

```typescript
// frontend/src/router/index.ts
import { createRouter } from 'vue-router';
import { getUserRoutes } from '@/api/user';

// 动态加载路由
export async function setupDynamicRoutes() {
  const { routes, menu } = await getUserRoutes();
  
  // 根据可访问路由生成前端路由配置
  const dynamicRoutes = routes.map((route: Route) => ({
    path: route.path.replace(/:(\w+)/g, (_, param) => `:${param}`),
    name: route.name,
    meta: {
      title: route.name,
      icon: route.icon,
      group: route.group,
      category: route.category,
      isMenu: route.isMenu,
      order: route.order,
    },
    component: () => import(`@/views/${route.group}/${route.name}.vue`),
  }));
  
  // 添加到路由器
  dynamicRoutes.forEach(route => {
    router.addRoute(route);
  });
  
  return menu;
}
```

### 3. 使用Vuex/Pinia存储权限

```typescript
// frontend/src/store/modules/user.ts
import { defineStore } from 'pinia';
import { getUserPermissions } from '@/api/user';

export const useUserStore = defineStore('user', {
  state: () => ({
    permissions: {} as UserPermissions,
    allPermissions: [] as string[],
    accessibleRoutes: [] as Route[],
  }),
  
  actions: {
    async fetchPermissions() {
      const permissions = await getUserPermissions();
      this.permissions = permissions;
      this.allPermissions = permissions.allPermissions;
      this.accessibleRoutes = permissions.accessibleRoutes;
    },
    
    hasPermission(permission: string): boolean {
      return this.allPermissions.includes(permission);
    },
    
    hasAnyPermission(permissions: string[]): boolean {
      return permissions.some(p => this.allPermissions.includes(p));
    },
    
    hasAllPermissions(permissions: string[]): boolean {
      return permissions.every(p => this.allPermissions.includes(p));
    },
  },
});
```

### 4. 在组件中使用权限

```vue
<template>
  <div>
    <!-- 基于权限显示/隐藏按钮 -->
    <el-button 
      v-if="hasPermission('build.trigger')"
      @click="triggerBuild">
      触发构建
    </el-button>
    
    <el-button 
      v-if="hasAnyPermission(['deploy.execute', 'deploy.approve'])"
      @click="deploy">
      部署
    </el-button>
  </div>
</template>

<script setup lang="ts">
import { useUserStore } from '@/store/modules/user';

const userStore = useUserStore();

const hasPermission = (permission: string) => {
  return userStore.hasPermission(permission);
};

const hasAnyPermission = (permissions: string[]) => {
  return userStore.hasAnyPermission(permissions);
};

const triggerBuild = () => {
  // 触发构建逻辑
};

const deploy = () => {
  // 部署逻辑
};
</script>
```

### 5. 权限指令

```typescript
// frontend/src/directives/permission.ts
import { Directive } from 'vue';
import { useUserStore } from '@/store/modules/user';

export const permissionDirective: Directive = {
  mounted(el, binding) {
    const { value } = binding;
    const userStore = useUserStore();
    
    if (Array.isArray(value)) {
      // v-permission="['perm1', 'perm2']" - 需要任意一个权限
      if (!userStore.hasAnyPermission(value)) {
        el.parentNode?.removeChild(el);
      }
    } else if (typeof value === 'string') {
      // v-permission="'perm1'" - 需要特定权限
      if (!userStore.hasPermission(value)) {
        el.parentNode?.removeChild(el);
      }
    }
  },
};

// 注册指令
app.directive('permission', permissionDirective);
```

使用指令：
```vue
<template>
  <el-button v-permission="'build.trigger'">触发构建</el-button>
  <el-button v-permission="['deploy.execute', 'deploy.approve']">部署</el-button>
</template>
```

## 性能优化

### 1. 缓存权限信息

```go
// 使用Redis缓存用户权限
func (s *UserPermissionsService) GetUserPermissionsWithCache(ctx context.Context, userId string) (*UserPermissionSummary, error) {
    cacheKey := fmt.Sprintf("user:permissions:%s", userId)
    
    // 尝试从缓存获取
    cached, err := s.ctx.Redis.Get(ctx, cacheKey).Result()
    if err == nil {
        var summary UserPermissionSummary
        if err := json.Unmarshal([]byte(cached), &summary); err == nil {
            return &summary, nil
        }
    }
    
    // 缓存未命中，查询数据库
    summary, err := s.GetUserPermissions(ctx, userId)
    if err != nil {
        return nil, err
    }
    
    // 写入缓存（TTL 5分钟）
    data, _ := json.Marshal(summary)
    s.ctx.Redis.Set(ctx, cacheKey, data, 5*time.Minute)
    
    return summary, nil
}
```

### 2. 权限变更时清理缓存

```go
// 当用户权限变更时，清理缓存
func (s *UserPermissionsService) InvalidateUserPermissionsCache(userId string) error {
    cacheKey := fmt.Sprintf("user:permissions:%s", userId)
    return s.ctx.Redis.Del(context.Background(), cacheKey).Err()
}

// 在添加/删除成员时调用
func (r *ProjectMemberRepo) AddMember(projectId, userId, roleId string) error {
    // 添加成员逻辑
    // ...
    
    // 清理用户权限缓存
    userPermService.InvalidateUserPermissionsCache(userId)
    
    return nil
}
```

## 最佳实践

### 1. 在登录后立即获取权限

```go
// 登录接口返回权限信息
func (s *AuthService) Login(username, password string) (*LoginResponse, error) {
    // 验证用户名密码
    user, err := s.validateUser(username, password)
    if err != nil {
        return nil, err
    }
    
    // 生成JWT Token
    token, err := s.jwtManager.GenerateToken(user.UserId, user.Username)
    if err != nil {
        return nil, err
    }
    
    // 获取用户权限
    permissions, err := s.userPermService.GetUserPermissions(context.Background(), user.UserId)
    if err != nil {
        log.Warnf("failed to get user permissions: %v", err)
    }
    
    return &LoginResponse{
        Token:       token,
        User:        user,
        Permissions: permissions,  // 直接返回权限信息
    }, nil
}
```

### 2. 定期刷新权限

前端每5分钟自动刷新一次权限信息，确保权限变更及时生效。

### 3. 权限变更通知

使用WebSocket推送权限变更通知：

```go
// 当用户权限变更时，通过WebSocket通知前端
func (s *UserPermissionsService) NotifyPermissionChange(userId string) {
    wsManager.SendToUser(userId, &Message{
        Type: "PERMISSION_CHANGED",
        Data: map[string]interface{}{
            "userId": userId,
            "timestamp": time.Now().Unix(),
        },
    })
}
```

## 故障排查

### 问题1：用户看不到某些路由

**原因**：
1. 路由权限配置错误
2. 用户没有对应权限
3. 缓存未更新

**解决**：
```bash
# 检查路由配置
SELECT * FROM t_router_permission WHERE route_id = 'xxx';

# 检查用户权限
curl -H "Authorization: Bearer <token>" \
     https://api.example.com/api/v1/user/permissions

# 清理缓存
redis-cli DEL "user:permissions:user_123"
```

### 问题2：权限更新不及时

**原因**：Redis缓存未失效

**解决**：
1. 降低缓存TTL
2. 在权限变更时主动清理缓存
3. 前端定期刷新权限

## 总结

本系统提供了一个完整的用户权限聚合和路由管理解决方案：

✅ **自动聚合** - 多层级权限自动汇总  
✅ **动态路由** - 基于权限动态生成路由  
✅ **灵活检查** - 多种权限检查方式  
✅ **前端集成** - 完整的前端集成方案  
✅ **性能优化** - 缓存机制提升性能  

通过这个系统，可以实现细粒度的权限控制和动态的前端菜单生成。

