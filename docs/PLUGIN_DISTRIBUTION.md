# 插件分发架构设计文档

## 问题描述

**插件存储在服务端，但任务执行在Agent端，如何实现插件的分发和执行？**

## 架构方案对比

### 方案一：按需下载（推荐）

**原理**：Agent在执行任务前，从服务端下载所需插件

**优点**：
- ✅ 插件版本实时更新
- ✅ Agent轻量化，不需要预装所有插件
- ✅ 支持插件热更新
- ✅ 集中管理，安全可控

**缺点**：
- ❌ 首次执行需要下载时间
- ❌ 需要网络连接

### 方案二：对象存储分发

**原理**：插件上传到S3/OSS等对象存储，Agent从对象存储下载

**优点**：
- ✅ 减轻服务端压力
- ✅ 支持CDN加速
- ✅ 高可用性

**缺点**：
- ❌ 依赖外部存储服务
- ❌ 增加成本

### 方案三：镜像预装

**原理**：将插件打包到Agent Docker镜像中

**优点**：
- ✅ 执行速度快，无需下载
- ✅ 离线环境可用

**缺点**：
- ❌ 更新插件需要重新打包镜像
- ❌ 镜像体积大
- ❌ 灵活性差

### 方案四：混合方案（最佳实践）

**原理**：常用插件预装 + 按需下载 + 本地缓存

**优点**：
- ✅ 综合了前三种方案的优点
- ✅ 性能好，灵活性高
- ✅ 支持离线和在线模式

**缺点**：
- ❌ 实现复杂度较高

## 推荐方案：混合架构

### 整体架构

```
┌─────────────────────────────────────────────────────────┐
│                      服务端 (Server)                      │
│                                                           │
│  ┌─────────────┐    ┌──────────────┐   ┌──────────────┐ │
│  │  Plugin DB  │    │ Plugin Files │   │ Object Store │ │
│  │  (元数据)    │    │  (.so files) │   │ (可选)       │ │
│  └─────────────┘    └──────────────┘   └──────────────┘ │
│         │                   │                   │         │
└─────────┼───────────────────┼───────────────────┼─────────┘
          │                   │                   │
          │ gRPC              │ HTTP/gRPC         │
          ▼                   ▼                   ▼
┌─────────────────────────────────────────────────────────┐
│                      Agent 端                             │
│                                                           │
│  ┌──────────────┐   ┌─────────────┐   ┌──────────────┐ │
│  │  Plugin      │   │   Plugin    │   │    Job       │ │
│  │  Downloader  │──▶│   Cache     │──▶│   Executor   │ │
│  └──────────────┘   └─────────────┘   └──────────────┘ │
│                                                           │
│  插件存储路径: ~/.arcade/plugins/                          │
│  ├── builtin/          (预装插件)                          │
│  ├── downloaded/       (下载的插件)                         │
│  └── temp/             (临时文件)                          │
└─────────────────────────────────────────────────────────┘
```

### 工作流程

```
1. Agent注册时，上报预装插件列表
   ┌───────┐                ┌────────┐
   │ Agent │───Register────▶│ Server │
   └───────┘                └────────┘
   plugins: ["notify/stdout", "build/docker"]

2. Server分配任务，携带插件信息
   ┌────────┐                ┌───────┐
   │ Server │───FetchJob────▶│ Agent │
   └────────┘                └───────┘
   job: {
     plugins: [
       {id: "notify/slack", version: "1.0.0", checksum: "abc123"},
       {id: "build/maven", version: "2.1.0", checksum: "def456"}
     ]
   }

3. Agent检查本地插件
   ┌───────┐      ┌──────────┐
   │ Agent │─────▶│ 本地缓存  │
   └───────┘      └──────────┘
   - notify/slack v1.0.0: 不存在 ❌
   - build/maven v2.1.0: 已缓存 ✅

4. Agent下载缺失插件
   ┌───────┐                ┌────────┐
   │ Agent │───Download────▶│ Server │
   └───────┘                └────────┘
   请求: plugin_id="notify/slack", version="1.0.0"

5. Agent验证插件完整性
   ┌───────┐
   │ Agent │
   └───┬───┘
       │ 计算SHA256
       │ 对比checksum
       └──▶ ✅ 验证通过

6. Agent加载插件并执行任务
   ┌───────┐      ┌────────┐      ┌──────┐
   │ Agent │─────▶│ Plugin │─────▶│ Job  │
   └───────┘      └────────┘      └──────┘
```

