-- ============================================================================
-- 数据库迁移脚本：task_logs → step_run_logs (ClickHouse)
-- 说明：将日志表从 task_logs 迁移为 step_run_logs
-- 执行前请备份数据库！
-- ============================================================================

-- ============================================================================
-- ClickHouse 数据库迁移
-- ============================================================================

-- 1. 创建新表 step_run_logs
-- ----------------------------------------------------------------------------
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

-- 2. 迁移旧数据（如果 task_logs 表存在）
-- ----------------------------------------------------------------------------
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

-- 3. 验证数据迁移
-- ----------------------------------------------------------------------------
-- 检查新表记录数
SELECT COUNT(*) as new_table_count FROM step_run_logs;

-- 检查旧表记录数（如果存在）
-- SELECT COUNT(*) as old_table_count FROM task_logs;

-- 4. 删除旧表（确认数据迁移成功后执行）
-- ----------------------------------------------------------------------------
-- 警告：执行前请确认数据已成功迁移！
-- DROP TABLE IF EXISTS task_logs;

-- ============================================================================
-- 回滚脚本（如果需要）
-- ============================================================================

-- 如果需要回滚，可以执行以下操作：
-- 1. 重新创建旧表
-- CREATE TABLE IF NOT EXISTS task_logs (
--     task_id String,
--     timestamp Int64,
--     line_number Int32,
--     level String,
--     content String,
--     stream String,
--     plugin_name String,
--     agent_id String
-- ) ENGINE = MergeTree()
-- ORDER BY (task_id, line_number, timestamp)
-- PRIMARY KEY (task_id, line_number)
-- SETTINGS index_granularity = 8192;
--
-- 2. 从新表迁移数据回旧表
-- INSERT INTO task_logs 
-- SELECT 
--     step_run_id as task_id,
--     timestamp,
--     line_number,
--     level,
--     content,
--     stream,
--     plugin_name,
--     agent_id
-- FROM step_run_logs;
--
-- 3. 删除新表
-- DROP TABLE IF EXISTS step_run_logs;

