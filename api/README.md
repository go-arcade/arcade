# Arcade API

Arcade 项目的 gRPC API 定义，使用 Protocol Buffers 定义，通过 Buf 进行管理。

## 概述

本目录包含了 Arcade 项目所有的 gRPC API 定义，分为四个主要服务模块：

- **Agent Service** - Agent 端与 Server 端通信的核心接口
- **Pipeline Service** - 流水线管理接口
- **Task Service** - 任务管理接口
- **Stream Service** - 实时数据流传输接口

## 目录结构

```
api/
├── buf.yaml                    # Buf 配置文件（lint 和 breaking change 检查）
├── buf.gen.yaml                # 代码生成配置文件
├── agent/v1/                   # Agent 服务 API
│   ├── agent.proto             # Proto 定义文件
│   ├── agent.pb.go             # 生成的 Go 消息代码
│   └── agent_grpc.pb.go        # 生成的 gRPC 服务代码
├── pipeline/v1/                # Pipeline 服务 API
│   ├── pipeline.proto
│   ├── pipeline.pb.go
│   └── pipeline_grpc.pb.go
├── stream/v1/                  # Stream 服务 API
│   ├── stream.proto
│   ├── stream.pb.go
│   └── stream_grpc.pb.go
└── task/v1/                    # Task 服务 API
    ├── task.proto
    ├── task.pb.go
    └── task_grpc.pb.go
```

## API 服务说明

### 1. Agent Service (`agent/v1`)

Agent 端与 Server 端通信的主要接口，负责 Agent 的生命周期管理和任务执行。

**主要功能：**
- **心跳保持** (`Heartbeat`) - Agent 定期向 Server 发送心跳
- **Agent 注册/注销** (`Register`/`Unregister`) - Agent 的生命周期管理
- **任务获取** (`FetchTask`) - Agent 主动拉取待执行的任务
- **状态上报** (`ReportTaskStatus`) - 上报任务执行状态
- **日志上报** (`ReportTaskLog`) - 批量上报任务执行日志
- **任务取消** (`CancelTask`) - Server 通知 Agent 取消任务
- **标签更新** (`UpdateLabels`) - 动态更新 Agent 的标签和标记
- **插件管理** (`DownloadPlugin`, `ListAvailablePlugins`) - 插件分发和查询

**核心特性：**
- 支持标签选择器（Label Selector）进行智能任务路由
- 支持插件动态分发
- 完善的指标上报机制

### 2. Pipeline Service (`pipeline/v1`)

流水线管理接口，负责 CI/CD 流水线的创建、执行和管理。

**主要功能：**
- **创建流水线** (`CreatePipeline`) - 定义流水线配置
- **获取流水线** (`GetPipeline`) - 获取流水线详情
- **列出流水线** (`ListPipelines`) - 分页查询流水线列表
- **触发执行** (`TriggerPipeline`) - 触发流水线执行
- **停止流水线** (`StopPipeline`) - 停止正在运行的流水线

**支持的触发方式：**
- 手动触发 (Manual)
- Webhook 触发 (Webhook)
- 定时触发 (Schedule/Cron)
- API 触发 (API)

**流水线状态：**
- PENDING（等待执行）
- RUNNING（执行中）
- SUCCESS（执行成功）
- FAILED（执行失败）
- CANCELLED（已取消）
- PARTIAL（部分成功）

### 3. Task Service (`task/v1`)

任务管理接口，负责单个任务的 CRUD 操作和执行管理。

**主要功能：**
- **创建任务** (`CreateTask`) - 创建新任务
- **获取任务** (`GetTask`) - 获取任务详情
- **列出任务** (`ListTasks`) - 分页查询任务列表
- **更新任务** (`UpdateTask`) - 更新任务配置
- **删除任务** (`DeleteTask`) - 删除任务
- **取消任务** (`CancelTask`) - 取消正在执行的任务
- **重试任务** (`RetryTask`) - 重新执行失败的任务
- **获取日志** (`GetTaskLog`) - 获取任务执行日志
- **产物管理** (`ListTaskArtifacts`) - 管理任务产物

**任务状态：**
- PENDING（等待执行）
- QUEUED（已入队）
- RUNNING（执行中）
- SUCCESS（执行成功）
- FAILED（执行失败）
- CANCELLED（已取消）
- TIMEOUT（超时）
- SKIPPED（已跳过）

**核心特性：**
- 支持任务依赖关系
- 支持失败重试机制
- 支持产物收集和管理
- 支持标签选择器路由

### 4. Stream Service (`stream/v1`)

实时数据流传输接口，提供双向流式通信能力。

