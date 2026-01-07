-- ============================================================================
-- 数据库迁移脚本：Task → StepRun
-- 说明：将 Task 相关表结构迁移为 StepRun
-- 执行前请备份数据库！
-- ============================================================================

-- ============================================================================
-- ClickHouse 数据库迁移
-- ============================================================================

-- 1. 日志表：task_logs → step_run_logs
-- ----------------------------------------------------------------------------

-- 创建新表
CREATE TABLE IF NOT EXISTS step_run_logs (
    step_run_id String,
    timestamp Int64,
    line_number Int32,
    level String,
    content String,
    stream String,
    plugin_name String,
    agent_id String
) ENGINE = MergeTree()
ORDER BY (step_run_id, line_number, timestamp)
PRIMARY KEY (step_run_id, line_number)
SETTINGS index_granularity = 8192;

-- 迁移旧数据（如果存在）
-- 注意：如果旧表不存在或为空，此步骤会报错，可以忽略
INSERT INTO step_run_logs 
SELECT 
    task_id as step_run_id,
    timestamp,
    line_number,
    level,
    content,
    stream,
    plugin_name,
    agent_id
FROM task_logs;

-- 删除旧表（确认数据迁移成功后执行）
-- DROP TABLE IF EXISTS task_logs;

-- ============================================================================
-- 2. 任务队列记录表：l_task_queue_records → l_step_run_queue_records
-- ----------------------------------------------------------------------------

-- 创建新表
CREATE TABLE IF NOT EXISTS l_step_run_queue_records (
    step_run_id String,
    step_run_type String,
    status String,
    queue String,
    priority Int32,
    pipeline_id String,
    pipeline_run_id String,
    stage_id String,
    job_id String,
    job_run_id String,
    agent_id String,
    payload String,
    create_time DateTime,
    start_time Nullable(DateTime),
    end_time Nullable(DateTime),
    duration Nullable(Int64),
    retry_count Int32,
    current_retry Int32,
    error_message Nullable(String)
) ENGINE = MergeTree()
ORDER BY (step_run_id, create_time)
PRIMARY KEY step_run_id
SETTINGS index_granularity = 8192;

-- 迁移旧数据（如果存在）
-- 注意：task_type 映射为 step_run_type，job_id 和 job_run_id 需要根据实际情况填充
INSERT INTO l_step_run_queue_records 
SELECT 
    task_id as step_run_id,
    task_type as step_run_type,
    status,
    queue,
    priority,
    pipeline_id,
    pipeline_run_id,
    stage_id,
    '' as job_id,  -- 需要根据实际情况填充
    '' as job_run_id,  -- 需要根据实际情况填充
    agent_id,
    payload,
    create_time,
    start_time,
    end_time,
    duration,
    retry_count,
    current_retry,
    error_message
FROM l_task_queue_records;

-- 删除旧表（确认数据迁移成功后执行）
-- DROP TABLE IF EXISTS l_task_queue_records;

-- ============================================================================
-- MySQL 数据库迁移
-- ============================================================================

-- 3. 任务表：t_task → t_step_run
-- ----------------------------------------------------------------------------

