# Arcade CI/CD 平台数据库设计文档

## 概述

本文档描述了 Arcade CI/CD 平台的完整数据库表结构设计。

## 数据库信息

- **数据库名称**: `arcade_ci_meta`
- **字符集**: `utf8mb4`
- **排序规则**: `utf8mb4_unicode_ci`
- **存储引擎**: InnoDB

## 模块划分

### 1. 用户和权限管理模块

#### 1.1 用户表 (t_user)

存储系统用户基本信息。

| 字段名 | 类型 | 说明 | 索引 |
|--------|------|------|------|
| id | INT | 主键ID | PK |
| user_id | VARCHAR(64) | 用户唯一标识 | UK |
| username | VARCHAR(64) | 用户名 | UK |
| nick_name | VARCHAR(128) | 昵称 | |
| password | VARCHAR(255) | 密码(加密) | |
| avatar | VARCHAR(512) | 头像URL | |
| email | VARCHAR(128) | 邮箱 | UK |
| phone | VARCHAR(32) | 手机号 | |
| is_enabled | TINYINT | 是否启用 | IDX |

#### 1.2 角色表 (t_role)

定义系统角色。

| 字段名 | 类型 | 说明 | 索引 |
|--------|------|------|------|
| id | INT | 主键ID | PK |
| role_id | VARCHAR(64) | 角色唯一标识 | UK |
| role_name | VARCHAR(128) | 角色名称 | |
| role_code | VARCHAR(64) | 角色编码 | UK |
| role_desc | VARCHAR(512) | 角色描述 | |

**默认角色**:
- ADMIN - 管理员
- DEVELOPER - 开发者
- OPERATOR - 运维人员
- VIEWER - 查看者

#### 1.3 用户组表 (t_user_group)

用户组织管理。

| 字段名 | 类型 | 说明 | 索引 |
|--------|------|------|------|
| id | INT | 主键ID | PK |
| group_id | VARCHAR(64) | 用户组唯一标识 | UK |
| group_name | VARCHAR(128) | 用户组名称 | UK |
| group_desc | VARCHAR(512) | 用户组描述 | |

#### 1.4 角色关系表 (t_role_relation)

用户-角色、用户组-角色的关系映射。

| 字段名 | 类型 | 说明 | 索引 |
|--------|------|------|------|
| id | INT | 主键ID | PK |
| role_id | VARCHAR(64) | 角色ID | IDX |
| user_id | VARCHAR(64) | 用户ID | IDX |
| group_id | VARCHAR(64) | 用户组ID | IDX |

#### 1.5 SSO认证提供者表 (t_sso_provider)

统一的SSO认证配置，支持多种认证方式。

| 字段名 | 类型 | 说明 | 索引 |
|--------|------|------|------|
| id | INT | 主键ID | PK |
| provider_id | VARCHAR(64) | 提供者唯一标识 | UK |
| name | VARCHAR(128) | 提供者名称 | |
| provider_type | VARCHAR(32) | 提供者类型 | IDX |
| config | JSON | 配置内容 | |
| description | VARCHAR(512) | 描述 | |
| priority | INT | 优先级 | IDX |
| is_enabled | TINYINT | 是否启用 | |

**支持的认证类型**:
- oauth - OAuth 2.0 (GitHub, GitLab, Google 等)
- ldap - LDAP/Active Directory
- oidc - OpenID Connect (Keycloak, Auth0 等)
- saml - SAML 2.0

**配置结构示例**:

OAuth:
```json
{
  "clientId": "xxx",
  "clientSecret": "xxx",
  "authURL": "https://...",
  "tokenURL": "https://...",
  "userInfoURL": "https://...",
  "scopes": ["read:user", "user:email"]
}
```

LDAP:
```json
{
  "host": "ldap.example.com",
  "port": 389,
  "baseDN": "dc=example,dc=com",
  "bindDN": "cn=admin,dc=example,dc=com",
  "userFilter": "(uid=%s)",
  "attributes": {
    "username": "uid",
    "email": "mail"
  }
}
```

OIDC:
```json
{
  "issuer": "https://...",
  "clientId": "xxx",
  "clientSecret": "xxx",
  "scopes": ["openid", "profile", "email"]
}
```

### 2. Agent管理模块

#### 2.1 Agent表 (t_agent)

任务执行器管理。