## API 设计

### 1. 扩展 Agent Proto 定义

在 `api/agent/v1/proto/agent.proto` 中添加：

```protobuf
// 插件信息
message PluginInfo {
  string plugin_id = 1;              // 插件ID
  string name = 2;                   // 插件名称
  string version = 3;                // 版本号
  string checksum = 4;               // SHA256校验和
  int64 size = 5;                    // 文件大小（字节）
  string download_url = 6;           // 下载地址（可选）
  PluginLocation location = 7;       // 插件位置
}

// 插件位置
enum PluginLocation {
  PLUGIN_LOCATION_UNKNOWN = 0;
  PLUGIN_LOCATION_SERVER = 1;        // 服务端文件系统
  PLUGIN_LOCATION_STORAGE = 2;       // 对象存储
  PLUGIN_LOCATION_REGISTRY = 3;      // 插件仓库
}

// 扩展 RegisterRequest
message RegisterRequest {
  // ... 原有字段 ...
  repeated string installed_plugins = 10;  // Agent已安装的插件列表
}

// 扩展 Job 消息
message Job {
  // ... 原有字段 ...
  repeated PluginInfo plugins = 14;  // 任务所需插件列表
}

// 新增：下载插件请求
message DownloadPluginRequest {
  string agent_id = 1;
  string plugin_id = 2;
  string version = 3;
}

// 新增：下载插件响应
message DownloadPluginResponse {
  bool success = 1;
  string message = 2;
  bytes plugin_data = 3;             // 插件二进制数据
  string checksum = 4;               // SHA256校验和
  int64 size = 5;                    // 文件大小
}

// 新增：Agent服务扩展
service Agent {
  // ... 原有方法 ...
  
  // 下载插件
  rpc DownloadPlugin(DownloadPluginRequest) returns (DownloadPluginResponse) {}
  
  // 列出可用插件
  rpc ListAvailablePlugins(ListAvailablePluginsRequest) returns (ListAvailablePluginsResponse) {}
}

// 列出可用插件请求
message ListAvailablePluginsRequest {
  string agent_id = 1;
  string plugin_type = 2;            // 可选：按类型过滤
}

// 列出可用插件响应
message ListAvailablePluginsResponse {
  bool success = 1;
  string message = 2;
  repeated PluginInfo plugins = 3;
}
```

### 2. Job Proto 扩展

在 `api/job/v1/proto/job.proto` 中添加：

```protobuf
// 扩展 CreateJobRequest
message CreateJobRequest {
  // ... 原有字段 ...
  repeated string required_plugins = 16;  // 所需插件ID列表
}

// 扩展 JobDetail
message JobDetail {
  // ... 原有字段 ...
  repeated PluginExecutionResult plugin_results = 27;  // 插件执行结果
}

// 插件执行结果
message PluginExecutionResult {
  string plugin_id = 1;
  string plugin_name = 2;
  string version = 3;
  string stage = 4;                  // before/after/on_success/on_failure
  bool success = 5;
  string error_message = 6;
  int64 start_time = 7;
  int64 end_time = 8;
  string result = 9;                 // 执行结果（JSON）
}
```

## 服务端实现

### 1. 插件下载服务

创建 `internal/engine/service/agent/service_plugin.go`:

