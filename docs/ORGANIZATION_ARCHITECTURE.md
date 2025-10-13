# Arcade CI/CD 组织架构设计

## 概述

Arcade CI/CD 平台采用企业级多租户架构，通过 **组织（Organization）** → **团队（Team）** → **用户（User）** 的层级结构，提供灵活的权限管理和资源隔离。

## 架构层级

```
┌─────────────────────────────────────────────────────────┐
│                      组织 (Organization)                  │
│  ┌───────────────────────────────────────────────────┐  │
│  │  - 组织所有者 (Owner)                              │  │
│  │  - 组织管理员 (Admin)                              │  │
│  │  - 组织成员 (Member)                               │  │
│  └───────────────────────────────────────────────────┘  │
│                                                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │   团队 A     │  │   团队 B     │  │   团队 C     │  │
│  ├──────────────┤  ├──────────────┤  ├──────────────┤  │
│  │ Team Owner   │  │ Maintainer   │  │ Developer    │  │
│  │ Developer    │  │ Developer    │  │ Reporter     │  │
│  │ Guest        │  │ Guest        │  │ Guest        │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
│         │                  │                  │          │
│         ▼                  ▼                  ▼          │
│  ┌──────────────────────────────────────────────────┐  │
│  │              项目 (Projects)                     │  │
│  │  - 项目 1 (关联团队 A)                           │  │
│  │  - 项目 2 (关联团队 B, C)                        │  │
│  │  - 项目 3 (组织级别，所有成员可访问)               │  │
│  └──────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

## 核心概念

### 1. 组织 (Organization)

组织是最高层级的资源容器，代表一个公司、部门或团体。

**特性**:
- 每个组织有唯一的名称标识
- 支持订阅计划：free（免费版）、pro（专业版）、enterprise（企业版）
- 可配置最大成员数、项目数、团队数等限制
- 支持 SAML、LDAP 等企业级认证方式

**角色**:
- **Owner（所有者）**: 最高权限，可删除组织
- **Admin（管理员）**: 管理组织、成员、团队、项目
- **Member（成员）**: 基础访问权限

### 2. 团队 (Team)

团队是组织内的工作单元，用于组织用户并分配项目权限。

**特性**:
- 属于某个组织
- 支持层级嵌套（父团队、子团队）
- 可见性控制：私有、内部、公开
- 可配置默认角色、成员邀请等策略

**角色**:
- **Owner（所有者）**: 完全控制团队
- **Maintainer（维护者）**: 管理团队成员和项目
- **Developer（开发者）**: 参与开发
- **Reporter（报告者）**: 报告问题
- **Guest（访客）**: 仅查看

### 3. 项目 (Project)

项目是 CI/CD 的核心实体，包含代码仓库、流水线等资源。

**特性**:
- 必须属于某个组织
- 拥有唯一的命名空间（org_name/project_name）
- 可以关联多个团队
- 可以直接添加用户成员
- 支持三种访问级别：owner（仅所有者）、team（团队成员）、org（组织成员）

**权限来源**:
- **直接成员**: 直接添加到项目的用户
- **团队成员**: 通过团队关联获得权限
- **组织成员**: 通过组织级别获得权限

## 权限模型

### 权限继承关系

```
组织级别权限
    ├─ 组织所有者: 所有资源的完全控制权
    ├─ 组织管理员: 管理所有项目、团队、成员
    └─ 组织成员: 查看公开项目

团队级别权限
    ├─ 团队所有者: 团队及关联项目的完全控制权
    ├─ 团队维护者: 管理团队成员和项目
    └─ 团队开发者: 访问团队关联的项目

项目级别权限
    ├─ 项目所有者: 项目的完全控制权
    ├─ 项目维护者: 管理项目和成员
    └─ 项目开发者: 参与开发
