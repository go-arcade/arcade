# 插件开发快速参考

这是一个快速参考手册，包含插件开发的常用代码片段和命令。

## 插件类型速查

| 类型 | 常量 | 主要方法 | 使用场景 |
|------|------|----------|----------|
| CI | `plugin.TypeCI` | `Build()`, `Test()`, `Lint()` | 构建、测试、代码检查 |
| CD | `plugin.TypeCD` | `Deploy()`, `Rollback()` | 部署、回滚 |
| Security | `plugin.TypeSecurity` | `Scan()`, `Audit()` | 安全扫描、审计 |
| Notify | `plugin.TypeNotify` | `Send()`, `SendTemplate()` | 消息通知 |
| Storage | `plugin.TypeStorage` | `Save()`, `Load()`, `Delete()` | 数据存储 |
| Custom | `plugin.TypeCustom` | `Execute()` | 自定义功能 |

## 插件模板

### 最小插件模板

```go
package main

import "context"

type MinimalPlugin struct{}

func (p *MinimalPlugin) Name() string        { return "minimal" }
func (p *MinimalPlugin) Description() string { return "最小插件示例" }
func (p *MinimalPlugin) Version() string     { return "1.0.0" }
func (p *MinimalPlugin) Type() string        { return "custom" }
func (p *MinimalPlugin) Init(ctx context.Context, config any) error { return nil }
func (p *MinimalPlugin) Cleanup() error { return nil }
func (p *MinimalPlugin) Execute(ctx context.Context, action string, params any) (any, error) {
    return "OK", nil
}

var Plugin = &MinimalPlugin{}
```

### 完整插件模板

```go
package main

import (
    "context"
    "fmt"
    "sync"
)

type FullPlugin struct {
    mu     sync.RWMutex
    config *Config
    client *Client
    done   chan struct{}
}

type Config struct {
    APIKey  string `mapstructure:"api_key"`
    Timeout int    `mapstructure:"timeout"`
}

func (p *FullPlugin) Name() string        { return "full-example" }
func (p *FullPlugin) Description() string { return "完整插件示例" }
func (p *FullPlugin) Version() string     { return "1.0.0" }
func (p *FullPlugin) Type() string        { return "custom" }

func (p *FullPlugin) Init(ctx context.Context, config any) error {
    p.config = &Config{Timeout: 30}
    
    if cfg, ok := config.(map[string]interface{}); ok {
        if apiKey, ok := cfg["api_key"].(string); ok {
            p.config.APIKey = apiKey
        }
        if timeout, ok := cfg["timeout"].(float64); ok {
            p.config.Timeout = int(timeout)
        }
    }
    
    p.client = NewClient(p.config)
    p.done = make(chan struct{})
    
    return nil
}

func (p *FullPlugin) Cleanup() error {
    close(p.done)
    if p.client != nil {
        p.client.Close()
    }
    return nil
}

func (p *FullPlugin) Execute(ctx context.Context, action string, params any) (any, error) {
    p.mu.RLock()
    defer p.mu.RUnlock()
    
    return p.client.Do(ctx, action, params)
}

var Plugin = &FullPlugin{}
```

## 常用代码片段

### 配置解析

```go
func (p *MyPlugin) Init(ctx context.Context, config any) error {
    cfg, ok := config.(map[string]interface{})
    if !ok {
        return fmt.Errorf("invalid config type")
    }
    
    // 字符串
    if val, ok := cfg["api_key"].(string); ok {
        p.apiKey = val
    }
    
    // 整数
    if val, ok := cfg["timeout"].(float64); ok {
        p.timeout = int(val)
    }
    
    // 布尔值
    if val, ok := cfg["enabled"].(bool); ok {
        p.enabled = val
    }
    
    // 数组
    if val, ok := cfg["hosts"].([]interface{}); ok {
        for _, v := range val {
            if host, ok := v.(string); ok {
                p.hosts = append(p.hosts, host)
            }
        }
    }
    
    // 嵌套对象
    if val, ok := cfg["auth"].(map[string]interface{}); ok {
        if username, ok := val["username"].(string); ok {
            p.username = username
        }
    }
    
    return nil
}
```

### HTTP 客户端