```go
package agent

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	v1 "github.com/observabil/arcade/api/agent/v1"
	"github.com/observabil/arcade/internal/engine/model"
	"github.com/observabil/arcade/internal/engine/repo"
	"github.com/observabil/arcade/pkg/ctx"
	"github.com/observabil/arcade/pkg/log"
)

type PluginDownloadService struct {
	pluginRepo *repo.PluginRepo
	ctx        *ctx.Context
}

func NewPluginDownloadService(ctx *ctx.Context, pluginRepo *repo.PluginRepo) *PluginDownloadService {
	return &PluginDownloadService{
		pluginRepo: pluginRepo,
		ctx:        ctx,
	}
}

// GetPluginForDownload 获取插件用于下载
func (s *PluginDownloadService) GetPluginForDownload(pluginID, version string) (*v1.DownloadPluginResponse, error) {
	// 1. 从数据库获取插件信息
	plugin, err := s.pluginRepo.GetPluginByID(pluginID)
	if err != nil {
		return &v1.DownloadPluginResponse{
			Success: false,
			Message: fmt.Sprintf("plugin not found: %v", err),
		}, nil
	}

	// 2. 检查版本
	if version != "" && plugin.Version != version {
		return &v1.DownloadPluginResponse{
			Success: false,
			Message: fmt.Sprintf("version mismatch: requested %s, available %s", version, plugin.Version),
		}, nil
	}

	// 3. 读取插件文件
	pluginPath := plugin.InstallPath
	if !filepath.IsAbs(pluginPath) {
		workDir, _ := os.Getwd()
		pluginPath = filepath.Join(workDir, pluginPath)
	}

	pluginData, err := os.ReadFile(pluginPath)
	if err != nil {
		log.Errorf("failed to read plugin file %s: %v", pluginPath, err)
		return &v1.DownloadPluginResponse{
			Success: false,
			Message: fmt.Sprintf("failed to read plugin file: %v", err),
		}, nil
	}

	// 4. 计算校验和
	hash := sha256.Sum256(pluginData)
	checksum := hex.EncodeToString(hash[:])

	// 5. 验证校验和（如果数据库中有记录）
	if plugin.Checksum != "" && plugin.Checksum != checksum {
		log.Warnf("plugin checksum mismatch: db=%s, file=%s", plugin.Checksum, checksum)
	}

	log.Infof("plugin %s (v%s) ready for download, size=%d bytes", pluginID, plugin.Version, len(pluginData))

	return &v1.DownloadPluginResponse{
		Success:    true,
		Message:    "plugin downloaded successfully",
		PluginData: pluginData,
		Checksum:   checksum,
		Size:       int64(len(pluginData)),
	}, nil
}

// ListAvailablePlugins 列出可用插件
func (s *PluginDownloadService) ListAvailablePlugins(pluginType string) ([]*v1.PluginInfo, error) {
	var plugins []model.Plugin
	var err error

	if pluginType != "" {
		plugins, err = s.pluginRepo.GetPluginsByType(pluginType)
	} else {
		plugins, err = s.pluginRepo.GetEnabledPlugins()
	}

	if err != nil {
		return nil, err
	}

	result := make([]*v1.PluginInfo, 0, len(plugins))
	for _, p := range plugins {
		result = append(result, &v1.PluginInfo{
			PluginId: p.PluginId,
			Name:     p.Name,
			Version:  p.Version,
			Checksum: p.Checksum,
			Location: v1.PluginLocation_PLUGIN_LOCATION_SERVER,
		})
	}

	return result, nil
}

// GetPluginsForJob 获取任务所需的插件列表
func (s *PluginDownloadService) GetPluginsForJob(jobID string) ([]*v1.PluginInfo, error) {
	// 从任务插件关联表获取
	jobPlugins, err := s.pluginRepo.GetJobPlugins(jobID)
	if err != nil {
		return nil, err
	}

	result := make([]*v1.PluginInfo, 0, len(jobPlugins))
	for _, jp := range jobPlugins {
		plugin, err := s.pluginRepo.GetPluginByID(jp.PluginId)
		if err != nil {
			log.Warnf("failed to get plugin %s: %v", jp.PluginId, err)
			continue
		}

		result = append(result, &v1.PluginInfo{
			PluginId: plugin.PluginId,
			Name:     plugin.Name,
			Version:  plugin.Version,
			Checksum: plugin.Checksum,
			Location: v1.PluginLocation_PLUGIN_LOCATION_SERVER,
		})
	}

	return result, nil
}
```

### 2. 扩展 Agent gRPC 服务

在 `internal/engine/service/agent/service_agent_pb.go` 中添加：

```go
// DownloadPlugin 实现插件下载接口
func (s *AgentServicePB) DownloadPlugin(ctx context.Context, req *v1.DownloadPluginRequest) (*v1.DownloadPluginResponse, error) {
	log.Infof("agent %s requesting plugin: %s (v%s)", req.AgentId, req.PluginId, req.Version)

	// 调用插件下载服务
	resp, err := s.pluginDownloadService.GetPluginForDownload(req.PluginId, req.Version)
	if err != nil {
		log.Errorf("failed to get plugin for download: %v", err)
		return &v1.DownloadPluginResponse{
			Success: false,
			Message: fmt.Sprintf("failed to get plugin: %v", err),
		}, nil
	}

	return resp, nil
}

// ListAvailablePlugins 列出可用插件
func (s *AgentServicePB) ListAvailablePlugins(ctx context.Context, req *v1.ListAvailablePluginsRequest) (*v1.ListAvailablePluginsResponse, error) {
	plugins, err := s.pluginDownloadService.ListAvailablePlugins(req.PluginType)
	if err != nil {
		return &v1.ListAvailablePluginsResponse{
			Success: false,
			Message: fmt.Sprintf("failed to list plugins: %v", err),
		}, nil
	}

	return &v1.ListAvailablePluginsResponse{
		Success: true,
		Message: "success",
		Plugins: plugins,
	}, nil
}
```