| 字段名 | 类型 | 说明 | 索引 |
|--------|------|------|------|
| id | INT | 主键ID | PK |
| agent_id | VARCHAR(64) | Agent唯一标识 | UK |
| agent_name | VARCHAR(128) | Agent名称 | |
| hostname | VARCHAR(255) | 主机名 | |
| address | VARCHAR(255) | Agent地址 | |
| port | VARCHAR(10) | Agent端口 | |
| username | VARCHAR(64) | SSH用户名 | |
| auth_type | TINYINT | 认证类型(0:密码,1:密钥) | |
| os | VARCHAR(32) | 操作系统 | |
| arch | VARCHAR(32) | 架构 | |
| version | VARCHAR(32) | Agent版本 | |
| status | TINYINT | Agent状态 | IDX |
| max_concurrent_jobs | INT | 最大并发任务数 | |
| running_jobs_count | INT | 正在执行的任务数 | |
| labels | JSON | Agent标签 | |
| metrics | JSON | Agent指标 | |
| last_heartbeat | DATETIME | 最后心跳时间 | IDX |
| is_enabled | TINYINT | 是否启用 | IDX |

**Agent状态枚举**:
- 0: UNKNOWN - 未知
- 1: ONLINE - 在线
- 2: OFFLINE - 离线
- 3: BUSY - 忙碌
- 4: IDLE - 空闲

#### 2.2 Agent配置表 (t_agent_config)

Agent 特定的配置项（每个Agent一条记录）。

| 字段名 | 类型 | 说明 | 索引 |
|--------|------|------|------|
| id | INT | 主键ID | PK |
| agent_id | VARCHAR(64) | Agent唯一标识 | UK |
| config_items | JSON | 所有配置项(JSON格式) | |
| description | VARCHAR(512) | 配置描述 | |

**config_items JSON 结构**:
```json
{
  "heartbeat_interval": 30,
  "max_concurrent_jobs": 5,
  "job_timeout": 3600,
  "workspace_dir": "/var/lib/arcade/workspace",
  "temp_dir": "/var/lib/arcade/temp",
  "log_level": "INFO",
  "enable_docker": true,
  "docker_network": "bridge",
  "resource_limits": {
    "cpu": "2",
    "memory": "4G"
  },
  "allowed_commands": ["docker", "kubectl", "npm", "yarn"],
  "env_vars": {
    "PATH": "/usr/local/bin:/usr/bin:/bin"
  },
  "cache_dir": "/var/lib/arcade/cache",
  "cleanup_policy": {
    "max_age_days": 7,
    "max_size_gb": 50
  }
}
```

**配置项说明**:
- `heartbeat_interval` - 心跳间隔(秒)
- `max_concurrent_jobs` - 最大并发任务数
- `job_timeout` - 任务超时时间(秒)
- `workspace_dir` - 工作目录
- `temp_dir` - 临时目录
- `log_level` - 日志级别
- `enable_docker` - 是否启用Docker
- `docker_network` - Docker网络模式
- `resource_limits` - 资源限制(JSON)
- `allowed_commands` - 允许执行的命令白名单(JSON)
- `env_vars` - 环境变量(JSON)
- `ssh_key` - SSH私钥(加密)
- `proxy_url` - 代理地址
- `cache_dir` - 缓存目录
- `cleanup_policy` - 清理策略(JSON)

### 3. 组织和团队管理模块

#### 3.1 组织表 (t_organization)

多租户组织管理，支持企业级权限控制。

| 字段名 | 类型 | 说明 | 索引 |
|--------|------|------|------|
| id | INT | 主键ID | PK |
| org_id | VARCHAR(64) | 组织唯一标识 | UK |
| name | VARCHAR(128) | 组织名称(英文标识) | UK |
| display_name | VARCHAR(255) | 组织显示名称 | |
| description | TEXT | 组织描述 | |
| logo | VARCHAR(512) | 组织Logo URL | |
| website | VARCHAR(512) | 组织官网 | |
| email | VARCHAR(128) | 组织联系邮箱 | |
| phone | VARCHAR(32) | 组织联系电话 | |
| address | VARCHAR(512) | 组织地址 | |
| settings | JSON | 组织设置 | |
| plan | VARCHAR(32) | 订阅计划 | IDX |
| status | TINYINT | 组织状态 | IDX |
| owner_user_id | VARCHAR(64) | 组织所有者用户ID | IDX |
| is_enabled | TINYINT | 是否启用 | |
| total_members | INT | 成员总数 | |
| total_teams | INT | 团队总数 | |
| total_projects | INT | 项目总数 | |

**订阅计划**:
- free - 免费版(基础功能)
- pro - 专业版(高级功能)
- enterprise - 企业版(完整功能+支持)

**组织状态枚举**:
- 0: INACTIVE - 未激活
- 1: ACTIVE - 正常
- 2: FROZEN - 冻结
- 3: DELETED - 已删除

