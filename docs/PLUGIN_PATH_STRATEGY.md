# 插件路径管理策略

## 概述

从 v1.0.0 版本开始，Arcade 插件系统采用**动态路径生成**策略，不再在数据库中存储本地安装路径。

## 设计理念

### 问题

旧的设计在数据库中存储 `install_path` 字段，存在以下问题：
1. 路径变更时需要更新数据库
2. 不同环境下路径可能不一致
3. 数据库存储冗余信息
4. 难以批量管理和清理

### 解决方案

采用**统一路径生成规则**，根据 `plugin_id` 和 `version` 动态生成本地路径。

## 路径生成规则

### 本地缓存路径

```go
func getLocalPath(pluginID, version string) string {
    filename := fmt.Sprintf("%s_%s.so", pluginID, version)
    return filepath.Join(localCacheDir, filename)
}
```

**示例：**
```
/var/lib/arcade/plugins/plugin_slack_abc123_1.0.0.so
/var/lib/arcade/plugins/plugin_k8s-deploy_xyz456_2.1.0.so
```

**格式：**
```
{localCacheDir}/{plugin_id}_{version}.so
```

### S3 存储路径

```go
func getS3Path(pluginID, version string) string {
    return fmt.Sprintf("plugins/%s/%s/%s_%s.so", pluginID, version, pluginID, version)
}
```

**示例：**
```
plugins/plugin_slack_abc123/1.0.0/plugin_slack_abc123_1.0.0.so
plugins/plugin_k8s-deploy_xyz456/2.1.0/plugin_k8s-deploy_xyz456_2.1.0.so
```

**格式：**
```
plugins/{plugin_id}/{version}/{plugin_id}_{version}.so
```

## 配置说明

### 本地缓存目录配置

在 `conf.d/config.toml` 中配置：

```toml
[plugin]
# 本地缓存目录
local_cache_dir = "/var/lib/arcade/plugins"

# 确保目录存在且有写权限
# mkdir -p /var/lib/arcade/plugins
# chmod 755 /var/lib/arcade/plugins
```

### 目录权限

```bash
# 创建目录
sudo mkdir -p /var/lib/arcade/plugins

# 设置权限
sudo chown -R arcade:arcade /var/lib/arcade/plugins
sudo chmod 755 /var/lib/arcade/plugins
```

## 代码实现

### Service 层

```go
// PluginService 中的路径生成方法
func (s *PluginService) getLocalPath(pluginID, version string) string {
    filename := fmt.Sprintf("%s_%s.so", pluginID, version)
    return filepath.Join(s.localCacheDir, filename)
}

func (s *PluginService) getS3Path(pluginID, version string) string {
    return fmt.Sprintf("plugins/%s/%s/%s_%s.so", pluginID, version, pluginID, version)
}
```

### Agent 下载服务

```go
// PluginDownloadService 中的路径生成方法
func (s *PluginDownloadService) getPluginLocalPath(pluginID, version string) string {
    localCacheDir := "/var/lib/arcade/plugins"
    filename := fmt.Sprintf("%s_%s.so", pluginID, version)
    return filepath.Join(localCacheDir, filename)
}
```

### Plugin Manager

```go
// LoadPluginsFromDatabase 加载插件时动态生成路径
for _, dbPlugin := range plugins {
    filename := fmt.Sprintf("%s_%s.so", dbPlugin.PluginId, dbPlugin.Version)
    absPath := filepath.Join(localCacheDir, filename)
    
    // 加载插件
    m.Register(absPath, dbPlugin.PluginId, config)
}
```

## 操作流程

### 安装插件

```
1. 上传插件文件 (plugin.so)
2. 生成 plugin_id (plugin_name_shortid)
3. 计算 checksum (SHA256)
4. 生成本地路径: /var/lib/arcade/plugins/{plugin_id}_{version}.so
5. 保存到本地缓存
6. 生成 S3 路径: plugins/{plugin_id}/{version}/{plugin_id}_{version}.so
7. 上传到 S3
8. 保存元数据到数据库（不包含 install_path）
9. 热加载到内存
```

### 启用插件

```
1. 从数据库读取插件信息（plugin_id, version）
2. 动态生成本地路径
3. 检查文件是否存在
4. 加载到内存
5. 更新数据库状态
```

