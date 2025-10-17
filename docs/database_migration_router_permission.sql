-- ============================================
-- 路由权限映射表迁移脚本
-- 用途：支持动态路由权限管理
-- 作者: gagral.x@gmail.com
-- 日期: 2025/01/14
-- ============================================

-- 创建路由权限映射表
CREATE TABLE IF NOT EXISTS `t_router_permission` (
    `id` BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
    `route_id` VARCHAR(128) NOT NULL UNIQUE COMMENT '路由ID（唯一标识）',
    `path` VARCHAR(255) NOT NULL COMMENT '路由路径（如：/projects）',
    `method` VARCHAR(16) NOT NULL COMMENT 'HTTP方法（GET/POST/PUT/DELETE等）',
    `name` VARCHAR(128) NOT NULL COMMENT '路由名称（用于显示）',
    `group` VARCHAR(64) DEFAULT NULL COMMENT '路由分组（如：project/org/team）',
    `category` VARCHAR(64) DEFAULT NULL COMMENT '路由分类（如：项目管理/CI/CD/监控）',
    `required_permissions` JSON DEFAULT NULL COMMENT '所需权限列表（JSON数组）',
    `icon` VARCHAR(64) DEFAULT NULL COMMENT '图标名称',
    `order` INT DEFAULT 0 COMMENT '排序顺序（数字越小越靠前）',
    `is_menu` TINYINT DEFAULT 0 COMMENT '是否显示在菜单 0:否 1:是',
    `is_enabled` TINYINT DEFAULT 1 COMMENT '是否启用 0:禁用 1:启用',
    `description` VARCHAR(512) DEFAULT NULL COMMENT '描述信息',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX `idx_path_method` (`path`, `method`) COMMENT '路径和方法索引',
    INDEX `idx_group` (`group`) COMMENT '分组索引',
    INDEX `idx_category` (`category`) COMMENT '分类索引',
    INDEX `idx_is_menu` (`is_menu`) COMMENT '菜单标志索引',
    INDEX `idx_is_enabled` (`is_enabled`) COMMENT '启用状态索引',
    INDEX `idx_order` (`order`) COMMENT '排序索引'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='路由权限映射表';

-- ============================================
-- 插入内置路由配置
-- ============================================

-- 项目管理路由
INSERT INTO `t_router_permission` (`route_id`, `path`, `method`, `name`, `group`, `category`, `required_permissions`, `icon`, `order`, `is_menu`, `is_enabled`, `description`) VALUES
(UPPER(REPLACE(UUID(), '-', '')), '/projects', 'GET', '项目列表', 'project', '项目管理', '["project.view"]', 'project', 100, 1, 1, '查看项目列表'),
(UPPER(REPLACE(UUID(), '-', '')), '/projects/:projectId', 'GET', '项目详情', 'project', '项目管理', '["project.view"]', '', 101, 0, 1, '查看项目详情'),
(UPPER(REPLACE(UUID(), '-', '')), '/projects', 'POST', '创建项目', 'project', '项目管理', '[]', '', 102, 0, 1, '创建新项目'),
(UPPER(REPLACE(UUID(), '-', '')), '/projects/:projectId', 'PUT', '更新项目', 'project', '项目管理', '["project.edit"]', '', 103, 0, 1, '更新项目设置'),
(UPPER(REPLACE(UUID(), '-', '')), '/projects/:projectId', 'DELETE', '删除项目', 'project', '项目管理', '["project.delete"]', '', 104, 0, 1, '删除项目'),
(UPPER(REPLACE(UUID(), '-', '')), '/projects/:projectId/settings', 'GET', '项目设置', 'project', '项目管理', '["project.settings"]', 'settings', 105, 1, 1, '查看和修改项目设置');

-- 流水线路由
INSERT INTO `t_router_permission` (`route_id`, `path`, `method`, `name`, `group`, `category`, `required_permissions`, `icon`, `order`, `is_menu`, `is_enabled`, `description`) VALUES
(UPPER(REPLACE(UUID(), '-', '')), '/projects/:projectId/pipelines', 'GET', '流水线列表', 'pipeline', 'CI/CD', '["pipeline.view"]', 'pipeline', 200, 1, 1, '查看流水线列表'),
(UPPER(REPLACE(UUID(), '-', '')), '/projects/:projectId/pipelines', 'POST', '创建流水线', 'pipeline', 'CI/CD', '["pipeline.create"]', '', 201, 0, 1, '创建新流水线'),
(UPPER(REPLACE(UUID(), '-', '')), '/projects/:projectId/pipelines/:pipelineId', 'GET', '流水线详情', 'pipeline', 'CI/CD', '["pipeline.view"]', '', 202, 0, 1, '查看流水线详情'),
(UPPER(REPLACE(UUID(), '-', '')), '/projects/:projectId/pipelines/:pipelineId', 'PUT', '更新流水线', 'pipeline', 'CI/CD', '["pipeline.edit"]', '', 203, 0, 1, '更新流水线配置'),
(UPPER(REPLACE(UUID(), '-', '')), '/projects/:projectId/pipelines/:pipelineId', 'DELETE', '删除流水线', 'pipeline', 'CI/CD', '["pipeline.delete"]', '', 204, 0, 1, '删除流水线'),
(UPPER(REPLACE(UUID(), '-', '')), '/projects/:projectId/pipelines/:pipelineId/run', 'POST', '运行流水线', 'pipeline', 'CI/CD', '["pipeline.run"]', '', 205, 0, 1, '触发流水线运行');

-- 构建路由
INSERT INTO `t_router_permission` (`route_id`, `path`, `method`, `name`, `group`, `category`, `required_permissions`, `icon`, `order`, `is_menu`, `is_enabled`, `description`) VALUES
(UPPER(REPLACE(UUID(), '-', '')), '/projects/:projectId/builds', 'GET', '构建列表', 'build', 'CI/CD', '["build.view"]', 'build', 300, 1, 1, '查看构建列表'),
(UPPER(REPLACE(UUID(), '-', '')), '/projects/:projectId/builds/:buildId', 'GET', '构建详情', 'build', 'CI/CD', '["build.view"]', '', 301, 0, 1, '查看构建详情'),
(UPPER(REPLACE(UUID(), '-', '')), '/projects/:projectId/builds/trigger', 'POST', '触发构建', 'build', 'CI/CD', '["build.trigger"]', '', 302, 0, 1, '手动触发构建'),
(UPPER(REPLACE(UUID(), '-', '')), '/projects/:projectId/builds/:buildId/cancel', 'POST', '取消构建', 'build', 'CI/CD', '["build.cancel"]', '', 303, 0, 1, '取消正在运行的构建'),
(UPPER(REPLACE(UUID(), '-', '')), '/projects/:projectId/builds/:buildId/retry', 'POST', '重试构建', 'build', 'CI/CD', '["build.retry"]', '', 304, 0, 1, '重试失败的构建'),
(UPPER(REPLACE(UUID(), '-', '')), '/projects/:projectId/builds/:buildId/logs', 'GET', '构建日志', 'build', 'CI/CD', '["build.log"]', '', 305, 0, 1, '查看构建日志'),
(UPPER(REPLACE(UUID(), '-', '')), '/projects/:projectId/builds/:buildId/artifacts', 'GET', '构建产物', 'build', 'CI/CD', '["build.artifact"]', '', 306, 0, 1, '下载构建产物');

-- 部署路由
INSERT INTO `t_router_permission` (`route_id`, `path`, `method`, `name`, `group`, `category`, `required_permissions`, `icon`, `order`, `is_menu`, `is_enabled`, `description`) VALUES
(UPPER(REPLACE(UUID(), '-', '')), '/projects/:projectId/deploys', 'GET', '部署列表', 'deploy', '部署', '["deploy.view"]', 'deploy', 400, 1, 1, '查看部署列表'),
(UPPER(REPLACE(UUID(), '-', '')), '/projects/:projectId/deploys/:deployId', 'GET', '部署详情', 'deploy', '部署', '["deploy.view"]', '', 401, 0, 1, '查看部署详情'),
(UPPER(REPLACE(UUID(), '-', '')), '/projects/:projectId/deploys', 'POST', '创建部署', 'deploy', '部署', '["deploy.create"]', '', 402, 0, 1, '创建部署计划'),
(UPPER(REPLACE(UUID(), '-', '')), '/projects/:projectId/deploys/:deployId/execute', 'POST', '执行部署', 'deploy', '部署', '["deploy.execute"]', '', 403, 0, 1, '执行部署'),
(UPPER(REPLACE(UUID(), '-', '')), '/projects/:projectId/deploys/:deployId/rollback', 'POST', '回滚部署', 'deploy', '部署', '["deploy.rollback"]', '', 404, 0, 1, '回滚到上一版本'),
(UPPER(REPLACE(UUID(), '-', '')), '/projects/:projectId/deploys/:deployId/approve', 'POST', '审批部署', 'deploy', '部署', '["deploy.approve"]', '', 405, 0, 1, '审批部署申请');

-- 成员管理路由
INSERT INTO `t_router_permission` (`route_id`, `path`, `method`, `name`, `group`, `category`, `required_permissions`, `icon`, `order`, `is_menu`, `is_enabled`, `description`) VALUES
(UPPER(REPLACE(UUID(), '-', '')), '/projects/:projectId/members', 'GET', '成员列表', 'member', '成员管理', '["member.view"]', 'team', 500, 1, 1, '查看成员列表'),
(UPPER(REPLACE(UUID(), '-', '')), '/projects/:projectId/members', 'POST', '邀请成员', 'member', '成员管理', '["member.invite"]', '', 501, 0, 1, '邀请新成员'),
(UPPER(REPLACE(UUID(), '-', '')), '/projects/:projectId/members/:userId', 'PUT', '更新成员角色', 'member', '成员管理', '["member.edit"]', '', 502, 0, 1, '修改成员角色'),
(UPPER(REPLACE(UUID(), '-', '')), '/projects/:projectId/members/:userId', 'DELETE', '移除成员', 'member', '成员管理', '["member.remove"]', '', 503, 0, 1, '移除项目成员');

-- 监控路由
INSERT INTO `t_router_permission` (`route_id`, `path`, `method`, `name`, `group`, `category`, `required_permissions`, `icon`, `order`, `is_menu`, `is_enabled`, `description`) VALUES
(UPPER(REPLACE(UUID(), '-', '')), '/projects/:projectId/monitor/dashboard', 'GET', '监控仪表板', 'monitor', '监控', '["monitor.view"]', 'dashboard', 600, 1, 1, '查看监控仪表板'),
(UPPER(REPLACE(UUID(), '-', '')), '/projects/:projectId/monitor/metrics', 'GET', '指标监控', 'monitor', '监控', '["monitor.metrics"]', '', 601, 0, 1, '查看监控指标'),
(UPPER(REPLACE(UUID(), '-', '')), '/projects/:projectId/monitor/logs', 'GET', '日志查询', 'monitor', '监控', '["monitor.logs"]', '', 602, 0, 1, '查询系统日志'),
(UPPER(REPLACE(UUID(), '-', '')), '/projects/:projectId/monitor/alerts', 'GET', '告警管理', 'monitor', '监控', '["monitor.alert"]', '', 603, 0, 1, '查看和管理告警');

-- 组织路由
INSERT INTO `t_router_permission` (`route_id`, `path`, `method`, `name`, `group`, `category`, `required_permissions`, `icon`, `order`, `is_menu`, `is_enabled`, `description`) VALUES
(UPPER(REPLACE(UUID(), '-', '')), '/orgs', 'GET', '组织列表', 'org', '组织管理', '[]', 'organization', 700, 1, 1, '查看组织列表'),
(UPPER(REPLACE(UUID(), '-', '')), '/orgs/:orgId', 'GET', '组织详情', 'org', '组织管理', '[]', '', 701, 0, 1, '查看组织详情'),
(UPPER(REPLACE(UUID(), '-', '')), '/orgs', 'POST', '创建组织', 'org', '组织管理', '[]', '', 702, 0, 1, '创建新组织'),
(UPPER(REPLACE(UUID(), '-', '')), '/orgs/:orgId', 'PUT', '更新组织', 'org', '组织管理', '[]', '', 703, 0, 1, '更新组织信息'),
(UPPER(REPLACE(UUID(), '-', '')), '/orgs/:orgId/members', 'GET', '组织成员', 'org', '组织管理', '[]', '', 704, 0, 1, '查看组织成员');

-- 团队路由
INSERT INTO `t_router_permission` (`route_id`, `path`, `method`, `name`, `group`, `category`, `required_permissions`, `icon`, `order`, `is_menu`, `is_enabled`, `description`) VALUES
(UPPER(REPLACE(UUID(), '-', '')), '/orgs/:orgId/teams', 'GET', '团队列表', 'team', '团队管理', '["team.view"]', 'team', 800, 1, 1, '查看团队列表'),
(UPPER(REPLACE(UUID(), '-', '')), '/orgs/:orgId/teams/:teamId', 'GET', '团队详情', 'team', '团队管理', '["team.view"]', '', 801, 0, 1, '查看团队详情'),
(UPPER(REPLACE(UUID(), '-', '')), '/orgs/:orgId/teams', 'POST', '创建团队', 'team', '团队管理', '["team.create"]', '', 802, 0, 1, '创建新团队'),
(UPPER(REPLACE(UUID(), '-', '')), '/orgs/:orgId/teams/:teamId', 'PUT', '更新团队', 'team', '团队管理', '["team.edit"]', '', 803, 0, 1, '更新团队信息'),
(UPPER(REPLACE(UUID(), '-', '')), '/orgs/:orgId/teams/:teamId', 'DELETE', '删除团队', 'team', '团队管理', '["team.delete"]', '', 804, 0, 1, '删除团队'),
(UPPER(REPLACE(UUID(), '-', '')), '/orgs/:orgId/teams/:teamId/members', 'GET', '团队成员', 'team', '团队管理', '["team.member"]', '', 805, 0, 1, '管理团队成员');

-- 用户权限相关路由（用于获取当前用户权限信息）
INSERT INTO `t_router_permission` (`route_id`, `path`, `method`, `name`, `group`, `category`, `required_permissions`, `icon`, `order`, `is_menu`, `is_enabled`, `description`) VALUES
(UPPER(REPLACE(UUID(), '-', '')), '/user/permissions', 'GET', '我的权限', 'user', '用户中心', '[]', '', 900, 0, 1, '获取当前用户权限信息'),
(UPPER(REPLACE(UUID(), '-', '')), '/user/routes', 'GET', '可访问路由', 'user', '用户中心', '[]', '', 901, 0, 1, '获取当前用户可访问的路由列表'),
(UPPER(REPLACE(UUID(), '-', '')), '/user/permissions/summary', 'GET', '权限摘要', 'user', '用户中心', '[]', '', 902, 0, 1, '获取当前用户权限摘要');

-- ============================================
-- 数据验证
-- ============================================

-- 查看插入的路由数量
SELECT 
    category AS '分类',
    COUNT(*) AS '路由数量',
    SUM(CASE WHEN is_menu = 1 THEN 1 ELSE 0 END) AS '菜单项数量'
FROM t_router_permission 
WHERE is_enabled = 1
GROUP BY category
ORDER BY category;

-- 查看所有菜单项
SELECT 
    category AS '分类',
    name AS '名称',
    path AS '路径',
    method AS '方法',
    icon AS '图标',
    `order` AS '排序'
FROM t_router_permission 
WHERE is_enabled = 1 AND is_menu = 1
ORDER BY category, `order`;

-- ============================================
-- 完成提示
-- ============================================
SELECT '路由权限表创建完成！' AS '状态', 
       COUNT(*) AS '路由总数',
       SUM(CASE WHEN is_menu = 1 THEN 1 ELSE 0 END) AS '菜单项数量'
FROM t_router_permission;