**settings JSON结构**:
```json
{
  "allow_public_projects": true,
  "require_two_factor": false,
  "allowed_domains": ["example.com"],
  "max_members": 100,
  "max_projects": 50,
  "max_teams": 20,
  "max_agents": 10,
  "default_visibility": "private",
  "enable_saml": false,
  "enable_ldap": false
}
```

#### 3.2 组织成员表 (t_organization_member)

组织成员及角色管理。

| 字段名 | 类型 | 说明 | 索引 |
|--------|------|------|------|
| id | INT | 主键ID | PK |
| org_id | VARCHAR(64) | 组织ID | IDX |
| user_id | VARCHAR(64) | 用户ID | IDX |
| role | VARCHAR(32) | 组织角色 | IDX |
| username | VARCHAR(64) | 用户名(冗余) | |
| email | VARCHAR(128) | 邮箱(冗余) | |
| invited_by | VARCHAR(64) | 邀请人用户ID | |
| status | TINYINT | 状态 | IDX |

**组织成员角色**:
- owner - 所有者(最高权限,可删除组织)
- admin - 管理员(管理组织、成员、团队、项目)
- member - 普通成员(基础访问权限)

**成员状态**:
- 0: PENDING - 待接受
- 1: ACTIVE - 正常
- 2: DISABLED - 禁用

#### 3.3 团队表 (t_team)

团队管理，支持层级嵌套。

| 字段名 | 类型 | 说明 | 索引 |
|--------|------|------|------|
| id | INT | 主键ID | PK |
| team_id | VARCHAR(64) | 团队唯一标识 | UK |
| org_id | VARCHAR(64) | 所属组织ID | IDX |
| name | VARCHAR(128) | 团队名称(英文标识) | |
| display_name | VARCHAR(255) | 团队显示名称 | |
| description | TEXT | 团队描述 | |
| avatar | VARCHAR(512) | 团队头像 | |
| parent_team_id | VARCHAR(64) | 父团队ID | IDX |
| path | VARCHAR(512) | 团队路径 | IDX |
| level | INT | 团队层级 | |
| settings | JSON | 团队设置 | |
| visibility | TINYINT | 可见性 | IDX |
| is_enabled | TINYINT | 是否启用 | |
| total_members | INT | 成员总数 | |
| total_projects | INT | 项目总数 | |

**团队可见性枚举**:
- 0: PRIVATE - 私有(仅成员可见)
- 1: INTERNAL - 内部(组织内可见)
- 2: PUBLIC - 公开(所有人可见)

**settings JSON结构**:
```json
{
  "default_role": "developer",
  "allow_member_invite": false,
  "require_approval": true,
  "max_members": 50
}
```

#### 3.4 团队成员表 (t_team_member)

团队成员及权限管理。

| 字段名 | 类型 | 说明 | 索引 |
|--------|------|------|------|
| id | INT | 主键ID | PK |
| team_id | VARCHAR(64) | 团队ID | IDX |
| user_id | VARCHAR(64) | 用户ID | IDX |
| role | VARCHAR(32) | 团队角色 | IDX |
| username | VARCHAR(64) | 用户名(冗余) | |
| invited_by | VARCHAR(64) | 邀请人用户ID | |

**团队成员角色**:
- owner - 所有者(完全控制团队)
- maintainer - 维护者(管理团队成员和项目)
- developer - 开发者(参与开发)
- reporter - 报告者(报告问题)
- guest - 访客(仅查看)

#### 3.5 项目团队关联表 (t_project_team_relation)

项目与团队的关联关系。

| 字段名 | 类型 | 说明 | 索引 |
|--------|------|------|------|
| id | INT | 主键ID | PK |
| project_id | VARCHAR(64) | 项目ID | IDX |
| team_id | VARCHAR(64) | 团队ID | IDX |
| access | VARCHAR(32) | 访问权限 | |

**访问权限**:
- read - 只读(查看代码和构建)
- write - 读写(提交代码、触发构建)
- admin - 管理员(完全控制项目)

#### 3.6 组织邀请表 (t_organization_invitation)

组织成员邀请管理。

| 字段名 | 类型 | 说明 | 索引 |
|--------|------|------|------|
| id | INT | 主键ID | PK |
| invitation_id | VARCHAR(64) | 邀请唯一标识 | UK |
| org_id | VARCHAR(64) | 组织ID | IDX |
| email | VARCHAR(128) | 被邀请人邮箱 | IDX |
| role | VARCHAR(32) | 角色 | |
| token | VARCHAR(255) | 邀请令牌 | UK |
| invited_by | VARCHAR(64) | 邀请人用户ID | |
| status | TINYINT | 状态 | IDX |
| expires_at | DATETIME | 过期时间 | IDX |

