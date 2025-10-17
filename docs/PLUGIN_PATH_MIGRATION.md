# 插件路径管理迁移指南

## 变更说明

从 v1.0.0 版本开始，插件系统不再在数据库中存储 `install_path` 字段，改为使用动态路径生成策略。

## 变更原因

### 旧的设计问题

```sql
-- 旧设计：在数据库中存储路径
CREATE TABLE t_plugin (
    ...
    install_path VARCHAR(500),  -- 存储完整路径
    ...
);
```

**存在的问题：**
1. ❌ 路径硬编码，不便于迁移
2. ❌ 不同环境路径不一致
3. ❌ 路径变更需要更新数据库
4. ❌ 数据冗余，路径可以通过规则生成

### 新的设计优势

```go
// 新设计：动态生成路径
func getLocalPath(pluginID, version string) string {
    filename := fmt.Sprintf("%s_%s.so", pluginID, version)
    return filepath.Join(localCacheDir, filename)
}
```

**优势：**
1. ✅ 路径统一管理
2. ✅ 环境无关，便于迁移
3. ✅ 配置灵活，修改配置即可
4. ✅ 数据库更简洁

## 迁移步骤

### 步骤 1：备份现有数据

```bash
# 备份插件文件
cp -r /old/plugin/path /backup/plugins/

# 备份数据库
mysqldump -u root -p arcade t_plugin > t_plugin_backup.sql
```

### 步骤 2：创建新的缓存目录

```bash
# 创建标准缓存目录
sudo mkdir -p /var/lib/arcade/plugins

# 设置权限
sudo chown -R arcade:arcade /var/lib/arcade/plugins
sudo chmod 755 /var/lib/arcade/plugins
```

### 步骤 3：重组插件文件

```bash
#!/bin/bash
# migrate_plugins.sh - 插件文件迁移脚本

MYSQL_USER="root"
MYSQL_PASS="your_password"
MYSQL_DB="arcade"
NEW_CACHE_DIR="/var/lib/arcade/plugins"

# 从数据库获取插件信息
mysql -u$MYSQL_USER -p$MYSQL_PASS $MYSQL_DB -N -e \
  "SELECT plugin_id, name, version, install_path FROM t_plugin WHERE install_path IS NOT NULL" | \
while IFS=$'\t' read -r plugin_id name version old_path; do
    if [ -f "$old_path" ]; then
        # 生成新文件名
        new_filename="${plugin_id}_${version}.so"
        new_path="${NEW_CACHE_DIR}/${new_filename}"
        
        # 复制文件到新位置
        cp "$old_path" "$new_path"
        echo "Migrated: $name ($version) -> $new_path"
    else
        echo "Warning: File not found for $name ($version): $old_path"
    fi
done

echo "Migration completed!"
```

### 步骤 4：验证文件完整性

```bash
#!/bin/bash
# verify_plugins.sh - 验证插件文件完整性

MYSQL_USER="root"
MYSQL_PASS="your_password"
MYSQL_DB="arcade"
CACHE_DIR="/var/lib/arcade/plugins"

mysql -u$MYSQL_USER -p$MYSQL_PASS $MYSQL_DB -N -e \
  "SELECT plugin_id, version, checksum FROM t_plugin" | \
while IFS=$'\t' read -r plugin_id version expected_checksum; do
    file_path="${CACHE_DIR}/${plugin_id}_${version}.so"
    
    if [ ! -f "$file_path" ]; then
        echo "ERROR: File not found: $file_path"
        continue
    fi
    
    # 计算实际校验和
    actual_checksum=$(sha256sum "$file_path" | awk '{print $1}')
    
    if [ "$actual_checksum" != "$expected_checksum" ]; then
        echo "ERROR: Checksum mismatch for $plugin_id ($version)"
        echo "  Expected: $expected_checksum"
        echo "  Actual:   $actual_checksum"
    else
        echo "OK: $plugin_id ($version)"
    fi
done
```

### 步骤 5：更新配置文件

在 `conf.d/config.toml` 中添加：

```toml
[plugin]
# 本地缓存目录
local_cache_dir = "/var/lib/arcade/plugins"
```

### 步骤 6：运行数据库迁移

```bash
mysql -u root -p arcade < docs/database_migration_plugin_refactor.sql
```

### 步骤 7：重启服务并测试

```bash
# 重启服务
systemctl restart arcade

# 测试插件列表
curl http://localhost:8080/api/v1/plugins

# 检查日志
tail -f /var/log/arcade/server.log | grep -i plugin
```

