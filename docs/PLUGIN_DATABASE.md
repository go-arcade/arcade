# 插件数据库管理指南

## 概述

Arcade 插件管理器现在支持从数据库加载插件配置，实现插件的动态管理和配置。

## 架构设计

### 三层结构

```
┌─────────────────────────────────────┐
│   Plugin Manager (插件管理器)         │
│   - 加载和管理插件实例                 │
│   - 执行插件生命周期                   │
└─────────────────┬───────────────────┘
                  │
                  ↓
┌─────────────────────────────────────┐
│   Plugin Repository (插件仓库)        │
│   - 从数据库读取插件配置                │
│   - 管理插件参数和配置项                │
└─────────────────┬───────────────────┘
                  │
                  ↓
┌─────────────────────────────────────┐
│   Database Tables (数据库表)         │
│   - t_plugin (插件表)                │
│   - t_plugin_config (配置表)         │
│   - t_job_plugin (任务关联表)        │
└─────────────────────────────────────┘
```

## 数据库表说明

### 1. t_plugin - 插件表

存储插件的元信息和 Schema 定义。

**关键字段**:
- `plugin_id` - 插件唯一标识
- `plugin_type` - 插件类型 (notify/deploy/build/test/custom)
- `entry_point` - 插件入口文件路径
- **`config_schema`** - 配置项的 JSON Schema 定义
- **`params_schema`** - 参数的 JSON Schema 定义
- `default_config` - 默认配置值
- `is_builtin` - 是否内置插件
- `is_enabled` - 是否启用

### 2. t_plugin_config - 插件配置表

存储插件的具体配置实例。

**关键字段**:
- `plugin_id` - 关联的插件ID
- **`config_items`** - 配置项内容 (JSON格式)
- `scope` - 作用域 (global/pipeline/user)
- `scope_id` - 作用域ID
- `is_default` - 是否默认配置

### 3. t_job_plugin - 任务插件关联表

任务与插件的关联及执行参数。

**关键字段**:
- `job_id` - 任务ID
- `plugin_id` - 插件ID
- `plugin_config_id` - 使用的配置ID
- **`params`** - 任务特定的参数 (JSON格式)
- `execution_stage` - 执行阶段
- `execution_order` - 执行顺序
- `status` - 执行状态

## Config vs Params 说明

### Config Items (配置项)
- **定义在**: `t_plugin_config.config_items`
- **作用**: 插件的**全局配置**，通常包含连接信息、认证信息等
- **作用域**: 可以是全局、流水线级或用户级
- **示例**: SMTP 服务器地址、Webhook URL、API Token

### Params (参数)
- **定义在**: `t_job_plugin.params`
- **作用**: **任务特定**的执行参数
- **作用域**: 仅针对单个任务
- **示例**: 邮件收件人、通知消息内容、部署命名空间

### 示例对比

#### Slack 通知插件

**Config Items** (全局配置):
```json
{
  "webhook_url": "https://hooks.slack.com/services/xxx",
  "channel": "#ci-notifications",
  "username": "Arcade CI"
}
```

**Params** (任务参数):
```json
{
  "message": "Build #123 completed successfully",
  "mention_users": ["@john", "@jane"]
}
```

#### Email 通知插件

**Config Items** (全局配置):
```json
{
  "smtp_host": "smtp.gmail.com",
  "smtp_port": 587,
  "smtp_user": "noreply@example.com",
  "smtp_password": "xxx",
  "from_address": "ci@example.com"
}
```

**Params** (任务参数):
```json
{
  "to": ["dev@example.com", "qa@example.com"],
  "subject": "Build Failed: Project X",
  "body": "详细错误信息..."
}
```

## 启动流程

### 1. 应用启动时

```go
// main.go 或 app.go
func main() {
    // Wire 自动注入
    app, cleanup, err := initApp(configPath, appCtx, logger)
    
    // 插件管理器已经在 ProvidePluginManager 中：
    // 1. 从数据库加载插件
    // 2. 初始化所有插件
    // 3. 如果数据库失败，回退到文件系统加载
}
```

### 2. 插件管理器启动流程

```
1. ProvidePluginManager 被调用
   ↓
2. SetPluginRepository (设置数据库仓库)
   ↓
3. LoadPluginsFromDatabase (从数据库加载)
   ├─ 查询所有 is_enabled=1 的插件
   ├─ 读取 entry_point 或 install_path
   ├─ 加载 .so 插件文件
   ├─ 应用 default_config
   └─ 注册到管理器
   ↓
4. Init (初始化所有插件)
   ↓
5. 插件就绪，可供使用
```

## 插件注册示例

### 路径说明