**邀请状态**:
- 0: PENDING - 待接受
- 1: ACCEPTED - 已接受
- 2: REJECTED - 已拒绝
- 3: EXPIRED - 已过期

### 4. 项目管理模块

#### 4.1 项目表 (t_project)

项目定义，包含代码仓库信息、触发配置等。

| 字段名 | 类型 | 说明 | 索引 |
|--------|------|------|------|
| id | INT | 主键ID | PK |
| project_id | VARCHAR(64) | 项目唯一标识 | UK |
| org_id | VARCHAR(64) | 所属组织ID | IDX |
| name | VARCHAR(128) | 项目名称(英文标识) | UK |
| display_name | VARCHAR(255) | 项目显示名称 | |
| namespace | VARCHAR(255) | 项目命名空间 | UK |
| description | TEXT | 项目描述 | |
| repo_url | VARCHAR(512) | 代码仓库URL | |
| repo_type | VARCHAR(32) | 仓库类型 | IDX |
| default_branch | VARCHAR(128) | 默认分支 | |
| auth_type | TINYINT | 认证类型 | |
| credential | TEXT | 认证凭证(加密) | |
| trigger_mode | INT | 触发模式(位掩码) | |
| webhook_secret | VARCHAR(255) | Webhook密钥 | |
| cron_expr | VARCHAR(128) | Cron表达式 | |
| build_config | JSON | 构建配置 | |
| env_vars | JSON | 环境变量 | |
| settings | JSON | 项目设置 | |
| tags | VARCHAR(512) | 项目标签 | |
| language | VARCHAR(64) | 主要编程语言 | |
| framework | VARCHAR(128) | 使用的框架 | |
| status | TINYINT | 项目状态 | IDX |
| visibility | TINYINT | 可见性 | IDX |
| group_id | VARCHAR(64) | 所属用户组ID | IDX |
| created_by | VARCHAR(64) | 创建者用户ID | IDX |
| is_enabled | TINYINT | 是否启用 | IDX |
| icon | VARCHAR(512) | 项目图标URL | |
| homepage | VARCHAR(512) | 项目主页 | |
| total_pipelines | INT | 流水线总数 | |
| total_builds | INT | 构建总次数 | |
| success_builds | INT | 成功构建次数 | |
| failed_builds | INT | 失败构建次数 | |

**触发模式枚举(位掩码)**:
- 1: MANUAL - 手动触发
- 2: WEBHOOK - Webhook触发
- 4: SCHEDULE - 定时触发
- 8: PUSH - Push触发
- 16: MR - Merge Request触发
- 32: TAG - Tag触发

可组合使用，例如: 7 (1+2+4) = 手动 + Webhook + 定时

**认证类型枚举**:
- 0: NONE - 无认证(公开仓库)
- 1: PASSWORD - 用户名密码
- 2: TOKEN - Token/PAT
- 3: SSH_KEY - SSH密钥

**仓库类型**:
- git - 通用Git仓库
- github - GitHub
- gitlab - GitLab
- gitee - Gitee码云
- svn - SVN

**项目状态枚举**:
- 0: INACTIVE - 未激活
- 1: ACTIVE - 正常
- 2: ARCHIVED - 归档
- 3: DISABLED - 禁用

**可见性枚举**:
- 0: PRIVATE - 私有(仅成员可见)
- 1: INTERNAL - 内部(登录用户可见)
- 2: PUBLIC - 公开(所有人可见)

**build_config JSON结构**:
```json
{
  "dockerfile": "Dockerfile",
  "build_context": ".",
  "build_args": {
    "VERSION": "1.0.0"
  },
  "cache_enabled": true,
  "test_enabled": true,
  "lint_enabled": true,
  "scan_enabled": false,
  "artifact_paths": ["dist/", "build/"],
  "artifact_expire": 30
}
```

**settings JSON结构**:
```json
{
  "auto_cancel": true,
  "timeout": 3600,
  "max_concurrent": 3,
  "retry_count": 1,
  "notify_on_success": false,
  "notify_on_failure": true,
  "clean_workspace": true,
  "allowed_branches": ["main", "develop"],
  "ignored_branches": ["temp/*"],
  "allowed_paths": ["src/**"],
  "ignored_paths": ["docs/**", "*.md"],
  "badge_enabled": true
}
```

#### 3.2 项目成员表 (t_project_member)

项目成员及权限管理（直接添加的用户）。

