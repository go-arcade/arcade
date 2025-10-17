-- ==========================================
-- 插件系统重构 - 数据库迁移脚本
-- ==========================================
-- 版本: v1.0.0
-- 日期: 2025-01-16
-- 描述: 为插件系统添加新字段以支持完整的生命周期管理
-- ==========================================

-- 1. 添加新字段到 t_plugin 表
ALTER TABLE t_plugin 
ADD COLUMN source VARCHAR(20) DEFAULT 'local' COMMENT '插件来源: local/market' AFTER checksum,
ADD COLUMN s3_path VARCHAR(500) DEFAULT '' COMMENT 'S3存储路径' AFTER source,
ADD COLUMN manifest JSON DEFAULT NULL COMMENT '插件清单' AFTER s3_path,
ADD COLUMN install_time DATETIME DEFAULT NULL COMMENT '安装时间' AFTER manifest,
ADD COLUMN update_time DATETIME DEFAULT NULL COMMENT '更新时间' AFTER install_time;

-- 2. 更新 is_enabled 字段注释（支持错误状态）
ALTER TABLE t_plugin MODIFY COLUMN is_enabled INT DEFAULT 1 COMMENT '状态: 0-禁用 1-启用 2-错误';

-- 2.1 移除 install_path 字段（本地路径改为动态生成，不再存储在数据库）
-- 说明：本地路径格式为 {localCacheDir}/{plugin_id}_{version}.so
-- 如果确认不再需要 install_path 字段，可以执行以下语句：
-- ALTER TABLE t_plugin DROP COLUMN install_path;
-- 注意：建议先备份数据，确认系统正常运行后再执行删除操作

-- 3. 更新现有数据（如果有旧数据需要迁移）
-- 为所有插件设置来源（所有插件都是下载安装的，没有内置插件）
UPDATE t_plugin SET source = 'local' WHERE (source IS NULL OR source = '');

-- 设置安装时间（使用创建时间）
UPDATE t_plugin SET install_time = created_time WHERE install_time IS NULL;

-- 注意：旧的 install_path 字段如果存在，可以保留用于迁移参考
-- 迁移完成后再删除该字段

-- 4. 添加索引以提高查询性能
CREATE INDEX idx_plugin_source ON t_plugin(source);
CREATE INDEX idx_plugin_status ON t_plugin(is_enabled);
CREATE INDEX idx_plugin_install_time ON t_plugin(install_time);

-- 5. 验证迁移
-- 检查新字段是否添加成功
SELECT 
    COUNT(*) as total_plugins,
    SUM(CASE WHEN source = 'local' THEN 1 ELSE 0 END) as local_plugins,
    SUM(CASE WHEN source = 'market' THEN 1 ELSE 0 END) as market_plugins,
    SUM(CASE WHEN is_enabled = 1 THEN 1 ELSE 0 END) as enabled_plugins,
    SUM(CASE WHEN is_enabled = 0 THEN 1 ELSE 0 END) as disabled_plugins
FROM t_plugin;

-- 6. 创建插件市场相关表（可选，为未来功能预留）
CREATE TABLE IF NOT EXISTS t_plugin_market (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    market_id VARCHAR(100) NOT NULL UNIQUE COMMENT '市场插件ID',
    name VARCHAR(100) NOT NULL COMMENT '插件名称',
    version VARCHAR(20) NOT NULL COMMENT '版本号',
    description TEXT COMMENT '描述',
    author VARCHAR(100) COMMENT '作者',
    plugin_type VARCHAR(50) COMMENT '插件类型',
    icon VARCHAR(500) COMMENT '图标URL',
    repository VARCHAR(500) COMMENT '仓库地址',
    homepage VARCHAR(500) COMMENT '主页',
    download_url VARCHAR(500) COMMENT '下载地址',
    download_count INT DEFAULT 0 COMMENT '下载次数',
    rating DECIMAL(3,2) DEFAULT 0.00 COMMENT '评分(0-5)',
    manifest JSON COMMENT '插件清单',
    checksum VARCHAR(64) COMMENT 'SHA256校验和',
    file_size BIGINT COMMENT '文件大小(字节)',
    is_verified TINYINT DEFAULT 0 COMMENT '是否官方验证',
    is_active TINYINT DEFAULT 1 COMMENT '是否激活',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_market_name (name),
    INDEX idx_market_type (plugin_type),
    INDEX idx_market_rating (rating),
    INDEX idx_market_downloads (download_count)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='插件市场表';

-- 7. 创建插件安装历史表（可选，用于审计）
CREATE TABLE IF NOT EXISTS t_plugin_install_history (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    plugin_id VARCHAR(100) NOT NULL COMMENT '插件ID',
    plugin_name VARCHAR(100) NOT NULL COMMENT '插件名称',
    version VARCHAR(20) NOT NULL COMMENT '版本号',
    action VARCHAR(20) NOT NULL COMMENT '操作: install/uninstall/enable/disable/update',
    source VARCHAR(20) COMMENT '来源: local/market/builtin',
    operator VARCHAR(100) COMMENT '操作人',
    status VARCHAR(20) COMMENT '状态: success/failed',
    error_message TEXT COMMENT '错误信息',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_plugin_history_id (plugin_id),
    INDEX idx_plugin_history_action (action),
    INDEX idx_plugin_history_time (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='插件安装历史表';

-- 8. 创建插件下载统计表（可选，用于监控）
CREATE TABLE IF NOT EXISTS t_plugin_download_stats (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    plugin_id VARCHAR(100) NOT NULL COMMENT '插件ID',
    agent_id VARCHAR(100) COMMENT 'Agent ID',
    download_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '下载时间',
    file_size BIGINT COMMENT '文件大小',
    duration_ms INT COMMENT '下载耗时(毫秒)',
    success TINYINT DEFAULT 1 COMMENT '是否成功',
    error_message TEXT COMMENT '错误信息',
    INDEX idx_download_plugin (plugin_id),
    INDEX idx_download_agent (agent_id),
    INDEX idx_download_time (download_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='插件下载统计表';

-- ==========================================
-- 回滚脚本（如需回滚，请谨慎使用）
-- ==========================================

-- ROLLBACK: 删除新增的字段
-- ALTER TABLE t_plugin 
-- DROP COLUMN source,
-- DROP COLUMN s3_path,
-- DROP COLUMN manifest,
-- DROP COLUMN install_time,
-- DROP COLUMN update_time;

-- ROLLBACK: 删除索引
-- DROP INDEX idx_plugin_source ON t_plugin;
-- DROP INDEX idx_plugin_status ON t_plugin;
-- DROP INDEX idx_plugin_type ON t_plugin;
-- DROP INDEX idx_plugin_install_time ON t_plugin;

-- ROLLBACK: 删除新表（如果创建了）
-- DROP TABLE IF EXISTS t_plugin_market;
-- DROP TABLE IF EXISTS t_plugin_install_history;
-- DROP TABLE IF EXISTS t_plugin_download_stats;

-- ==========================================
-- 迁移完成
-- ==========================================