### 3. FetchJob 返回插件信息

修改 `internal/engine/service/agent/service_agent_pb.go` 中的 `FetchJob`:

```go
func (s *AgentServicePB) FetchJob(ctx context.Context, req *v1.FetchJobRequest) (*v1.FetchJobResponse, error) {
	// ... 原有逻辑 ...

	// 为每个任务附加插件信息
	for _, job := range jobs {
		plugins, err := s.pluginDownloadService.GetPluginsForJob(job.JobId)
		if err != nil {
			log.Warnf("failed to get plugins for job %s: %v", job.JobId, err)
			continue
		}
		job.Plugins = plugins
	}

	return &v1.FetchJobResponse{
		Success: true,
		Message: "success",
		Jobs:    jobs,
	}, nil
}
```

## Agent 端实现

### 1. 插件下载管理器

创建 Agent 端插件管理器（这是Agent项目的代码，需要单独维护）:

```go
// agent/internal/plugin/downloader.go
package plugin

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	v1 "github.com/observabil/arcade/api/agent/v1"
	"github.com/observabil/arcade/pkg/log"
)

type PluginDownloader struct {
	agentClient v1.AgentClient
	pluginDir   string
	cache       map[string]*CachedPlugin // key: pluginID@version
	mu          sync.RWMutex
}

type CachedPlugin struct {
	PluginID  string
	Version   string
	Path      string
	Checksum  string
	Size      int64
	CachedAt  int64
}

func NewPluginDownloader(agentClient v1.AgentClient, pluginDir string) *PluginDownloader {
	return &PluginDownloader{
		agentClient: agentClient,
		pluginDir:   pluginDir,
		cache:       make(map[string]*CachedPlugin),
	}
}

// Init 初始化插件目录
func (d *PluginDownloader) Init() error {
	// 创建目录结构
	dirs := []string{
		filepath.Join(d.pluginDir, "builtin"),
		filepath.Join(d.pluginDir, "downloaded"),
		filepath.Join(d.pluginDir, "temp"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create dir %s: %w", dir, err)
		}
	}

	// 扫描已缓存的插件
	return d.scanCachedPlugins()
}

// scanCachedPlugins 扫描本地缓存
func (d *PluginDownloader) scanCachedPlugins() error {
	downloadedDir := filepath.Join(d.pluginDir, "downloaded")
	entries, err := os.ReadDir(downloadedDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || !filepath.Ext(entry.Name()) == ".so" {
			continue
		}

		pluginPath := filepath.Join(downloadedDir, entry.Name())
		info, _ := entry.Info()
		
		// 计算校验和
		checksum, err := calculateChecksum(pluginPath)
		if err != nil {
			log.Warnf("failed to calculate checksum for %s: %v", pluginPath, err)
			continue
		}

		// 解析插件ID和版本（假设文件名格式：pluginID@version.so）
		// 例如：notify-slack@1.0.0.so
		name := entry.Name()[:len(entry.Name())-3] // 去掉.so
		
		cached := &CachedPlugin{
			Path:     pluginPath,
			Checksum: checksum,
			Size:     info.Size(),
			CachedAt: info.ModTime().Unix(),
		}

		d.mu.Lock()
		d.cache[name] = cached
		d.mu.Unlock()

		log.Infof("found cached plugin: %s (checksum=%s)", name, checksum[:8])
	}

	return nil
}

// EnsurePlugins 确保所有需要的插件都已就绪
func (d *PluginDownloader) EnsurePlugins(ctx context.Context, agentID string, plugins []*v1.PluginInfo) ([]string, error) {
	pluginPaths := make([]string, 0, len(plugins))

	for _, plugin := range plugins {
		path, err := d.ensurePlugin(ctx, agentID, plugin)
		if err != nil {
			return nil, fmt.Errorf("failed to ensure plugin %s: %w", plugin.PluginId, err)
		}
		pluginPaths = append(pluginPaths, path)
	}

	return pluginPaths, nil
}

// ensurePlugin 确保单个插件就绪
func (d *PluginDownloader) ensurePlugin(ctx context.Context, agentID string, plugin *v1.PluginInfo) (string, error) {
	cacheKey := fmt.Sprintf("%s@%s", plugin.PluginId, plugin.Version)

	// 检查缓存
	d.mu.RLock()
	cached, exists := d.cache[cacheKey]
	d.mu.RUnlock()

	if exists {
		// 验证校验和
		if cached.Checksum == plugin.Checksum {
			log.Infof("plugin %s already cached", cacheKey)
			return cached.Path, nil
		}
		log.Warnf("plugin %s checksum mismatch, re-downloading", cacheKey)
	}

	// 下载插件
	log.Infof("downloading plugin %s (v%s)", plugin.PluginId, plugin.Version)
	return d.downloadPlugin(ctx, agentID, plugin)
}

// downloadPlugin 从服务端下载插件
func (d *PluginDownloader) downloadPlugin(ctx context.Context, agentID string, plugin *v1.PluginInfo) (string, error) {
	// 调用gRPC下载
	resp, err := d.agentClient.DownloadPlugin(ctx, &v1.DownloadPluginRequest{
		AgentId:  agentID,
		PluginId: plugin.PluginId,
		Version:  plugin.Version,
	})
	if err != nil {
		return "", fmt.Errorf("grpc call failed: %w", err)
	}

	if !resp.Success {
		return "", fmt.Errorf("download failed: %s", resp.Message)
	}

	// 验证校验和
	actualChecksum := calculateChecksumFromBytes(resp.PluginData)
	if actualChecksum != resp.Checksum {
		return "", fmt.Errorf("checksum verification failed: expected %s, got %s", 
			resp.Checksum, actualChecksum)
	}

	// 保存到本地
	filename := fmt.Sprintf("%s@%s.so", plugin.PluginId, plugin.Version)
	pluginPath := filepath.Join(d.pluginDir, "downloaded", filename)

	if err := os.WriteFile(pluginPath, resp.PluginData, 0755); err != nil {
		return "", fmt.Errorf("failed to save plugin: %w", err)
	}

	// 更新缓存
	cacheKey := fmt.Sprintf("%s@%s", plugin.PluginId, plugin.Version)
	d.mu.Lock()
	d.cache[cacheKey] = &CachedPlugin{
		PluginID:  plugin.PluginId,
		Version:   plugin.Version,
		Path:      pluginPath,
		Checksum:  resp.Checksum,
		Size:      resp.Size,
		CachedAt:  time.Now().Unix(),
	}
	d.mu.Unlock()

	log.Infof("plugin %s downloaded successfully: %s", plugin.PluginId, pluginPath)
	return pluginPath, nil
}

// GetInstalledPlugins 获取已安装插件列表
func (d *PluginDownloader) GetInstalledPlugins() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	installed := make([]string, 0, len(d.cache))
	for key := range d.cache {
		installed = append(installed, key)
	}
	return installed
}

// 辅助函数
func calculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func calculateChecksumFromBytes(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
```

