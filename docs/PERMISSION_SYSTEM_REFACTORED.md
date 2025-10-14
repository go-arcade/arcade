# 权限系统重构总结

## 核心变更

### ✅ 从JSON字段改为关联表设计

**原设计（旧）**：
```
t_role
  - permissions JSON  // 权限存储在JSON字段中
```

**新设计**：
```
t_permission           (权限点表)
  - permission_id
  - code              (如: project.view)
  - name
  - category
  
t_role_permission      (角色-权限关联表)
  - role_id
  - permission_id
```

## 数据库架构

### 表结构关系图

```
┌─────────────┐         ┌──────────────────┐         ┌─────────────┐
│   t_role    │         │ t_role_permission│         │ t_permission│
│             │         │                  │         │             │
│ role_id (PK)│◄────────│ role_id          │         │ permission_ │
│ name        │         │ permission_id    │────────►│   id (PK)   │
│ scope       │         │                  │         │ code        │
│ priority    │         │                  │         │ name        │
└─────────────┘         └──────────────────┘         │ category    │
                                                      └─────────────┘
```

## 迁移步骤

### 1. 运行数据库迁移（按顺序）

```bash
# Step 1: 创建角色表（已移除permissions字段）
mysql -u root -p arcade < docs/database_migration_roles.sql

# Step 2: 创建权限点表和角色权限关联表
mysql -u root -p arcade < docs/database_migration_permissions.sql

# Step 3: 创建路由权限表
mysql -u root -p arcade < docs/database_migration_router_permission.sql
```

### 2. 更新代码依赖

新增文件：
- ✅ `internal/engine/model/model_permission_new.go` - 权限模型
- ✅ `internal/engine/repo/repo_permission.go` - 权限仓库
- ✅ `internal/engine/service/service_user_permissions.go` - 用户权限聚合服务
- ✅ `pkg/http/middleware/user_permissions.go` - 权限中间件

修改文件：
- ✅ `internal/engine/service/service_permission.go` - 使用关联表查询
- ✅ `internal/engine/repo/repo_router_permission.go` - 路由权限仓库

## 权限查询方式变更

### 旧方式（JSON字段）

```go
// 从JSON字段解析
var permissions []string
json.Unmarshal([]byte(role.Permissions), &permissions)
```

### 新方式（关联表）

```go
// 通过JOIN查询
var permissionCodes []string
db.Table("t_role_permission rp").
    Select("p.code").
    Joins("JOIN t_permission p ON rp.permission_id = p.permission_id").
    Where("rp.role_id = ? AND p.is_enabled = ?", roleId, 1).
    Pluck("code", &permissionCodes)
```

## 权限管理API

### 查询角色权限

```go
// 使用PermissionRepo
permRepo := repo.NewPermissionRepo(ctx)

// 获取角色的所有权限代码
permissions, err := permRepo.GetRolePermissions("project_developer")
// 返回: ["project.view", "build.trigger", ...]

// 获取角色的详细权限信息
permissions, err := permRepo.GetRolePermissionsDetailed("project_developer")
// 返回完整的Permission对象列表
```

### 为角色添加/删除权限

```go
permRepo := repo.NewPermissionRepo(ctx)

// 添加权限
permRepo.AddRolePermissionByCode("custom_role", "build.trigger")

// 删除权限
permRepo.RemoveRolePermissionByCode("custom_role", "build.trigger")

// 设置权限（替换所有）
permRepo.SetRolePermissions("custom_role", []string{
    "project.view",
    "build.view",
    "build.trigger",
})

// 检查权限
hasPermission, err := permRepo.HasPermission("project_developer", "build.trigger")
```

### 创建自定义角色

```sql
-- 1. 创建角色
INSERT INTO t_role (role_id, name, display_name, scope, priority, is_builtin, is_enabled)
VALUES ('custom_build_manager', 'build_manager', '构建管理员', 'project', 25, 0, 1);

-- 2. 分配权限
INSERT INTO t_role_permission (role_id, permission_id)
SELECT 'custom_build_manager', permission_id 
FROM t_permission 
WHERE code IN ('project.view', 'build.view', 'build.trigger', 'build.cancel', 'build.retry');
```

或通过代码：

```go
roleRepo := repo.NewRoleRepo(ctx)
permRepo := repo.NewPermissionRepo(ctx)

// 创建角色
role := &model.Role{
    RoleId:      "custom_build_manager",
    Name:        "build_manager",
    DisplayName: "构建管理员",
    Scope:       "project",
    Priority:    25,
    IsBuiltin:   0,
    IsEnabled:   1,
}
roleRepo.CreateRole(role)

// 分配权限
permissions := []string{
    "project.view",
    "build.view",
    "build.trigger",
    "build.cancel",
    "build.retry",
}
permRepo.SetRolePermissions("custom_build_manager", permissions)
```

