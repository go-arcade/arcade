# 插件系统快速入门

本文档帮助你快速上手 Arcade 插件系统的安装、使用和管理。

## 前置条件

- Arcade Server 已安装并运行
- 配置了对象存储（S3/MinIO/OSS/COS/GCS）
- 具有管理员权限

## 第一步：准备插件包

插件必须打包为 **zip 格式**，包含以下文件：
- `plugin.so` - 插件二进制文件（必需）
- `manifest.json` - 插件清单文件（必需）

### 方式一：开发自己的插件

1. 创建插件代码：

```go
package main

import (
    "context"
    "github.com/observabil/arcade/pkg/plugin"
)

type MyNotifyPlugin struct{}

func (p *MyNotifyPlugin) Name() string        { return "my-notify" }
func (p *MyNotifyPlugin) Description() string { return "My notification plugin" }
func (p *MyNotifyPlugin) Version() string     { return "1.0.0" }
func (p *MyNotifyPlugin) Type() plugin.PluginType { return plugin.TypeNotify }

func (p *MyNotifyPlugin) Init(ctx context.Context, config any) error {
    // 初始化逻辑
    return nil
}

func (p *MyNotifyPlugin) Cleanup() error {
    // 清理逻辑
    return nil
}

func (p *MyNotifyPlugin) Send(ctx context.Context, message any, opts ...plugin.Option) error {
    // 发送通知逻辑
    return nil
}

func (p *MyNotifyPlugin) SendTemplate(ctx context.Context, template string, data any, opts ...plugin.Option) error {
    // 发送模板通知逻辑
    return nil
}

// 导出插件实例
var Plugin plugin.NotifyPlugin = &MyNotifyPlugin{}
```

2. 编译为 .so 文件：

```bash
go build -buildmode=plugin -o plugin.so main.go
```

3. 创建 `manifest.json` 文件：

```json
{
  "name": "my-notify",
  "version": "1.0.0",
  "description": "My custom notification plugin",
  "author": "Your Name",
  "pluginType": "notify",
  "entryPoint": "my-notify.so",
  "configSchema": {
    "type": "object",
    "properties": {
      "webhook_url": {
        "type": "string",
        "description": "Webhook URL"
      },
      "timeout": {
        "type": "integer",
        "description": "Request timeout in seconds",
        "default": 30
      }
    },
    "required": ["webhook_url"]
  },
  "defaultConfig": {
    "timeout": 30
  },
  "icon": "https://example.com/icon.png",
  "tags": ["notify", "webhook"],
  "resources": {
    "cpu": "100m",
    "memory": "128Mi",
    "disk": "10Mi"
  }
}
```

4. 打包为 zip 文件：

```bash
# 创建 zip 包（包含 plugin.so 和 manifest.json）
zip my-notify-plugin.zip plugin.so manifest.json

# 验证 zip 包内容
unzip -l my-notify-plugin.zip
```

输出应该显示：
```
Archive:  my-notify-plugin.zip
  Length      Date    Time    Name
---------  ---------- -----   ----
  xxxxxxx  2025-01-16 10:30   plugin.so
     xxxx  2025-01-16 10:30   manifest.json
```

### 方式二：从插件市场下载

（功能开发中，敬请期待）

## 第三步：安装插件

### 使用 cURL

```bash
curl -X POST http://localhost:8080/api/v1/plugins/install \
  -F "source=local" \
  -F "file=@my-notify-plugin.zip"
```

**注意：** 现在只需要上传 zip 包，系统会自动解压并解析 manifest.json

### 使用 Web UI

1. 登录 Arcade Web UI
2. 导航到 "插件管理" 页面
3. 点击 "安装插件" 按钮
4. 选择 "本地上传"
5. 选择 zip 包文件（会自动解析 manifest.json）
6. 点击 "安装"

### 响应示例

```json
{
  "success": true,
  "message": "plugin installed successfully",
  "pluginId": "plugin_my-notify_abc123",
  "version": "1.0.0"
}
```

## 第四步：查看插件列表

```bash
curl http://localhost:8080/api/v1/plugins
```