### 2. Agent 任务执行器集成

```go
// agent/internal/executor/job_executor.go
package executor

import (
	"context"
	"fmt"

	v1 "github.com/observabil/arcade/api/agent/v1"
	"github.com/observabil/arcade/agent/internal/plugin"
	pluginmgr "github.com/observabil/arcade/pkg/plugin"
	"github.com/observabil/arcade/pkg/log"
)

type JobExecutor struct {
	agentID          string
	pluginDownloader *plugin.PluginDownloader
	pluginManager    *pluginmgr.Manager
}

func NewJobExecutor(agentID string, downloader *plugin.PluginDownloader, manager *pluginmgr.Manager) *JobExecutor {
	return &JobExecutor{
		agentID:          agentID,
		pluginDownloader: downloader,
		pluginManager:    manager,
	}
}

// ExecuteJob 执行任务
func (e *JobExecutor) ExecuteJob(ctx context.Context, job *v1.Job) error {
	log.Infof("executing job %s", job.JobId)

	// 1. 确保所有插件就绪
	if len(job.Plugins) > 0 {
		log.Infof("job requires %d plugins", len(job.Plugins))
		pluginPaths, err := e.pluginDownloader.EnsurePlugins(ctx, e.agentID, job.Plugins)
		if err != nil {
			return fmt.Errorf("failed to ensure plugins: %w", err)
		}

		// 2. 加载插件
		for i, pluginPath := range pluginPaths {
			pluginInfo := job.Plugins[i]
			pluginName := fmt.Sprintf("%s@%s", pluginInfo.PluginId, pluginInfo.Version)
			
			if err := e.pluginManager.Register(pluginPath, pluginName, nil); err != nil {
				log.Warnf("failed to register plugin %s: %v", pluginName, err)
				// 可能已经加载过了，继续执行
			} else {
				log.Infof("plugin %s loaded successfully", pluginName)
			}
		}

		// 3. 初始化插件
		if err := e.pluginManager.Init(ctx); err != nil {
			return fmt.Errorf("failed to init plugins: %w", err)
		}
	}

	// 4. 执行before阶段插件
	if err := e.executePluginStage(ctx, job, "before"); err != nil {
		return fmt.Errorf("before stage failed: %w", err)
	}

	// 5. 执行任务命令
	jobErr := e.executeCommands(ctx, job)

	// 6. 根据任务结果执行相应插件
	if jobErr != nil {
		e.executePluginStage(ctx, job, "on_failure")
		return jobErr
	} else {
		e.executePluginStage(ctx, job, "on_success")
	}

	// 7. 执行after阶段插件
	if err := e.executePluginStage(ctx, job, "after"); err != nil {
		log.Warnf("after stage failed: %v", err)
	}

	return nil
}

// executePluginStage 执行特定阶段的插件
func (e *JobExecutor) executePluginStage(ctx context.Context, job *v1.Job, stage string) error {
	// 从数据库或任务配置获取该阶段需要执行的插件
	// 这里简化处理
	for _, pluginInfo := range job.Plugins {
		pluginName := fmt.Sprintf("%s@%s", pluginInfo.PluginId, pluginInfo.Version)
		
		// 获取插件并执行
		// 根据插件类型调用不同的接口
		if notifyPlugin, err := e.pluginManager.GetNotifyPlugin(pluginName); err == nil {
			message := fmt.Sprintf("Job %s %s", job.JobId, stage)
			if err := notifyPlugin.Send(ctx, message); err != nil {
				log.Errorf("failed to execute notify plugin %s: %v", pluginName, err)
			}
		}
	}

	return nil
}

// executeCommands 执行任务命令
func (e *JobExecutor) executeCommands(ctx context.Context, job *v1.Job) error {
	for _, cmd := range job.Commands {
		log.Infof("executing command: %s", cmd)
		// 执行命令逻辑...
	}
	return nil
}
```