**数据库中存储相对路径**:
```sql
entry_point = 'plugins/notify/slack.so'
install_path = 'plugins/notify/slack.so'
```

**加载时转换为绝对路径**:
- 程序运行目录: `/Users/gagral/go/src/arcade`
- 相对路径: `plugins/notify/slack.so`
- 转换后绝对路径: `/Users/gagral/go/src/arcade/plugins/notify/slack.so`

**支持两种路径**:
1. **相对路径** (推荐): `plugins/notify/slack.so` - 基于程序运行目录
2. **绝对路径**: `/opt/arcade/plugins/slack.so` - 直接使用

### 添加新插件到数据库

```sql
INSERT INTO `t_plugin` (
  `plugin_id`, 
  `name`, 
  `version`, 
  `description`,
  `plugin_type`, 
  `entry_point`,  -- 相对路径
  `config_schema`,
  `params_schema`,
  `default_config`,
  `is_builtin`,
  `is_enabled`
) VALUES (
  'notify_custom',
  'Custom Notification',
  '1.0.0',
  '自定义通知插件',
  'notify',
  'plugins/notify/custom.so',
  '{
    "type": "object",
    "properties": {
      "api_url": {"type": "string", "description": "API endpoint"},
      "api_key": {"type": "string", "description": "API key"}
    },
    "required": ["api_url", "api_key"]
  }',
  '{
    "type": "object",
    "properties": {
      "title": {"type": "string"},
      "body": {"type": "string"}
    }
  }',
  '{
    "timeout": 30,
    "retry": 3
  }',
  0,  -- 非内置
  1   -- 启用
);
```

### 添加插件配置

```sql
INSERT INTO `t_plugin_config` (
  `config_id`,
  `plugin_id`,
  `name`,
  `config_items`,
  `scope`,
  `is_default`
) VALUES (
  'config_slack_prod',
  'notify_slack',
  'Production Slack',
  '{
    "webhook_url": "https://hooks.slack.com/services/xxx",
    "channel": "#production",
    "username": "Arcade Prod CI"
  }',
  'global',
  1  -- 设为默认
);
```

### 为任务添加插件

```sql
INSERT INTO `t_job_plugin` (
  `job_id`,
  `plugin_id`,
  `plugin_config_id`,
  `params`,
  `execution_order`,
  `execution_stage`
) VALUES (
  'job_123',
  'notify_slack',
  'config_slack_prod',
  '{
    "message": "Build completed!",
    "mention_users": ["@devops"]
  }',
  1,
  'on_success'  -- 成功后执行
);
```

## 插件执行阶段

| 阶段 | 说明 | 使用场景 |
|------|------|---------|
| before | 任务执行前 | 环境准备、依赖检查 |
| after | 任务执行后 | 清理资源、收集日志 |
| on_success | 成功后 | 发送成功通知、部署应用 |
| on_failure | 失败后 | 发送失败通知、回滚操作 |

## 插件类型

| 类型 | 说明 | 示例 |
|------|------|------|
| notify | 通知插件 | Slack, Email, 钉钉 |
| deploy | 部署插件 | Kubernetes, Docker |
| build | 构建插件 | Docker Build, Maven |
| test | 测试插件 | JUnit, Coverage |
| custom | 自定义插件 | 任意功能 |

## 管理操作

### 启用/禁用插件

```sql
-- 禁用插件
UPDATE `t_plugin` SET `is_enabled` = 0 WHERE `plugin_id` = 'notify_old';

-- 启用插件
UPDATE `t_plugin` SET `is_enabled` = 1 WHERE `plugin_id` = 'notify_new';
```

### 升级插件版本

```sql
-- 插入新版本
INSERT INTO `t_plugin` (..., `version`, ...) VALUES (..., '2.0.0', ...);

-- 禁用旧版本
UPDATE `t_plugin` 
SET `is_enabled` = 0 
WHERE `plugin_id` = 'notify_slack' AND `version` = '1.0.0';
```

### 查询插件使用情况

```sql
-- 查看哪些任务使用了某个插件
SELECT jp.job_id, j.name, jp.execution_stage, jp.status
FROM t_job_plugin jp
LEFT JOIN t_job j ON jp.job_id = j.job_id
WHERE jp.plugin_id = 'notify_slack';

-- 查看插件的配置使用情况
SELECT pc.name, pc.scope, COUNT(*) as usage_count
FROM t_plugin_config pc
LEFT JOIN t_job_plugin jp ON pc.config_id = jp.plugin_config_id
WHERE pc.plugin_id = 'notify_slack'
GROUP BY pc.config_id;
```

## 最佳实践

### 1. 配置分离
- 敏感信息（密钥、密码）放在 `config_items`
- 业务参数（收件人、消息）放在 `params`