```

### 权限计算规则

用户对项目的最终权限 = **MAX**（直接项目权限, 团队继承权限, 组织继承权限）

**示例**:
- 用户 A 是组织管理员 → 拥有组织内所有项目的管理权限
- 用户 B 是团队开发者 → 拥有团队关联项目的读写权限
- 用户 C 直接被添加为项目维护者 → 拥有该项目的管理权限

## 典型使用场景

### 场景 1: 小型初创公司

```
组织: Startup Inc.
├─ 团队: Engineering
│  ├─ 项目: backend-api
│  ├─ 项目: frontend-web
│  └─ 项目: mobile-app
└─ 团队: Ops
   └─ 项目: infrastructure
```

**特点**:
- 单一组织，少量团队
- 成员可能跨团队协作
- 项目访问级别设为 `org`，方便协作

### 场景 2: 中型企业

```
组织: Tech Corp
├─ 团队: Product Team A
│  ├─ 子团队: Frontend
│  │  └─ 项目: product-a-web
│  └─ 子团队: Backend
│     └─ 项目: product-a-api
├─ 团队: Product Team B
│  └─ 项目: product-b
└─ 团队: Platform
   ├─ 项目: ci-cd-platform
   └─ 项目: monitoring
```

**特点**:
- 多个产品团队，层级结构
- 团队之间权限隔离
- 平台团队的项目供所有团队使用

### 场景 3: 大型企业集团

```
组织: Enterprise Group
├─ 团队: Business Unit 1
│  ├─ 子团队: Development
│  └─ 子团队: QA
├─ 团队: Business Unit 2
│  ├─ 子团队: Development
│  └─ 子团队: QA
└─ 团队: Shared Services
   ├─ 项目: auth-service (共享)
   ├─ 项目: api-gateway (共享)
   └─ 项目: logging-platform (共享)
```

**特点**:
- 多个业务单元，完全隔离
- 共享服务团队提供公共组件
- 细粒度的权限控制

## 项目访问控制

### 访问级别 (access_level)

项目可以设置默认访问级别：

1. **owner**: 仅项目所有者可访问
   - 适用于敏感项目、实验性项目
   
2. **team**: 关联团队成员可访问（默认）
   - 适用于普通团队项目
   - 最常用的模式
   
3. **org**: 组织所有成员可访问
   - 适用于公共库、文档、工具类项目

### 团队关联权限 (project_team_relation)

项目可以关联多个团队，每个团队有不同的访问权限：

- **read**: 只读（查看代码和构建）
- **write**: 读写（提交代码、触发构建）
- **admin**: 管理员（完全控制项目）

**示例**:
```
项目: microservice-api
├─ 团队 Backend (admin) - 完全控制
├─ 团队 Frontend (write) - 可以查看和触发构建
└─ 团队 QA (read) - 只能查看
```

## 成员邀请流程

### 组织邀请

1. 组织管理员发起邀请
2. 系统生成邀请链接（带令牌）
3. 被邀请人通过邮件接收邀请
4. 点击链接加入组织
5. 自动分配指定角色

### 团队添加

1. 团队所有者或维护者添加成员
2. 可以从组织现有成员中选择
3. 分配团队角色
4. 成员立即获得团队关联项目的权限

## 最佳实践

### 1. 组织设计

- ✅ 一个公司/部门创建一个组织
- ✅ 合理设置成员数和项目数限制
- ✅ 启用必要的认证方式（SAML/LDAP）
- ❌ 避免创建过多顶级组织

### 2. 团队划分

- ✅ 按产品线或业务单元划分团队
- ✅ 使用子团队细化职责（前端、后端、测试）
- ✅ 保持团队规模适中（5-20人）
- ❌ 避免团队层级过深（建议不超过3层）

### 3. 项目管理

- ✅ 使用规范的命名空间（org/project）
- ✅ 合理设置项目访问级别
- ✅ 定期审查项目成员和权限
- ❌ 避免过度开放权限

### 4. 权限分配

- ✅ 遵循最小权限原则
- ✅ 优先通过团队分配权限
- ✅ 直接成员仅用于特殊情况
- ❌ 避免给所有人 admin 权限

## 数据库表关系

```sql
-- 组织
t_organization
  ├─ t_organization_member (组织成员)
  ├─ t_organization_invitation (邀请)
  └─ t_team (团队)
      ├─ t_team_member (团队成员)
      └─ t_project_team_relation (项目关联)
          └─ t_project (项目)
              ├─ t_project_member (项目成员)
              ├─ t_project_webhook (Webhook)
              ├─ t_project_variable (变量)
              └─ t_pipeline (流水线)