响应：

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "pluginId": "plugin_my-notify_abc123",
      "name": "my-notify",
      "version": "1.0.0",
      "pluginType": "notify",
      "isEnabled": 1,
      "installPath": "/var/lib/arcade/plugins/plugin_my-notify_abc123_1.0.0.so",
      "checksum": "a1b2c3d4...",
      "source": "local"
    }
  ]
}
```

## 第五步：配置插件

### 创建插件配置

```bash
curl -X POST http://localhost:8080/api/v1/plugins/plugin_my-notify_abc123/configs \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Production Webhook",
    "configItems": {
      "webhook_url": "https://hooks.example.com/notify",
      "timeout": 60
    },
    "scope": "global",
    "isDefault": 1
  }'
```

## 第六步：在任务中使用插件

### 在流水线配置中引用插件

```yaml
stages:
  - name: build
    jobs:
      - name: build-app
        commands:
          - go build
        plugins:
          - plugin_id: plugin_my-notify_abc123
            execution_stage: on_success
            params:
              message: "Build completed successfully!"
```

## 管理操作

### 禁用插件

```bash
curl -X POST http://localhost:8080/api/v1/plugins/plugin_my-notify_abc123/disable
```

### 启用插件

```bash
curl -X POST http://localhost:8080/api/v1/plugins/plugin_my-notify_abc123/enable
```

### 更新插件

```bash
curl -X PUT http://localhost:8080/api/v1/plugins/plugin_my-notify_abc123 \
  -F "source=local" \
  -F "file=@my-notify-v2.so" \
  -F "manifest=$(cat manifest-v2.json)"
```

### 卸载插件

```bash
curl -X DELETE http://localhost:8080/api/v1/plugins/plugin_my-notify_abc123
```

## 监控插件

### 查看插件运行状态

```bash
curl http://localhost:8080/api/v1/plugins/metrics
```

响应：

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "name": "my-notify",
      "type": "notify",
      "version": "1.0.0",
      "path": "/var/lib/arcade/plugins/plugin_my-notify_abc123_1.0.0.so",
      "description": "My notification plugin"
    }
  ]
}
```

### 查看插件详情

```bash
curl http://localhost:8080/api/v1/plugins/plugin_my-notify_abc123
```

## 常见问题

### Q: 插件安装失败怎么办？

A: 检查以下几点：
1. 插件文件格式是否正确（必须是 .so 文件）
2. 清单文件是否符合规范
3. 服务器日志中的详细错误信息
4. 存储配置是否正确

### Q: 如何验证插件是否正常工作？

A: 可以通过以下方式验证：
1. 查看插件列表，确认状态为"启用"
2. 查看插件运行指标
3. 在测试任务中使用插件
4. 查看任务日志中的插件执行记录

### Q: 插件可以热更新吗？

A: 是的，插件系统支持热更新：
- 启用/禁用插件会立即生效
- 更新插件会自动热加载新版本
- 无需重启服务器

### Q: 插件可以卸载吗？

A: 是的，所有插件都可以卸载。系统没有内置插件，所有插件都是通过安装获得的，可以随时卸载。

### Q: 如何备份和恢复插件？

A: 插件文件存储在以下位置：
- 本地缓存：`/var/lib/arcade/plugins/`
- S3 存储：`plugins/{plugin_id}/{version}/`
- 数据库：`t_plugin` 表

备份这三个位置的数据即可。

## 下一步

- [插件开发指南](./PLUGIN_DEVELOPMENT.md) - 学习如何开发自己的插件
- [插件系统架构](./PLUGIN_SYSTEM_REFACTOR.md) - 了解插件系统的完整架构
- [API 文档](../api/README_CN.md) - 查看完整的 API 文档

## 获取帮助

如果遇到问题，可以：
- 查看服务器日志：`/var/log/arcade/server.log`
- 查看插件目录：`/var/lib/arcade/plugins/`
- 提交 Issue：https://github.com/observabil/arcade/issues
- 加入社区讨论

---

**文档版本：** v1.0.0  
**最后更新：** 2025-01-16  
**维护者：** Arcade Team