## HTTP API端点

### 获取用户完整权限

```http
GET /api/v1/user/permissions
Authorization: Bearer <token>

Response:
{
  "userId": "user_123",
  "organizations": [...],
  "teams": [...],
  "projects": [...],
  "allPermissions": [
    "project.view",
    "build.trigger",
    "pipeline.run"
  ],
  "accessibleRoutes": [...],
  "accessibleResources": {
    "organizations": ["org_001"],
    "teams": ["team_001"],
    "projects": ["proj_001"]
  }
}
```

### 获取可访问路由

```http
GET /api/v1/user/routes
Authorization: Bearer <token>

Response:
{
  "routes": [...],
  "menu": {
    "项目管理": [...],
    "CI/CD": [...],
    "部署": [...]
  }
}
```

## 性能优化

### 1. 权限缓存

```go
// 缓存角色权限（5分钟）
cacheKey := fmt.Sprintf("role:permissions:%s", roleId)
cached, err := redis.Get(ctx, cacheKey).Result()
if err == nil {
    var permissions []string
    json.Unmarshal([]byte(cached), &permissions)
    return permissions
}

// 查询数据库
permissions := permRepo.GetRolePermissions(roleId)

// 写入缓存
data, _ := json.Marshal(permissions)
redis.Set(ctx, cacheKey, data, 5*time.Minute)
```

### 2. 批量查询优化

```go
// 一次性查询多个角色的权限
func (r *PermissionRepo) GetMultiRolePermissions(roleIds []string) (map[string][]string, error) {
    type Result struct {
        RoleId string
        Code   string
    }
    
    var results []Result
    err := r.Ctx.DB.Table("t_role_permission rp").
        Select("rp.role_id, p.code").
        Joins("JOIN t_permission p ON rp.permission_id = p.permission_id").
        Where("rp.role_id IN ? AND p.is_enabled = ?", roleIds, 1).
        Scan(&results).Error
    
    if err != nil {
        return nil, err
    }
    
    // 组织结果
    permMap := make(map[string][]string)
    for _, r := range results {
        permMap[r.RoleId] = append(permMap[r.RoleId], r.Code)
    }
    
    return permMap, nil
}
```

## 迁移检查清单

- [x] 数据库表创建（t_permission, t_role_permission）
- [x] 移除 t_role.permissions 字段
- [x] 插入权限点数据
- [x] 为内置角色分配权限
- [x] 创建权限仓库（repo_permission.go）
- [x] 更新权限服务（service_permission.go）
- [x] 更新用户权限服务（service_user_permissions.go）
- [x] 创建路由权限表
- [ ] 更新API路由注册
- [ ] 编写单元测试
- [ ] 前端集成

## 常见操作

### 查询角色的所有权限

```sql
SELECT 
    r.name AS '角色',
    p.category AS '权限分类',
    p.code AS '权限代码',
    p.name AS '权限名称'
FROM t_role r
JOIN t_role_permission rp ON r.role_id = rp.role_id
JOIN t_permission p ON rp.permission_id = p.permission_id
WHERE r.role_id = 'project_developer'
AND p.is_enabled = 1
ORDER BY p.category, p.code;
```

### 为角色批量添加权限

```sql
INSERT INTO t_role_permission (role_id, permission_id)
SELECT 'custom_role', permission_id
FROM t_permission
WHERE code IN ('project.view', 'build.view', 'build.trigger')
AND is_enabled = 1;
```

### 复制角色权限

```sql
-- 将 project_developer 的权限复制给新角色
INSERT INTO t_role_permission (role_id, permission_id)
SELECT 'new_custom_role', permission_id
FROM t_role_permission
WHERE role_id = 'project_developer';
```

### 查找拥有特定权限的所有角色

```sql
SELECT DISTINCT r.role_id, r.name, r.scope
FROM t_role r
JOIN t_role_permission rp ON r.role_id = rp.role_id
JOIN t_permission p ON rp.permission_id = p.permission_id
WHERE p.code = 'build.trigger'
AND r.is_enabled = 1
AND p.is_enabled = 1;
```

## 总结

✅ **标准化设计** - 采用关联表，符合数据库范式  
✅ **灵活扩展** - 新增权限点不需要修改表结构  
✅ **性能优化** - 可通过索引和缓存提升查询速度  
✅ **易于维护** - 权限管理更加清晰  
✅ **向后兼容** - 保留了权限常量，代码改动最小  

权限系统现在更加规范和灵活！🎉