```

## API 设计建议

### 组织相关

```
GET    /api/v1/organizations                 # 列表
POST   /api/v1/organizations                 # 创建
GET    /api/v1/organizations/:org_id         # 详情
PATCH  /api/v1/organizations/:org_id         # 更新
DELETE /api/v1/organizations/:org_id         # 删除

GET    /api/v1/organizations/:org_id/members # 成员列表
POST   /api/v1/organizations/:org_id/members # 添加成员
DELETE /api/v1/organizations/:org_id/members/:user_id # 移除成员

POST   /api/v1/organizations/:org_id/invitations # 发送邀请
GET    /api/v1/invitations/:token            # 查看邀请
POST   /api/v1/invitations/:token/accept     # 接受邀请
```

### 团队相关

```
GET    /api/v1/organizations/:org_id/teams   # 列表
POST   /api/v1/organizations/:org_id/teams   # 创建
GET    /api/v1/teams/:team_id                # 详情
PATCH  /api/v1/teams/:team_id                # 更新
DELETE /api/v1/teams/:team_id                # 删除

GET    /api/v1/teams/:team_id/members        # 成员列表
POST   /api/v1/teams/:team_id/members        # 添加成员
DELETE /api/v1/teams/:team_id/members/:user_id # 移除成员

GET    /api/v1/teams/:team_id/projects       # 关联项目
POST   /api/v1/teams/:team_id/projects       # 关联项目
DELETE /api/v1/teams/:team_id/projects/:project_id # 取消关联
```

### 项目相关

```
GET    /api/v1/organizations/:org_id/projects # 列表
POST   /api/v1/organizations/:org_id/projects # 创建
GET    /api/v1/projects/:project_id          # 详情
PATCH  /api/v1/projects/:project_id          # 更新
DELETE /api/v1/projects/:project_id          # 删除

GET    /api/v1/projects/:project_id/members  # 成员列表
POST   /api/v1/projects/:project_id/members  # 添加成员
DELETE /api/v1/projects/:project_id/members/:user_id # 移除成员

GET    /api/v1/projects/:project_id/teams    # 关联团队
POST   /api/v1/projects/:project_id/teams    # 关联团队
PATCH  /api/v1/projects/:project_id/teams/:team_id # 更新权限
DELETE /api/v1/projects/:project_id/teams/:team_id # 取消关联
```

## 迁移指南

### 从简单模式迁移到组织模式

如果系统已经有项目和用户，迁移步骤：

1. **创建默认组织**
   ```sql
   INSERT INTO t_organization (org_id, name, display_name, owner_user_id)
   VALUES ('org_default', 'default-org', '默认组织', 'admin_user_id');
   ```

2. **迁移所有用户到组织**
   ```sql
   INSERT INTO t_organization_member (org_id, user_id, role)
   SELECT 'org_default', user_id, 'member' FROM t_user;
   ```

3. **创建默认团队**
   ```sql
   INSERT INTO t_team (team_id, org_id, name, display_name)
   VALUES ('team_default', 'org_default', 'default-team', '默认团队');
   ```

4. **迁移项目**
   ```sql
   UPDATE t_project 
   SET org_id = 'org_default',
       namespace = CONCAT('default-org/', name),
       access_level = 'org';
   ```

5. **关联团队和项目**
   ```sql
   INSERT INTO t_project_team_relation (project_id, team_id, access)
   SELECT project_id, 'team_default', 'admin' FROM t_project;
   ```

## 总结

Arcade CI/CD 的组织架构设计提供了：

- ✅ **灵活的层级结构**: 组织 → 团队 → 项目
- ✅ **细粒度的权限控制**: 多层级角色和权限
- ✅ **多租户隔离**: 完整的资源隔离和配额管理
- ✅ **企业级功能**: SAML/LDAP 认证、审计日志
- ✅ **易于扩展**: 支持从小团队到大型企业

这种架构既适合小型团队的快速上手，也能满足大型企业的复杂需求。