```go
import (
    "bytes"
    "context"
    "encoding/json"
    "net/http"
    "time"
)

type HTTPPlugin struct {
    client *http.Client
    baseURL string
}

func (p *HTTPPlugin) Init(ctx context.Context, config any) error {
    p.client = &http.Client{
        Timeout: 30 * time.Second,
    }
    
    cfg, _ := config.(map[string]interface{})
    p.baseURL, _ = cfg["base_url"].(string)
    
    return nil
}

func (p *HTTPPlugin) Cleanup() error {
    if p.client != nil {
        p.client.CloseIdleConnections()
    }
    return nil
}

func (p *HTTPPlugin) post(ctx context.Context, path string, data interface{}) error {
    body, _ := json.Marshal(data)
    req, _ := http.NewRequestWithContext(ctx, "POST", p.baseURL+path, bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := p.client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode >= 400 {
        return fmt.Errorf("HTTP %d", resp.StatusCode)
    }
    
    return nil
}
```

### 数据库连接

```go
import (
    "context"
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
)

type DBPlugin struct {
    db *sql.DB
}

func (p *DBPlugin) Init(ctx context.Context, config any) error {
    cfg, _ := config.(map[string]interface{})
    dsn, _ := cfg["dsn"].(string)
    
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return err
    }
    
    db.SetMaxOpenConns(10)
    db.SetMaxIdleConns(5)
    
    if err := db.PingContext(ctx); err != nil {
        return err
    }
    
    p.db = db
    return nil
}

func (p *DBPlugin) Cleanup() error {
    if p.db != nil {
        return p.db.Close()
    }
    return nil
}
```

### 后台任务

```go
type BackgroundPlugin struct {
    done   chan struct{}
    ticker *time.Ticker
}

func (p *BackgroundPlugin) Init(ctx context.Context, config any) error {
    p.done = make(chan struct{})
    p.ticker = time.NewTicker(1 * time.Minute)
    
    go p.backgroundJob()
    
    return nil
}

func (p *BackgroundPlugin) Cleanup() error {
    if p.ticker != nil {
        p.ticker.Stop()
    }
    close(p.done)
    return nil
}

func (p *BackgroundPlugin) backgroundJob() {
    for {
        select {
        case <-p.done:
            return
        case <-p.ticker.C:
            // 执行定期任务
            p.doWork()
        }
    }
}
```

### 并发安全

```go
import "sync"

type SafePlugin struct {
    mu    sync.RWMutex
    cache map[string]interface{}
}

func (p *SafePlugin) Init(ctx context.Context, config any) error {
    p.cache = make(map[string]interface{})
    return nil
}

func (p *SafePlugin) Get(key string) (interface{}, bool) {
    p.mu.RLock()
    defer p.mu.RUnlock()
    val, ok := p.cache[key]
    return val, ok
}

func (p *SafePlugin) Set(key string, val interface{}) {
    p.mu.Lock()
    defer p.mu.Unlock()
    p.cache[key] = val
}
```

### 错误处理

```go
import "fmt"

func (p *MyPlugin) Execute(ctx context.Context, action string, params any) (any, error) {
    // 参数验证
    if action == "" {
        return nil, fmt.Errorf("action is required")
    }
    
    // 类型检查
    data, ok := params.(map[string]interface{})
    if !ok {
        return nil, fmt.Errorf("invalid params type: expected map, got %T", params)
    }
    
    // 业务逻辑
    result, err := p.doSomething(data)
    if err != nil {
        return nil, fmt.Errorf("execute action %s: %w", action, err)
    }
    
    return result, nil
}
```

## 常用命令

### 编译

```bash
# 编译单个插件
go build -buildmode=plugin -o plugins/myplugin.so myplugin.go

# 编译所有插件
make plugins

# 清理插件
make plugins-clean
```

### 测试

```bash
# 运行测试
go test -v ./pkg/plugins/myplugin/

# 测试覆盖率
go test -cover ./pkg/plugins/myplugin/

# 生成覆盖率报告
go test -coverprofile=coverage.out ./pkg/plugins/myplugin/
go tool cover -html=coverage.out
```

### 调试

```bash
# 查看符号
go tool nm plugins/myplugin.so | grep Plugin

# 运行示例
go run examples/plugin_autowatch/main.go

# 查看日志
tail -f logs/arcade.log
```

## 配置示例

### 基础配置

```yaml
plugins:
  - path: ./plugins/myplugin.so
    name: myplugin
    type: notify
    version: "1.0.0"
    config:
      api_key: "xxx"
      timeout: 30
```

### 复杂配置