| 字段名 | 类型 | 说明 | 索引 |
|--------|------|------|------|
| id | INT | 主键ID | PK |
| project_id | VARCHAR(64) | 项目ID | IDX |
| user_id | VARCHAR(64) | 用户ID | IDX |
| role | VARCHAR(32) | 成员角色 | IDX |
| username | VARCHAR(64) | 用户名(冗余) | |
| source | VARCHAR(32) | 来源 | |

**成员角色**:
- owner - 所有者(完全控制)
- maintainer - 维护者(管理项目、成员)
- developer - 开发者(创建分支、提交代码)
- reporter - 报告者(创建问题、查看)
- guest - 访客(仅查看)

**成员来源**:
- direct - 直接添加到项目
- team - 通过团队获得权限
- org - 通过组织获得权限

> 注：项目实际权限由 `直接成员` + `团队成员` + `组织成员` 共同决定

#### 3.3 项目Webhook表 (t_project_webhook)

项目级Webhook配置。

| 字段名 | 类型 | 说明 | 索引 |
|--------|------|------|------|
| id | INT | 主键ID | PK |
| webhook_id | VARCHAR(64) | Webhook唯一标识 | UK |
| project_id | VARCHAR(64) | 项目ID | IDX |
| name | VARCHAR(128) | Webhook名称 | |
| url | VARCHAR(512) | Webhook URL | |
| secret | VARCHAR(255) | 密钥 | |
| events | JSON | 触发事件列表 | |
| is_enabled | TINYINT | 是否启用 | IDX |
| description | VARCHAR(512) | 描述 | |

**触发事件**:
- push - 代码推送
- pull_request - PR/MR
- tag - 标签创建
- issue - 问题
- release - 发布

#### 3.4 项目变量表 (t_project_variable)

项目级环境变量和密钥管理。

| 字段名 | 类型 | 说明 | 索引 |
|--------|------|------|------|
| id | INT | 主键ID | PK |
| variable_id | VARCHAR(64) | 变量唯一标识 | UK |
| project_id | VARCHAR(64) | 项目ID | IDX |
| key | VARCHAR(255) | 变量键 | |
| value | TEXT | 变量值(加密) | |
| type | VARCHAR(32) | 变量类型 | IDX |
| protected | TINYINT | 是否保护 | |
| masked | TINYINT | 是否掩码 | |
| description | VARCHAR(512) | 描述 | |

**变量类型**:
- env - 环境变量
- secret - 密钥(敏感信息)
- file - 文件内容

### 4. 流水线和任务管理模块

#### 4.1 流水线定义表 (t_pipeline)

流水线配置信息。

| 字段名 | 类型 | 说明 | 索引 |
|--------|------|------|------|
| id | INT | 主键ID | PK |
| pipeline_id | VARCHAR(64) | 流水线唯一标识 | UK |
| project_id | VARCHAR(64) | 所属项目ID | IDX |
| name | VARCHAR(255) | 流水线名称 | IDX |
| description | TEXT | 流水线描述 | |
| repo_url | VARCHAR(512) | 代码仓库URL | |
| branch | VARCHAR(128) | 分支 | |
| status | TINYINT | 流水线状态 | IDX |
| trigger_type | TINYINT | 触发类型 | |
| cron | VARCHAR(128) | Cron表达式 | |
| env | JSON | 全局环境变量 | |
| total_runs | INT | 总执行次数 | |
| success_runs | INT | 成功次数 | |
| failed_runs | INT | 失败次数 | |
| created_by | VARCHAR(64) | 创建者用户ID | IDX |
| is_enabled | TINYINT | 是否启用 | IDX |

**触发类型枚举**:
- 1: MANUAL - 手动触发
- 2: WEBHOOK - Webhook触发
- 3: SCHEDULE - 定时触发
- 4: API - API触发

#### 4.2 流水线执行记录表 (t_pipeline_run)

流水线每次执行的记录。

| 字段名 | 类型 | 说明 | 索引 |
|--------|------|------|------|
| id | INT | 主键ID | PK |
| run_id | VARCHAR(64) | 执行唯一标识 | UK |
| pipeline_id | VARCHAR(64) | 流水线ID | IDX |
| pipeline_name | VARCHAR(255) | 流水线名称 | |
| branch | VARCHAR(128) | 分支 | |
| commit_sha | VARCHAR(64) | Commit SHA | |
| status | TINYINT | 执行状态 | IDX |
| trigger_type | TINYINT | 触发类型 | |
| triggered_by | VARCHAR(64) | 触发者用户ID | IDX |
| env | JSON | 环境变量 | |
| total_jobs | INT | 总任务数 | |
| completed_jobs | INT | 已完成任务数 | |
| failed_jobs | INT | 失败任务数 | |
| running_jobs | INT | 运行中任务数 | |
| current_stage | INT | 当前阶段 | |
| total_stages | INT | 总阶段数 | |
| start_time | DATETIME | 开始时间 | IDX |
| end_time | DATETIME | 结束时间 | |
| duration | BIGINT | 执行时长(毫秒) | |

