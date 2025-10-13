-- ================================================================
-- 项目管理模块表结构设计
-- ================================================================

-- 项目表
CREATE TABLE IF NOT EXISTS `t_project` (
  `id` INT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `project_id` VARCHAR(64) NOT NULL COMMENT '项目唯一标识',
  `name` VARCHAR(128) NOT NULL COMMENT '项目名称(英文标识)',
  `display_name` VARCHAR(255) NOT NULL COMMENT '项目显示名称',
  `description` TEXT DEFAULT NULL COMMENT '项目描述',
  `repo_url` VARCHAR(512) NOT NULL COMMENT '代码仓库URL',
  `repo_type` VARCHAR(32) NOT NULL DEFAULT 'git' COMMENT '仓库类型(git/github/gitlab/gitee/svn)',
  `default_branch` VARCHAR(128) NOT NULL DEFAULT 'main' COMMENT '默认分支',
  `auth_type` TINYINT NOT NULL DEFAULT 0 COMMENT '认证类型: 0-无, 1-用户名密码, 2-Token, 3-SSH密钥',
  `credential` TEXT DEFAULT NULL COMMENT '认证凭证(加密存储)',
  `trigger_mode` INT NOT NULL DEFAULT 1 COMMENT '触发模式(位掩码): 1-手动, 2-Webhook, 4-定时, 8-Push, 16-MR, 32-Tag',
  `webhook_secret` VARCHAR(255) DEFAULT NULL COMMENT 'Webhook密钥',
  `cron_expr` VARCHAR(128) DEFAULT NULL COMMENT '定时任务Cron表达式',
  `build_config` JSON DEFAULT NULL COMMENT '构建配置(JSON格式)',
  `env_vars` JSON DEFAULT NULL COMMENT '环境变量(JSON格式)',
  `settings` JSON DEFAULT NULL COMMENT '项目设置(JSON格式)',
  `tags` VARCHAR(512) DEFAULT NULL COMMENT '项目标签(逗号分隔)',
  `language` VARCHAR(64) DEFAULT NULL COMMENT '主要编程语言(Go/Java/Python/Node.js等)',
  `framework` VARCHAR(128) DEFAULT NULL COMMENT '使用的框架',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '项目状态: 0-未激活, 1-正常, 2-归档, 3-禁用',
  `visibility` TINYINT NOT NULL DEFAULT 0 COMMENT '可见性: 0-私有, 1-内部, 2-公开',
  `group_id` VARCHAR(64) DEFAULT NULL COMMENT '所属用户组ID',
  `created_by` VARCHAR(64) NOT NULL COMMENT '创建者用户ID',
  `is_enabled` TINYINT NOT NULL DEFAULT 1 COMMENT '是否启用: 0-禁用, 1-启用',
  `icon` VARCHAR(512) DEFAULT NULL COMMENT '项目图标URL',
  `homepage` VARCHAR(512) DEFAULT NULL COMMENT '项目主页',
  `total_pipelines` INT NOT NULL DEFAULT 0 COMMENT '流水线总数',
  `total_builds` INT NOT NULL DEFAULT 0 COMMENT '构建总次数',
  `success_builds` INT NOT NULL DEFAULT 0 COMMENT '成功构建次数',
  `failed_builds` INT NOT NULL DEFAULT 0 COMMENT '失败构建次数',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_project_id` (`project_id`),
  UNIQUE KEY `uk_name` (`name`),
  KEY `idx_status` (`status`),
  KEY `idx_visibility` (`visibility`),
  KEY `idx_group_id` (`group_id`),
  KEY `idx_created_by` (`created_by`),
  KEY `idx_is_enabled` (`is_enabled`),
  KEY `idx_repo_type` (`repo_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='项目表';

-- 项目成员表
CREATE TABLE IF NOT EXISTS `t_project_member` (
  `id` INT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `project_id` VARCHAR(64) NOT NULL COMMENT '项目ID',
  `user_id` VARCHAR(64) NOT NULL COMMENT '用户ID',
  `role` VARCHAR(32) NOT NULL COMMENT '角色(owner/maintainer/developer/reporter/guest)',
  `username` VARCHAR(64) DEFAULT NULL COMMENT '用户名(冗余)',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_project_user` (`project_id`, `user_id`),
  KEY `idx_project_id` (`project_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_role` (`role`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='项目成员表';

-- 项目Webhook表
CREATE TABLE IF NOT EXISTS `t_project_webhook` (
  `id` INT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `webhook_id` VARCHAR(64) NOT NULL COMMENT 'Webhook唯一标识',
  `project_id` VARCHAR(64) NOT NULL COMMENT '项目ID',
  `name` VARCHAR(128) NOT NULL COMMENT 'Webhook名称',
  `url` VARCHAR(512) NOT NULL COMMENT 'Webhook URL',
  `secret` VARCHAR(255) DEFAULT NULL COMMENT '密钥',
  `events` JSON NOT NULL COMMENT '触发事件列表(push/merge_request/tag等)',
  `is_enabled` TINYINT NOT NULL DEFAULT 1 COMMENT '是否启用: 0-禁用, 1-启用',
  `description` VARCHAR(512) DEFAULT NULL COMMENT '描述',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_webhook_id` (`webhook_id`),
  KEY `idx_project_id` (`project_id`),
  KEY `idx_is_enabled` (`is_enabled`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='项目Webhook表';

-- 项目变量表
CREATE TABLE IF NOT EXISTS `t_project_variable` (
  `id` INT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `variable_id` VARCHAR(64) NOT NULL COMMENT '变量唯一标识',
  `project_id` VARCHAR(64) NOT NULL COMMENT '项目ID',
  `key` VARCHAR(255) NOT NULL COMMENT '变量键',
  `value` TEXT NOT NULL COMMENT '变量值(敏感信息加密存储)',
  `type` VARCHAR(32) NOT NULL DEFAULT 'env' COMMENT '类型(env/secret/file)',
  `protected` TINYINT NOT NULL DEFAULT 0 COMMENT '是否保护(仅在保护分支可用): 0-否, 1-是',
  `masked` TINYINT NOT NULL DEFAULT 0 COMMENT '是否掩码(日志中隐藏): 0-否, 1-是',
  `description` VARCHAR(512) DEFAULT NULL COMMENT '描述',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_variable_id` (`variable_id`),
  UNIQUE KEY `uk_project_key` (`project_id`, `key`),
  KEY `idx_project_id` (`project_id`),
  KEY `idx_type` (`type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='项目变量表';

-- 修改流水线表，增加项目关联
ALTER TABLE `t_pipeline` ADD COLUMN `project_id` VARCHAR(64) DEFAULT NULL COMMENT '所属项目ID' AFTER `pipeline_id`;
ALTER TABLE `t_pipeline` ADD KEY `idx_project_id` (`project_id`);

-- ================================================================
-- 示例数据
-- ================================================================

-- 插入示例项目
INSERT INTO `t_project` (
  `project_id`, 
  `name`, 
  `display_name`, 
  `description`, 
  `repo_url`, 
  `repo_type`, 
  `default_branch`, 
  `auth_type`,
  `trigger_mode`,
  `build_config`,
  `env_vars`,
  `settings`,
  `tags`,
  `language`,
  `framework`,
  `status`,
  `visibility`,
  `created_by`
) VALUES (
  'proj_demo',
  'demo-project',
  '演示项目',
  '这是一个示例项目，用于演示 Arcade CI/CD 平台的功能',
  'https://github.com/example/demo-project.git',
  'github',
  'main',
  2,
  7,
  JSON_OBJECT(
    'dockerfile', 'Dockerfile',
    'build_context', '.',
    'cache_enabled', true,
    'test_enabled', true,
    'lint_enabled', true,
    'scan_enabled', false,
    'artifact_paths', JSON_ARRAY('dist/', 'build/'),
    'artifact_expire', 30
  ),
  JSON_OBJECT(
    'NODE_ENV', 'production',
    'APP_NAME', 'Demo App'
  ),
  JSON_OBJECT(
    'auto_cancel', true,
    'timeout', 3600,
    'max_concurrent', 3,
    'retry_count', 1,
    'notify_on_success', false,
    'notify_on_failure', true,
    'clean_workspace', true,
    'allowed_branches', JSON_ARRAY('main', 'develop', 'release/*'),
    'badge_enabled', true
  ),
  'demo,example,ci-cd',
  'Go',
  'Gin',
  1,
  1,
  'user_admin'
) ON DUPLICATE KEY UPDATE `name` = `name`;

-- 插入项目成员
INSERT INTO `t_project_member` (`project_id`, `user_id`, `role`, `username`)
VALUES 
  ('proj_demo', 'user_admin', 'owner', 'admin')
ON DUPLICATE KEY UPDATE `role` = `role`;

-- 插入项目变量
INSERT INTO `t_project_variable` (`variable_id`, `project_id`, `key`, `value`, `type`, `protected`, `masked`, `description`)
VALUES 
  ('var_001', 'proj_demo', 'DATABASE_URL', 'mysql://user:pass@localhost:3306/db', 'secret', 1, 1, '数据库连接URL'),
  ('var_002', 'proj_demo', 'API_KEY', 'your-api-key-here', 'secret', 1, 1, 'API密钥'),
  ('var_003', 'proj_demo', 'BUILD_ENV', 'production', 'env', 0, 0, '构建环境')
ON DUPLICATE KEY UPDATE `key` = `key`;

-- 插入项目Webhook
INSERT INTO `t_project_webhook` (`webhook_id`, `project_id`, `name`, `url`, `events`, `description`)
VALUES (
  'webhook_001',
  'proj_demo',
  'GitHub Webhook',
  'https://api.github.com/repos/example/demo-project/hooks',
  JSON_ARRAY('push', 'pull_request', 'tag'),
  'GitHub 自动触发 Webhook'
) ON DUPLICATE KEY UPDATE `name` = `name`;
