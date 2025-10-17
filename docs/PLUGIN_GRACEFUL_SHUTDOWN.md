# 插件优雅关闭与注册状态管理

## 概述

本文档描述了插件系统的注册状态管理和优雅关闭机制的实现。

## 一、插件注册状态管理

### 1.1 数据库字段

在 `t_plugin` 表中新增了两个字段：

- `register_status` (INT): 插件注册状态
  - `0` - 未注册
  - `1` - 注册中
  - `2` - 已注册
  - `3` - 注册失败
  
- `register_error` (TEXT): 注册错误信息（当注册失败时记录详细错误）

### 1.2 状态常量

在 `model_plugin.go` 中定义了状态常量：

```go
const (
    PluginRegisterStatusUnregistered = 0 // 未注册
    PluginRegisterStatusRegistering  = 1 // 注册中
    PluginRegisterStatusRegistered   = 2 // 已注册
    PluginRegisterStatusFailed       = 3 // 注册失败
)
```

### 1.3 仓储方法

在 `repo_plugin.go` 中新增了更新注册状态的方法：

```go
func (r *PluginRepo) UpdatePluginRegistrationStatus(pluginID string, status int, errorMsg string) error
```

### 1.4 服务层集成

在 `service_plugin.go` 的 `hotReloadPlugin` 方法中集成了状态更新：

1. **注册前**：将状态更新为 `注册中`
2. **注册成功**：将状态更新为 `已注册`，清空错误信息
3. **注册失败**：将状态更新为 `注册失败`，记录错误信息

## 二、插件重载机制优化

### 2.1 停止逻辑

在 `Manager.ReloadPlugin` 方法中实现了明确的停止流程：

```go
// Step 1: 调用插件的 Cleanup 方法
m.cleanupPlugin(oldClient)

// Step 2: 终止插件进程
oldClient.pluginClient.Kill()

// Step 3: 等待进程优雅退出（最多5秒）
等待 oldClient.pluginClient.Exited() 返回 true

// Step 4: 从注册表中移除
delete(m.plugins, name)
delete(m.clients, name)

// Step 5: 重新注册（支持重试3次）
m.RegisterPlugin(name, pluginPath, pluginConfig)
```

### 2.2 关键改进

- ✅ **显式停止**：明确调用 Cleanup 和 Kill
- ✅ **优雅等待**：等待进程自然退出（超时保护）
- ✅ **重试机制**：注册失败时自动重试3次
- ✅ **详细日志**：记录每个步骤的执行情况

## 三、主程序优雅关闭

### 3.1 信号处理

在 `main.go` 中实现了完整的信号处理：

```go
// 监听系统信号
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

// 等待信号
sig := <-quit

// 按顺序关闭
1. HTTP 服务器
2. 插件管理器（调用 cleanup()）
3. gRPC 服务器
```

### 3.2 关闭顺序

1. **HTTP 服务器**：停止接收新请求，处理完现有请求
2. **插件管理器**：
   - 对每个插件调用 `Cleanup()` 方法
   - 终止所有插件进程
   - 清理内部注册表
3. **gRPC 服务器**：停止 gRPC 服务

### 3.3 App cleanup 函数

在 `app.go` 中的 cleanup 函数包含：

```go
cleanup := func() {
    // 停止所有插件
    if pluginMgr != nil {
        pluginMgr.Close()
    }
    
    // 停止 gRPC 服务器
    if grpcServer != nil {
        grpcServer.Stop()
    }
}
```

## 四、数据库迁移

### 4.1 迁移脚本

执行 SQL 脚本：`docs/database_migration_plugin_register_status.sql`

```sql
ALTER TABLE `t_plugin` 
ADD COLUMN `register_status` INT NOT NULL DEFAULT 0 COMMENT '注册状态: 0=未注册 1=注册中 2=已注册 3=注册失败',
ADD COLUMN `register_error` TEXT NULL COMMENT '注册错误信息';

CREATE INDEX `idx_register_status` ON `t_plugin` (`register_status`);
```

### 4.2 回滚脚本

如需回滚，执行：

```sql
DROP INDEX `idx_register_status` ON `t_plugin`;
ALTER TABLE `t_plugin` 
DROP COLUMN `register_error`,
DROP COLUMN `register_status`;
```

## 五、测试验证

### 5.1 正常关闭测试

```bash
# 启动服务
./arcade

# 发送 SIGTERM 信号
kill -TERM <pid>

# 预期日志：
# Received signal: terminated, shutting down gracefully...
# Shutting down plugin manager...
# [plugin] cleanup plugin xxx...
# [plugin] plugin xxx process stopped
# Plugin manager stopped successfully
# Shutting down gRPC server...
# Server shutdown complete
```

### 5.2 插件重载测试

```bash
# 通过 API 重载插件
curl -X POST http://localhost:8080/api/plugin/reload -d '{"name":"stdout"}'

# 预期日志：
# [plugin] reloading stdout from /path/to/plugin ...
# [plugin] cleanup failed for stdout: <error if any>
# [plugin] wait for old stdout to exit timeout, continuing reload (如果超时)
# [plugin] reloaded stdout successfully after 1 attempt(s)
```

### 5.3 注册状态测试

```sql
-- 查询所有插件的注册状态
SELECT plugin_id, name, version, register_status, register_error 
FROM t_plugin;

-- 预期结果：
-- register_status = 2 (已注册) 表示插件运行正常
-- register_status = 3 (注册失败) 时 register_error 字段包含错误详情
```

## 六、监控和告警

### 6.1 监控指标

建议监控以下指标：

- 注册状态为 3（失败）的插件数量
- 插件进程的存活状态
- 插件 Cleanup 调用成功率

### 6.2 告警规则

- 当有插件注册失败时发送告警
- 当主程序关闭超时时发送告警（如插件无法停止）

## 七、最佳实践

1. **插件开发**：
   - 插件必须正确实现 `Cleanup()` 方法
   - 清理资源时应设置合理的超时时间
   
2. **运维部署**：
   - 使用 SIGTERM 而非 SIGKILL 关闭服务
   - 监控插件注册状态
   - 定期检查插件进程是否有僵尸进程

3. **故障排查**：
   - 查看 `register_error` 字段了解注册失败原因
   - 检查日志中的插件停止过程
   - 使用 `ps` 命令确认插件进程是否正确停止

## 八、相关文件

- `/Users/gagral/code/go/src/arcade/internal/engine/model/model_plugin.go` - 插件模型定义
- `/Users/gagral/code/go/src/arcade/internal/engine/repo/repo_plugin.go` - 插件仓储方法
- `/Users/gagral/code/go/src/arcade/internal/engine/service/plugin/service_plugin.go` - 插件服务层
- `/Users/gagral/code/go/src/arcade/pkg/plugin/manager.go` - 插件管理器
- `/Users/gagral/code/go/src/arcade/cmd/arcade/main.go` - 主程序入口
- `/Users/gagral/code/go/src/arcade/internal/app/app.go` - 应用初始化
- `/Users/gagral/code/go/src/arcade/docs/database_migration_plugin_register_status.sql` - 数据库迁移脚本

## 九、版本历史

- **2025-10-17**: 初始版本
  - 添加插件注册状态管理
  - 实现插件重载前的停止逻辑
  - 优化主程序优雅关闭流程

