-- ============================================
-- 权限系统数据库迁移脚本
-- 用途：权限点独立管理，通过关联表实现角色-权限关系
-- 作者: gagral.x@gmail.com
-- 日期: 2025/01/14
-- ============================================

-- ============================================
-- 1. 权限点表
-- ============================================
CREATE TABLE IF NOT EXISTS `t_permission` (
    `id` BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
    `permission_id` VARCHAR(64) NOT NULL UNIQUE COMMENT '权限点ID（唯一标识）',
    `code` VARCHAR(64) NOT NULL UNIQUE COMMENT '权限点代码（如：project.view）',
    `name` VARCHAR(128) NOT NULL COMMENT '权限点名称',
    `category` VARCHAR(32) NOT NULL COMMENT '权限分类（project/build/pipeline/deploy/member/team/monitor/security）',
    `description` VARCHAR(512) DEFAULT NULL COMMENT '描述信息',
    `is_enabled` TINYINT DEFAULT 1 COMMENT '是否启用 0:禁用 1:启用',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX `idx_code` (`code`) COMMENT '权限代码索引',
    INDEX `idx_category` (`category`) COMMENT '分类索引',
    INDEX `idx_is_enabled` (`is_enabled`) COMMENT '启用状态索引'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='权限点表';

-- ============================================
-- 2. 角色权限关联表
-- ============================================
CREATE TABLE IF NOT EXISTS `t_role_permission` (
    `id` BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
    `role_id` VARCHAR(64) NOT NULL COMMENT '角色ID',
    `permission_id` VARCHAR(64) NOT NULL COMMENT '权限点ID',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    UNIQUE KEY `uk_role_permission` (`role_id`, `permission_id`) COMMENT '角色-权限唯一索引',
    INDEX `idx_role_id` (`role_id`) COMMENT '角色ID索引',
    INDEX `idx_permission_id` (`permission_id`) COMMENT '权限ID索引'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色权限关联表';

-- ============================================
-- 插入权限点数据
-- ============================================

-- 项目权限
INSERT INTO `t_permission` (`permission_id`, `code`, `name`, `category`, `description`, `is_enabled`) VALUES
(UPPER(REPLACE(UUID(), '-', '')), 'project.view', '查看项目', 'project', '查看项目基本信息和列表', 1),
(UPPER(REPLACE(UUID(), '-', '')), 'project.edit', '编辑项目', 'project', '编辑项目基本信息', 1),
(UPPER(REPLACE(UUID(), '-', '')), 'project.delete', '删除项目', 'project', '删除项目（高危操作）', 1),
(UPPER(REPLACE(UUID(), '-', '')), 'project.transfer', '转移项目', 'project', '转移项目所有权', 1),
(UPPER(REPLACE(UUID(), '-', '')), 'project.archive', '归档项目', 'project', '归档或恢复项目', 1),
(UPPER(REPLACE(UUID(), '-', '')), 'project.settings', '项目设置', 'project', '修改项目高级设置', 1),
(UPPER(REPLACE(UUID(), '-', '')), 'project.variables', '管理变量', 'project', '管理项目环境变量', 1),
(UPPER(REPLACE(UUID(), '-', '')), 'project.webhook', '管理Webhook', 'project', '配置项目Webhook', 1),
(UPPER(REPLACE(UUID(), '-', '')), 'project.integration', '管理集成', 'project', '管理第三方集成', 1);

-- 构建权限
INSERT INTO `t_permission` (`permission_id`, `code`, `name`, `category`, `description`, `is_enabled`) VALUES
(UPPER(REPLACE(UUID(), '-', '')), 'build.view', '查看构建', 'build', '查看构建记录和状态', 1),
(UPPER(REPLACE(UUID(), '-', '')), 'build.trigger', '触发构建', 'build', '手动触发构建', 1),
(UPPER(REPLACE(UUID(), '-', '')), 'build.cancel', '取消构建', 'build', '取消正在运行的构建', 1),
(UPPER(REPLACE(UUID(), '-', '')), 'build.retry', '重试构建', 'build', '重试失败的构建', 1),
(UPPER(REPLACE(UUID(), '-', '')), 'build.delete', '删除构建', 'build', '删除构建记录', 1),
(UPPER(REPLACE(UUID(), '-', '')), 'build.artifact', '下载产物', 'build', '下载构建产物', 1),
(UPPER(REPLACE(UUID(), '-', '')), 'build.log', '查看日志', 'build', '查看构建日志', 1);

-- 流水线权限
INSERT INTO `t_permission` (`permission_id`, `code`, `name`, `category`, `description`, `is_enabled`) VALUES
(UPPER(REPLACE(UUID(), '-', '')), 'pipeline.view', '查看流水线', 'pipeline', '查看流水线配置', 1),
(UPPER(REPLACE(UUID(), '-', '')), 'pipeline.create', '创建流水线', 'pipeline', '创建新流水线', 1),
(UPPER(REPLACE(UUID(), '-', '')), 'pipeline.edit', '编辑流水线', 'pipeline', '编辑流水线配置', 1),
(UPPER(REPLACE(UUID(), '-', '')), 'pipeline.delete', '删除流水线', 'pipeline', '删除流水线', 1),
(UPPER(REPLACE(UUID(), '-', '')), 'pipeline.run', '运行流水线', 'pipeline', '触发流水线运行', 1),
(UPPER(REPLACE(UUID(), '-', '')), 'pipeline.stop', '停止流水线', 'pipeline', '停止运行中的流水线', 1);

-- 部署权限
INSERT INTO `t_permission` (`permission_id`, `code`, `name`, `category`, `description`, `is_enabled`) VALUES
(UPPER(REPLACE(UUID(), '-', '')), 'deploy.view', '查看部署', 'deploy', '查看部署记录', 1),
(UPPER(REPLACE(UUID(), '-', '')), 'deploy.create', '创建部署', 'deploy', '创建部署计划', 1),
(UPPER(REPLACE(UUID(), '-', '')), 'deploy.execute', '执行部署', 'deploy', '执行部署操作', 1),
(UPPER(REPLACE(UUID(), '-', '')), 'deploy.rollback', '回滚部署', 'deploy', '回滚到上一版本', 1),
(UPPER(REPLACE(UUID(), '-', '')), 'deploy.approve', '审批部署', 'deploy', '审批部署申请', 1);

-- 成员权限
INSERT INTO `t_permission` (`permission_id`, `code`, `name`, `category`, `description`, `is_enabled`) VALUES
(UPPER(REPLACE(UUID(), '-', '')), 'member.view', '查看成员', 'member', '查看项目成员列表', 1),
(UPPER(REPLACE(UUID(), '-', '')), 'member.invite', '邀请成员', 'member', '邀请新成员加入', 1),
(UPPER(REPLACE(UUID(), '-', '')), 'member.edit', '编辑成员', 'member', '修改成员角色', 1),
(UPPER(REPLACE(UUID(), '-', '')), 'member.remove', '移除成员', 'member', '移除项目成员', 1);

-- 团队权限
INSERT INTO `t_permission` (`permission_id`, `code`, `name`, `category`, `description`, `is_enabled`) VALUES
(UPPER(REPLACE(UUID(), '-', '')), 'team.view', '查看团队', 'team', '查看团队信息', 1),
(UPPER(REPLACE(UUID(), '-', '')), 'team.create', '创建团队', 'team', '创建新团队', 1),
(UPPER(REPLACE(UUID(), '-', '')), 'team.edit', '编辑团队', 'team', '编辑团队信息', 1),
(UPPER(REPLACE(UUID(), '-', '')), 'team.delete', '删除团队', 'team', '删除团队', 1),
(UPPER(REPLACE(UUID(), '-', '')), 'team.member', '管理团队成员', 'team', '添加/移除团队成员', 1),
(UPPER(REPLACE(UUID(), '-', '')), 'team.project', '管理团队项目', 'team', '管理团队可访问的项目', 1),
(UPPER(REPLACE(UUID(), '-', '')), 'team.settings', '团队设置', 'team', '修改团队设置', 1);

-- 安全权限
INSERT INTO `t_permission` (`permission_id`, `code`, `name`, `category`, `description`, `is_enabled`) VALUES
(UPPER(REPLACE(UUID(), '-', '')), 'security.scan', '安全扫描', 'security', '执行安全扫描', 1),
(UPPER(REPLACE(UUID(), '-', '')), 'security.audit', '安全审计', 'security', '查看安全审计日志', 1),
(UPPER(REPLACE(UUID(), '-', '')), 'security.policy', '安全策略', 'security', '管理安全策略', 1);

-- ============================================
-- 为内置角色分配权限
-- ============================================

-- 获取权限ID的辅助变量（使用时需要替换）
SET @perm_project_view = (SELECT permission_id FROM t_permission WHERE code = 'project.view');
SET @perm_project_edit = (SELECT permission_id FROM t_permission WHERE code = 'project.edit');
SET @perm_project_delete = (SELECT permission_id FROM t_permission WHERE code = 'project.delete');
SET @perm_project_transfer = (SELECT permission_id FROM t_permission WHERE code = 'project.transfer');
SET @perm_project_archive = (SELECT permission_id FROM t_permission WHERE code = 'project.archive');
SET @perm_project_settings = (SELECT permission_id FROM t_permission WHERE code = 'project.settings');
SET @perm_project_variables = (SELECT permission_id FROM t_permission WHERE code = 'project.variables');
SET @perm_project_webhook = (SELECT permission_id FROM t_permission WHERE code = 'project.webhook');
SET @perm_project_integration = (SELECT permission_id FROM t_permission WHERE code = 'project.integration');

SET @perm_build_view = (SELECT permission_id FROM t_permission WHERE code = 'build.view');
SET @perm_build_trigger = (SELECT permission_id FROM t_permission WHERE code = 'build.trigger');
SET @perm_build_cancel = (SELECT permission_id FROM t_permission WHERE code = 'build.cancel');
SET @perm_build_retry = (SELECT permission_id FROM t_permission WHERE code = 'build.retry');
SET @perm_build_delete = (SELECT permission_id FROM t_permission WHERE code = 'build.delete');
SET @perm_build_artifact = (SELECT permission_id FROM t_permission WHERE code = 'build.artifact');
SET @perm_build_log = (SELECT permission_id FROM t_permission WHERE code = 'build.log');

SET @perm_pipeline_view = (SELECT permission_id FROM t_permission WHERE code = 'pipeline.view');
SET @perm_pipeline_create = (SELECT permission_id FROM t_permission WHERE code = 'pipeline.create');
SET @perm_pipeline_edit = (SELECT permission_id FROM t_permission WHERE code = 'pipeline.edit');
SET @perm_pipeline_delete = (SELECT permission_id FROM t_permission WHERE code = 'pipeline.delete');
SET @perm_pipeline_run = (SELECT permission_id FROM t_permission WHERE code = 'pipeline.run');
SET @perm_pipeline_stop = (SELECT permission_id FROM t_permission WHERE code = 'pipeline.stop');

SET @perm_deploy_view = (SELECT permission_id FROM t_permission WHERE code = 'deploy.view');
SET @perm_deploy_create = (SELECT permission_id FROM t_permission WHERE code = 'deploy.create');
SET @perm_deploy_execute = (SELECT permission_id FROM t_permission WHERE code = 'deploy.execute');
SET @perm_deploy_rollback = (SELECT permission_id FROM t_permission WHERE code = 'deploy.rollback');
SET @perm_deploy_approve = (SELECT permission_id FROM t_permission WHERE code = 'deploy.approve');

SET @perm_member_view = (SELECT permission_id FROM t_permission WHERE code = 'member.view');
SET @perm_member_invite = (SELECT permission_id FROM t_permission WHERE code = 'member.invite');
SET @perm_member_edit = (SELECT permission_id FROM t_permission WHERE code = 'member.edit');
SET @perm_member_remove = (SELECT permission_id FROM t_permission WHERE code = 'member.remove');

SET @perm_team_view = (SELECT permission_id FROM t_permission WHERE code = 'team.view');
SET @perm_team_create = (SELECT permission_id FROM t_permission WHERE code = 'team.create');
SET @perm_team_edit = (SELECT permission_id FROM t_permission WHERE code = 'team.edit');
SET @perm_team_delete = (SELECT permission_id FROM t_permission WHERE code = 'team.delete');
SET @perm_team_member = (SELECT permission_id FROM t_permission WHERE code = 'team.member');
SET @perm_team_project = (SELECT permission_id FROM t_permission WHERE code = 'team.project');
SET @perm_team_settings = (SELECT permission_id FROM t_permission WHERE code = 'team.settings');

SET @perm_monitor_view = (SELECT permission_id FROM t_permission WHERE code = 'monitor.view');
SET @perm_monitor_metrics = (SELECT permission_id FROM t_permission WHERE code = 'monitor.metrics');
SET @perm_monitor_logs = (SELECT permission_id FROM t_permission WHERE code = 'monitor.logs');
SET @perm_monitor_alert = (SELECT permission_id FROM t_permission WHERE code = 'monitor.alert');
SET @perm_monitor_dashboard = (SELECT permission_id FROM t_permission WHERE code = 'monitor.dashboard');

SET @perm_security_scan = (SELECT permission_id FROM t_permission WHERE code = 'security.scan');
SET @perm_security_audit = (SELECT permission_id FROM t_permission WHERE code = 'security.audit');
SET @perm_security_policy = (SELECT permission_id FROM t_permission WHERE code = 'security.policy');

-- 项目Owner角色 - 所有权限
INSERT INTO `t_role_permission` (`role_id`, `permission_id`) 
SELECT 'project_owner', permission_id FROM t_permission WHERE is_enabled = 1;

-- 项目Maintainer角色
INSERT INTO `t_role_permission` (`role_id`, `permission_id`) VALUES
('project_maintainer', @perm_project_view),
('project_maintainer', @perm_project_edit),
('project_maintainer', @perm_project_archive),
('project_maintainer', @perm_project_settings),
('project_maintainer', @perm_project_variables),
('project_maintainer', @perm_project_webhook),
('project_maintainer', @perm_project_integration),
('project_maintainer', @perm_build_view),
('project_maintainer', @perm_build_trigger),
('project_maintainer', @perm_build_cancel),
('project_maintainer', @perm_build_retry),
('project_maintainer', @perm_build_delete),
('project_maintainer', @perm_build_artifact),
('project_maintainer', @perm_build_log),
('project_maintainer', @perm_pipeline_view),
('project_maintainer', @perm_pipeline_create),
('project_maintainer', @perm_pipeline_edit),
('project_maintainer', @perm_pipeline_delete),
('project_maintainer', @perm_pipeline_run),
('project_maintainer', @perm_pipeline_stop),
('project_maintainer', @perm_deploy_view),
('project_maintainer', @perm_deploy_create),
('project_maintainer', @perm_deploy_execute),
('project_maintainer', @perm_deploy_rollback),
('project_maintainer', @perm_deploy_approve),
('project_maintainer', @perm_member_view),
('project_maintainer', @perm_member_invite),
('project_maintainer', @perm_member_edit),
('project_maintainer', @perm_member_remove),
('project_maintainer', @perm_team_view),
('project_maintainer', @perm_team_project),
('project_maintainer', @perm_team_settings),
('project_maintainer', @perm_monitor_view),
('project_maintainer', @perm_monitor_metrics),
('project_maintainer', @perm_monitor_logs),
('project_maintainer', @perm_monitor_alert),
('project_maintainer', @perm_monitor_dashboard),
('project_maintainer', @perm_security_scan),
('project_maintainer', @perm_security_audit);

-- 项目Developer角色
INSERT INTO `t_role_permission` (`role_id`, `permission_id`) VALUES
('project_developer', @perm_project_view),
('project_developer', @perm_build_view),
('project_developer', @perm_build_trigger),
('project_developer', @perm_build_cancel),
('project_developer', @perm_build_retry),
('project_developer', @perm_build_artifact),
('project_developer', @perm_build_log),
('project_developer', @perm_pipeline_view),
('project_developer', @perm_pipeline_create),
('project_developer', @perm_pipeline_edit),
('project_developer', @perm_pipeline_run),
('project_developer', @perm_pipeline_stop),
('project_developer', @perm_deploy_view),
('project_developer', @perm_deploy_create),
('project_developer', @perm_deploy_execute),
('project_developer', @perm_member_view),
('project_developer', @perm_team_view),
('project_developer', @perm_monitor_view),
('project_developer', @perm_monitor_metrics),
('project_developer', @perm_monitor_logs),
('project_developer', @perm_security_scan);

-- 项目Reporter角色
INSERT INTO `t_role_permission` (`role_id`, `permission_id`) VALUES
('project_reporter', @perm_project_view),
('project_reporter', @perm_build_view),
('project_reporter', @perm_build_artifact),
('project_reporter', @perm_build_log),
('project_reporter', @perm_pipeline_view),
('project_reporter', @perm_deploy_view),
('project_reporter', @perm_member_view),
('project_reporter', @perm_team_view),
('project_reporter', @perm_monitor_view),
('project_reporter', @perm_monitor_metrics),
('project_reporter', @perm_monitor_logs);

-- 项目Guest角色
INSERT INTO `t_role_permission` (`role_id`, `permission_id`) VALUES
('project_guest', @perm_project_view),
('project_guest', @perm_build_view),
('project_guest', @perm_build_log),
('project_guest', @perm_pipeline_view),
('project_guest', @perm_deploy_view),
('project_guest', @perm_member_view),
('project_guest', @perm_monitor_view);

-- 团队Owner角色
INSERT INTO `t_role_permission` (`role_id`, `permission_id`) VALUES
('team_owner', @perm_team_view),
('team_owner', @perm_team_edit),
('team_owner', @perm_team_delete),
('team_owner', @perm_team_member),
('team_owner', @perm_team_project),
('team_owner', @perm_team_settings);

-- 团队Maintainer角色
INSERT INTO `t_role_permission` (`role_id`, `permission_id`) VALUES
('team_maintainer', @perm_team_view),
('team_maintainer', @perm_team_edit),
('team_maintainer', @perm_team_member),
('team_maintainer', @perm_team_project),
('team_maintainer', @perm_team_settings);

-- 团队Developer角色
INSERT INTO `t_role_permission` (`role_id`, `permission_id`) VALUES
('team_developer', @perm_team_view),
('team_developer', @perm_team_project);

-- 团队Reporter和Guest角色
INSERT INTO `t_role_permission` (`role_id`, `permission_id`) VALUES
('team_reporter', @perm_team_view),
('team_guest', @perm_team_view);

-- 组织Owner角色
INSERT INTO `t_role_permission` (`role_id`, `permission_id`) VALUES
('org_owner', @perm_team_view),
('org_owner', @perm_team_create),
('org_owner', @perm_team_edit),
('org_owner', @perm_team_delete),
('org_owner', @perm_team_member),
('org_owner', @perm_team_project),
('org_owner', @perm_team_settings);

-- 组织Admin角色
INSERT INTO `t_role_permission` (`role_id`, `permission_id`) VALUES
('org_admin', @perm_team_view),
('org_admin', @perm_team_create),
('org_admin', @perm_team_edit),
('org_admin', @perm_team_member),
('org_admin', @perm_team_project),
('org_admin', @perm_team_settings);

-- 组织Member角色
INSERT INTO `t_role_permission` (`role_id`, `permission_id`) VALUES
('org_member', @perm_team_view);

-- ============================================
-- 数据验证
-- ============================================

-- 查看权限点统计
SELECT 
    category AS '权限分类',
    COUNT(*) AS '权限数量'
FROM t_permission
WHERE is_enabled = 1
GROUP BY category
ORDER BY category;

-- 查看角色权限统计
SELECT 
    r.name AS '角色名称',
    r.scope AS '作用域',
    COUNT(rp.permission_id) AS '权限数量'
FROM t_role r
LEFT JOIN t_role_permission rp ON r.role_id = rp.role_id
WHERE r.is_enabled = 1
GROUP BY r.role_id, r.name, r.scope
ORDER BY r.scope, r.priority DESC;

-- 查看所有权限点
SELECT 
    category AS '分类',
    code AS '权限代码',
    name AS '权限名称',
    description AS '描述'
FROM t_permission
WHERE is_enabled = 1
ORDER BY category, code;

-- ============================================
-- 完成提示
-- ============================================
SELECT '权限系统初始化完成！' AS '状态',
       (SELECT COUNT(*) FROM t_permission WHERE is_enabled = 1) AS '权限点总数',
       (SELECT COUNT(DISTINCT role_id) FROM t_role_permission) AS '已配置角色数',
       (SELECT COUNT(*) FROM t_role_permission) AS '角色权限关联总数';

