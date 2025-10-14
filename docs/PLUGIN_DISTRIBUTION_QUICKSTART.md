# 插件分发快速入门

## 核心概念

### 问题
**插件存储在服务端，但任务执行在Agent端**

### 解决方案
**按需下载 + 本地缓存 + 完整性校验**

## 工作流程（5个步骤）

```
1. Agent注册 → 上报已安装插件列表
   ↓
2. 拉取任务 → Server返回任务+插件信息
   ↓
3. 检查本地 → 验证插件是否存在且版本匹配
   ↓
4. 下载插件 → 缺失插件从Server下载
   ↓
5. 执行任务 → 加载插件并运行任务
```

## 快速实现步骤

### Step 1: 扩展 Proto 定义

在 `api/agent/v1/proto/agent.proto` 添加：

```protobuf
// 插件信息
message PluginInfo {
  string plugin_id = 1;
  string version = 2;
  string checksum = 3;
  int64 size = 4;
}

// 扩展Job消息
message Job {
  // ... 原有字段 ...
  repeated PluginInfo plugins = 14;  // 新增
}

// 下载插件
service Agent {
  rpc DownloadPlugin(DownloadPluginRequest) returns (DownloadPluginResponse) {}
}

message DownloadPluginRequest {
  string agent_id = 1;
  string plugin_id = 2;
  string version = 3;
}

message DownloadPluginResponse {
  bool success = 1;
  bytes plugin_data = 2;
  string checksum = 3;
}
```

重新生成代码：
```bash
cd api
buf generate
```

### Step 2: 服务端实现插件下载接口

```go
// internal/engine/service/agent/service_plugin.go
func (s *AgentServicePB) DownloadPlugin(ctx context.Context, req *v1.DownloadPluginRequest) (*v1.DownloadPluginResponse, error) {
    // 1. 从数据库获取插件信息
    plugin, err := s.pluginRepo.GetPluginByID(req.PluginId)
    if err != nil {
        return &v1.DownloadPluginResponse{Success: false}, nil
    }

    // 2. 读取插件文件
    pluginData, err := os.ReadFile(plugin.InstallPath)
    if err != nil {
        return &v1.DownloadPluginResponse{Success: false}, nil
    }

    // 3. 计算校验和
    hash := sha256.Sum256(pluginData)
    checksum := hex.EncodeToString(hash[:])

    // 4. 返回插件数据
    return &v1.DownloadPluginResponse{
        Success:    true,
        PluginData: pluginData,
        Checksum:   checksum,
    }, nil
}
```

### Step 3: 服务端在分发任务时附加插件信息

```go
// internal/engine/service/agent/service_agent_pb.go
func (s *AgentServicePB) FetchJob(ctx context.Context, req *v1.FetchJobRequest) (*v1.FetchJobResponse, error) {
    jobs := s.getJobsForAgent(req.AgentId)
    
    // 为每个任务附加插件信息
    for _, job := range jobs {
        jobPlugins, _ := s.pluginRepo.GetJobPlugins(job.JobId)
        
        for _, jp := range jobPlugins {
            plugin, _ := s.pluginRepo.GetPluginByID(jp.PluginId)
            job.Plugins = append(job.Plugins, &v1.PluginInfo{
                PluginId: plugin.PluginId,
                Version:  plugin.Version,
                Checksum: plugin.Checksum,
                Size:     calculateFileSize(plugin.InstallPath),
            })
        }
    }
    
    return &v1.FetchJobResponse{
        Success: true,
        Jobs:    jobs,
    }, nil
}
```

### Step 4: Agent端实现插件下载器

```go
// agent/internal/plugin/downloader.go
type PluginDownloader struct {
    client    v1.AgentClient
    pluginDir string
    cache     map[string]string  // pluginID@version -> localPath
}

func (d *PluginDownloader) EnsurePlugins(ctx context.Context, agentID string, plugins []*v1.PluginInfo) error {
    for _, plugin := range plugins {
        cacheKey := fmt.Sprintf("%s@%s", plugin.PluginId, plugin.Version)
        
        // 检查本地缓存
        if localPath, exists := d.cache[cacheKey]; exists {
            if d.verifyChecksum(localPath, plugin.Checksum) {
                log.Infof("plugin %s already cached", cacheKey)
                continue
            }
        }
        
        // 下载插件
        log.Infof("downloading plugin %s", plugin.PluginId)
        resp, err := d.client.DownloadPlugin(ctx, &v1.DownloadPluginRequest{
            AgentId:  agentID,
            PluginId: plugin.PluginId,
            Version:  plugin.Version,
        })
        if err != nil {
            return err
        }
        
        // 验证校验和
        actualChecksum := sha256.Sum256(resp.PluginData)
        if hex.EncodeToString(actualChecksum[:]) != resp.Checksum {
            return fmt.Errorf("checksum mismatch")
        }
        
        // 保存到本地
        filename := fmt.Sprintf("%s.so", cacheKey)
        localPath := filepath.Join(d.pluginDir, "downloaded", filename)
        os.WriteFile(localPath, resp.PluginData, 0755)
        
        // 更新缓存
        d.cache[cacheKey] = localPath
        log.Infof("plugin %s downloaded successfully", plugin.PluginId)
    }
    
    return nil
}

func (d *PluginDownloader) verifyChecksum(filePath string, expectedChecksum string) bool {
    data, _ := os.ReadFile(filePath)
    actualChecksum := sha256.Sum256(data)
    return hex.EncodeToString(actualChecksum[:]) == expectedChecksum
}
```

