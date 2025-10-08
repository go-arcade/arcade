# Arcade 插件开发指南

本指南详细介绍如何为 Arcade CI/CD 平台开发自定义插件。

## 目录

- [插件系统概述](#插件系统概述)
- [插件类型](#插件类型)
- [开发环境准备](#开发环境准备)
- [快速开始](#快速开始)
- [详细开发步骤](#详细开发步骤)
- [插件接口详解](#插件接口详解)
- [实战示例](#实战示例)
- [编译和调试](#编译和调试)
- [最佳实践](#最佳实践)
- [常见问题](#常见问题)

## 插件系统概述

Arcade 使用 Go 的 `plugin` 包实现动态插件系统，支持：

- ✅ 运行时动态加载插件
- ✅ 热更新（自动监控文件变化）
- ✅ 多种插件类型
- ✅ 配置驱动
- ✅ 生命周期管理

### 工作原理

```
┌─────────────┐
│  主程序      │
│  (arcade)   │
└──────┬──────┘
       │
       │ 加载
       ▼
┌─────────────┐      ┌──────────────┐
│ 插件管理器    │─────▶│ 文件监控器    │
│ (Manager)   │      │ (Watcher)    │
└──────┬──────┘      └──────────────┘
       │
       │ 调用
       ▼
┌─────────────┐
│  插件.so     │
│  (Plugin)   │
└─────────────┘
```

## 插件类型

Arcade 支持多种插件类型：

### 1. CI 插件 (CIPlugin)

用于持续集成任务。

**接口方法：**
- `Build()` - 构建项目
- `Test()` - 运行测试
- `Lint()` - 代码检查

**使用场景：**
- 编译 Go/Java/Node.js 项目
- 运行单元测试
- 代码质量检查
- 代码覆盖率统计

### 2. CD 插件 (CDPlugin)

用于持续部署。

**接口方法：**
- `Deploy()` - 部署应用
- `Rollback()` - 回滚部署

**使用场景：**
- Kubernetes 部署
- Docker 镜像推送
- FTP/SFTP 上传
- 云服务部署（AWS、阿里云等）

### 3. Security 插件 (SecurityPlugin)

用于安全检查。

**接口方法：**
- `Scan()` - 安全扫描
- `Audit()` - 安全审计

**使用场景：**
- 依赖漏洞扫描
- 代码安全审计
- 镜像安全检查
- 合规性检查

### 4. Notify 插件 (NotifyPlugin)

用于消息通知。

**接口方法：**
- `Send()` - 发送通知
- `SendTemplate()` - 发送模板通知

**使用场景：**
- Slack 通知
- 钉钉通知
- 企业微信通知
- 邮件通知
- 短信通知

### 5. Storage 插件 (StoragePlugin)

用于数据存储。

**接口方法：**
- `Save()` - 保存数据
- `Load()` - 加载数据
- `Delete()` - 删除数据

**使用场景：**
- 产物存储
- 日志归档
- 缓存管理
- 对象存储（S3、OSS 等）

### 6. Custom 插件 (CustomPlugin)

用于自定义功能。

**接口方法：**
- `Execute()` - 执行自定义操作

**使用场景：**
- 任何不属于以上类型的功能
- 特殊业务逻辑
- 第三方系统集成

## 开发环境准备

### 必需条件

1. **Go 版本**：必须与主程序相同（当前为 Go 1.24.0）
2. **操作系统**：Linux 或 macOS（plugin 包不支持 Windows）
3. **依赖版本**：必须与主程序的依赖版本完全一致

### 检查环境

```bash
# 检查 Go 版本
go version

# 查看主程序 Go 版本
grep "^go " go.mod

# 克隆项目
git clone <repository>
cd arcade
```

## 快速开始

### 5 分钟创建你的第一个插件

#### 1. 创建插件文件

```go
// hello_plugin.go
package main

import (
    "context"
    "fmt"
)

// HelloPlugin 示例插件
type HelloPlugin struct {
    config map[string]interface{}
}

// Name 插件名称
func (p *HelloPlugin) Name() string {
    return "hello"
}

// Description 插件描述
func (p *HelloPlugin) Description() string {
    return "一个简单的示例插件"
}

// Version 插件版本
func (p *HelloPlugin) Version() string {
    return "1.0.0"
}

// Type 插件类型
func (p *HelloPlugin) Type() string {
    return "custom"
}

// Init 初始化插件
func (p *HelloPlugin) Init(ctx context.Context, config any) error {
    fmt.Println("Hello Plugin 初始化")
    if cfg, ok := config.(map[string]interface{}); ok {
        p.config = cfg
        fmt.Printf("配置: %+v\n", cfg)
    }
    return nil
}

// Cleanup 清理插件资源
func (p *HelloPlugin) Cleanup() error {
    fmt.Println("Hello Plugin 清理")
    return nil
}

// Execute 执行自定义操作
func (p *HelloPlugin) Execute(ctx context.Context, action string, params any) (any, error) {
    msg := fmt.Sprintf("Hello from plugin! Action: %s", action)
    fmt.Println(msg)
    return msg, nil
}

// Plugin 导出插件实例
var Plugin = &HelloPlugin{}
```

#### 2. 编译插件

```bash
go build -buildmode=plugin -o plugins/hello.so hello_plugin.go
```

#### 3. 配置插件

编辑 `conf.d/plugins.yaml`:

```yaml
plugins:
  - path: ./plugins/hello.so
    name: hello
    type: custom
    version: "1.0.0"
    config:
      greeting: "你好"
```

#### 4. 运行

启动 Arcade 服务，插件会自动加载！

## 详细开发步骤

### 步骤 1：设计插件

在开始编码前，明确：

1. **插件类型**：选择合适的插件类型
2. **功能范围**：插件要实现什么功能
3. **配置项**：需要哪些配置参数
4. **依赖项**：需要哪些外部依赖

### 步骤 2：创建项目结构

推荐的目录结构：

```
pkg/plugins/
└── notify/
    └── dingtalk/
        ├── dingtalk.go       # 主要实现
        ├── config.go         # 配置结构
        ├── client.go         # HTTP 客户端
        └── README.md         # 插件文档
```

### 步骤 3：实现基础接口

所有插件必须实现 `BasePlugin` 接口：

```go
package main

import (
    "context"
    "github.com/observabil/arcade/pkg/plugin"
)

type MyPlugin struct {
    name        string
    version     string
    description string
    config      *MyConfig
    // 其他字段...
}

// BasePlugin 接口实现
func (p *MyPlugin) Name() string        { return p.name }
func (p *MyPlugin) Description() string { return p.description }
func (p *MyPlugin) Version() string     { return p.version }
func (p *MyPlugin) Type() plugin.PluginType { return plugin.TypeNotify }

func (p *MyPlugin) Init(ctx context.Context, config any) error {
    // 解析配置
    // 初始化资源
    return nil
}

func (p *MyPlugin) Cleanup() error {
    // 清理资源
    return nil
}
```

### 步骤 4：实现特定类型接口

根据插件类型实现相应接口。

#### Notify 插件示例

```go
func (p *MyPlugin) Send(ctx context.Context, message any) error {
    // 发送消息逻辑
    return nil
}

func (p *MyPlugin) SendTemplate(ctx context.Context, template string, data any) error {
    // 发送模板消息逻辑
    return nil
}
```

### 步骤 5：导出插件实例

在文件末尾导出插件：

```go
// Plugin 导出的插件实例
var Plugin = &MyPlugin{
    name:        "my-plugin",
    version:     "1.0.0",
    description: "我的插件",
}
```

## 插件接口详解

### BasePlugin - 基础接口

所有插件都必须实现：

```go
type BasePlugin interface {
    Name() string                                 // 插件名称（唯一标识）
    Description() string                          // 插件描述
    Version() string                              // 插件版本
    Type() PluginType                            // 插件类型
    Init(ctx context.Context, config any) error  // 初始化
    Cleanup() error                              // 清理资源
}
```

**重要说明：**

- `Name()` 返回的名称必须唯一
- `Type()` 必须返回正确的插件类型常量
- `Init()` 在插件加载后自动调用
- `Cleanup()` 在插件卸载前自动调用

### CIPlugin - CI 插件

```go
type CIPlugin interface {
    BasePlugin
    Build(ctx context.Context, projectConfig any, opts ...Option) error
    Test(ctx context.Context, projectConfig any, opts ...Option) error
    Lint(ctx context.Context, projectConfig any, opts ...Option) error
}
```

### CDPlugin - CD 插件

```go
type CDPlugin interface {
    BasePlugin
    Deploy(ctx context.Context, projectConfig any, opts ...Option) error
    Rollback(ctx context.Context, projectConfig any, opts ...Option) error
}
```

### SecurityPlugin - 安全插件

```go
type SecurityPlugin interface {
    BasePlugin
    Scan(ctx context.Context, projectConfig any, opts ...Option) error
    Audit(ctx context.Context, projectConfig any, opts ...Option) error
}
```

### NotifyPlugin - 通知插件

```go
type NotifyPlugin interface {
    BasePlugin
    Send(ctx context.Context, message any, opts ...Option) error
    SendTemplate(ctx context.Context, template string, data any, opts ...Option) error
}
```

### StoragePlugin - 存储插件

```go
type StoragePlugin interface {
    BasePlugin
    Save(ctx context.Context, key string, data any, opts ...Option) error
    Load(ctx context.Context, key string, opts ...Option) (any, error)
    Delete(ctx context.Context, key string, opts ...Option) error
}
```

### CustomPlugin - 自定义插件

```go
type CustomPlugin interface {
    BasePlugin
    Execute(ctx context.Context, action string, params any, opts ...Option) (any, error)
}
```

## 实战示例

### 示例 1：钉钉通知插件

完整实现一个钉钉通知插件：

```go
package main

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

// DingTalkPlugin 钉钉通知插件
type DingTalkPlugin struct {
    webhook string
    secret  string
    client  *http.Client
}

func (p *DingTalkPlugin) Name() string        { return "dingtalk" }
func (p *DingTalkPlugin) Description() string { return "钉钉消息通知插件" }
func (p *DingTalkPlugin) Version() string     { return "1.0.0" }
func (p *DingTalkPlugin) Type() string        { return "notify" }

func (p *DingTalkPlugin) Init(ctx context.Context, config any) error {
    cfg, ok := config.(map[string]interface{})
    if !ok {
        return fmt.Errorf("invalid config format")
    }

    webhook, ok := cfg["webhook"].(string)
    if !ok || webhook == "" {
        return fmt.Errorf("webhook is required")
    }
    p.webhook = webhook

    if secret, ok := cfg["secret"].(string); ok {
        p.secret = secret
    }

    p.client = &http.Client{
        Timeout: 10 * time.Second,
    }

    return nil
}

func (p *DingTalkPlugin) Cleanup() error {
    if p.client != nil {
        p.client.CloseIdleConnections()
    }
    return nil
}

func (p *DingTalkPlugin) Send(ctx context.Context, message any) error {
    msg, ok := message.(string)
    if !ok {
        return fmt.Errorf("message must be string")
    }

    payload := map[string]interface{}{
        "msgtype": "text",
        "text": map[string]string{
            "content": msg,
        },
    }

    return p.sendRequest(ctx, payload)
}

func (p *DingTalkPlugin) SendTemplate(ctx context.Context, template string, data any) error {
    // 实现模板消息
    payload := map[string]interface{}{
        "msgtype": "markdown",
        "markdown": map[string]string{
            "title": "通知",
            "text":  template,
        },
    }

    return p.sendRequest(ctx, payload)
}

func (p *DingTalkPlugin) sendRequest(ctx context.Context, payload map[string]interface{}) error {
    body, err := json.Marshal(payload)
    if err != nil {
        return fmt.Errorf("marshal payload: %w", err)
    }

    req, err := http.NewRequestWithContext(ctx, "POST", p.webhook, bytes.NewReader(body))
    if err != nil {
        return fmt.Errorf("create request: %w", err)
    }

    req.Header.Set("Content-Type", "application/json")

    resp, err := p.client.Do(req)
    if err != nil {
        return fmt.Errorf("send request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
    }

    return nil
}

var Plugin = &DingTalkPlugin{}
```

配置文件：

```yaml
plugins:
  - path: ./plugins/dingtalk.so
    name: dingtalk
    type: notify
    version: "1.0.0"
    config:
      webhook: "https://oapi.dingtalk.com/robot/send?access_token=xxx"
      secret: "SEC..."
```

### 示例 2：Docker 构建插件

```go
package main

import (
    "context"
    "fmt"
    "os/exec"
)

// DockerBuildPlugin Docker 构建插件
type DockerBuildPlugin struct {
    registry string
}

func (p *DockerBuildPlugin) Name() string        { return "docker-build" }
func (p *DockerBuildPlugin) Description() string { return "Docker 镜像构建插件" }
func (p *DockerBuildPlugin) Version() string     { return "1.0.0" }
func (p *DockerBuildPlugin) Type() string        { return "ci" }

func (p *DockerBuildPlugin) Init(ctx context.Context, config any) error {
    if cfg, ok := config.(map[string]interface{}); ok {
        if registry, ok := cfg["registry"].(string); ok {
            p.registry = registry
        }
    }
    return nil
}

func (p *DockerBuildPlugin) Cleanup() error {
    return nil
}

func (p *DockerBuildPlugin) Build(ctx context.Context, projectConfig any) error {
    cfg, ok := projectConfig.(map[string]interface{})
    if !ok {
        return fmt.Errorf("invalid project config")
    }

    imageName, _ := cfg["image"].(string)
    tag, _ := cfg["tag"].(string)

    if imageName == "" {
        return fmt.Errorf("image name is required")
    }
    if tag == "" {
        tag = "latest"
    }

    fullImage := fmt.Sprintf("%s/%s:%s", p.registry, imageName, tag)

    // 构建镜像
    cmd := exec.CommandContext(ctx, "docker", "build", "-t", fullImage, ".")
    cmd.Stdout = nil // 根据需要重定向输出
    cmd.Stderr = nil

    if err := cmd.Run(); err != nil {
        return fmt.Errorf("docker build failed: %w", err)
    }

    // 推送镜像
    cmd = exec.CommandContext(ctx, "docker", "push", fullImage)
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("docker push failed: %w", err)
    }

    return nil
}

func (p *DockerBuildPlugin) Test(ctx context.Context, projectConfig any) error {
    // 实现测试逻辑
    return nil
}

func (p *DockerBuildPlugin) Lint(ctx context.Context, projectConfig any) error {
    // 实现 Lint 逻辑
    return nil
}

var Plugin = &DockerBuildPlugin{}
```

## 编译和调试

### 编译插件

```bash
# 基本编译
go build -buildmode=plugin -o plugins/myplugin.so myplugin.go

# 指定输出目录
go build -buildmode=plugin -o /path/to/plugins/myplugin.so myplugin.go

# 使用 Makefile（推荐）
make plugins
```

### 常见编译问题

#### 问题 1：版本不匹配

```
plugin was built with a different version of package...
```

**解决方法：**
```bash
# 确保 Go 版本一致
go version

# 清理缓存
go clean -cache

# 重新编译
make plugins-clean
make plugins
```

#### 问题 2：符号未导出

```
plugin: symbol Plugin not found
```

**解决方法：**
确保导出插件实例：
```go
var Plugin = &MyPlugin{}  // 必须首字母大写
```

### 调试插件

#### 方法 1：使用日志

```go
import "github.com/observabil/arcade/pkg/log"

func (p *MyPlugin) Init(ctx context.Context, config any) error {
    log.Infof("插件初始化: %s", p.Name())
    log.Debugf("配置: %+v", config)
    return nil
}
```

#### 方法 2：单独测试

创建测试程序：

```go
// test_plugin.go
package main

import (
    "context"
    "plugin"
)

func main() {
    // 加载插件
    p, err := plugin.Open("./plugins/myplugin.so")
    if err != nil {
        panic(err)
    }

    // 查找符号
    sym, err := p.Lookup("Plugin")
    if err != nil {
        panic(err)
    }

    // 类型断言
    myPlugin := sym.(MyPluginInterface)

    // 测试
    myPlugin.Init(context.Background(), nil)
    // ... 其他测试
}
```

运行测试：
```bash
go run test_plugin.go
```

## 最佳实践

### 1. 配置管理

**定义配置结构：**

```go
type MyPluginConfig struct {
    APIKey    string        `mapstructure:"api_key"`
    Timeout   time.Duration `mapstructure:"timeout"`
    RetryMax  int           `mapstructure:"retry_max"`
    EnableSSL bool          `mapstructure:"enable_ssl"`
}

func (p *MyPlugin) Init(ctx context.Context, config any) error {
    cfg := &MyPluginConfig{
        Timeout:  10 * time.Second,  // 默认值
        RetryMax: 3,
        EnableSSL: true,
    }

    // 解析配置
    if configMap, ok := config.(map[string]interface{}); ok {
        // 手动解析或使用 mapstructure
        if apiKey, ok := configMap["api_key"].(string); ok {
            cfg.APIKey = apiKey
        }
        // ...
    }

    p.config = cfg
    return nil
}
```

### 2. 错误处理

```go
import "fmt"

func (p *MyPlugin) Execute(ctx context.Context, action string, params any) (any, error) {
    // 参数验证
    if action == "" {
        return nil, fmt.Errorf("action is required")
    }

    // 业务逻辑错误
    result, err := p.doSomething(ctx, action)
    if err != nil {
        return nil, fmt.Errorf("execute %s failed: %w", action, err)
    }

    return result, nil
}
```

### 3. 资源管理

```go
type MyPlugin struct {
    conn   *sql.DB
    cache  *redis.Client
    done   chan struct{}
}

func (p *MyPlugin) Init(ctx context.Context, config any) error {
    // 初始化资源
    var err error
    p.conn, err = sql.Open("mysql", "...")
    if err != nil {
        return err
    }

    p.done = make(chan struct{})

    // 启动后台任务
    go p.backgroundTask()

    return nil
}

func (p *MyPlugin) Cleanup() error {
    // 停止后台任务
    close(p.done)

    // 关闭连接
    if p.conn != nil {
        p.conn.Close()
    }

    if p.cache != nil {
        p.cache.Close()
    }

    return nil
}

func (p *MyPlugin) backgroundTask() {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()

    for {
        select {
        case <-p.done:
            return
        case <-ticker.C:
            // 定期任务
        }
    }
}
```

### 4. 并发安全

```go
import "sync"

type MyPlugin struct {
    mu    sync.RWMutex
    cache map[string]interface{}
}

func (p *MyPlugin) Get(key string) (interface{}, bool) {
    p.mu.RLock()
    defer p.mu.RUnlock()

    val, ok := p.cache[key]
    return val, ok
}

func (p *MyPlugin) Set(key string, val interface{}) {
    p.mu.Lock()
    defer p.mu.Unlock()

    p.cache[key] = val
}
```

### 5. 日志规范

```go
import "github.com/observabil/arcade/pkg/log"

func (p *MyPlugin) Execute(ctx context.Context, action string, params any) (any, error) {
    log.Infof("[%s] 开始执行: action=%s", p.Name(), action)

    result, err := p.doWork(action, params)
    if err != nil {
        log.Errorf("[%s] 执行失败: %v", p.Name(), err)
        return nil, err
    }

    log.Infof("[%s] 执行成功: result=%+v", p.Name(), result)
    return result, nil
}
```

### 6. 上下文处理

```go
func (p *MyPlugin) Execute(ctx context.Context, action string, params any) (any, error) {
    // 检查上下文取消
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }

    // 传递上下文
    result, err := p.callExternalAPI(ctx, action)
    if err != nil {
        return nil, err
    }

    return result, nil
}

func (p *MyPlugin) callExternalAPI(ctx context.Context, action string) (interface{}, error) {
    req, _ := http.NewRequestWithContext(ctx, "POST", "...", nil)
    // ...
    return nil, nil
}
```

### 7. 单元测试

```go
// myplugin_test.go
package main

import (
    "context"
    "testing"
)

func TestMyPlugin_Init(t *testing.T) {
    plugin := &MyPlugin{}

    config := map[string]interface{}{
        "api_key": "test-key",
    }

    err := plugin.Init(context.Background(), config)
    if err != nil {
        t.Fatalf("Init failed: %v", err)
    }

    if plugin.config.APIKey != "test-key" {
        t.Errorf("Expected api_key=test-key, got %s", plugin.config.APIKey)
    }
}

func TestMyPlugin_Execute(t *testing.T) {
    plugin := &MyPlugin{}
    plugin.Init(context.Background(), nil)
    defer plugin.Cleanup()

    result, err := plugin.Execute(context.Background(), "test", nil)
    if err != nil {
        t.Fatalf("Execute failed: %v", err)
    }

    // 断言结果
    // ...
}
```

运行测试：
```bash
go test -v ./pkg/plugins/...
```

## 常见问题

### Q1: 插件加载失败

**错误：**
```
plugin: symbol Plugin not found
```

**原因：**
- 没有导出 `Plugin` 变量
- 变量名称错误（必须是 `Plugin`）

**解决：**
```go
// ✅ 正确
var Plugin = &MyPlugin{}

// ❌ 错误
var plugin = &MyPlugin{}  // 小写
var MyPlugin = &MyPlugin{} // 名称错误
```

### Q2: 版本冲突

**错误：**
```
plugin was built with a different version of package
```

**解决：**
1. 使用相同的 Go 版本
2. 依赖版本必须完全一致
3. 清理缓存后重新编译

```bash
go clean -cache -modcache
go mod tidy
make plugins
```

### Q3: 插件无法卸载

**问题：**
Go 的 plugin 包不支持真正卸载插件。

**解决方案：**
1. 重启服务（最彻底）
2. 使用多进程架构
3. 接受这个限制，只在索引中移除

### Q4: 跨平台问题

**问题：**
plugin 包只支持 Linux 和 macOS。

**解决方案：**
- 开发：在 Linux/macOS 上开发
- Windows：使用 WSL2 或虚拟机

### Q5: 配置解析失败

**问题：**
配置类型断言失败。

**解决：**
```go
func (p *MyPlugin) Init(ctx context.Context, config any) error {
    if config == nil {
        return fmt.Errorf("config is required")
    }

    configMap, ok := config.(map[string]interface{})
    if !ok {
        return fmt.Errorf("invalid config type: %T", config)
    }

    // 安全地获取配置值
    if val, ok := configMap["key"]; ok {
        if strVal, ok := val.(string); ok {
            p.someField = strVal
        }
    }

    return nil
}
```

## 插件发布

### 1. 编写文档

创建插件 README：

```markdown
# MyPlugin

简短描述

## 功能

- 功能 1
- 功能 2

## 配置

\`\`\`yaml
plugins:
  - path: ./plugins/myplugin.so
    name: myplugin
    type: custom
    version: "1.0.0"
    config:
      key: value
\`\`\`

## 使用示例

...

## 更新日志

### v1.0.0
- 初始版本
```

### 2. 版本管理

使用语义化版本：

- `1.0.0` - 主版本.次版本.修订号
- 破坏性变更增加主版本
- 新功能增加次版本
- Bug 修复增加修订号

### 3. 测试清单

发布前确认：

- [ ] 所有功能正常工作
- [ ] 单元测试通过
- [ ] 配置文档完整
- [ ] 错误处理完善
- [ ] 资源正确清理
- [ ] 日志输出合理
- [ ] 性能可接受

## 更多资源

- [插件快速开始](./PLUGIN_QUICKSTART.md)
- [插件自动加载](./PLUGIN_AUTO_LOAD.md)
- [项目 README](../README.md)
- [示例插件](../pkg/plugins/)

## 获取帮助

遇到问题？

1. 查看本文档的常见问题部分
2. 查看示例插件代码
3. 提交 Issue
4. 联系维护者

祝开发愉快！🎉
