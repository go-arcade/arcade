-- ================================================================
-- Arcade CI/CD 平台数据库表结构设计
-- 数据库: arcade_ci_meta
-- 字符集: utf8mb4
-- 排序规则: utf8mb4_unicode_ci
-- ================================================================

-- ================================================================
-- 用户和权限管理模块
-- ================================================================

-- 用户表
CREATE TABLE IF NOT EXISTS `t_user` (
  `id` INT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` VARCHAR(64) NOT NULL COMMENT '用户唯一标识',
  `username` VARCHAR(64) NOT NULL COMMENT '用户名',
  `nick_name` VARCHAR(128) DEFAULT NULL COMMENT '昵称',
  `password` VARCHAR(255) NOT NULL COMMENT '密码(加密)',
  `avatar` VARCHAR(512) DEFAULT NULL COMMENT '头像URL',
  `email` VARCHAR(128) DEFAULT NULL COMMENT '邮箱',
  `phone` VARCHAR(32) DEFAULT NULL COMMENT '手机号',
  `is_enabled` TINYINT NOT NULL DEFAULT 0 COMMENT '是否启用: 0-启用, 1-禁用',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_id` (`user_id`),
  UNIQUE KEY `uk_username` (`username`),
  UNIQUE KEY `uk_email` (`email`),
  KEY `idx_is_enabled` (`is_enabled`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 角色表
CREATE TABLE IF NOT EXISTS `t_role` (
  `id` INT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `role_id` VARCHAR(64) NOT NULL COMMENT '角色唯一标识',
  `role_name` VARCHAR(128) NOT NULL COMMENT '角色名称',
  `role_code` VARCHAR(64) NOT NULL COMMENT '角色编码',
  `role_desc` VARCHAR(512) DEFAULT NULL COMMENT '角色描述',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_role_id` (`role_id`),
  UNIQUE KEY `uk_role_code` (`role_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色表';

-- 用户组表
CREATE TABLE IF NOT EXISTS `t_user_group` (
  `id` INT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `group_id` VARCHAR(64) NOT NULL COMMENT '用户组唯一标识',
  `group_name` VARCHAR(128) NOT NULL COMMENT '用户组名称',
  `group_desc` VARCHAR(512) DEFAULT NULL COMMENT '用户组描述',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_group_id` (`group_id`),
  UNIQUE KEY `uk_group_name` (`group_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户组表';

-- 角色关系表（用户-角色、用户组-角色）
CREATE TABLE IF NOT EXISTS `t_role_relation` (
  `id` INT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `role_id` VARCHAR(64) NOT NULL COMMENT '角色ID',
  `user_id` VARCHAR(64) DEFAULT NULL COMMENT '用户ID',
  `group_id` VARCHAR(64) DEFAULT NULL COMMENT '用户组ID',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_role_id` (`role_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_group_id` (`group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色关系表';

-- SSO认证提供者表
CREATE TABLE IF NOT EXISTS `t_sso_provider` (
  `id` INT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `provider_id` VARCHAR(64) NOT NULL COMMENT '提供者唯一标识',
  `name` VARCHAR(128) NOT NULL COMMENT '提供者名称',
  `provider_type` VARCHAR(32) NOT NULL COMMENT '提供者类型(oauth/ldap/oidc/saml)',
  `config` JSON NOT NULL COMMENT '配置内容(根据type不同,内容结构不同)',
  `description` VARCHAR(512) DEFAULT NULL COMMENT '描述',
  `priority` INT NOT NULL DEFAULT 0 COMMENT '优先级(数字越小优先级越高)',
  `is_enabled` TINYINT NOT NULL DEFAULT 1 COMMENT '是否启用: 0-禁用, 1-启用',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_provider_id` (`provider_id`),
  KEY `idx_provider_type` (`provider_type`),
  KEY `idx_priority` (`priority`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='SSO认证提供者表';

-- ================================================================
-- Agent管理模块
-- ================================================================

-- Agent表（执行器）
CREATE TABLE IF NOT EXISTS `t_agent` (
  `id` INT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `agent_id` VARCHAR(64) NOT NULL COMMENT 'Agent唯一标识',
  `agent_name` VARCHAR(128) NOT NULL COMMENT 'Agent名称',
  `hostname` VARCHAR(255) DEFAULT NULL COMMENT '主机名',
  `address` VARCHAR(255) NOT NULL COMMENT 'Agent地址',
  `port` VARCHAR(10) NOT NULL COMMENT 'Agent端口',
  `username` VARCHAR(64) DEFAULT NULL COMMENT 'SSH用户名',
  `auth_type` TINYINT NOT NULL DEFAULT 0 COMMENT '认证类型: 0-密码, 1-密钥',
  `os` VARCHAR(32) DEFAULT NULL COMMENT '操作系统',
  `arch` VARCHAR(32) DEFAULT NULL COMMENT '架构(amd64/arm64)',
  `version` VARCHAR(32) DEFAULT NULL COMMENT 'Agent版本',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT 'Agent状态: 0-未知, 1-在线, 2-离线, 3-忙碌, 4-空闲',
  `max_concurrent_jobs` INT NOT NULL DEFAULT 1 COMMENT '最大并发任务数',
  `running_jobs_count` INT NOT NULL DEFAULT 0 COMMENT '正在执行的任务数',
  `labels` JSON DEFAULT NULL COMMENT 'Agent标签(JSON格式)',
  `metrics` JSON DEFAULT NULL COMMENT 'Agent指标(CPU/内存等)',
  `last_heartbeat` DATETIME DEFAULT NULL COMMENT '最后心跳时间',
  `is_enabled` TINYINT NOT NULL DEFAULT 1 COMMENT '是否启用: 0-禁用, 1-启用',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_agent_id` (`agent_id`),
  KEY `idx_status` (`status`),
  KEY `idx_is_enabled` (`is_enabled`),
  KEY `idx_last_heartbeat` (`last_heartbeat`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Agent表';

-- Agent配置表（每个Agent一条记录）
CREATE TABLE IF NOT EXISTS `t_agent_config` (
  `id` INT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `agent_id` VARCHAR(64) NOT NULL COMMENT 'Agent唯一标识',
  `config_items` JSON NOT NULL COMMENT '配置项',
  `description` VARCHAR(512) DEFAULT NULL COMMENT '配置描述',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_agent_id` (`agent_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Agent配置表';

-- ================================================================
-- 流水线和任务管理模块
-- ================================================================

-- 流水线定义表
CREATE TABLE IF NOT EXISTS `t_pipeline` (
  `id` INT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `pipeline_id` VARCHAR(64) NOT NULL COMMENT '流水线唯一标识',
  `name` VARCHAR(255) NOT NULL COMMENT '流水线名称',
  `description` TEXT DEFAULT NULL COMMENT '流水线描述',
  `repo_url` VARCHAR(512) DEFAULT NULL COMMENT '代码仓库URL',
  `branch` VARCHAR(128) DEFAULT 'main' COMMENT '分支',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '流水线状态: 0-未知, 1-等待, 2-运行中, 3-成功, 4-失败, 5-已取消, 6-部分成功',
  `trigger_type` TINYINT NOT NULL DEFAULT 1 COMMENT '触发类型: 0-未知, 1-手动, 2-Webhook, 3-定时, 4-API',
  `cron` VARCHAR(128) DEFAULT NULL COMMENT 'Cron表达式(定时触发)',
  `env` JSON DEFAULT NULL COMMENT '全局环境变量(JSON格式)',
  `total_runs` INT NOT NULL DEFAULT 0 COMMENT '总执行次数',
  `success_runs` INT NOT NULL DEFAULT 0 COMMENT '成功次数',
  `failed_runs` INT NOT NULL DEFAULT 0 COMMENT '失败次数',
  `created_by` VARCHAR(64) NOT NULL COMMENT '创建者用户ID',
  `is_enabled` TINYINT NOT NULL DEFAULT 1 COMMENT '是否启用: 0-禁用, 1-启用',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_pipeline_id` (`pipeline_id`),
  KEY `idx_name` (`name`(191)),
  KEY `idx_status` (`status`),
  KEY `idx_created_by` (`created_by`),
  KEY `idx_is_enabled` (`is_enabled`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='流水线定义表';

-- 流水线执行记录表
CREATE TABLE IF NOT EXISTS `t_pipeline_run` (
  `id` INT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `run_id` VARCHAR(64) NOT NULL COMMENT '流水线执行唯一标识',
  `pipeline_id` VARCHAR(64) NOT NULL COMMENT '流水线ID',
  `pipeline_name` VARCHAR(255) NOT NULL COMMENT '流水线名称(冗余)',
  `branch` VARCHAR(128) DEFAULT NULL COMMENT '分支',
  `commit_sha` VARCHAR(64) DEFAULT NULL COMMENT 'Commit SHA',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '执行状态: 0-未知, 1-等待, 2-运行中, 3-成功, 4-失败, 5-已取消, 6-部分成功',
  `trigger_type` TINYINT NOT NULL DEFAULT 1 COMMENT '触发类型: 0-未知, 1-手动, 2-Webhook, 3-定时, 4-API',
  `triggered_by` VARCHAR(64) DEFAULT NULL COMMENT '触发者用户ID',
  `env` JSON DEFAULT NULL COMMENT '环境变量(JSON格式)',
  `total_jobs` INT NOT NULL DEFAULT 0 COMMENT '总任务数',
  `completed_jobs` INT NOT NULL DEFAULT 0 COMMENT '已完成任务数',
  `failed_jobs` INT NOT NULL DEFAULT 0 COMMENT '失败任务数',
  `running_jobs` INT NOT NULL DEFAULT 0 COMMENT '运行中任务数',
  `current_stage` INT NOT NULL DEFAULT 0 COMMENT '当前阶段',
  `total_stages` INT NOT NULL DEFAULT 0 COMMENT '总阶段数',
  `start_time` DATETIME DEFAULT NULL COMMENT '开始时间',
  `end_time` DATETIME DEFAULT NULL COMMENT '结束时间',
  `duration` BIGINT DEFAULT NULL COMMENT '执行时长(毫秒)',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_run_id` (`run_id`),
  KEY `idx_pipeline_id` (`pipeline_id`),
  KEY `idx_status` (`status`),
  KEY `idx_triggered_by` (`triggered_by`),
  KEY `idx_start_time` (`start_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='流水线执行记录表';

-- 流水线阶段表
CREATE TABLE IF NOT EXISTS `t_pipeline_stage` (
  `id` INT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `stage_id` VARCHAR(64) NOT NULL COMMENT '阶段唯一标识',
  `pipeline_id` VARCHAR(64) NOT NULL COMMENT '流水线ID',
  `name` VARCHAR(255) NOT NULL COMMENT '阶段名称',
  `stage_order` INT NOT NULL DEFAULT 0 COMMENT '阶段顺序',
  `parallel` TINYINT NOT NULL DEFAULT 0 COMMENT '是否并行执行: 0-否, 1-是',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_stage_id` (`stage_id`),
  KEY `idx_pipeline_id` (`pipeline_id`),
  KEY `idx_stage_order` (`stage_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='流水线阶段表';

-- 任务表
CREATE TABLE IF NOT EXISTS `t_job` (
  `id` INT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `job_id` VARCHAR(64) NOT NULL COMMENT '任务唯一标识',
  `name` VARCHAR(255) NOT NULL COMMENT '任务名称',
  `pipeline_id` VARCHAR(64) DEFAULT NULL COMMENT '所属流水线ID',
  `pipeline_run_id` VARCHAR(64) DEFAULT NULL COMMENT '所属流水线执行ID',
  `stage_id` VARCHAR(64) DEFAULT NULL COMMENT '所属阶段ID',
  `stage` INT NOT NULL DEFAULT 0 COMMENT '阶段序号',
  `agent_id` VARCHAR(64) DEFAULT NULL COMMENT '执行的Agent ID',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '任务状态: 0-未知, 1-等待, 2-入队, 3-运行中, 4-成功, 5-失败, 6-已取消, 7-超时, 8-已跳过',
  `priority` INT NOT NULL DEFAULT 5 COMMENT '优先级: 1-最高, 5-普通, 10-最低',
  `image` VARCHAR(255) DEFAULT NULL COMMENT 'Docker镜像',
  `commands` TEXT DEFAULT NULL COMMENT '执行命令列表(JSON数组)',
  `workspace` VARCHAR(512) DEFAULT NULL COMMENT '工作目录',
  `env` JSON DEFAULT NULL COMMENT '环境变量(JSON格式)',
  `secrets` JSON DEFAULT NULL COMMENT '密钥信息(JSON格式)',
  `timeout` INT NOT NULL DEFAULT 3600 COMMENT '超时时间(秒)',
  `retry_count` INT NOT NULL DEFAULT 0 COMMENT '重试次数',
  `current_retry` INT NOT NULL DEFAULT 0 COMMENT '当前重试次数',
  `allow_failure` TINYINT NOT NULL DEFAULT 0 COMMENT '是否允许失败: 0-否, 1-是',
  `label_selector` JSON DEFAULT NULL COMMENT '标签选择器(JSON格式)',
  `tags` VARCHAR(512) DEFAULT NULL COMMENT '任务标签(逗号分隔,已废弃)',
  `depends_on` VARCHAR(512) DEFAULT NULL COMMENT '依赖的任务ID列表(逗号分隔)',
  `exit_code` INT DEFAULT NULL COMMENT '退出码',
  `error_message` TEXT DEFAULT NULL COMMENT '错误信息',
  `start_time` DATETIME DEFAULT NULL COMMENT '开始时间',
  `end_time` DATETIME DEFAULT NULL COMMENT '结束时间',
  `duration` BIGINT DEFAULT NULL COMMENT '执行时长(毫秒)',
  `created_by` VARCHAR(64) DEFAULT NULL COMMENT '创建者用户ID',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_job_id` (`job_id`),
  KEY `idx_pipeline_id` (`pipeline_id`),
  KEY `idx_pipeline_run_id` (`pipeline_run_id`),
  KEY `idx_agent_id` (`agent_id`),
  KEY `idx_status` (`status`),
  KEY `idx_priority` (`priority`),
  KEY `idx_start_time` (`start_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务表';

-- 任务产物表
CREATE TABLE IF NOT EXISTS `t_job_artifact` (
  `id` INT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `artifact_id` VARCHAR(64) NOT NULL COMMENT '产物唯一标识',
  `job_id` VARCHAR(64) NOT NULL COMMENT '任务ID',
  `pipeline_run_id` VARCHAR(64) DEFAULT NULL COMMENT '流水线执行ID',
  `name` VARCHAR(255) NOT NULL COMMENT '产物名称',
  `path` VARCHAR(1024) NOT NULL COMMENT '产物路径(支持glob模式)',
  `destination` VARCHAR(1024) DEFAULT NULL COMMENT '目标存储路径',
  `size` BIGINT DEFAULT NULL COMMENT '文件大小(字节)',
  `storage_type` VARCHAR(32) DEFAULT 'minio' COMMENT '存储类型(minio/s3/oss/gcs/cos)',
  `storage_path` VARCHAR(1024) DEFAULT NULL COMMENT '实际存储路径',
  `expire` TINYINT NOT NULL DEFAULT 0 COMMENT '是否过期: 0-否, 1-是',
  `expire_days` INT DEFAULT NULL COMMENT '过期天数',
  `expired_at` DATETIME DEFAULT NULL COMMENT '过期时间',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_artifact_id` (`artifact_id`),
  KEY `idx_job_id` (`job_id`),
  KEY `idx_pipeline_run_id` (`pipeline_run_id`),
  KEY `idx_expire` (`expire`, `expired_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务产物表';

-- ================================================================
-- 日志和事件模块（MongoDB Collections）
-- ================================================================

-- MongoDB Collection 设计说明
-- 数据库名称: job_log
-- 以下为各 Collection 的结构定义

-- Collection 1: job_logs (任务执行日志)
-- 用途: 存储 Job 执行过程中的日志输出
-- 索引: job_id, timestamp, agent_id
-- {
--   "_id": ObjectId,
--   "log_id": "uuid",
--   "job_id": "job_xxx",
--   "pipeline_run_id": "run_xxx",
--   "agent_id": "agent_xxx",
--   "line_number": 1,
--   "content": "log content line",
--   "timestamp": ISODate("2024-01-01T00:00:00Z"),
--   "level": "INFO|WARN|ERROR|DEBUG",
--   "created_at": ISODate("2024-01-01T00:00:00Z")
-- }

-- Collection 2: terminal_logs (终端日志/构建日志)
-- 用途: 存储构建、发布等各环境产生的完整终端日志
-- 索引: session_id, environment, timestamp, user_id
-- {
--   "_id": ObjectId,
--   "session_id": "uuid",
--   "session_type": "build|deploy|release|debug",
--   "environment": "dev|test|staging|prod",
--   "job_id": "job_xxx",
--   "pipeline_id": "pipeline_xxx",
--   "pipeline_run_id": "run_xxx",
--   "user_id": "user_xxx",
--   "hostname": "agent-hostname",
--   "working_directory": "/path/to/workspace",
--   "command": "npm run build",
--   "exit_code": 0,
--   "logs": [
--     {
--       "line": 1,
--       "timestamp": ISODate("2024-01-01T00:00:00.000Z"),
--       "content": "Starting build process...",
--       "stream": "stdout|stderr"
--     },
--     {
--       "line": 2,
--       "timestamp": ISODate("2024-01-01T00:00:01.000Z"),
--       "content": "Compiling source files...",
--       "stream": "stdout"
--     }
--   ],
--   "metadata": {
--     "total_lines": 150,
--     "duration_ms": 5000,
--     "output_size_bytes": 10240
--   },
--   "status": "running|completed|failed|timeout",
--   "start_time": ISODate("2024-01-01T00:00:00Z"),
--   "end_time": ISODate("2024-01-01T00:05:00Z"),
--   "created_at": ISODate("2024-01-01T00:00:00Z"),
--   "updated_at": ISODate("2024-01-01T00:05:00Z")
-- }

-- Collection 3: build_artifacts_logs (产物构建日志)
-- 用途: 记录产物上传、下载等操作日志
-- 索引: artifact_id, operation, timestamp
-- {
--   "_id": ObjectId,
--   "artifact_id": "artifact_xxx",
--   "job_id": "job_xxx",
--   "operation": "upload|download|delete",
--   "file_name": "app-v1.0.0.tar.gz",
--   "file_size": 1024000,
--   "storage_type": "minio|s3|oss|gcs|cos",
--   "storage_path": "artifacts/2024/01/01/xxx.tar.gz",
--   "user_id": "user_xxx",
--   "status": "success|failed",
--   "error_message": "error details if failed",
--   "duration_ms": 2000,
--   "timestamp": ISODate("2024-01-01T00:00:00Z")
-- }

-- 建议的 MongoDB 索引
-- db.job_logs.createIndex({ "job_id": 1, "timestamp": 1 })
-- db.job_logs.createIndex({ "agent_id": 1, "timestamp": -1 })
-- db.job_logs.createIndex({ "pipeline_run_id": 1, "timestamp": 1 })
-- db.job_logs.createIndex({ "timestamp": 1 }, { expireAfterSeconds: 7776000 }) // 90天后自动删除
--
-- db.terminal_logs.createIndex({ "session_id": 1 })
-- db.terminal_logs.createIndex({ "environment": 1, "timestamp": -1 })
-- db.terminal_logs.createIndex({ "job_id": 1, "timestamp": -1 })
-- db.terminal_logs.createIndex({ "pipeline_run_id": 1, "timestamp": -1 })
-- db.terminal_logs.createIndex({ "user_id": 1, "timestamp": -1 })
-- db.terminal_logs.createIndex({ "status": 1, "timestamp": -1 })
-- db.terminal_logs.createIndex({ "created_at": 1 }, { expireAfterSeconds: 15552000 }) // 180天后自动删除
--
-- db.build_artifacts_logs.createIndex({ "artifact_id": 1, "timestamp": -1 })
-- db.build_artifacts_logs.createIndex({ "job_id": 1, "timestamp": -1 })
-- db.build_artifacts_logs.createIndex({ "operation": 1, "timestamp": -1 })
-- db.build_artifacts_logs.createIndex({ "timestamp": 1 }, { expireAfterSeconds: 7776000 }) // 90天后自动删除

-- 系统事件表
CREATE TABLE IF NOT EXISTS `t_system_event` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `event_id` VARCHAR(64) NOT NULL COMMENT '事件唯一标识',
  `event_type` TINYINT NOT NULL COMMENT '事件类型: 1-任务创建, 2-任务开始, 3-任务完成, 4-任务失败, 5-Agent上线, 6-流水线开始, 7-流水线完成, 8-流水线失败',
  `resource_type` VARCHAR(32) NOT NULL COMMENT '资源类型(job/pipeline/agent)',
  `resource_id` VARCHAR(64) NOT NULL COMMENT '资源ID',
  `resource_name` VARCHAR(255) DEFAULT NULL COMMENT '资源名称',
  `message` TEXT DEFAULT NULL COMMENT '事件消息',
  `metadata` JSON DEFAULT NULL COMMENT '事件元数据(JSON格式)',
  `user_id` VARCHAR(64) DEFAULT NULL COMMENT '关联用户ID',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_event_id` (`event_id`),
  KEY `idx_event_type` (`event_type`),
  KEY `idx_resource` (`resource_type`, `resource_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_create_time` (`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统事件表';

-- ================================================================
-- 插件管理模块
-- ================================================================

-- 插件表
CREATE TABLE IF NOT EXISTS `t_plugin` (
  `id` INT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `plugin_id` VARCHAR(64) NOT NULL COMMENT '插件唯一标识',
  `name` VARCHAR(128) NOT NULL COMMENT '插件名称',
  `version` VARCHAR(32) NOT NULL COMMENT '插件版本',
  `description` TEXT DEFAULT NULL COMMENT '插件描述',
  `author` VARCHAR(128) DEFAULT NULL COMMENT '插件作者',
  `plugin_type` VARCHAR(32) NOT NULL COMMENT '插件类型(notify/deploy/test/build/custom)',
  `entry_point` VARCHAR(255) NOT NULL COMMENT '插件入口(文件路径或命令)',
  `config_schema` JSON DEFAULT NULL COMMENT '配置项Schema定义(JSON Schema格式)',
  `params_schema` JSON DEFAULT NULL COMMENT '参数Schema定义(JSON Schema格式)',
  `default_config` JSON DEFAULT NULL COMMENT '默认配置值',
  `icon` VARCHAR(512) DEFAULT NULL COMMENT '插件图标URL',
  `repository` VARCHAR(512) DEFAULT NULL COMMENT '代码仓库地址',
  `documentation` TEXT DEFAULT NULL COMMENT '文档说明',
  `is_enabled` TINYINT NOT NULL DEFAULT 1 COMMENT '是否启用: 0-禁用, 1-启用',
  `is_builtin` TINYINT NOT NULL DEFAULT 0 COMMENT '是否内置插件: 0-否, 1-是',
  `install_path` VARCHAR(512) DEFAULT NULL COMMENT '插件安装路径',
  `checksum` VARCHAR(128) DEFAULT NULL COMMENT '插件文件校验和(SHA256)',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_plugin_id_version` (`plugin_id`, `version`),
  KEY `idx_plugin_type` (`plugin_type`),
  KEY `idx_is_enabled` (`is_enabled`),
  KEY `idx_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='插件表';

-- 插件配置表
CREATE TABLE IF NOT EXISTS `t_plugin_config` (
  `id` INT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `config_id` VARCHAR(64) NOT NULL COMMENT '配置唯一标识',
  `plugin_id` VARCHAR(64) NOT NULL COMMENT '插件ID',
  `name` VARCHAR(128) NOT NULL COMMENT '配置名称',
  `config_items` JSON NOT NULL COMMENT '配置项(key-value格式)',
  `description` VARCHAR(512) DEFAULT NULL COMMENT '配置描述',
  `scope` VARCHAR(32) NOT NULL DEFAULT 'global' COMMENT '作用域(global/pipeline/user)',
  `scope_id` VARCHAR(64) DEFAULT NULL COMMENT '作用域ID',
  `is_default` TINYINT NOT NULL DEFAULT 0 COMMENT '是否默认配置: 0-否, 1-是',
  `created_by` VARCHAR(64) DEFAULT NULL COMMENT '创建者用户ID',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_config_id` (`config_id`),
  KEY `idx_plugin_id` (`plugin_id`),
  KEY `idx_scope` (`scope`, `scope_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='插件配置表';

-- 任务插件关联表
CREATE TABLE IF NOT EXISTS `t_job_plugin` (
  `id` INT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `job_id` VARCHAR(64) NOT NULL COMMENT '任务ID',
  `plugin_id` VARCHAR(64) NOT NULL COMMENT '插件ID',
  `plugin_config_id` VARCHAR(64) DEFAULT NULL COMMENT '插件配置ID',
  `params` JSON DEFAULT NULL COMMENT '任务特定的插件参数',
  `execution_order` INT NOT NULL DEFAULT 0 COMMENT '执行顺序',
  `execution_stage` VARCHAR(32) NOT NULL COMMENT '执行阶段(before/after/on_success/on_failure)',
  `status` TINYINT DEFAULT 0 COMMENT '执行状态: 0-未执行, 1-执行中, 2-成功, 3-失败',
  `result` TEXT DEFAULT NULL COMMENT '执行结果',
  `error_message` TEXT DEFAULT NULL COMMENT '错误信息',
  `started_at` DATETIME DEFAULT NULL COMMENT '开始执行时间',
  `completed_at` DATETIME DEFAULT NULL COMMENT '完成时间',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_job_id` (`job_id`),
  KEY `idx_plugin_id` (`plugin_id`),
  KEY `idx_execution_order` (`execution_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务插件关联表';

-- 插件权限表
CREATE TABLE IF NOT EXISTS `t_plugin_permission` (
  `id` INT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `plugin_id` VARCHAR(64) NOT NULL COMMENT '插件ID',
  `permission_type` VARCHAR(32) NOT NULL COMMENT '权限类型(file_system/network/process/database/environment)',
  `resource` VARCHAR(255) DEFAULT NULL COMMENT '资源路径或标识',
  `action` VARCHAR(32) NOT NULL COMMENT '操作(read/write/execute)',
  `is_allowed` TINYINT NOT NULL DEFAULT 0 COMMENT '是否允许: 0-否, 1-是',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_plugin_id` (`plugin_id`),
  KEY `idx_permission_type` (`permission_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='插件权限表';

-- 插件资源配额表
CREATE TABLE IF NOT EXISTS `t_plugin_quota` (
  `id` INT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `plugin_id` VARCHAR(64) NOT NULL COMMENT '插件ID',
  `max_cpu_percent` INT DEFAULT 50 COMMENT '最大CPU使用率(%)',
  `max_memory_mb` INT DEFAULT 512 COMMENT '最大内存(MB)',
  `max_execution_time` INT DEFAULT 300 COMMENT '最大执行时间(秒)',
  `max_disk_io_mbps` INT DEFAULT 100 COMMENT '最大磁盘IO(MB/s)',
  `max_network_mbps` INT DEFAULT 10 COMMENT '最大网络带宽(MB/s)',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_plugin_id` (`plugin_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='插件资源配额表';

-- 插件执行审计表
CREATE TABLE IF NOT EXISTS `t_plugin_audit` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `plugin_id` VARCHAR(64) NOT NULL COMMENT '插件ID',
  `job_id` VARCHAR(64) DEFAULT NULL COMMENT '关联任务ID',
  `user_id` VARCHAR(64) DEFAULT NULL COMMENT '触发用户ID',
  `action` VARCHAR(64) NOT NULL COMMENT '操作(load/execute/unload)',
  `status` VARCHAR(32) NOT NULL COMMENT '状态(success/failed)',
  `error_message` TEXT DEFAULT NULL COMMENT '错误信息',
  `execution_time_ms` INT DEFAULT NULL COMMENT '执行时长(毫秒)',
  `resource_usage` JSON DEFAULT NULL COMMENT '资源使用情况',
  `network_calls` JSON DEFAULT NULL COMMENT '网络调用记录',
  `file_access` JSON DEFAULT NULL COMMENT '文件访问记录',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_plugin_id` (`plugin_id`),
  KEY `idx_job_id` (`job_id`),
  KEY `idx_action` (`action`),
  KEY `idx_create_time` (`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='插件执行审计表';

-- 插件来源表
CREATE TABLE IF NOT EXISTS `t_plugin_source` (
  `id` INT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `source_id` VARCHAR(64) NOT NULL COMMENT '来源唯一标识',
  `name` VARCHAR(128) NOT NULL COMMENT '来源名称',
  `source_type` VARCHAR(32) NOT NULL COMMENT '来源类型(official/verified/community)',
  `repository` VARCHAR(512) NOT NULL COMMENT '仓库地址',
  `public_key` TEXT DEFAULT NULL COMMENT '签名公钥',
  `is_trusted` TINYINT NOT NULL DEFAULT 0 COMMENT '是否信任: 0-否, 1-是',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_source_id` (`source_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='插件来源表';

-- ================================================================
-- 配置管理模块
-- ================================================================

-- 对象存储配置表
CREATE TABLE IF NOT EXISTS `t_storage_config` (
  `id` INT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `storage_id` VARCHAR(64) NOT NULL COMMENT '存储唯一标识',
  `name` VARCHAR(128) NOT NULL COMMENT '存储名称',
  `storage_type` VARCHAR(32) NOT NULL COMMENT '存储类型(minio/s3/oss/gcs/cos)',
  `config` JSON NOT NULL COMMENT '存储配置(根据type不同,内容结构不同)',
  `description` VARCHAR(512) DEFAULT NULL COMMENT '描述',
  `is_default` TINYINT NOT NULL DEFAULT 0 COMMENT '是否默认: 0-否, 1-是',
  `is_enabled` TINYINT NOT NULL DEFAULT 1 COMMENT '是否启用: 0-禁用, 1-启用',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_storage_id` (`storage_id`),
  KEY `idx_storage_type` (`storage_type`),
  KEY `idx_is_default` (`is_default`),
  KEY `idx_is_enabled` (`is_enabled`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='对象存储配置表';

-- 系统配置表
CREATE TABLE IF NOT EXISTS `t_system_config` (
  `id` INT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `config_key` VARCHAR(128) NOT NULL COMMENT '配置键',
  `config_value` TEXT NOT NULL COMMENT '配置值',
  `config_type` VARCHAR(32) NOT NULL DEFAULT 'string' COMMENT '配置类型(string/json/int/bool)',
  `description` VARCHAR(512) DEFAULT NULL COMMENT '配置描述',
  `is_encrypted` TINYINT NOT NULL DEFAULT 0 COMMENT '是否加密: 0-否, 1-是',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_config_key` (`config_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统配置表';

-- 密钥管理表
CREATE TABLE IF NOT EXISTS `t_secret` (
  `id` INT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `secret_id` VARCHAR(64) NOT NULL COMMENT '密钥唯一标识',
  `name` VARCHAR(255) NOT NULL COMMENT '密钥名称',
  `secret_type` VARCHAR(32) NOT NULL COMMENT '密钥类型(password/token/ssh_key/env)',
  `secret_value` TEXT NOT NULL COMMENT '密钥值(加密存储)',
  `description` VARCHAR(512) DEFAULT NULL COMMENT '密钥描述',
  `scope` VARCHAR(32) NOT NULL DEFAULT 'global' COMMENT '作用域(global/pipeline/user)',
  `scope_id` VARCHAR(64) DEFAULT NULL COMMENT '作用域ID',
  `created_by` VARCHAR(64) NOT NULL COMMENT '创建者用户ID',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_secret_id` (`secret_id`),
  KEY `idx_name` (`name`(191)),
  KEY `idx_scope` (`scope`, `scope_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='密钥管理表';

-- ================================================================
-- 审计日志模块
-- ================================================================

-- 操作审计日志表
CREATE TABLE IF NOT EXISTS `t_audit_log` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` VARCHAR(64) NOT NULL COMMENT '操作用户ID',
  `username` VARCHAR(64) NOT NULL COMMENT '操作用户名',
  `action` VARCHAR(64) NOT NULL COMMENT '操作动作(create/update/delete/execute)',
  `resource_type` VARCHAR(32) NOT NULL COMMENT '资源类型(pipeline/job/agent/user)',
  `resource_id` VARCHAR(64) DEFAULT NULL COMMENT '资源ID',
  `resource_name` VARCHAR(255) DEFAULT NULL COMMENT '资源名称',
  `ip_address` VARCHAR(64) DEFAULT NULL COMMENT 'IP地址',
  `user_agent` VARCHAR(512) DEFAULT NULL COMMENT 'User Agent',
  `request_params` JSON DEFAULT NULL COMMENT '请求参数(JSON格式)',
  `response_status` INT DEFAULT NULL COMMENT '响应状态码',
  `error_message` TEXT DEFAULT NULL COMMENT '错误信息',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '操作时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_action` (`action`),
  KEY `idx_resource` (`resource_type`, `resource_id`),
  KEY `idx_create_time` (`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='操作审计日志表';

-- ================================================================
-- 初始化数据
-- ================================================================

-- 插入默认管理员用户（密码: admin123，需要在应用层加密）
INSERT INTO `t_user` (`user_id`, `username`, `nick_name`, `password`, `email`, `is_enabled`)
VALUES ('user_admin', 'admin', '系统管理员', 'TO_BE_ENCRYPTED', 'admin@arcade.local', 0)
ON DUPLICATE KEY UPDATE `username` = `username`;

-- 插入默认角色
INSERT INTO `t_role` (`role_id`, `role_name`, `role_code`, `role_desc`)
VALUES
  ('role_admin', '管理员', 'ADMIN', '系统管理员角色'),
  ('role_developer', '开发者', 'DEVELOPER', '开发者角色'),
  ('role_operator', '运维人员', 'OPERATOR', '运维人员角色'),
  ('role_viewer', '查看者', 'VIEWER', '只读查看角色')
ON DUPLICATE KEY UPDATE `role_name` = `role_name`;

-- 绑定管理员角色
INSERT INTO `t_role_relation` (`role_id`, `user_id`)
VALUES ('role_admin', 'user_admin')
ON DUPLICATE KEY UPDATE `role_id` = `role_id`;

-- 插入示例 SSO 提供者配置
-- OAuth (GitHub)
INSERT INTO `t_sso_provider` (`provider_id`, `name`, `provider_type`, `config`, `description`, `priority`, `is_enabled`)
VALUES (
  'oauth_github',
  'GitHub',
  'oauth',
  JSON_OBJECT(
    'clientId', 'YOUR_GITHUB_CLIENT_ID',
    'clientSecret', 'YOUR_GITHUB_CLIENT_SECRET',
    'authURL', 'https://github.com/login/oauth/authorize',
    'tokenURL', 'https://github.com/login/oauth/access_token',
    'userInfoURL', 'https://api.github.com/user',
    'scopes', JSON_ARRAY('read:user', 'user:email'),
    'redirectURL', 'http://localhost:8080/api/v1/auth/oauth/callback'
  ),
  'GitHub OAuth',
  1,
  0
) ON DUPLICATE KEY UPDATE `name` = `name`;

-- OAuth (GitLab)
INSERT INTO `t_sso_provider` (`provider_id`, `name`, `provider_type`, `config`, `description`, `priority`, `is_enabled`)
VALUES (
  'oauth_gitlab',
  'GitLab',
  'oauth',
  JSON_OBJECT(
    'clientId', 'YOUR_GITLAB_CLIENT_ID',
    'clientSecret', 'YOUR_GITLAB_CLIENT_SECRET',
    'authURL', 'https://gitlab.com/oauth/authorize',
    'tokenURL', 'https://gitlab.com/oauth/token',
    'userInfoURL', 'https://gitlab.com/api/v4/user',
    'scopes', JSON_ARRAY('read_user'),
    'redirectURL', 'http://localhost:8080/api/v1/auth/oauth/callback'
  ),
  'GitLab OAuth',
  2,
  0
) ON DUPLICATE KEY UPDATE `name` = `name`;

-- OAuth (Google)
INSERT INTO `t_sso_provider` (`provider_id`, `name`, `provider_type`, `config`, `description`, `priority`, `is_enabled`)
VALUES (
  'oauth_google',
  'Google',
  'oauth',
  JSON_OBJECT(
    'clientId', 'YOUR_GOOGLE_CLIENT_ID',
    'clientSecret', 'YOUR_GOOGLE_CLIENT_SECRET',
    'authURL', 'https://accounts.google.com/o/oauth2/v2/auth',
    'tokenURL', 'https://oauth2.googleapis.com/token',
    'userInfoURL', 'https://www.googleapis.com/oauth2/v2/userinfo',
    'scopes', JSON_ARRAY('openid', 'profile', 'email'),
    'redirectURL', 'http://localhost:8080/api/v1/auth/oauth/callback'
  ),
  'Google OAuth',
  5,
  0
) ON DUPLICATE KEY UPDATE `name` = `name`;

-- OAuth (Slack)
INSERT INTO `t_sso_provider` (`provider_id`, `name`, `provider_type`, `config`, `description`, `priority`, `is_enabled`)
VALUES (
  'oauth_slack',
  'Slack',
  'oauth',
  JSON_OBJECT(
    'clientId', 'YOUR_SLACK_CLIENT_ID',
    'clientSecret', 'YOUR_SLACK_CLIENT_SECRET',
    'authURL', 'https://slack.com/oauth/v2/authorize',
    'tokenURL', 'https://slack.com/api/oauth.v2.access',
    'userInfoURL', 'https://slack.com/api/users.identity',
    'scopes', JSON_ARRAY('identity.basic', 'identity.email', 'identity.avatar'),
    'redirectURL', 'http://localhost:8080/api/v1/auth/oauth/callback'
  ),
  'Slack OAuth',
  6,
  0
) ON DUPLICATE KEY UPDATE `name` = `name`;

-- LDAP
INSERT INTO `t_sso_provider` (`provider_id`, `name`, `provider_type`, `config`, `description`, `priority`, `is_enabled`)
VALUES (
  'ldap_default',
  'LDAP',
  'ldap',
  JSON_OBJECT(
    'host', 'ldap.example.com',
    'port', 389,
    'useTLS', false,
    'skipVerify', false,
    'baseDN', 'dc=example,dc=com',
    'bindDN', 'cn=admin,dc=example,dc=com',
    'bindPassword', 'YOUR_BIND_PASSWORD',
    'userFilter', '(uid=%s)',
    'userDN', 'ou=users,dc=example,dc=com',
    'groupFilter', '(memberUid=%s)',
    'groupDN', 'ou=groups,dc=example,dc=com',
    'attributes', JSON_OBJECT(
      'username', 'uid',
      'email', 'mail',
      'displayName', 'cn',
      'groups', 'memberOf'
    )
  ),
  'LDAP',
  10,
  0
) ON DUPLICATE KEY UPDATE `name` = `name`;

-- OIDC (OpenID Connect)
INSERT INTO `t_sso_provider` (`provider_id`, `name`, `provider_type`, `config`, `description`, `priority`, `is_enabled`)
VALUES (
  'oidc_keycloak',
  'Keycloak',
  'oidc',
  JSON_OBJECT(
    'issuer', 'https://keycloak.example.com/realms/arcade',
    'clientId', 'YOUR_CLIENT_ID',
    'clientSecret', 'YOUR_CLIENT_SECRET',
    'redirectURL', 'http://localhost:8080/api/v1/auth/oidc/callback',
    'scopes', JSON_ARRAY('openid', 'profile', 'email'),
    'userInfoURL', 'https://keycloak.example.com/realms/arcade/protocol/openid-connect/userinfo',
    'skipVerify', false
  ),
  'Keycloak OIDC',
  3,
  0
) ON DUPLICATE KEY UPDATE `name` = `name`;

-- OIDC (Google Workspace)
INSERT INTO `t_sso_provider` (`provider_id`, `name`, `provider_type`, `config`, `description`, `priority`, `is_enabled`)
VALUES (
  'oidc_google',
  'Google Workspace',
  'oidc',
  JSON_OBJECT(
    'issuer', 'https://accounts.google.com',
    'clientId', 'YOUR_GOOGLE_CLIENT_ID',
    'clientSecret', 'YOUR_GOOGLE_CLIENT_SECRET',
    'redirectURL', 'http://localhost:8080/api/v1/auth/oidc/callback',
    'scopes', JSON_ARRAY('openid', 'profile', 'email'),
    'hostedDomain', 'example.com'
  ),
  'Google Workspace OIDC',
  7,
  0
) ON DUPLICATE KEY UPDATE `name` = `name`;

-- 插入示例对象存储配置
-- MinIO 配置
INSERT INTO `t_storage_config` (`storage_id`, `name`, `storage_type`, `config`, `description`, `is_default`, `is_enabled`)
VALUES (
  'storage_minio_default',
  'MinIO Default',
  'minio',
  JSON_OBJECT(
    'endpoint', 'localhost:9000',
    'accessKey', 'YOUR_MINIO_ACCESS_KEY',
    'secretKey', 'YOUR_MINIO_SECRET_KEY',
    'bucket', 'arcade',
    'region', 'us-east-1',
    'useTLS', false,
    'basePath', 'artifacts'
  ),
  '默认 MinIO 对象存储',
  1,
  0
) ON DUPLICATE KEY UPDATE `name` = `name`;

-- AWS S3 配置
INSERT INTO `t_storage_config` (`storage_id`, `name`, `storage_type`, `config`, `description`, `is_default`, `is_enabled`)
VALUES (
  'storage_s3_default',
  'AWS S3',
  's3',
  JSON_OBJECT(
    'endpoint', 'https://s3.amazonaws.com',
    'accessKey', 'YOUR_AWS_ACCESS_KEY',
    'secretKey', 'YOUR_AWS_SECRET_KEY',
    'bucket', 'arcade-artifacts',
    'region', 'us-east-1',
    'useTLS', true,
    'basePath', 'ci-artifacts'
  ),
  'AWS S3 对象存储',
  0,
  0
) ON DUPLICATE KEY UPDATE `name` = `name`;

-- 阿里云 OSS 配置
INSERT INTO `t_storage_config` (`storage_id`, `name`, `storage_type`, `config`, `description`, `is_default`, `is_enabled`)
VALUES (
  'storage_oss_default',
  'Aliyun OSS',
  'oss',
  JSON_OBJECT(
    'endpoint', 'oss-cn-hangzhou.aliyuncs.com',
    'accessKey', 'YOUR_OSS_ACCESS_KEY',
    'secretKey', 'YOUR_OSS_SECRET_KEY',
    'bucket', 'arcade-artifacts',
    'region', 'cn-hangzhou',
    'useTLS', true,
    'basePath', 'build-artifacts'
  ),
  '阿里云 OSS 对象存储',
  0,
  0
) ON DUPLICATE KEY UPDATE `name` = `name`;

-- Google Cloud Storage 配置
INSERT INTO `t_storage_config` (`storage_id`, `name`, `storage_type`, `config`, `description`, `is_default`, `is_enabled`)
VALUES (
  'storage_gcs_default',
  'Google Cloud Storage',
  'gcs',
  JSON_OBJECT(
    'endpoint', 'https://storage.googleapis.com',
    'accessKey', '/path/to/service-account-key.json',
    'bucket', 'arcade-artifacts',
    'region', 'us-central1',
    'basePath', 'ci-builds'
  ),
  'Google Cloud Storage',
  0,
  0
) ON DUPLICATE KEY UPDATE `name` = `name`;

-- 腾讯云 COS 配置
INSERT INTO `t_storage_config` (`storage_id`, `name`, `storage_type`, `config`, `description`, `is_default`, `is_enabled`)
VALUES (
  'storage_cos_default',
  'Tencent COS',
  'cos',
  JSON_OBJECT(
    'endpoint', 'https://cos.ap-guangzhou.myqcloud.com',
    'accessKey', 'YOUR_COS_SECRET_ID',
    'secretKey', 'YOUR_COS_SECRET_KEY',
    'bucket', 'arcade-artifacts',
    'region', 'ap-guangzhou',
    'useTLS', true,
    'basePath', 'pipeline-artifacts'
  ),
  '腾讯云 COS 对象存储',
  0,
  0
) ON DUPLICATE KEY UPDATE `name` = `name`;

-- 插入示例插件
-- Slack 通知插件
INSERT INTO `t_plugin` (`plugin_id`, `name`, `version`, `description`, `author`, `plugin_type`, `entry_point`, `config_schema`, `params_schema`, `default_config`, `is_builtin`)
VALUES (
  'notify_slack',
  'Slack Notification',
  '1.0.0',
  'Send notifications to Slack channels',
  'Arcade Team',
  'notify',
  'plugins/notify/slack.so',
  JSON_OBJECT(
    'type', 'object',
    'properties', JSON_OBJECT(
      'webhook_url', JSON_OBJECT('type', 'string', 'description', 'Slack Webhook URL'),
      'channel', JSON_OBJECT('type', 'string', 'description', 'Target channel'),
      'username', JSON_OBJECT('type', 'string', 'description', 'Bot username', 'default', 'Arcade CI')
    ),
    'required', JSON_ARRAY('webhook_url')
  ),
  JSON_OBJECT(
    'type', 'object',
    'properties', JSON_OBJECT(
      'message', JSON_OBJECT('type', 'string', 'description', 'Custom message'),
      'mention_users', JSON_OBJECT('type', 'array', 'items', JSON_OBJECT('type', 'string'))
    )
  ),
  JSON_OBJECT(
    'username', 'Arcade CI',
    'icon_emoji', ':rocket:'
  ),
  1
) ON DUPLICATE KEY UPDATE `version` = `version`;

-- Email 通知插件
INSERT INTO `t_plugin` (`plugin_id`, `name`, `version`, `description`, `author`, `plugin_type`, `entry_point`, `config_schema`, `params_schema`, `default_config`, `is_builtin`)
VALUES (
  'notify_email',
  'Email Notification',
  '1.0.0',
  'Send email notifications',
  'Arcade Team',
  'notify',
  'plugins/notify/email.so',
  JSON_OBJECT(
    'type', 'object',
    'properties', JSON_OBJECT(
      'smtp_host', JSON_OBJECT('type', 'string', 'description', 'SMTP server host'),
      'smtp_port', JSON_OBJECT('type', 'integer', 'description', 'SMTP server port'),
      'smtp_user', JSON_OBJECT('type', 'string', 'description', 'SMTP username'),
      'smtp_password', JSON_OBJECT('type', 'string', 'description', 'SMTP password'),
      'from_address', JSON_OBJECT('type', 'string', 'description', 'From email address')
    ),
    'required', JSON_ARRAY('smtp_host', 'smtp_port', 'from_address')
  ),
  JSON_OBJECT(
    'type', 'object',
    'properties', JSON_OBJECT(
      'to', JSON_OBJECT('type', 'array', 'items', JSON_OBJECT('type', 'string')),
      'subject', JSON_OBJECT('type', 'string', 'description', 'Email subject'),
      'body', JSON_OBJECT('type', 'string', 'description', 'Email body')
    ),
    'required', JSON_ARRAY('to')
  ),
  JSON_OBJECT(
    'smtp_port', 587
  ),
  1
) ON DUPLICATE KEY UPDATE `version` = `version`;

-- Docker 构建插件
INSERT INTO `t_plugin` (`plugin_id`, `name`, `version`, `description`, `author`, `plugin_type`, `entry_point`, `config_schema`, `params_schema`, `is_builtin`)
VALUES (
  'build_docker',
  'Docker Build',
  '1.0.0',
  'Build and push Docker images',
  'Arcade Team',
  'build',
  'plugins/build/docker.so',
  JSON_OBJECT(
    'type', 'object',
    'properties', JSON_OBJECT(
      'registry', JSON_OBJECT('type', 'string', 'description', 'Docker registry URL'),
      'username', JSON_OBJECT('type', 'string', 'description', 'Registry username'),
      'password', JSON_OBJECT('type', 'string', 'description', 'Registry password')
    )
  ),
  JSON_OBJECT(
    'type', 'object',
    'properties', JSON_OBJECT(
      'dockerfile', JSON_OBJECT('type', 'string', 'description', 'Dockerfile path', 'default', 'Dockerfile'),
      'context', JSON_OBJECT('type', 'string', 'description', 'Build context', 'default', '.'),
      'tags', JSON_OBJECT('type', 'array', 'items', JSON_OBJECT('type', 'string')),
      'build_args', JSON_OBJECT('type', 'object', 'description', 'Build arguments'),
      'push', JSON_OBJECT('type', 'boolean', 'description', 'Push after build', 'default', true)
    ),
    'required', JSON_ARRAY('tags')
  ),
  1
) ON DUPLICATE KEY UPDATE `version` = `version`;

-- 钉钉通知插件
INSERT INTO `t_plugin` (`plugin_id`, `name`, `version`, `description`, `author`, `plugin_type`, `entry_point`, `config_schema`, `params_schema`, `is_builtin`)
VALUES (
  'notify_dingtalk',
  'DingTalk Notification',
  '1.0.0',
  'Send notifications to DingTalk群',
  'Arcade Team',
  'notify',
  'plugins/notify/dingtalk.so',
  JSON_OBJECT(
    'type', 'object',
    'properties', JSON_OBJECT(
      'webhook_url', JSON_OBJECT('type', 'string', 'description', 'DingTalk Webhook URL'),
      'secret', JSON_OBJECT('type', 'string', 'description', 'Webhook secret for signature')
    ),
    'required', JSON_ARRAY('webhook_url')
  ),
  JSON_OBJECT(
    'type', 'object',
    'properties', JSON_OBJECT(
      'message', JSON_OBJECT('type', 'string', 'description', 'Custom message'),
      'at_mobiles', JSON_OBJECT('type', 'array', 'items', JSON_OBJECT('type', 'string')),
      'at_all', JSON_OBJECT('type', 'boolean', 'description', '@所有人', 'default', false)
    )
  ),
  1
) ON DUPLICATE KEY UPDATE `version` = `version`;

-- Kubernetes 部署插件
INSERT INTO `t_plugin` (`plugin_id`, `name`, `version`, `description`, `author`, `plugin_type`, `entry_point`, `config_schema`, `params_schema`, `is_builtin`)
VALUES (
  'deploy_kubernetes',
  'Kubernetes Deploy',
  '1.0.0',
  'Deploy applications to Kubernetes',
  'Arcade Team',
  'deploy',
  'plugins/deploy/kubernetes.so',
  JSON_OBJECT(
    'type', 'object',
    'properties', JSON_OBJECT(
      'kubeconfig', JSON_OBJECT('type', 'string', 'description', 'Kubeconfig file path or content'),
      'context', JSON_OBJECT('type', 'string', 'description', 'Kubernetes context name')
    ),
    'required', JSON_ARRAY('kubeconfig')
  ),
  JSON_OBJECT(
    'type', 'object',
    'properties', JSON_OBJECT(
      'namespace', JSON_OBJECT('type', 'string', 'description', 'Target namespace', 'default', 'default'),
      'manifests', JSON_OBJECT('type', 'array', 'items', JSON_OBJECT('type', 'string'), 'description', 'YAML manifest files'),
      'wait', JSON_OBJECT('type', 'boolean', 'description', 'Wait for deployment', 'default', true),
      'timeout', JSON_OBJECT('type', 'integer', 'description', 'Timeout in seconds', 'default', 300)
    ),
    'required', JSON_ARRAY('manifests')
  ),
  1
) ON DUPLICATE KEY UPDATE `version` = `version`;

-- 插入插件权限配置
-- Slack 插件权限
INSERT INTO `t_plugin_permission` (`plugin_id`, `permission_type`, `resource`, `action`, `is_allowed`)
VALUES
  ('notify_slack', 'network', 'https://hooks.slack.com', 'write', 1),
  ('notify_slack', 'file_system', '/tmp/slack_cache', 'write', 1)
ON DUPLICATE KEY UPDATE `is_allowed` = `is_allowed`;

-- Email 插件权限
INSERT INTO `t_plugin_permission` (`plugin_id`, `permission_type`, `resource`, `action`, `is_allowed`)
VALUES
  ('notify_email', 'network', '*:587', 'write', 1),
  ('notify_email', 'network', '*:465', 'write', 1)
ON DUPLICATE KEY UPDATE `is_allowed` = `is_allowed`;

-- Docker 插件权限
INSERT INTO `t_plugin_permission` (`plugin_id`, `permission_type`, `resource`, `action`, `is_allowed`)
VALUES
  ('build_docker', 'process', '/usr/bin/docker', 'execute', 1),
  ('build_docker', 'network', '*', 'write', 1),
  ('build_docker', 'file_system', '/var/run/docker.sock', 'read', 1),
  ('build_docker', 'file_system', '/tmp/docker_build', 'write', 1)
ON DUPLICATE KEY UPDATE `is_allowed` = `is_allowed`;

-- DingTalk 插件权限
INSERT INTO `t_plugin_permission` (`plugin_id`, `permission_type`, `resource`, `action`, `is_allowed`)
VALUES
  ('notify_dingtalk', 'network', 'https://oapi.dingtalk.com', 'write', 1)
ON DUPLICATE KEY UPDATE `is_allowed` = `is_allowed`;

-- Kubernetes 插件权限
INSERT INTO `t_plugin_permission` (`plugin_id`, `permission_type`, `resource`, `action`, `is_allowed`)
VALUES
  ('deploy_kubernetes', 'process', '/usr/bin/kubectl', 'execute', 1),
  ('deploy_kubernetes', 'network', '*:443', 'write', 1),
  ('deploy_kubernetes', 'file_system', '/tmp/k8s_manifests', 'read', 1)
ON DUPLICATE KEY UPDATE `is_allowed` = `is_allowed`;

-- 插入插件资源配额
INSERT INTO `t_plugin_quota` (`plugin_id`, `max_cpu_percent`, `max_memory_mb`, `max_execution_time`, `max_network_mbps`)
VALUES
  ('notify_slack', 10, 64, 30, 1),
  ('notify_email', 10, 64, 60, 1),
  ('build_docker', 80, 2048, 1800, 50),
  ('notify_dingtalk', 10, 64, 30, 1),
  ('deploy_kubernetes', 50, 512, 600, 10)
ON DUPLICATE KEY UPDATE `max_cpu_percent` = `max_cpu_percent`;

-- 插入官方插件来源
INSERT INTO `t_plugin_source` (`source_id`, `name`, `source_type`, `repository`, `is_trusted`)
VALUES 
  ('source_arcade_official', 'Arcade Official', 'official', 'https://github.com/observabil/arcade-plugins', 1),
  ('source_community', 'Community', 'community', 'https://plugins.arcade.io', 0)
ON DUPLICATE KEY UPDATE `name` = `name`;

-- 插入示例 Agent 配置（Agent注册时自动创建）
-- 每个 Agent 一条记录，所有配置在 config_items JSON 中
INSERT INTO `t_agent_config` (`agent_id`, `config_items`, `description`)
VALUES (
  'agent_001',
  JSON_OBJECT(
    'heartbeat_interval', 30,
    'max_concurrent_jobs', 5,
    'job_timeout', 3600,
    'workspace_dir', '/var/lib/arcade/workspace',
    'temp_dir', '/var/lib/arcade/temp',
    'log_level', 'INFO',
    'enable_docker', true,
    'docker_network', 'bridge',
    'resource_limits', JSON_OBJECT('cpu', '2', 'memory', '4G'),
    'allowed_commands', JSON_ARRAY('docker', 'kubectl', 'npm', 'yarn', 'go', 'python'),
    'env_vars', JSON_OBJECT('PATH', '/usr/local/bin:/usr/bin:/bin'),
    'cache_dir', '/var/lib/arcade/cache',
    'cleanup_policy', JSON_OBJECT('max_age_days', 7, 'max_size_gb', 50)
  ),
  'Agent 001 默认配置'
) ON DUPLICATE KEY UPDATE `config_items` = `config_items`;