## 配置示例

### 服务端配置 (config.toml)

```toml
[plugin]
# 插件存储路径
storage_path = "./plugins"

# 最大插件大小 (MB)
max_plugin_size = 100

# 是否启用插件缓存
enable_cache = true

# 缓存过期时间 (秒)
cache_ttl = 3600

[plugin.security]
# 是否强制校验插件签名
require_checksum = true

# 是否允许不安全的插件
allow_unsafe = false
```

### Agent 配置

```yaml
# agent配置
agent:
  id: "agent-001"
  plugins:
    # 插件存储目录
    directory: "/var/lib/arcade/plugins"
    
    # 下载配置
    download:
      # 并发下载数
      concurrent: 3
      # 下载超时 (秒)
      timeout: 300
      # 失败重试次数
      retry: 3
    
    # 缓存配置
    cache:
      # 最大缓存大小 (MB)
      max_size: 1000
      # 缓存清理策略: lru/lfu/fifo
      eviction: "lru"
      # 缓存过期时间 (小时)
      ttl: 24
    
    # 预装插件列表
    builtin:
      - "notify/stdout"
      - "build/docker"
```

## 安全考虑

### 1. 插件完整性验证

```go
// 强制校验和验证
func verifyPlugin(pluginData []byte, expectedChecksum string) error {
    actualChecksum := calculateChecksumFromBytes(pluginData)
    if actualChecksum != expectedChecksum {
        return fmt.Errorf("security: checksum mismatch")
    }
    return nil
}
```

### 2. 插件签名（可选，高安全场景）

```go
// 使用GPG签名验证插件
func verifyPluginSignature(pluginPath string, signaturePath string, publicKeyPath string) error {
    // 实现GPG签名验证
    // ...
}
```