### Step 5: Agent集成到任务执行流程

```go
// agent/internal/executor/job_executor.go
func (e *JobExecutor) ExecuteJob(ctx context.Context, job *v1.Job) error {
    // 1. 确保插件就绪
    if len(job.Plugins) > 0 {
        if err := e.pluginDownloader.EnsurePlugins(ctx, e.agentID, job.Plugins); err != nil {
            return fmt.Errorf("failed to prepare plugins: %w", err)
        }
        
        // 2. 加载插件
        for _, plugin := range job.Plugins {
            cacheKey := fmt.Sprintf("%s@%s", plugin.PluginId, plugin.Version)
            pluginPath := e.pluginDownloader.GetPath(cacheKey)
            e.pluginManager.Register(pluginPath, cacheKey, nil)
        }
        
        // 3. 初始化插件
        e.pluginManager.Init(ctx)
    }
    
    // 4. 执行before插件
    e.executePluginStage(ctx, job, "before")
    
    // 5. 执行任务命令
    jobErr := e.executeCommands(ctx, job)
    
    // 6. 执行after插件
    if jobErr != nil {
        e.executePluginStage(ctx, job, "on_failure")
    } else {
        e.executePluginStage(ctx, job, "on_success")
    }
    
    e.executePluginStage(ctx, job, "after")
    
    return jobErr
}
```

## 目录结构

### 服务端

```
internal/engine/
├── service/
│   └── agent/
│       ├── service_agent.go           # Agent基础服务
│       ├── service_agent_pb.go        # gRPC实现
│       └── service_plugin.go          # 插件下载服务（新增）
├── repo/
│   └── repo_plugin.go                 # 插件仓库
└── model/
    └── model_plugin.go                # 插件模型
```

### Agent端（独立项目）

```
agent/
├── internal/
│   ├── plugin/
│   │   ├── downloader.go              # 插件下载器（新增）
│   │   └── manager.go                 # 插件管理器
│   ├── executor/
│   │   └── job_executor.go            # 任务执行器（修改）
│   └── config/
│       └── config.go
└── main.go

本地插件目录：
~/.arcade/plugins/
├── builtin/                           # 预装插件
│   ├── notify-stdout.so
│   └── build-docker.so
├── downloaded/                        # 下载的插件
│   ├── notify-slack@1.0.0.so
│   └── deploy-k8s@2.1.0.so
└── temp/                              # 临时文件
```

## 配置示例

### 服务端 `config.toml`

```toml
[plugin]
storage_path = "./plugins"
max_plugin_size = 100  # MB
require_checksum = true
```

### Agent 配置

```yaml
agent:
  id: "agent-001"
  server: "server.example.com:9090"
  
  plugins:
    directory: "/var/lib/arcade/plugins"
    download_timeout: 300  # 秒
    cache_size: 1000       # MB
    
    # 预装插件
    builtin:
      - "notify/stdout"
      - "build/docker"
```

## 数据库表设计

### 任务插件关联表

```sql
CREATE TABLE t_job_plugin (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    job_id VARCHAR(64) NOT NULL,
    plugin_id VARCHAR(128) NOT NULL,
    plugin_config_id VARCHAR(64),
    execution_order INT DEFAULT 0,
    execution_stage VARCHAR(32),  -- before/after/on_success/on_failure
    status INT DEFAULT 0,         -- 0:未执行 1:执行中 2:成功 3:失败
    result TEXT,
    error_message TEXT,
    started_at DATETIME,
    completed_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_job_id (job_id),
    INDEX idx_plugin_id (plugin_id)
);
```

## 监控指标

```go
// Prometheus指标
plugin_download_total{plugin_id, version, status}       // 下载次数
plugin_download_duration_seconds{plugin_id}             // 下载耗时
plugin_download_size_bytes{plugin_id}                   // 下载大小
plugin_cache_hit_total{plugin_id}                       // 缓存命中
plugin_cache_miss_total{plugin_id}                      // 缓存未命中
```

