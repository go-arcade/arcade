-- ================================================================
-- 组织和团队管理模块表结构设计
-- ================================================================

-- 组织表
CREATE TABLE IF NOT EXISTS `t_organization` (
  `id` INT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `org_id` VARCHAR(64) NOT NULL COMMENT '组织唯一标识',
  `name` VARCHAR(128) NOT NULL COMMENT '组织名称(英文标识)',
  `display_name` VARCHAR(255) NOT NULL COMMENT '组织显示名称',
  `description` TEXT DEFAULT NULL COMMENT '组织描述',
  `logo` VARCHAR(512) DEFAULT NULL COMMENT '组织Logo URL',
  `website` VARCHAR(512) DEFAULT NULL COMMENT '组织官网',
  `email` VARCHAR(128) DEFAULT NULL COMMENT '组织联系邮箱',
  `phone` VARCHAR(32) DEFAULT NULL COMMENT '组织联系电话',
  `address` VARCHAR(512) DEFAULT NULL COMMENT '组织地址',
  `settings` JSON DEFAULT NULL COMMENT '组织设置(JSON格式)',
  `plan` VARCHAR(32) NOT NULL DEFAULT 'free' COMMENT '订阅计划(free/pro/enterprise)',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态: 0-未激活, 1-正常, 2-冻结, 3-已删除',
  `owner_user_id` VARCHAR(64) NOT NULL COMMENT '组织所有者用户ID',
  `is_enabled` TINYINT NOT NULL DEFAULT 1 COMMENT '是否启用: 0-禁用, 1-启用',
  `total_members` INT NOT NULL DEFAULT 0 COMMENT '成员总数',
  `total_teams` INT NOT NULL DEFAULT 0 COMMENT '团队总数',
  `total_projects` INT NOT NULL DEFAULT 0 COMMENT '项目总数',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_org_id` (`org_id`),
  UNIQUE KEY `uk_name` (`name`),
  KEY `idx_status` (`status`),
  KEY `idx_owner_user_id` (`owner_user_id`),
  KEY `idx_plan` (`plan`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='组织表';

-- 组织成员表
CREATE TABLE IF NOT EXISTS `t_organization_member` (
  `id` INT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `org_id` VARCHAR(64) NOT NULL COMMENT '组织ID',
  `user_id` VARCHAR(64) NOT NULL COMMENT '用户ID',
  `role` VARCHAR(32) NOT NULL COMMENT '组织角色(owner/admin/member)',
  `username` VARCHAR(64) DEFAULT NULL COMMENT '用户名(冗余)',
  `email` VARCHAR(128) DEFAULT NULL COMMENT '邮箱(冗余)',
  `invited_by` VARCHAR(64) DEFAULT NULL COMMENT '邀请人用户ID',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态: 0-待接受, 1-正常, 2-禁用',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_org_user` (`org_id`, `user_id`),
  KEY `idx_org_id` (`org_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_role` (`role`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='组织成员表';

-- 团队表
CREATE TABLE IF NOT EXISTS `t_team` (
  `id` INT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `team_id` VARCHAR(64) NOT NULL COMMENT '团队唯一标识',
  `org_id` VARCHAR(64) NOT NULL COMMENT '所属组织ID',
  `name` VARCHAR(128) NOT NULL COMMENT '团队名称(英文标识)',
  `display_name` VARCHAR(255) NOT NULL COMMENT '团队显示名称',
  `description` TEXT DEFAULT NULL COMMENT '团队描述',
  `avatar` VARCHAR(512) DEFAULT NULL COMMENT '团队头像',
  `parent_team_id` VARCHAR(64) DEFAULT NULL COMMENT '父团队ID(支持嵌套)',
  `path` VARCHAR(512) DEFAULT NULL COMMENT '团队路径(用于层级关系,如:/parent/child)',
  `level` INT NOT NULL DEFAULT 1 COMMENT '团队层级(1为顶层)',
  `settings` JSON DEFAULT NULL COMMENT '团队设置(JSON格式)',
  `visibility` TINYINT NOT NULL DEFAULT 0 COMMENT '可见性: 0-私有, 1-内部, 2-公开',
  `is_enabled` TINYINT NOT NULL DEFAULT 1 COMMENT '是否启用: 0-禁用, 1-启用',
  `total_members` INT NOT NULL DEFAULT 0 COMMENT '成员总数',
  `total_projects` INT NOT NULL DEFAULT 0 COMMENT '项目总数',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_team_id` (`team_id`),
  UNIQUE KEY `uk_org_name` (`org_id`, `name`),
  KEY `idx_org_id` (`org_id`),
  KEY `idx_parent_team_id` (`parent_team_id`),
  KEY `idx_visibility` (`visibility`),
  KEY `idx_path` (`path`(191))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='团队表';

-- 团队成员表
CREATE TABLE IF NOT EXISTS `t_team_member` (
  `id` INT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `team_id` VARCHAR(64) NOT NULL COMMENT '团队ID',
  `user_id` VARCHAR(64) NOT NULL COMMENT '用户ID',
  `role` VARCHAR(32) NOT NULL COMMENT '团队角色(owner/maintainer/developer/reporter/guest)',
  `username` VARCHAR(64) DEFAULT NULL COMMENT '用户名(冗余)',
  `invited_by` VARCHAR(64) DEFAULT NULL COMMENT '邀请人用户ID',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_team_user` (`team_id`, `user_id`),
  KEY `idx_team_id` (`team_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_role` (`role`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='团队成员表';

-- 项目团队关联表
CREATE TABLE IF NOT EXISTS `t_project_team_relation` (
  `id` INT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `project_id` VARCHAR(64) NOT NULL COMMENT '项目ID',
  `team_id` VARCHAR(64) NOT NULL COMMENT '团队ID',
  `access` VARCHAR(32) NOT NULL DEFAULT 'read' COMMENT '访问权限(read/write/admin)',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_project_team` (`project_id`, `team_id`),
  KEY `idx_project_id` (`project_id`),
  KEY `idx_team_id` (`team_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='项目团队关联表';

-- 组织邀请表
CREATE TABLE IF NOT EXISTS `t_organization_invitation` (
  `id` INT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `invitation_id` VARCHAR(64) NOT NULL COMMENT '邀请唯一标识',
  `org_id` VARCHAR(64) NOT NULL COMMENT '组织ID',
  `email` VARCHAR(128) NOT NULL COMMENT '被邀请人邮箱',
  `role` VARCHAR(32) NOT NULL DEFAULT 'member' COMMENT '角色(owner/admin/member)',
  `token` VARCHAR(255) NOT NULL COMMENT '邀请令牌',
  `invited_by` VARCHAR(64) NOT NULL COMMENT '邀请人用户ID',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态: 0-待接受, 1-已接受, 2-已拒绝, 3-已过期',
  `expires_at` DATETIME NOT NULL COMMENT '过期时间',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_invitation_id` (`invitation_id`),
  UNIQUE KEY `uk_token` (`token`),
  KEY `idx_org_id` (`org_id`),
  KEY `idx_email` (`email`),
  KEY `idx_status` (`status`),
  KEY `idx_expires_at` (`expires_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='组织邀请表';

-- 修改项目表，添加组织关联
ALTER TABLE `t_project` 
  ADD COLUMN `org_id` VARCHAR(64) NOT NULL COMMENT '所属组织ID' AFTER `project_id`,
  ADD COLUMN `namespace` VARCHAR(255) NOT NULL COMMENT '项目命名空间(org_name/project_name)' AFTER `display_name`,
  ADD COLUMN `access_level` VARCHAR(32) NOT NULL DEFAULT 'team' COMMENT '默认访问级别(owner/team/org)' AFTER `visibility`,
  DROP COLUMN `group_id`,
  ADD KEY `idx_org_id` (`org_id`),
  ADD UNIQUE KEY `uk_namespace` (`namespace`);

-- 修改项目成员表，添加来源字段
ALTER TABLE `t_project_member`
  ADD COLUMN `source` VARCHAR(32) NOT NULL DEFAULT 'direct' COMMENT '来源(direct/team/org)' AFTER `username`;

-- ================================================================
-- 示例数据
-- ================================================================

-- 插入示例组织
INSERT INTO `t_organization` (
  `org_id`,
  `name`,
  `display_name`,
  `description`,
  `logo`,
  `website`,
  `email`,
  `settings`,
  `plan`,
  `status`,
  `owner_user_id`
) VALUES (
  'org_default',
  'default-org',
  '默认组织',
  '这是系统默认组织',
  NULL,
  'https://arcade.example.com',
  'admin@example.com',
  JSON_OBJECT(
    'allow_public_projects', true,
    'require_two_factor', false,
    'allowed_domains', JSON_ARRAY('example.com'),
    'max_members', 100,
    'max_projects', 50,
    'max_teams', 20,
    'max_agents', 10,
    'default_visibility', 'private',
    'enable_saml', false,
    'enable_ldap', false
  ),
  'free',
  1,
  'user_admin'
) ON DUPLICATE KEY UPDATE `name` = `name`;

-- 插入组织成员
INSERT INTO `t_organization_member` (`org_id`, `user_id`, `role`, `username`, `email`, `status`)
VALUES 
  ('org_default', 'user_admin', 'owner', 'admin', 'admin@example.com', 1)
ON DUPLICATE KEY UPDATE `role` = `role`;

-- 插入示例团队
INSERT INTO `t_team` (
  `team_id`,
  `org_id`,
  `name`,
  `display_name`,
  `description`,
  `parent_team_id`,
  `path`,
  `level`,
  `settings`,
  `visibility`
) VALUES (
  'team_dev',
  'org_default',
  'development',
  '开发团队',
  '负责产品研发的核心团队',
  NULL,
  '/development',
  1,
  JSON_OBJECT(
    'default_role', 'developer',
    'allow_member_invite', false,
    'require_approval', true,
    'max_members', 50
  ),
  1
),
(
  'team_ops',
  'org_default',
  'operations',
  '运维团队',
  '负责系统运维和基础设施管理',
  NULL,
  '/operations',
  1,
  JSON_OBJECT(
    'default_role', 'developer',
    'allow_member_invite', false,
    'require_approval', true,
    'max_members', 30
  ),
  1
),
(
  'team_frontend',
  'org_default',
  'frontend',
  '前端团队',
  '开发团队下的前端小组',
  'team_dev',
  '/development/frontend',
  2,
  JSON_OBJECT(
    'default_role', 'developer',
    'allow_member_invite', true,
    'require_approval', false,
    'max_members', 20
  ),
  1
)
ON DUPLICATE KEY UPDATE `name` = `name`;

-- 插入团队成员
INSERT INTO `t_team_member` (`team_id`, `user_id`, `role`, `username`)
VALUES 
  ('team_dev', 'user_admin', 'owner', 'admin'),
  ('team_ops', 'user_admin', 'maintainer', 'admin'),
  ('team_frontend', 'user_admin', 'owner', 'admin')
ON DUPLICATE KEY UPDATE `role` = `role`;

-- 更新已有项目，关联到默认组织
UPDATE `t_project` 
SET 
  `org_id` = 'org_default',
  `namespace` = CONCAT('default-org/', `name`),
  `access_level` = 'team'
WHERE `org_id` IS NULL OR `org_id` = '';

-- 插入项目团队关联
INSERT INTO `t_project_team_relation` (`project_id`, `team_id`, `access`)
SELECT 'proj_demo', 'team_dev', 'admin'
WHERE EXISTS (SELECT 1 FROM `t_project` WHERE `project_id` = 'proj_demo')
  AND EXISTS (SELECT 1 FROM `t_team` WHERE `team_id` = 'team_dev')
ON DUPLICATE KEY UPDATE `access` = `access`;