**主要功能：**
- **任务日志流** (`StreamTaskLog`, `UploadTaskLog`) - 实时获取和上报任务日志
- **任务状态流** (`StreamTaskStatus`) - 实时推送任务状态变化
- **流水线状态流** (`StreamPipelineStatus`) - 实时推送流水线状态变化
- **Agent 通道** (`AgentChannel`) - Agent 与 Server 双向通信
- **Agent 状态流** (`StreamAgentStatus`) - 实时监控 Agent 状态
- **事件流** (`StreamEvents`) - 推送系统事件

**支持的事件类型：**
- 任务事件（创建、开始、完成、失败、取消）
- 流水线事件（开始、完成、失败）
- Agent 事件（注册、注销、离线）

## 快速开始

### 前置要求

- [Buf CLI](https://docs.buf.build/installation) >= 1.0.0
- [Go](https://golang.org/) >= 1.21
- [Protocol Buffers Compiler](https://grpc.io/docs/protoc-installation/)

### 安装 Buf

```bash
# macOS
brew install bufbuild/buf/buf

# Linux
curl -sSL "https://github.com/bufbuild/buf/releases/latest/download/buf-$(uname -s)-$(uname -m)" -o /usr/local/bin/buf
chmod +x /usr/local/bin/buf

# 验证安装
buf --version
```

### 生成代码

```bash
# 在项目根目录下执行
make proto

# 或者在 api 目录下直接使用 buf
cd api
buf generate
```

### 代码检查

```bash
# Lint 检查
buf lint

# Breaking change 检查
buf breaking --against '.git#branch=main'
```

### 格式化

```bash
# 格式化所有 proto 文件
buf format -w
```

## 配置说明

### buf.yaml

主配置文件，定义了：
- 模块名称：`buf.build/observabil/arcade`
- Lint 规则：使用 STANDARD 规则集，但允许流式 RPC
- Breaking change 检查：使用 FILE 级别检查

### buf.gen.yaml

代码生成配置，定义了：
- Go Package 前缀：`github.com/observabil/arcade/api`
- 插件配置：
  - `protocolbuffers/go` - 生成 Go 消息代码
  - `grpc/go` - 生成 gRPC 服务代码
- 路径模式：`source_relative`（相对于源文件生成）

## 使用示例

### 客户端调用示例

```go
package main

import (
    "context"
    "log"
    
    "google.golang.org/grpc"
    agentv1 "github.com/observabil/arcade/api/agent/v1"
)

func main() {
    // 连接到 gRPC 服务
    conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
    if err != nil {
        log.Fatalf("连接失败: %v", err)
    }
    defer conn.Close()
    
    // 创建客户端
    client := agentv1.NewAgentServiceClient(conn)
    
    // 调用 Register RPC
    req := &agentv1.RegisterRequest{
        Hostname:          "my-agent",
        Ip:                "192.168.1.100",
        Os:                "linux",
        Arch:              "amd64",
        Version:           "1.0.0",
        MaxConcurrentJobs: 5,
        Labels: map[string]string{
            "env":  "production",
            "zone": "us-west-1",
        },
    }
    
    resp, err := client.Register(context.Background(), req)
    if err != nil {
        log.Fatalf("注册失败: %v", err)
    }
    
    log.Printf("注册成功，Agent ID: %s", resp.AgentId)
}
```

### 服务端实现示例

```go
package main

import (
    "context"
    "log"
    "net"
    
    "google.golang.org/grpc"
    agentv1 "github.com/observabil/arcade/api/agent/v1"
)

type agentService struct {
    agentv1.UnimplementedAgentServiceServer
}

func (s *agentService) Register(ctx context.Context, req *agentv1.RegisterRequest) (*agentv1.RegisterResponse, error) {
    log.Printf("收到注册请求: %+v", req)
    
    return &agentv1.RegisterResponse{
        Success:           true,
        Message:           "注册成功",
        AgentId:           "agent-12345",
        HeartbeatInterval: 30,
    }, nil
}

func main() {
    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("监听失败: %v", err)
    }
    
    s := grpc.NewServer()
    agentv1.RegisterAgentServiceServer(s, &agentService{})
    
    log.Println("gRPC 服务启动在 :50051")
    if err := s.Serve(lis); err != nil {
        log.Fatalf("服务启动失败: %v", err)
    }
}
```

### 流式 RPC 示例

```go
// 客户端：实时接收任务日志
func streamTaskLog(client streamv1.StreamServiceClient, taskID string) {
    req := &streamv1.StreamTaskLogRequest{
        JobId:  taskID,
        Follow: true, // 持续跟踪，类似 tail -f
    }
    
    stream, err := client.StreamTaskLog(context.Background(), req)
    if err != nil {
        log.Fatalf("创建流失败: %v", err)
    }
    
    for {
        resp, err := stream.Recv()
        if err == io.EOF {
            break
        }
        if err != nil {
            log.Fatalf("接收失败: %v", err)
        }
        
        log.Printf("[%s] %s", resp.LogChunk.Level, resp.LogChunk.Content)
    }
}
```

## 标签选择器使用说明

标签选择器用于任务路由，可以精确控制任务在哪些 Agent 上执行。

### 简单匹配

```go
// 匹配具有特定标签的 Agent
labelSelector := &agentv1.LabelSelector{
    MatchLabels: map[string]string{
        "env":  "production",
        "zone": "us-west-1",
        "os":   "linux",
    },
}
```

### 表达式匹配

```go
// 使用更复杂的匹配规则
labelSelector := &agentv1.LabelSelector{
    MatchExpressions: []*agentv1.LabelSelectorRequirement{
        {
            Key:      "env",
            Operator: agentv1.LabelOperator_LABEL_OPERATOR_IN,
            Values:   []string{"production", "staging"},
        },
        {
            Key:      "gpu",
            Operator: agentv1.LabelOperator_LABEL_OPERATOR_EXISTS,
        },
        {
            Key:      "memory",
            Operator: agentv1.LabelOperator_LABEL_OPERATOR_GT,
            Values:   []string{"8192"}, // 内存大于 8GB
        },
    },
}
```

### 支持的操作符

- `IN` - 标签值在指定列表中
- `NOT_IN` - 标签值不在指定列表中
- `EXISTS` - 标签键存在
- `NOT_EXISTS` - 标签键不存在
- `GT` - 标签值大于指定值（用于数值比较）
- `LT` - 标签值小于指定值（用于数值比较）

## 插件分发机制

Agent Service 支持插件动态分发，支持三种插件位置：

1. **SERVER** - 服务端文件系统
2. **STORAGE** - 对象存储（S3/OSS/COS/GCS）
3. **REGISTRY** - 插件仓库

### 下载插件示例

```go
req := &agentv1.DownloadPluginRequest{
    AgentId:  "agent-123",
    PluginId: "notify",
    Version:  "1.0.0", // 可选，不指定则下载最新版本
}

resp, err := client.DownloadPlugin(context.Background(), req)
if err != nil {
    log.Fatalf("下载插件失败: %v", err)
}

// 验证校验和
if calculateSHA256(resp.PluginData) != resp.Checksum {
    log.Fatal("插件校验和不匹配")
}

// 保存插件
os.WriteFile("plugins/notify.so", resp.PluginData, 0755)
```

## 开发指南

### 修改 Proto 文件

1. 修改对应的 `.proto` 文件
2. 运行 `buf lint` 检查代码风格
3. 运行 `buf breaking --against '.git#branch=main'` 检查破坏性变更
4. 运行 `buf generate` 生成新代码
5. 提交代码

### 添加新的 RPC 方法

```protobuf
service YourService {
  // 添加新方法
  rpc NewMethod(NewMethodRequest) returns (NewMethodResponse) {}
}

message NewMethodRequest {
  string param = 1;
}

message NewMethodResponse {
  bool success = 1;
  string message = 2;
}
```

### 版本管理

API 使用语义化版本管理，遵循以下规则：

- **主版本号** (`v1`, `v2`) - 不兼容的 API 变更
- **次版本号** - 向后兼容的功能新增
- **修订号** - 向后兼容的问题修正

当需要引入破坏性变更时，创建新的版本目录（如 `agent/v2/`）。

## 常见问题

### 1. 如何处理大文件传输？

对于大文件（如插件二进制），建议：
- 使用流式 RPC 分块传输
- 或者返回预签名 URL，让客户端直接从对象存储下载

### 2. 如何处理长时间运行的任务？

使用 Stream Service 的流式接口：
- 实时推送任务状态更新
- 实时推送日志输出
- 使用双向流保持连接

### 3. 如何实现任务优先级？

在任务的 `labels` 中添加 `priority` 标签：
```go
labels: map[string]string{
    "priority": "high",
}
```

Agent 在 FetchTask 时可以按优先级排序。

### 4. 如何处理 Agent 断线重连？

Agent 应该：
1. 实现指数退避重连策略
2. 重连后重新注册
3. 上报未完成任务的状态

## 相关文档

- [Plugin Development Guide](../docs/PLUGIN_DEVELOPMENT.md)
- [Plugin Distribution Guide](../docs/PLUGIN_DISTRIBUTION.md)
- [Implementation Guide](../docs/IMPLEMENTATION_GUIDE.md)
- [Buf Documentation](https://docs.buf.build/)
- [gRPC Documentation](https://grpc.io/docs/)
- [Protocol Buffers Documentation](https://protobuf.dev/)

## 贡献指南

欢迎贡献！在提交 PR 之前，请确保：

1. ✅ 所有 proto 文件通过 `buf lint` 检查
2. ✅ 没有引入破坏性变更（或在新版本中）
3. ✅ 添加了充分的注释
4. ✅ 生成的代码已更新
5. ✅ 相关文档已更新

## 许可证

本项目使用 [LICENSE](../LICENSE) 文件中定义的许可证。