## 故障排查

### 问题1：插件下载失败

```bash
# 检查日志
tail -f /var/log/arcade/agent.log | grep "plugin"

# 常见原因：
1. 网络连接问题 → 检查Server连接
2. 插件文件不存在 → 检查数据库中的install_path
3. 权限问题 → 检查文件权限
4. 磁盘空间不足 → 检查磁盘使用情况
```

### 问题2：校验和验证失败

```bash
# 原因：
1. 插件文件被篡改
2. 传输过程中数据损坏
3. 数据库中的checksum不正确

# 解决：
# 重新生成校验和
cd /path/to/plugins
sha256sum plugin.so

# 更新数据库
UPDATE t_plugin SET checksum='新的校验和' WHERE plugin_id='xxx';
```

### 问题3：插件加载失败

```bash
# 原因：
1. Go版本不匹配
2. 依赖版本不一致
3. 插件接口未正确实现

# 解决：
# 检查Go版本
go version

# 重新编译插件
go build -buildmode=plugin -o plugin.so plugin.go
```

## API调用示例

### Agent注册时上报插件

```go
resp, err := client.Register(ctx, &v1.RegisterRequest{
    AgentId:  "agent-001",
    Hostname: "worker-01",
    InstalledPlugins: []string{
        "notify/stdout@1.0.0",
        "build/docker@2.0.0",
    },
})
```

### 获取任务和插件信息

```go
resp, err := client.FetchJob(ctx, &v1.FetchJobRequest{
    AgentId: "agent-001",
    MaxJobs: 1,
})

for _, job := range resp.Jobs {
    fmt.Printf("Job: %s\n", job.JobId)
    for _, plugin := range job.Plugins {
        fmt.Printf("  Plugin: %s v%s (checksum: %s)\n", 
            plugin.PluginId, plugin.Version, plugin.Checksum[:8])
    }
}
```

### 下载插件

```go
resp, err := client.DownloadPlugin(ctx, &v1.DownloadPluginRequest{
    AgentId:  "agent-001",
    PluginId: "notify/slack",
    Version:  "1.0.0",
})

if resp.Success {
    // 保存插件文件
    os.WriteFile("/path/to/plugin.so", resp.PluginData, 0755)
}
```

## 性能优化建议

### 1. 并发下载

```go
// 使用goroutine并发下载多个插件
var wg sync.WaitGroup
for _, plugin := range plugins {
    wg.Add(1)
    go func(p *v1.PluginInfo) {
        defer wg.Done()
        d.downloadPlugin(ctx, agentID, p)
    }(plugin)
}
wg.Wait()
```

### 2. 压缩传输

```go
// 使用gzip压缩插件数据
func compressPluginData(data []byte) []byte {
    var buf bytes.Buffer
    gz := gzip.NewWriter(&buf)
    gz.Write(data)
    gz.Close()
    return buf.Bytes()
}
```

### 3. 增量更新

```go
// 只传输变化的部分（类似rsync）
// 适用于大型插件的小幅更新
```

### 4. CDN分发

```yaml
# 将插件上传到CDN
plugin:
  cdn_enabled: true
  cdn_url: "https://cdn.example.com/plugins"
```

## 安全最佳实践

1. **强制校验和验证** - 防止插件被篡改
2. **访问控制** - 限制Agent只能下载授权的插件
3. **加密传输** - 使用TLS加密gRPC通信
4. **插件签名** - 使用GPG签名验证插件来源（可选）
5. **沙箱执行** - 在隔离环境中执行插件（可选）

## 测试

### 单元测试

```bash
# 测试插件下载器
go test -v ./agent/internal/plugin/...

# 测试服务端插件服务
go test -v ./internal/engine/service/agent/...
```

### 集成测试

```bash
# 启动测试Server
make test-server

# 启动测试Agent
make test-agent

# 运行集成测试
go test -v -tags=integration ./tests/...
```

## 总结

### 核心优势
✅ **按需下载** - 只下载需要的插件  
✅ **本地缓存** - 避免重复下载  
✅ **完整性校验** - 确保插件安全  
✅ **热更新** - 支持插件版本升级  
✅ **灵活配置** - 预装+下载混合模式  

### 适用场景
- ✅ 大量Agent节点
- ✅ 插件频繁更新
- ✅ 网络条件良好
- ✅ 需要集中管理插件

### 下一步
1. 扩展Proto定义
2. 实现服务端下载接口
3. 开发Agent下载器
4. 编写测试用例
5. 部署验证

## 参考资料

- [完整设计文档](./PLUGIN_DISTRIBUTION.md)
- [插件开发指南](./PLUGIN_DEVELOPMENT.md)
- [Agent架构文档](./AGENT_ARCHITECTURE.md)