### 3. 访问控制

```go
// 检查Agent是否有权限下载特定插件
func (s *PluginDownloadService) CheckPermission(agentID string, pluginID string) error {
    // 根据Agent标签、组织、权限等判断
    // ...
}
```

## 性能优化

### 1. 插件压缩传输

```go
// 使用gzip压缩插件数据
func compressPlugin(data []byte) ([]byte, error) {
    var buf bytes.Buffer
    gzipWriter := gzip.NewWriter(&buf)
    _, err := gzipWriter.Write(data)
    gzipWriter.Close()
    return buf.Bytes(), err
}
```

### 2. 增量更新

```go
// 只下载变化的部分（使用rsync算法或二进制diff）
func deltaUpdate(oldVersion, newVersion string) ([]byte, error) {
    // 实现增量更新逻辑
    // ...
}
```

### 3. CDN加速

```yaml
plugin:
  distribution:
    # 使用CDN分发插件
    cdn_enabled: true
    cdn_url: "https://cdn.example.com/plugins"
```

## 监控和日志

### 插件下载指标

```go
// Prometheus 指标
var (
    pluginDownloadTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "plugin_download_total",
            Help: "Total number of plugin downloads",
        },
        []string{"plugin_id", "version", "status"},
    )

    pluginDownloadDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "plugin_download_duration_seconds",
            Help:    "Plugin download duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"plugin_id"},
    )
)
```

### 日志记录

```go
log.Infof("[PluginDownload] agent=%s plugin=%s version=%s size=%d checksum=%s duration=%dms",
    agentID, pluginID, version, size, checksum[:8], duration)
```

## 故障处理

### 1. 下载失败重试

```go
func (d *PluginDownloader) downloadWithRetry(ctx context.Context, agentID string, plugin *v1.PluginInfo) (string, error) {
    maxRetries := 3
    for i := 0; i < maxRetries; i++ {
        path, err := d.downloadPlugin(ctx, agentID, plugin)
        if err == nil {
            return path, nil
        }
        
        log.Warnf("download failed (attempt %d/%d): %v", i+1, maxRetries, err)
        time.Sleep(time.Second * time.Duration(i+1))
    }
    
    return "", fmt.Errorf("download failed after %d retries", maxRetries)
}
```

### 2. 降级策略

```go
// 如果下载失败，尝试使用旧版本或跳过该插件
func (e *JobExecutor) executionWithFallback(ctx context.Context, job *v1.Job) error {
    err := e.ExecuteJob(ctx, job)
    if err != nil {
        // 降级：不使用插件执行
        log.Warn("executing job without plugins due to download failure")
        return e.executeCommandsOnly(ctx, job)
    }
    return nil
}
```

## 测试

### 单元测试

```go
func TestPluginDownloader_EnsurePlugins(t *testing.T) {
    // Mock gRPC client
    mockClient := &mockAgentClient{
        downloadResp: &v1.DownloadPluginResponse{
            Success: true,
            PluginData: []byte("mock plugin data"),
            Checksum: "abc123",
        },
    }

    downloader := NewPluginDownloader(mockClient, "/tmp/test-plugins")
    downloader.Init()

    plugins := []*v1.PluginInfo{
        {
            PluginId: "test-plugin",
            Version: "1.0.0",
            Checksum: "abc123",
        },
    }

    paths, err := downloader.EnsurePlugins(context.Background(), "agent-001", plugins)
    
    assert.NoError(t, err)
    assert.Len(t, paths, 1)
}
```

## 总结

本方案提供了一个完整的插件分发架构：

✅ **灵活性**：支持多种分发方式  
✅ **安全性**：校验和验证、访问控制  
✅ **性能**：本地缓存、按需下载  
✅ **可靠性**：重试机制、降级策略  
✅ **可扩展**：支持对象存储、CDN加速  

### 实施步骤

1. **Phase 1**：实现基础的按需下载功能
2. **Phase 2**：添加本地缓存和校验和验证
3. **Phase 3**：集成到任务执行流程
4. **Phase 4**：性能优化（压缩、CDN等）
5. **Phase 5**：监控和告警

### 下一步

1. 扩展 Proto 定义并重新生成代码
2. 实现服务端插件下载服务
3. 开发 Agent 端插件下载器
4. 集成到任务执行流程
5. 编写测试用例
6. 部署和验证