-- 创建新表
CREATE TABLE IF NOT EXISTS t_step_run (
    id INT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    step_run_id VARCHAR(64) NOT NULL COMMENT '步骤执行唯一标识',
    name VARCHAR(255) NOT NULL COMMENT '步骤名称',
    pipeline_id VARCHAR(64) DEFAULT NULL COMMENT '所属流水线ID',
    pipeline_run_id VARCHAR(64) DEFAULT NULL COMMENT '所属流水线执行ID',
    stage_id VARCHAR(64) DEFAULT NULL COMMENT '所属阶段ID',
    job_id VARCHAR(64) DEFAULT NULL COMMENT '所属作业ID',
    job_run_id VARCHAR(64) DEFAULT NULL COMMENT '所属作业执行ID',
    step_index INT NOT NULL DEFAULT 0 COMMENT '步骤序号',
    agent_id VARCHAR(64) DEFAULT NULL COMMENT '执行的Agent ID',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '步骤执行状态: 0-未知, 1-等待, 2-入队, 3-运行中, 4-成功, 5-失败, 6-已取消, 7-超时, 8-已跳过',
    priority INT NOT NULL DEFAULT 5 COMMENT '优先级: 1-最高, 5-普通, 10-最低',
    uses VARCHAR(255) DEFAULT NULL COMMENT '插件标识',
    action VARCHAR(255) DEFAULT NULL COMMENT '插件动作',
    args TEXT DEFAULT NULL COMMENT '插件参数(JSON格式)',
    workspace VARCHAR(512) DEFAULT NULL COMMENT '工作目录',
    env JSON DEFAULT NULL COMMENT '环境变量(JSON格式)',
    secrets JSON DEFAULT NULL COMMENT '密钥信息(JSON格式)',
    timeout INT NOT NULL DEFAULT 3600 COMMENT '超时时间(秒)',
    retry_count INT NOT NULL DEFAULT 0 COMMENT '重试次数',
    current_retry INT NOT NULL DEFAULT 0 COMMENT '当前重试次数',
    allow_failure TINYINT NOT NULL DEFAULT 0 COMMENT '是否允许失败: 0-否, 1-是',
    continue_on_error TINYINT NOT NULL DEFAULT 0 COMMENT '错误时继续: 0-否, 1-是',
    when_condition TEXT DEFAULT NULL COMMENT '条件表达式',
    label_selector JSON DEFAULT NULL COMMENT '标签选择器(JSON格式)',
    depends_on VARCHAR(512) DEFAULT NULL COMMENT '依赖的步骤执行ID列表(逗号分隔)',
    exit_code INT DEFAULT NULL COMMENT '退出码',
    error_message TEXT DEFAULT NULL COMMENT '错误信息',
    start_time DATETIME DEFAULT NULL COMMENT '开始时间',
    end_time DATETIME DEFAULT NULL COMMENT '结束时间',
    duration BIGINT DEFAULT NULL COMMENT '执行时长(毫秒)',
    created_by VARCHAR(64) DEFAULT NULL COMMENT '创建者用户ID',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at DATETIME DEFAULT NULL COMMENT '删除时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_step_run_id (step_run_id),
    KEY idx_pipeline_run_id (pipeline_run_id),
    KEY idx_job_run_id (job_run_id),
    KEY idx_agent_id (agent_id),
    KEY idx_status (status),
    KEY idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='步骤执行表';

-- 迁移旧数据（如果存在）
-- 注意：job_id 和 job_run_id 需要根据实际情况填充，step_index 可以从 stage 字段映射
INSERT INTO t_step_run (
    step_run_id,
    name,
    pipeline_id,
    pipeline_run_id,
    stage_id,
    job_id,
    job_run_id,
    step_index,
    agent_id,
    status,
    priority,
    workspace,
    env,
    secrets,
    timeout,
    retry_count,
    current_retry,
    allow_failure,
    label_selector,
    depends_on,
    exit_code,
    error_message,
    start_time,
    end_time,
    duration,
    created_by,
    created_at,
    updated_at,
    deleted_at
)
SELECT 
    task_id as step_run_id,
    name,
    pipeline_id,
    pipeline_run_id,
    stage_id,
    '' as job_id,  -- 需要根据实际情况填充
    '' as job_run_id,  -- 需要根据实际情况填充
    stage as step_index,  -- 临时映射，需要根据实际情况调整
    agent_id,
    status,
    priority,
    workspace,
    env,
    secrets,
    timeout,
    retry_count,
    current_retry,
    allow_failure,
    label_selector,
    depends_on,
    exit_code,
    error_message,
    start_time,
    end_time,
    duration,
    created_by,
    created_at,
    updated_at,
    deleted_at
FROM t_task;

-- 删除旧表（确认数据迁移成功后执行）
-- DROP TABLE IF EXISTS t_task;

-- ============================================================================
-- 4. 任务产物表：t_task_artifact → t_step_run_artifact
-- ----------------------------------------------------------------------------

-- 创建新表
CREATE TABLE IF NOT EXISTS t_step_run_artifact (
    id INT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    artifact_id VARCHAR(64) NOT NULL COMMENT '产物唯一标识',
    step_run_id VARCHAR(64) NOT NULL COMMENT '步骤执行ID',
    job_run_id VARCHAR(64) DEFAULT NULL COMMENT '作业执行ID',
    pipeline_run_id VARCHAR(64) DEFAULT NULL COMMENT '流水线执行ID',
    name VARCHAR(255) NOT NULL COMMENT '产物名称',
    path VARCHAR(512) NOT NULL COMMENT '产物路径',
    destination VARCHAR(512) DEFAULT NULL COMMENT '目标路径',
    size BIGINT NOT NULL DEFAULT 0 COMMENT '文件大小(字节)',
    storage_type VARCHAR(32) DEFAULT NULL COMMENT '存储类型: minio/s3/oss/gcs/cos',
    storage_path VARCHAR(512) DEFAULT NULL COMMENT '存储路径',
    expire TINYINT NOT NULL DEFAULT 0 COMMENT '是否过期: 0-否, 1-是',
    expire_days INT DEFAULT NULL COMMENT '过期天数',
    expired_at DATETIME DEFAULT NULL COMMENT '过期时间',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at DATETIME DEFAULT NULL COMMENT '删除时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_artifact_id (artifact_id),
    KEY idx_step_run_id (step_run_id),
    KEY idx_job_run_id (job_run_id),
    KEY idx_pipeline_run_id (pipeline_run_id),
    KEY idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='步骤执行产物表';

-- 迁移旧数据（如果存在）
INSERT INTO t_step_run_artifact (
    artifact_id,
    step_run_id,
    job_run_id,
    pipeline_run_id,
    name,
    path,
    destination,
    size,
    storage_type,
    storage_path,
    expire,
    expire_days,
    expired_at,
    created_at,
    updated_at,
    deleted_at
)
SELECT 
    artifact_id,
    task_id as step_run_id,
    '' as job_run_id,  -- 需要根据实际情况填充
    pipeline_run_id,
    name,
    path,
    destination,
    size,
    storage_type,
    storage_path,
    expire,
    expire_days,
    expired_at,
    created_at,
    updated_at,
    deleted_at
FROM t_task_artifact;

-- 删除旧表（确认数据迁移成功后执行）
-- DROP TABLE IF EXISTS t_task_artifact;

-- ============================================================================
-- 验证查询
-- ============================================================================

-- ClickHouse 验证
-- SELECT COUNT(*) FROM step_run_logs;
-- SELECT COUNT(*) FROM l_step_run_queue_records;

-- MySQL 验证
-- SELECT COUNT(*) FROM t_step_run;
-- SELECT COUNT(*) FROM t_step_run_artifact;

-- ============================================================================
-- 回滚脚本（如果需要）
-- ============================================================================

-- 如果需要回滚，可以执行以下操作：
-- 1. 恢复旧表名和字段名
-- 2. 从新表迁移数据回旧表
-- 3. 删除新表

-- 注意：回滚操作需要根据实际情况调整，这里仅提供参考