### 2. 作用域管理
- **global**: 公司级配置（如统一的 SMTP 服务器）
- **pipeline**: 项目级配置（如项目专属 Slack 频道）
- **user**: 个人配置（如个人邮箱通知）

### 3. 默认配置
- 每个插件至少有一个 `is_default=1` 的配置
- 简化任务配置，不需要每次都指定配置

### 4. 版本管理
- 使用语义化版本 (1.0.0)
- 保留旧版本以支持回滚
- 通过 `is_enabled` 控制版本切换

### 5. 安全建议
- 敏感配置项使用加密存储
- 限制插件的文件系统访问权限
- 验证插件文件的校验和 (checksum)

## 故障排查

### 插件加载失败

**检查清单**:
1. 插件文件路径是否正确
2. 插件文件是否存在且有执行权限
3. 插件版本是否兼容
4. 配置 Schema 是否正确

**查看日志**:
```
loading N plugins from database
loaded plugin XXX (vX.X.X) from database: /path/to/plugin.so
successfully loaded M/N plugins from database
```

### 配置验证失败

1. 检查 `config_items` 是否符合 `config_schema`
2. 使用 JSON Schema 验证工具验证
3. 查看必填字段是否都已提供

### 参数验证失败

1. 检查 `params` 是否符合 `params_schema`
2. 确认任务参数类型正确
3. 检查必填参数

## 监控和运维

### 关键指标

- 插件加载成功率
- 插件执行成功率
- 插件平均执行时长
- 插件错误率

### 日志监控

```bash
# 查看插件加载日志
grep "loaded plugin" logs/arcade.log

# 查看插件执行失败
grep "failed to load plugin" logs/arcade.log

# 查看插件初始化错误
grep "failed to initialize plugins" logs/arcade.log
```

## 迁移指南

### 从文件配置迁移到数据库

**步骤 1**: 导出现有插件配置
```yaml
# conf.d/plugins.yaml
plugins:
  - name: notify_slack
    path: plugins/notify/slack.so
    type: notify
    config:
      webhook_url: https://...
```

**步骤 2**: 导入到数据库
```sql
INSERT INTO t_plugin (...) VALUES (...);
INSERT INTO t_plugin_config (...) VALUES (...);
```

**步骤 3**: 重启应用
- 插件管理器会自动从数据库加载
- 文件配置作为后备方案

## 开发插件

### 插件接口

```go
type NotifyPlugin interface {
    BasePlugin
    Notify(ctx context.Context, params map[string]interface{}) error
}
```

### 注册插件到数据库

开发完成后，需要：

1. 编译插件为 `.so` 文件
2. 将插件文件放到指定目录
3. 在数据库中注册插件信息
4. 定义 config_schema 和 params_schema
5. 创建默认配置
6. 启用插件

### Schema 示例

**config_schema**:
```json
{
  "type": "object",
  "properties": {
    "api_url": {
      "type": "string",
      "description": "API endpoint URL"
    },
    "api_key": {
      "type": "string",
      "description": "API authentication key"
    },
    "timeout": {
      "type": "integer",
      "description": "Request timeout in seconds",
      "default": 30
    }
  },
  "required": ["api_url", "api_key"]
}
```

**params_schema**:
```json
{
  "type": "object",
  "properties": {
    "message": {
      "type": "string",
      "description": "Notification message"
    },
    "priority": {
      "type": "string",
      "enum": ["low", "normal", "high"],
      "default": "normal"
    }
  },
  "required": ["message"]
}
```

## API 集成示例

### 获取可用插件列表

```http
GET /api/v1/plugins?type=notify&enabled=1
```

### 获取插件详情

```http
GET /api/v1/plugins/{plugin_id}
```

### 创建插件配置

```http
POST /api/v1/plugins/{plugin_id}/configs
{
  "name": "My Slack Config",
  "config_items": {
    "webhook_url": "...",
    "channel": "#builds"
  },
  "scope": "pipeline",
  "scope_id": "pipeline_123"
}
```

### 为任务添加插件

```http
POST /api/v1/jobs/{job_id}/plugins
{
  "plugin_id": "notify_slack",
  "plugin_config_id": "config_slack_prod",
  "params": {
    "message": "Build completed!"
  },
  "execution_stage": "on_success"
}
```

## 参考文档

- [PLUGIN_DEVELOPMENT.md](./PLUGIN_DEVELOPMENT.md) - 插件开发指南
- [PLUGIN_REFERENCE.md](./PLUGIN_REFERENCE.md) - 插件API参考
- [JSON Schema](https://json-schema.org/) - JSON Schema 规范

