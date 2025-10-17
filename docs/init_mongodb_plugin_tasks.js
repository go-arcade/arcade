// ================================================================
// Arcade 插件管理系统 MongoDB 初始化脚本
// Collection: plugin_install_tasks
// 用途: 创建索引和配置
// ================================================================

// 使用数据库
db = db.getSiblingDB('arcade');

print('开始初始化 MongoDB Collection: plugin_install_tasks');
print('===================================');

// ================================================================
// Collection: plugin_install_tasks (插件安装任务)
// ================================================================
print('\n初始化 Collection: plugin_install_tasks');

// 1. 任务ID唯一索引（最重要的索引）
db.plugin_install_tasks.createIndex(
  { task_id: 1 },
  { 
    name: 'idx_task_id',
    unique: true,
    background: true 
  }
);
print('✓ 创建唯一索引: idx_task_id');

// 2. 状态索引（用于查询特定状态的任务）
db.plugin_install_tasks.createIndex(
  { status: 1 },
  { 
    name: 'idx_status',
    background: true 
  }
);
print('✓ 创建索引: idx_status');

// 3. 插件名称索引（用于查询特定插件的安装历史）
db.plugin_install_tasks.createIndex(
  { plugin_name: 1 },
  { 
    name: 'idx_plugin_name',
    background: true 
  }
);
print('✓ 创建索引: idx_plugin_name');

// 4. 创建时间倒序索引（用于列表查询，最新的在前）
db.plugin_install_tasks.createIndex(
  { create_time: -1 },
  { 
    name: 'idx_create_time_desc',
    background: true 
  }
);
print('✓ 创建索引: idx_create_time_desc');

// 5. 完成时间倒序索引（用于清理旧任务）
db.plugin_install_tasks.createIndex(
  { completed_time: -1 },
  { 
    name: 'idx_completed_time_desc',
    background: true,
    sparse: true  // 稀疏索引，因为completed_time可能为空
  }
);
print('✓ 创建稀疏索引: idx_completed_time_desc');

// 6. 复合索引：状态+创建时间（用于按状态查询并排序）
db.plugin_install_tasks.createIndex(
  { status: 1, create_time: -1 },
  { 
    name: 'idx_status_create_time',
    background: true 
  }
);
print('✓ 创建复合索引: idx_status_create_time');

// 7. 复合索引：插件名称+创建时间（用于查询插件安装历史）
db.plugin_install_tasks.createIndex(
  { plugin_name: 1, create_time: -1 },
  { 
    name: 'idx_plugin_name_create_time',
    background: true 
  }
);
print('✓ 创建复合索引: idx_plugin_name_create_time');

// ================================================================
// 插入示例数据（可选，仅用于测试）
// ================================================================
print('\n是否插入示例数据? (取消下方注释以启用)');

/*
db.plugin_install_tasks.insertOne({
  task_id: "example_task_001",
  plugin_name: "stdout",
  version: "1.0.0",
  status: "completed",
  progress: 100,
  message: "install success",
  error: "",
  plugin_id: "example_plugin_001",
  source: "local",
  create_time: new Date(),
  update_time: new Date(),
  completed_time: new Date()
});
print('✓ 插入示例数据: task_id=example_task_001');
*/

// ================================================================
// 显示集合信息
// ================================================================
print('\n===================================');
print('Collection 初始化完成！');

print('\nCollection: plugin_install_tasks');
var indexes = db.plugin_install_tasks.getIndexes();
print('索引数量: ' + indexes.length);
indexes.forEach(function(index) {
  var indexInfo = '  * ' + index.name;
  if (index.unique) {
    indexInfo += ' [UNIQUE]';
  }
  if (index.sparse) {
    indexInfo += ' [SPARSE]';
  }
  print(indexInfo);
});

// 显示集合统计
print('\nCollection 统计信息:');
var stats = db.plugin_install_tasks.stats();
print('  - 文档数量: ' + stats.count);
print('  - 存储大小: ' + (stats.size / 1024).toFixed(2) + ' KB');
print('  - 索引大小: ' + (stats.totalIndexSize / 1024).toFixed(2) + ' KB');

print('\n===================================');
print('提示:');
print('  1. task_id 字段有唯一索引，确保任务ID不重复');
print('  2. completed_time 使用稀疏索引，节省存储空间');
print('  3. 复合索引优化了常见的查询模式');
print('  4. 建议定期清理已完成的旧任务（保留24小时）');
print('\n执行完成！');