#### 4.3 流水线阶段表 (t_pipeline_stage)

流水线的阶段配置。

| 字段名 | 类型 | 说明 | 索引 |
|--------|------|------|------|
| id | INT | 主键ID | PK |
| stage_id | VARCHAR(64) | 阶段唯一标识 | UK |
| pipeline_id | VARCHAR(64) | 流水线ID | IDX |
| name | VARCHAR(255) | 阶段名称 | |
| stage_order | INT | 阶段顺序 | IDX |
| parallel | TINYINT | 是否并行执行 | |

#### 4.4 任务表 (t_job)

任务详细信息和执行状态。

| 字段名 | 类型 | 说明 | 索引 |
|--------|------|------|------|
| id | INT | 主键ID | PK |
| job_id | VARCHAR(64) | 任务唯一标识 | UK |
| name | VARCHAR(255) | 任务名称 | |
| pipeline_id | VARCHAR(64) | 所属流水线ID | IDX |
| pipeline_run_id | VARCHAR(64) | 所属流水线执行ID | IDX |
| stage_id | VARCHAR(64) | 所属阶段ID | |
| stage | INT | 阶段序号 | |
| agent_id | VARCHAR(64) | 执行的Agent ID | IDX |
| status | TINYINT | 任务状态 | IDX |
| priority | INT | 优先级 | IDX |
| image | VARCHAR(255) | Docker镜像 | |
| commands | TEXT | 执行命令列表 | |
| workspace | VARCHAR(512) | 工作目录 | |
| env | JSON | 环境变量 | |
| secrets | JSON | 密钥信息 | |
| timeout | INT | 超时时间(秒) | |
| retry_count | INT | 重试次数 | |
| current_retry | INT | 当前重试次数 | |
| allow_failure | TINYINT | 是否允许失败 | |
| label_selector | JSON | 标签选择器 | |
| depends_on | VARCHAR(512) | 依赖的任务ID | |
| exit_code | INT | 退出码 | |
| error_message | TEXT | 错误信息 | |
| start_time | DATETIME | 开始时间 | IDX |
| end_time | DATETIME | 结束时间 | |
| duration | BIGINT | 执行时长(毫秒) | |
| created_by | VARCHAR(64) | 创建者用户ID | |

**任务状态枚举**:
- 1: PENDING - 等待执行
- 2: QUEUED - 已入队
- 3: RUNNING - 执行中
- 4: SUCCESS - 执行成功
- 5: FAILED - 执行失败
- 6: CANCELLED - 已取消
- 7: TIMEOUT - 超时
- 8: SKIPPED - 已跳过

#### 4.5 任务产物表 (t_job_artifact)

任务产生的构建产物。

| 字段名 | 类型 | 说明 | 索引 |
|--------|------|------|------|
| id | INT | 主键ID | PK |
| artifact_id | VARCHAR(64) | 产物唯一标识 | UK |
| job_id | VARCHAR(64) | 任务ID | IDX |
| pipeline_run_id | VARCHAR(64) | 流水线执行ID | IDX |
| name | VARCHAR(255) | 产物名称 | |
| path | VARCHAR(1024) | 产物路径 | |
| destination | VARCHAR(1024) | 目标存储路径 | |
| size | BIGINT | 文件大小(字节) | |
| storage_type | VARCHAR(32) | 存储类型 | |
| storage_path | VARCHAR(1024) | 实际存储路径 | |
| expire | TINYINT | 是否过期 | IDX |
| expire_days | INT | 过期天数 | |
| expired_at | DATETIME | 过期时间 | IDX |

**支持的存储类型**:
- minio
- s3
- oss (阿里云)
- gcs (Google Cloud)
- cos (腾讯云)

### 5. 日志和事件模块

#### 5.1 任务日志 (MongoDB)

任务执行日志存储在 MongoDB 中以提高性能。

**Collection**: `job_logs`

```json
{
  "log_id": "uuid",
  "job_id": "job_xxx",
  "agent_id": "agent_xxx",
  "line_number": 1,
  "content": "log content",
  "timestamp": ISODate,
  "level": "INFO|WARN|ERROR"
}
```

#### 5.2 系统事件表 (t_system_event)

记录系统重要事件。