### 步骤 8：删除旧字段（可选）

确认系统运行正常后，可以删除 `install_path` 字段：

```sql
-- 再次备份
CREATE TABLE t_plugin_before_drop AS SELECT * FROM t_plugin;

-- 删除 install_path 字段
ALTER TABLE t_plugin DROP COLUMN install_path;

-- 验证
DESCRIBE t_plugin;
```

## 路径映射对照表

### 旧路径格式

```
/path/to/plugins/plugin.so
/home/arcade/plugins/slack-notify.so
./plugins/notify.so
```

### 新路径格式

```
/var/lib/arcade/plugins/plugin_slack_abc123_1.0.0.so
/var/lib/arcade/plugins/plugin_k8s-deploy_xyz456_2.1.0.so
/var/lib/arcade/plugins/plugin_email-notify_def789_1.2.0.so
```

### 命名规则

```
{plugin_id}_{version}.so
```

其中：
- `plugin_id`: 插件唯一标识（如 `plugin_slack_abc123`）
- `version`: 版本号（如 `1.0.0`）

## 常见问题

### Q1: 迁移会影响现有插件吗？

A: 不会。迁移过程只是重组文件，插件功能不受影响。建议在维护窗口期执行。

### Q2: 如何处理路径冲突？

A: 新的命名规则使用 `plugin_id` + `version` 组合，确保唯一性，不会产生冲突。

### Q3: 迁移失败怎么办？

A: 已经备份了所有数据，可以恢复：
```bash
# 恢复数据库
mysql -u root -p arcade < t_plugin_backup.sql

# 恢复文件
cp -r /backup/plugins/* /old/plugin/path/
```

### Q4: 迁移后能否回退？

A: 可以，只要保留了备份：
1. 恢复数据库备份
2. 恢复旧的插件文件
3. 回滚代码版本

### Q5: 多节点部署如何迁移？

A: 
1. 所有节点共享同一个 S3 存储
2. 每个节点独立维护本地缓存
3. 按节点逐个迁移，无需停机

## 迁移检查清单

- [ ] 备份数据库
- [ ] 备份插件文件
- [ ] 创建新的缓存目录
- [ ] 设置正确的目录权限
- [ ] 执行文件迁移脚本
- [ ] 验证文件完整性（校验和）
- [ ] 更新配置文件
- [ ] 运行数据库迁移
- [ ] 重启服务
- [ ] 测试插件列表 API
- [ ] 测试插件安装
- [ ] 测试插件启用/禁用
- [ ] 验证日志输出
- [ ] 确认系统稳定运行
- [ ] （可选）删除 install_path 字段

## 回滚方案

如果迁移后出现问题，按以下步骤回滚：

```bash
# 1. 停止服务
systemctl stop arcade

# 2. 恢复数据库
mysql -u root -p arcade < t_plugin_backup.sql

# 3. 恢复文件
cp -r /backup/plugins/* /old/plugin/path/

# 4. 回滚代码
git checkout v0.9.x

# 5. 启动服务
systemctl start arcade

# 6. 验证
curl http://localhost:8080/api/v1/plugins
```

## 注意事项

1. **测试环境先行**
   - 在测试环境完整测试迁移流程
   - 确认无问题后再在生产环境执行

2. **选择合适时间**
   - 在业务低峰期执行
   - 预留足够的维护时间窗口

3. **保留备份**
   - 至少保留一周的备份
   - 确认系统稳定后再清理

4. **监控告警**
   - 迁移后密切监控系统状态
   - 关注插件加载日志
   - 监控磁盘使用情况

5. **通知相关人员**
   - 提前通知开发团队和运维团队
   - 准备应急预案

## 迁移时间估算

| 插件数量 | 预计时间 | 建议维护窗口 |
|---------|---------|------------|
| < 10    | 10分钟  | 30分钟     |
| 10-50   | 30分钟  | 1小时      |
| 50-100  | 1小时   | 2小时      |
| > 100   | 2小时+  | 4小时      |

## 相关文档

- [插件路径管理策略](./PLUGIN_PATH_STRATEGY.md)
- [插件系统重构文档](./PLUGIN_SYSTEM_REFACTOR.md)
- [数据库迁移脚本](./database_migration_plugin_refactor.sql)

---

**文档版本：** v1.0.0  
**最后更新：** 2025-01-16  
**维护者：** Arcade Team

