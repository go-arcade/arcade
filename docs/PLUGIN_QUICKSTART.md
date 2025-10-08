# 插件系统快速开始

本指南帮助您快速上手 Arcade 插件自动加载系统。

## 5分钟快速体验

### 1. 编译示例插件

```bash
# 编译所有插件
make plugins

# 或手动编译单个插件
cd pkg/plugins/notify/stdout
go build -buildmode=plugin -o ../../../../plugins/stdout.so stdout.go
cd -
```

### 2. 运行演示程序

```bash
# 方式一：使用 Makefile（推荐）
make example-autoload

# 方式二：直接运行
go run examples/plugin_autowatch/main.go

# 方式三：使用测试脚本
bash scripts/test_plugin_autoload.sh
```

### 3. 测试自动加载

程序启动后，打开另一个终端窗口：

```bash
# 测试 1: 查看当前插件
ls plugins/

# 测试 2: 删除插件（观察自动卸载）
rm plugins/stdout.so
# 查看程序输出，应该看到："已卸载插件: stdout"

# 测试 3: 重新添加插件（观察自动加载）
make plugins
# 查看程序输出，应该看到："✓ 成功加载插件: ./plugins/stdout.so"

# 测试 4: 修改配置文件
vim conf.d/plugins.yaml
# 保存后查看输出，应该看到："✓ 配置文件重新加载成功"
```

### 4. 停止程序

在运行演示程序的终端按 `Ctrl+C` 优雅退出。

## 在自己的项目中使用

### 基础用法

```go
package main

import (
    "context"
    "github.com/observabil/arcade/pkg/plugin"
)

func main() {
    // 1. 创建管理器
    manager := plugin.NewManager()
    manager.SetContext(context.Background())
    
    // 2. 加载配置
    manager.LoadPluginsFromConfig("./conf.d/plugins.yaml")
    manager.Init(context.Background())
    
    // 3. 启动自动监控
    watchDirs := []string{"./plugins"}
    manager.StartAutoWatch(watchDirs, "./conf.d/plugins.yaml")
    defer manager.StopAutoWatch()
    
    // 4. 使用插件
    notifyPlugin, _ := manager.GetNotifyPlugin("stdout")
    notifyPlugin.Send(context.Background(), "Hello from plugin!")
    
    // 5. 保持运行...
    select {}
}
```

### 配置文件示例

创建 `conf.d/plugins.yaml`:

```yaml
plugins:
  - path: ./plugins/stdout.so
    name: stdout
    type: notify
    version: "1.0.0"
    config:
      prefix: "[通知]"
      
  - path: ./plugins/custom-plugin.so
    name: my-plugin
    type: custom
    version: "1.0.0"
    config:
      key: value
```

## 开发自己的插件

### 1. 创建插件代码

```go
// myplugin.go
package main

import "context"

type MyPlugin struct{}

func (p *MyPlugin) Name() string        { return "my-plugin" }
func (p *MyPlugin) Description() string { return "我的第一个插件" }
func (p *MyPlugin) Version() string     { return "1.0.0" }
func (p *MyPlugin) Type() string        { return "custom" }

func (p *MyPlugin) Init(ctx context.Context, config any) error {
    // 初始化逻辑
    return nil
}

func (p *MyPlugin) Cleanup() error {
    // 清理逻辑
    return nil
}

func (p *MyPlugin) Execute(ctx context.Context, action string, params any) (any, error) {
    // 执行自定义操作
    return "success", nil
}

// 导出插件实例
var Plugin = &MyPlugin{}
```

### 2. 编译插件

```bash
go build -buildmode=plugin -o plugins/myplugin.so myplugin.go
```

### 3. 自动加载

如果演示程序正在运行，插件会自动加载！否则将插件路径添加到配置文件即可。

## 常用命令

```bash
# 查看所有可用命令
make help

# 编译所有插件
make plugins

# 清理插件
make plugins-clean

# 运行演示
make example-autoload

# 运行测试脚本
make test-plugin-autoload

# 生成 proto 代码
make proto

# 完整构建
make all
```

## 故障排除

### 问题：插件编译失败

**错误信息：**
```
plugin was built with a different version of package...
```

**解决方法：**
确保插件使用与主程序相同的 Go 版本和依赖：

```bash
# 检查 Go 版本
go version

# 清理并重新编译
go clean -cache
make plugins-clean
make plugins
```

### 问题：插件未自动加载

**检查清单：**

1. 确认文件后缀是 `.so`
2. 确认文件在监控目录中
3. 查看程序日志输出
4. 确认插件编译成功

```bash
# 查看插件文件
ls -la plugins/

# 手动测试加载
go run examples/plugin_autowatch/main.go
```

### 问题：配置文件修改无效

**解决方法：**

1. 检查 YAML 格式
2. 确认配置文件路径正确
3. 查看错误日志

```bash
# 验证 YAML 格式
cat conf.d/plugins.yaml

# 或使用在线工具验证 YAML
```

## 下一步

- 📖 阅读 [插件自动加载详细文档](./PLUGIN_AUTO_LOAD.md)
- 🔧 查看 [插件开发指南](./PLUGIN_DEVELOPMENT.md)
- 📝 参考 [标签示例](./LABEL_EXAMPLES.md)
- 🚀 查看 [完整文档](./QUICKSTART.md)

## 实用资源

- **源代码目录：**
  - 插件管理器：`pkg/plugin/plugin_manager.go`
  - 自动监控器：`pkg/plugin/plugin_watcher.go`
  - 示例插件：`pkg/plugins/notify/stdout/`
  - 演示程序：`examples/plugin_autowatch/main.go`

- **配置文件：**
  - 插件配置：`conf.d/plugins.yaml`

- **测试脚本：**
  - 测试脚本：`scripts/test_plugin_autoload.sh`

## 核心特性

✅ **自动加载** - 放入插件文件即可自动加载  
✅ **自动卸载** - 删除插件文件自动卸载  
✅ **配置热重载** - 修改配置立即生效  
✅ **手动重载** - API 支持手动重载指定插件  
✅ **防抖机制** - 避免频繁重复操作  
✅ **并发安全** - 内部使用读写锁保护  
✅ **类型丰富** - 支持6种插件类型  
✅ **易于扩展** - 简单的插件接口  

## 性能提示

1. **避免频繁重载** - 防抖机制会延迟 500ms
2. **资源清理** - 实现完善的 `Cleanup()` 方法
3. **并发安全** - 插件方法需要考虑并发调用
4. **错误处理** - 插件错误不会影响其他插件

## 最佳实践

1. ✅ 使用语义化版本号
2. ✅ 编写清晰的文档
3. ✅ 实现完整的错误处理
4. ✅ 添加日志输出
5. ✅ 编写单元测试
6. ✅ 使用配置文件管理插件
7. ✅ 在测试环境充分验证

## 获取帮助

如有问题，请：

1. 查看详细文档
2. 检查日志输出
3. 运行测试脚本
4. 提交 Issue

祝使用愉快！🎉