| 字段名 | 类型 | 说明 | 索引 |
|--------|------|------|------|
| id | BIGINT | 主键ID | PK |
| event_id | VARCHAR(64) | 事件唯一标识 | UK |
| event_type | TINYINT | 事件类型 | IDX |
| resource_type | VARCHAR(32) | 资源类型 | IDX |
| resource_id | VARCHAR(64) | 资源ID | IDX |
| resource_name | VARCHAR(255) | 资源名称 | |
| message | TEXT | 事件消息 | |
| metadata | JSON | 事件元数据 | |
| user_id | VARCHAR(64) | 关联用户ID | IDX |
| create_time | DATETIME | 创建时间 | IDX |

**事件类型枚举**:
- 1: JOB_CREATED - 任务创建
- 2: JOB_STARTED - 任务开始
- 3: JOB_COMPLETED - 任务完成
- 4: JOB_FAILED - 任务失败
- 5: AGENT_ONLINE - Agent上线
- 6: PIPELINE_STARTED - 流水线开始
- 7: PIPELINE_COMPLETED - 流水线完成
- 8: PIPELINE_FAILED - 流水线失败

### 6. 配置管理模块

#### 6.1 对象存储配置表 (t_storage_config)

对象存储配置管理，支持多种存储后端。

| 字段名 | 类型 | 说明 | 索引 |
|--------|------|------|------|
| id | INT | 主键ID | PK |
| storage_id | VARCHAR(64) | 存储唯一标识 | UK |
| name | VARCHAR(128) | 存储名称 | |
| storage_type | VARCHAR(32) | 存储类型 | IDX |
| config | JSON | 存储配置 | |
| description | VARCHAR(512) | 描述 | |
| is_default | TINYINT | 是否默认 | IDX |
| is_enabled | TINYINT | 是否启用 | |

**支持的存储类型**:
- minio - MinIO
- s3 - AWS S3
- oss - 阿里云 OSS
- gcs - Google Cloud Storage
- cos - 腾讯云 COS

**配置结构示例**:

MinIO/S3/OSS/COS:
```json
{
  "endpoint": "localhost:9000",
  "accessKey": "xxx",
  "secretKey": "xxx",
  "bucket": "artifacts",
  "region": "us-east-1",
  "useTLS": false,
  "basePath": "ci-artifacts"
}
```

GCS:
```json
{
  "endpoint": "https://storage.googleapis.com",
  "accessKey": "/path/to/service-account-key.json",
  "bucket": "artifacts",
  "region": "us-central1",
  "basePath": "ci-builds"
}
```

#### 6.2 系统配置表 (t_system_config)

系统级配置项。

| 字段名 | 类型 | 说明 | 索引 |
|--------|------|------|------|
| id | INT | 主键ID | PK |
| config_key | VARCHAR(128) | 配置键 | UK |
| config_value | TEXT | 配置值 | |
| config_type | VARCHAR(32) | 配置类型 | |
| description | VARCHAR(512) | 配置描述 | |
| is_encrypted | TINYINT | 是否加密 | |

#### 6.3 密钥管理表 (t_secret)

敏感信息管理。

| 字段名 | 类型 | 说明 | 索引 |
|--------|------|------|------|
| id | INT | 主键ID | PK |
| secret_id | VARCHAR(64) | 密钥唯一标识 | UK |
| name | VARCHAR(255) | 密钥名称 | IDX |
| secret_type | VARCHAR(32) | 密钥类型 | |
| secret_value | TEXT | 密钥值(加密) | |
| description | VARCHAR(512) | 密钥描述 | |
| scope | VARCHAR(32) | 作用域 | IDX |
| scope_id | VARCHAR(64) | 作用域ID | IDX |
| created_by | VARCHAR(64) | 创建者用户ID | |

**密钥类型**:
- password - 密码
- token - Token
- ssh_key - SSH密钥
- env - 环境变量

**作用域**:
- global - 全局
- pipeline - 流水线级
- user - 用户级

### 7. 审计日志模块

#### 7.1 操作审计日志表 (t_audit_log)

记录用户操作行为。

| 字段名 | 类型 | 说明 | 索引 |
|--------|------|------|------|
| id | BIGINT | 主键ID | PK |
| user_id | VARCHAR(64) | 操作用户ID | IDX |
| username | VARCHAR(64) | 操作用户名 | |
| action | VARCHAR(64) | 操作动作 | IDX |
| resource_type | VARCHAR(32) | 资源类型 | IDX |
| resource_id | VARCHAR(64) | 资源ID | IDX |
| resource_name | VARCHAR(255) | 资源名称 | |
| ip_address | VARCHAR(64) | IP地址 | |
| user_agent | VARCHAR(512) | User Agent | |
| request_params | JSON | 请求参数 | |
| response_status | INT | 响应状态码 | |
| error_message | TEXT | 错误信息 | |
| create_time | DATETIME | 操作时间 | IDX |

