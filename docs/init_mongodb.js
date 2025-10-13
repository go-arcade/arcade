// ================================================================
// Arcade CI/CD 平台 MongoDB 初始化脚本
// 数据库: job_log
// 用途: 创建索引和配置 TTL
// ================================================================

// 使用数据库
db = db.getSiblingDB('job_log');

print('开始初始化 MongoDB 数据库: job_log');
print('===================================');

// ================================================================
// Collection 1: job_logs (任务执行日志)
// ================================================================
print('\n初始化 Collection: job_logs');

// 创建索引
db.job_logs.createIndex(
  { job_id: 1, timestamp: 1 },
  { name: 'idx_job_id_timestamp', background: true }
);
print('✓ 创建索引: idx_job_id_timestamp');

db.job_logs.createIndex(
  { agent_id: 1, timestamp: -1 },
  { name: 'idx_agent_id_timestamp', background: true }
);
print('✓ 创建索引: idx_agent_id_timestamp');

db.job_logs.createIndex(
  { pipeline_run_id: 1, timestamp: 1 },
  { name: 'idx_pipeline_run_id_timestamp', background: true }
);
print('✓ 创建索引: idx_pipeline_run_id_timestamp');

// TTL 索引 - 90天后自动删除
db.job_logs.createIndex(
  { timestamp: 1 },
  { 
    name: 'idx_timestamp_ttl',
    expireAfterSeconds: 7776000, // 90天 = 90 * 24 * 60 * 60
    background: true 
  }
);
print('✓ 创建 TTL 索引: idx_timestamp_ttl (90天过期)');

// ================================================================
// Collection 2: terminal_logs (终端日志/构建日志)
// ================================================================
print('\n初始化 Collection: terminal_logs');

// 创建索引
db.terminal_logs.createIndex(
  { session_id: 1 },
  { name: 'idx_session_id', unique: true, background: true }
);
print('✓ 创建唯一索引: idx_session_id');

db.terminal_logs.createIndex(
  { environment: 1, timestamp: -1 },
  { name: 'idx_environment_timestamp', background: true }
);
print('✓ 创建索引: idx_environment_timestamp');

db.terminal_logs.createIndex(
  { job_id: 1, timestamp: -1 },
  { name: 'idx_job_id_timestamp', background: true, sparse: true }
);
print('✓ 创建索引: idx_job_id_timestamp');

db.terminal_logs.createIndex(
  { pipeline_run_id: 1, timestamp: -1 },
  { name: 'idx_pipeline_run_id_timestamp', background: true, sparse: true }
);
print('✓ 创建索引: idx_pipeline_run_id_timestamp');

db.terminal_logs.createIndex(
  { user_id: 1, created_at: -1 },
  { name: 'idx_user_id_created_at', background: true }
);
print('✓ 创建索引: idx_user_id_created_at');

db.terminal_logs.createIndex(
  { status: 1, created_at: -1 },
  { name: 'idx_status_created_at', background: true }
);
print('✓ 创建索引: idx_status_created_at');

// TTL 索引 - 180天后自动删除
db.terminal_logs.createIndex(
  { created_at: 1 },
  { 
    name: 'idx_created_at_ttl',
    expireAfterSeconds: 15552000, // 180天 = 180 * 24 * 60 * 60
    background: true 
  }
);
print('✓ 创建 TTL 索引: idx_created_at_ttl (180天过期)');

// ================================================================
// Collection 3: build_artifacts_logs (产物构建日志)
// ================================================================
print('\n初始化 Collection: build_artifacts_logs');

// 创建索引
db.build_artifacts_logs.createIndex(
  { artifact_id: 1, timestamp: -1 },
  { name: 'idx_artifact_id_timestamp', background: true }
);
print('✓ 创建索引: idx_artifact_id_timestamp');

db.build_artifacts_logs.createIndex(
  { job_id: 1, timestamp: -1 },
  { name: 'idx_job_id_timestamp', background: true }
);
print('✓ 创建索引: idx_job_id_timestamp');

db.build_artifacts_logs.createIndex(
  { operation: 1, timestamp: -1 },
  { name: 'idx_operation_timestamp', background: true }
);
print('✓ 创建索引: idx_operation_timestamp');

db.build_artifacts_logs.createIndex(
  { user_id: 1, timestamp: -1 },
  { name: 'idx_user_id_timestamp', background: true }
);
print('✓ 创建索引: idx_user_id_timestamp');

// TTL 索引 - 90天后自动删除
db.build_artifacts_logs.createIndex(
  { timestamp: 1 },
  { 
    name: 'idx_timestamp_ttl',
    expireAfterSeconds: 7776000, // 90天 = 90 * 24 * 60 * 60
    background: true 
  }
);
print('✓ 创建 TTL 索引: idx_timestamp_ttl (90天过期)');

// ================================================================
// 列出所有集合和索引
// ================================================================
print('\n===================================');
print('数据库初始化完成！');
print('\n当前数据库中的集合:');
db.getCollectionNames().forEach(function(collName) {
  print('  - ' + collName);
  var indexes = db.getCollection(collName).getIndexes();
  print('    索引数量: ' + indexes.length);
  indexes.forEach(function(index) {
    print('      * ' + index.name);
  });
});

print('\n数据库统计信息:');
printjson(db.stats());

print('\n提示: TTL 索引将自动删除过期数据');
print('  - job_logs: 90天后自动删除');
print('  - terminal_logs: 180天后自动删除');
print('  - build_artifacts_logs: 90天后自动删除');

