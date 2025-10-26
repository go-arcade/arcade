# Arcade Agent API 中文文档

简体中文 | [English](./README.md)

Arcade 与 Agent 交互的 gRPC API 定义，使用 Protocol Buffers 定义，通过 Buf 进行管理。

## 概述

本目录包含了 Arcade 与 Agent 交互的所有的 gRPC API 定义，分为四个主要服务模块：

- **Agent Service** - Agent 端与 Server 端通信的核心接口
- **Pipeline Service** - 流水线管理接口
- **Task Service** - 任务管理接口
- **Stream Service** - 实时数据流传输接口

## 目录结构

```
api/
├── buf.yaml                    # Buf 配置文件（lint 和 breaking change 检查）
├── buf.gen.yaml                # 代码生成配置文件
├── README.md                   # 英文文档
├── README_CN.md               # 中文文档
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

[查看完整 Agent Service 文档](./agent/v1/README.md)

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

### 4. Stream Service (`stream/v1`)

实时数据流传输接口，提供双向流式通信能力。

**主要功能：**
- **任务日志流** (`StreamTaskLog`, `UploadTaskLog`) - 实时获取和上报任务日志
- **任务状态流** (`StreamTaskStatus`) - 实时推送任务状态变化
- **流水线状态流** (`StreamPipelineStatus`) - 实时推送流水线状态变化
- **Agent 通道** (`AgentChannel`) - Agent 与 Server 双向通信
- **Agent 状态流** (`StreamAgentStatus`) - 实时监控 Agent 状态
- **事件流** (`StreamEvents`) - 推送系统事件

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

## 相关文档

- [插件开发指南](../docs/PLUGIN_DEVELOPMENT.md)
- [插件分发指南](../docs/PLUGIN_DISTRIBUTION.md)
- [插件系统重构文档](../docs/PLUGIN_SYSTEM_REFACTOR.md)
- [实现指南](../docs/IMPLEMENTATION_GUIDE.md)
- [Buf 文档](https://docs.buf.build/)
- [gRPC 文档](https://grpc.io/docs/)
- [Protocol Buffers 文档](https://protobuf.dev/)

## 许可证

本项目使用 [LICENSE](../LICENSE) 文件中定义的许可证。