## 数据关系图

```
用户模块:
t_user ─┬─ t_role_relation ─── t_role
        └─ t_role_relation ─── t_user_group

组织和团队模块:
t_organization ─┬─ t_organization_member ─── t_user
                ├─ t_organization_invitation
                ├─ t_team ─┬─ t_team_member ─── t_user
                │          └─ t_project_team_relation ─── t_project
                └─ t_project (一对多)

项目模块:
t_project ─┬─ t_project_member ─── t_user
           ├─ t_project_webhook
           ├─ t_project_variable
           ├─ t_project_team_relation ─── t_team
           └─ t_pipeline (一对多)

流水线模块:
t_pipeline ─┬─ t_pipeline_run ─── t_job ─── t_job_artifact
            └─ t_pipeline_stage

Agent模块:
t_agent ─── t_job (执行关系)

事件模块:
t_system_event ─── 各资源表
t_audit_log ─── t_user
```

## 权限继承关系

项目的访问权限采用多层继承模型：

```
组织所有者 (org owner)
    └─> 拥有组织内所有资源的完全控制权

组织管理员 (org admin)
    └─> 可以管理组织、创建团队、管理所有项目

组织成员 (org member)
    └─> 基础访问权限，可以查看公开项目

团队所有者 (team owner)
    └─> 拥有团队及关联项目的完全控制权

团队维护者 (team maintainer)
    └─> 可以管理团队成员和关联的项目

团队开发者 (team developer)
    └─> 可以访问团队关联的项目并进行开发

项目所有者 (project owner)
    └─> 拥有项目的完全控制权

项目成员 (project member)
    └─> 根据分配的角色拥有对应权限
```

**权限计算规则**:
1. 用户的最终权限 = MAX(直接项目权限, 团队继承权限, 组织继承权限)
2. 组织所有者和管理员拥有组织内所有项目的管理权限
3. 团队成员通过团队关联获得项目访问权限
4. 项目可以设置 `access_level` 控制默认访问范围
   - owner: 仅项目所有者可访问
   - team: 关联团队成员可访问（默认）
   - org: 组织所有成员可访问

## 索引策略

### 主要索引
1. **主键索引**: 所有表的 `id` 字段
2. **唯一索引**: 所有表的业务主键(如 `user_id`, `job_id` 等)
3. **查询索引**: 高频查询字段(如 `status`, `create_time` 等)
4. **外键索引**: 关联查询字段(如 `pipeline_id`, `agent_id` 等)

### 复合索引建议
- `t_job`: (`pipeline_id`, `status`, `start_time`)
- `t_pipeline_run`: (`pipeline_id`, `status`, `start_time`)
- `t_system_event`: (`resource_type`, `resource_id`, `create_time`)

## 性能优化建议

### 1. 分区策略
- `t_audit_log`: 按月分区
- `t_system_event`: 按月分区
- `t_job`: 可考虑按季度分区

### 2. 归档策略
- 任务日志: 保留90天，超期归档到对象存储
- 审计日志: 保留1年，超期归档
- 系统事件: 保留180天，超期归档

### 3. 缓存策略
- 用户信息: Redis缓存，TTL 30分钟
- Agent状态: Redis缓存，TTL 1分钟
- 流水线配置: Redis缓存，TTL 10分钟
- 系统配置: Redis缓存，更新时清除

## 数据迁移

### 版本管理
使用数据库迁移工具(如 golang-migrate)管理表结构版本。

### 迁移脚本命名
```
000001_init_tables.up.sql
000001_init_tables.down.sql
000002_add_pipeline_tables.up.sql
000002_add_pipeline_tables.down.sql
```

## 安全考虑

1. **密码加密**: 使用 bcrypt 加密存储
2. **敏感字段加密**: `t_secret.secret_value` 使用 AES-256 加密
3. **SSO配置加密**: `t_sso_provider.config` 中的敏感信息(clientSecret, bindPassword等)需要加密
4. **SQL注入防护**: 使用参数化查询
5. **访问控制**: 基于角色的数据访问控制(RBAC)
6. **LDAP认证**: 支持TLS加密连接
7. **OIDC验证**: 支持JWT签名验证

## 备份策略

1. **全量备份**: 每天凌晨执行
2. **增量备份**: 每小时执行
3. **binlog备份**: 实时备份
4. **备份保留**: 全量备份保留30天，增量备份保留7天