### 卸载插件

```
1. 从数据库读取插件信息（plugin_id, version）
2. 动态生成本地路径
3. 从内存卸载
4. 删除本地文件
5. 动态生成 S3 路径
6. 删除 S3 文件
7. 删除数据库记录
```

## 优势

### ✅ 1. 简化数据模型

- 数据库不存储路径信息
- 减少数据冗余
- 避免路径不一致

### ✅ 2. 统一管理

- 路径生成规则统一
- 便于批量操作
- 易于迁移

### ✅ 3. 灵活配置

- 可通过配置文件修改缓存目录
- 无需更新数据库
- 支持多环境部署

### ✅ 4. 易于维护

- 清理旧版本更简单
- 批量操作更方便
- 问题排查更容易

## 文件清理

### 手动清理旧版本

```bash
# 列出所有插件文件
ls -lh /var/lib/arcade/plugins/

# 清理特定插件的旧版本
rm /var/lib/arcade/plugins/plugin_slack_abc123_1.0.0.so

# 清理所有 1.0.0 版本的插件
rm /var/lib/arcade/plugins/*_1.0.0.so
```

### 自动清理脚本

```bash
#!/bin/bash
# cleanup_old_plugins.sh - 清理超过 30 天未使用的插件文件

PLUGIN_DIR="/var/lib/arcade/plugins"
DAYS=30

find "$PLUGIN_DIR" -name "*.so" -type f -mtime +$DAYS -delete
echo "Cleaned up plugins older than $DAYS days"
```

## 迁移指南

### 从旧版本迁移

如果你的系统中有使用 `install_path` 的旧数据：

1. **备份数据库**
```sql
CREATE TABLE t_plugin_backup AS SELECT * FROM t_plugin;
```

2. **检查现有插件**
```sql
SELECT plugin_id, name, version, install_path FROM t_plugin;
```

3. **重新整理插件文件**
```bash
# 创建新的缓存目录
mkdir -p /var/lib/arcade/plugins

# 复制并重命名插件文件
# 格式：{plugin_id}_{version}.so
cp /old/path/plugin.so /var/lib/arcade/plugins/plugin_slack_abc123_1.0.0.so
```

4. **验证路径**
```bash
# 列出所有插件
ls -lh /var/lib/arcade/plugins/

# 确认命名符合规则：{plugin_id}_{version}.so
```

5. **删除旧字段（可选）**
```sql
ALTER TABLE t_plugin DROP COLUMN install_path;
```

## 常见问题

### Q1: 如何确定插件的本地路径？

A: 使用公式：`{localCacheDir}/{plugin_id}_{version}.so`

### Q2: 如果本地文件丢失怎么办？

A: 系统会自动从 S3 重新下载并缓存到本地。

### Q3: 多个服务器节点如何同步插件？

A: 所有节点共享同一个 S3 存储，每个节点独立维护本地缓存。首次启用插件时会自动从 S3 下载。

### Q4: 如何批量清理旧版本插件？

A: 使用文件名模式匹配：
```bash
# 删除所有 v1.0.0 版本
rm /var/lib/arcade/plugins/*_1.0.0.so

# 删除特定插件的所有版本
rm /var/lib/arcade/plugins/plugin_slack_*
```

### Q5: 路径生成规则可以自定义吗？

A: 可以，修改 `getLocalPath()` 和 `getS3Path()` 方法即可，但建议保持统一规则。

## 最佳实践

1. **统一缓存目录**
   - 所有环境使用相同的路径结构
   - 便于运维管理

2. **定期清理**
   - 设置定时任务清理旧版本
   - 保留最近 N 个版本

3. **监控磁盘使用**
   - 监控缓存目录大小
   - 设置告警阈值

4. **备份策略**
   - S3 作为主存储和备份
   - 本地缓存可随时重建

5. **版本管理**
   - 严格遵循语义化版本
   - 版本号作为路径组成部分

## 相关代码

- `internal/engine/service/service_plugin.go` - 路径生成逻辑
- `internal/engine/service/agent/service_plugin_download.go` - Agent 下载服务路径
- `pkg/plugin/plugin_manager.go` - 插件加载路径生成

---

**文档版本：** v1.0.0  
**最后更新：** 2025-01-16  
**维护者：** Arcade Team