```yaml
plugins:
  - path: ./plugins/advanced.so
    name: advanced
    type: custom
    version: "2.0.0"
    config:
      enabled: true
      hosts:
        - host1.example.com
        - host2.example.com
      auth:
        username: admin
        password: secret
      settings:
        retry_max: 3
        retry_delay: 1000
```

## 接口实现速查

### Notify 插件

```go
func (p *NotifyPlugin) Send(ctx context.Context, message any) error {
    msg, _ := message.(string)
    return p.sendMessage(ctx, msg)
}

func (p *NotifyPlugin) SendTemplate(ctx context.Context, template string, data any) error {
    content := p.renderTemplate(template, data)
    return p.sendMessage(ctx, content)
}
```

### CI 插件

```go
func (p *CIPlugin) Build(ctx context.Context, projectConfig any) error {
    return p.runCommand(ctx, "build")
}

func (p *CIPlugin) Test(ctx context.Context, projectConfig any) error {
    return p.runCommand(ctx, "test")
}

func (p *CIPlugin) Lint(ctx context.Context, projectConfig any) error {
    return p.runCommand(ctx, "lint")
}
```

### Storage 插件

```go
func (p *StoragePlugin) Save(ctx context.Context, key string, data any) error {
    return p.store.Put(ctx, key, data)
}

func (p *StoragePlugin) Load(ctx context.Context, key string) (any, error) {
    return p.store.Get(ctx, key)
}

func (p *StoragePlugin) Delete(ctx context.Context, key string) error {
    return p.store.Delete(ctx, key)
}
```

## 常见错误

### 编译错误

| 错误信息 | 原因 | 解决方法 |
|---------|------|----------|
| `plugin was built with a different version` | Go 版本不一致 | 使用相同 Go 版本 |
| `symbol Plugin not found` | 未导出 Plugin 变量 | 添加 `var Plugin = ...` |
| `package X is not in GOROOT` | 依赖缺失 | 运行 `go mod tidy` |

### 运行时错误

| 错误信息 | 原因 | 解决方法 |
|---------|------|----------|
| `plugin already registered` | 插件名称重复 | 使用唯一的插件名称 |
| `invalid config type` | 配置类型错误 | 检查配置文件格式 |
| `context canceled` | 上下文被取消 | 正确处理 context |

## 性能优化

### 连接池

```go
import "sync"

type PoolPlugin struct {
    pool sync.Pool
}

func (p *PoolPlugin) Init(ctx context.Context, config any) error {
    p.pool = sync.Pool{
        New: func() interface{} {
            return &Connection{/* ... */}
        },
    }
    return nil
}

func (p *PoolPlugin) Execute(ctx context.Context, action string, params any) (any, error) {
    conn := p.pool.Get().(*Connection)
    defer p.pool.Put(conn)
    
    return conn.Execute(action, params)
}
```

### 缓存

```go
import (
    "sync"
    "time"
)

type CachedPlugin struct {
    cache map[string]*cacheItem
    mu    sync.RWMutex
}

type cacheItem struct {
    value      interface{}
    expireTime time.Time
}

func (p *CachedPlugin) get(key string) (interface{}, bool) {
    p.mu.RLock()
    defer p.mu.RUnlock()
    
    item, ok := p.cache[key]
    if !ok || time.Now().After(item.expireTime) {
        return nil, false
    }
    
    return item.value, true
}

func (p *CachedPlugin) set(key string, value interface{}, ttl time.Duration) {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    p.cache[key] = &cacheItem{
        value:      value,
        expireTime: time.Now().Add(ttl),
    }
}
```

## 安全建议

1. **验证输入**
```go
func (p *Plugin) Execute(ctx context.Context, action string, params any) (any, error) {
    if action == "" {
        return nil, fmt.Errorf("action is required")
    }
    // 验证参数...
}
```

2. **使用 HTTPS**
```go
client := &http.Client{
    Transport: &http.Transport{
        TLSClientConfig: &tls.Config{
            MinVersion: tls.VersionTLS12,
        },
    },
}
```

3. **不在日志中输出敏感信息**
```go
// ❌ 错误
log.Infof("API Key: %s", apiKey)

// ✅ 正确
log.Info("API Key configured")
```

4. **使用上下文超时**
```go
ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
defer cancel()
```

## 更多资源

- [完整开发指南](./PLUGIN_DEVELOPMENT.md)
- [快速开始](./PLUGIN_QUICKSTART.md)
- [自动加载文档](./PLUGIN_AUTO_LOAD.md)

