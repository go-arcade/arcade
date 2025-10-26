# Arcade Agent API

English | [简体中文](./README_zh_CN.md)

gRPC API definitions for the Arcade and Agent interaction, defined using Protocol Buffers and managed through Buf.

## Overview

This directory contains all gRPC API definitions for the Arcade and Agent interaction, divided into four main service modules:

- **Agent Service** - Core interface for communication between Agent and Server
- **Pipeline Service** - Pipeline management interface
- **Task Service** - Task management interface
- **Stream Service** - Real-time data streaming interface

## Directory Structure

```
api/
├── buf.yaml                    # Buf configuration file (lint and breaking change checks)
├── buf.gen.yaml                # Code generation configuration file
├── README.md                   # English documentation
├── README_zh_CN.md             # Chinese documentation
├── agent/v1/                   # Agent Service API
│   ├── agent.proto             # Proto definition file
│   ├── agent.pb.go             # Generated Go message code
│   └── agent_grpc.pb.go        # Generated gRPC service code
├── pipeline/v1/                # Pipeline Service API
│   ├── pipeline.proto
│   ├── pipeline.pb.go
│   └── pipeline_grpc.pb.go
├── stream/v1/                  # Stream Service API
│   ├── stream.proto
│   ├── stream.pb.go
│   └── stream_grpc.pb.go
└── task/v1/                    # Task Service API
    ├── task.proto
    ├── task.pb.go
    └── task_grpc.pb.go
```

## API Service Description

### 1. Agent Service (`agent/v1`)

The main interface for communication between Agent and Server, responsible for Agent lifecycle management and task execution.

**Main Features:**
- **Heartbeat** (`Heartbeat`) - Agent periodically sends heartbeat to Server
- **Agent Registration/Unregistration** (`Register`/`Unregister`) - Agent lifecycle management
- **Task Fetching** (`FetchTask`) - Agent actively pulls tasks to execute
- **Status Reporting** (`ReportTaskStatus`) - Report task execution status
- **Log Reporting** (`ReportTaskLog`) - Batch report task execution logs
- **Task Cancellation** (`CancelTask`) - Server notifies Agent to cancel task
- **Label Updates** (`UpdateLabels`) - Dynamically update Agent's labels
- **Plugin Management** (`DownloadPlugin`, `ListAvailablePlugins`) - Plugin distribution and query

**Core Features:**
- Support label selector for intelligent task routing
- Support dynamic plugin distribution
- Comprehensive metrics reporting mechanism

### 2. Pipeline Service (`pipeline/v1`)

Pipeline management interface, responsible for creating, executing and managing CI/CD pipelines.

**Main Features:**
- **Create Pipeline** (`CreatePipeline`) - Define pipeline configuration
- **Get Pipeline** (`GetPipeline`) - Get pipeline details
- **List Pipelines** (`ListPipelines`) - Paginated pipeline list query
- **Trigger Execution** (`TriggerPipeline`) - Trigger pipeline execution
- **Stop Pipeline** (`StopPipeline`) - Stop running pipeline

**Supported Trigger Methods:**
- Manual trigger (Manual)
- Webhook trigger (Webhook)
- Schedule trigger (Schedule/Cron)
- API trigger (API)

**Pipeline Status:**
- PENDING (Pending)
- RUNNING (Running)
- SUCCESS (Success)
- FAILED (Failed)
- CANCELLED (Cancelled)
- PARTIAL (Partial success)

### 3. Task Service (`task/v1`)

Task management interface, responsible for CRUD operations and execution management of individual tasks.

**Main Features:**
- **Create Task** (`CreateTask`) - Create new task
- **Get Task** (`GetTask`) - Get task details
- **List Tasks** (`ListTasks`) - Paginated task list query
- **Update Task** (`UpdateTask`) - Update task configuration
- **Delete Task** (`DeleteTask`) - Delete task
- **Cancel Task** (`CancelTask`) - Cancel running task
- **Retry Task** (`RetryTask`) - Re-execute failed task
- **Get Log** (`GetTaskLog`) - Get task execution log
- **Artifact Management** (`ListTaskArtifacts`) - Manage task artifacts

**Task Status:**
- PENDING (Pending)
- QUEUED (Queued)
- RUNNING (Running)
- SUCCESS (Success)
- FAILED (Failed)
- CANCELLED (Cancelled)
- TIMEOUT (Timeout)
- SKIPPED (Skipped)

**Core Features:**
- Support task dependencies
- Support failure retry mechanism
- Support artifact collection and management
- Support label selector routing

### 4. Stream Service (`stream/v1`)

Real-time data streaming interface, providing bidirectional streaming communication capability.

**Main Features:**
- **Task Log Stream** (`StreamTaskLog`, `UploadTaskLog`) - Real-time get and report task logs
- **Task Status Stream** (`StreamTaskStatus`) - Real-time push task status changes
- **Pipeline Status Stream** (`StreamPipelineStatus`) - Real-time push pipeline status changes
- **Agent Channel** (`AgentChannel`) - Bidirectional communication between Agent and Server
- **Agent Status Stream** (`StreamAgentStatus`) - Real-time monitor Agent status
- **Event Stream** (`StreamEvents`) - Push system events

**Supported Event Types:**
- Task events (created, started, completed, failed, cancelled)
- Pipeline events (started, completed, failed)
- Agent events (registered, unregistered, offline)

## Quick Start

### Prerequisites

- [Buf CLI](https://docs.buf.build/installation) >= 1.0.0
- [Go](https://golang.org/) >= 1.21
- [Protocol Buffers Compiler](https://grpc.io/docs/protoc-installation/)

### Install Buf

```bash
# macOS
brew install bufbuild/buf/buf

# Linux
curl -sSL "https://github.com/bufbuild/buf/releases/latest/download/buf-$(uname -s)-$(uname -m)" -o /usr/local/bin/buf
chmod +x /usr/local/bin/buf

# Verify installation
buf --version
```

### Generate Code

```bash
# Execute in project root directory
make proto

# Or use buf directly in api directory
cd api
buf generate
```

### Code Check

```bash
# Lint check
buf lint

# Breaking change check
buf breaking --against '.git#branch=main'
```

### Format

```bash
# Format all proto files
buf format -w
```

## Configuration Description

### buf.yaml

Main configuration file, defines:
- Module name: `buf.build/observabil/arcade`
- Lint rules: Use STANDARD rule set, but allow streaming RPC
- Breaking change check: Use FILE level check

### buf.gen.yaml

Code generation configuration, defines:
- Go Package prefix: `github.com/go-arcade/arcade/api`
- Plugin configuration:
  - `protocolbuffers/go` - Generate Go message code
  - `grpc/go` - Generate gRPC service code
- Path mode: `source_relative` (relative to source file generation)

## Usage Examples

### Client Call Example

```go
package main

import (
    "context"
    "log"
    
    "google.golang.org/grpc"
    agentv1 "github.com/go-arcade/arcade/api/agent/v1"
)

func main() {
    // Connect to gRPC service
    conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
    if err != nil {
        log.Fatalf("Connection failed: %v", err)
    }
    defer conn.Close()
    
    // Create client
    client := agentv1.NewAgentServiceClient(conn)
    
    // Call Register RPC
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
        log.Fatalf("Registration failed: %v", err)
    }
    
    log.Printf("Registration successful, Agent ID: %s", resp.AgentId)
}
```

### Server Implementation Example

```go
package main

import (
    "context"
    "log"
    "net"
    
    "google.golang.org/grpc"
    agentv1 "github.com/go-arcade/arcade/api/agent/v1"
)

type agentService struct {
    agentv1.UnimplementedAgentServiceServer
}

func (s *agentService) Register(ctx context.Context, req *agentv1.RegisterRequest) (*agentv1.RegisterResponse, error) {
    log.Printf("Received registration request: %+v", req)
    
    return &agentv1.RegisterResponse{
        Success:           true,
        Message:           "Registration successful",
        AgentId:           "agent-12345",
        HeartbeatInterval: 30,
    }, nil
}

func main() {
    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("Listen failed: %v", err)
    }
    
    s := grpc.NewServer()
    agentv1.RegisterAgentServiceServer(s, &agentService{})
    
    log.Println("gRPC service started on :50051")
    if err := s.Serve(lis); err != nil {
        log.Fatalf("Service failed to start: %v", err)
    }
}
```

### Streaming RPC Example

```go
// Client: Receive task logs in real-time
func streamTaskLog(client streamv1.StreamServiceClient, taskID string) {
    req := &streamv1.StreamTaskLogRequest{
        JobId:  taskID,
        Follow: true, // Continuously track, similar to tail -f
    }
    
    stream, err := client.StreamTaskLog(context.Background(), req)
    if err != nil {
        log.Fatalf("Failed to create stream: %v", err)
    }
    
    for {
        resp, err := stream.Recv()
        if err == io.EOF {
            break
        }
        if err != nil {
            log.Fatalf("Receive failed: %v", err)
        }
        
        log.Printf("[%s] %s", resp.LogChunk.Level, resp.LogChunk.Content)
    }
}
```

## Label Selector Usage

Label selectors are used for task routing, allowing precise control over which Agents execute tasks.

### Simple Matching

```go
// Match Agents with specific labels
labelSelector := &agentv1.LabelSelector{
    MatchLabels: map[string]string{
        "env":  "production",
        "zone": "us-west-1",
        "os":   "linux",
    },
}
```

### Expression Matching

```go
// Use more complex matching rules
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
            Values:   []string{"8192"}, // Memory greater than 8GB
        },
    },
}
```

### Supported Operators

- `IN` - Label value is in the specified list
- `NOT_IN` - Label value is not in the specified list
- `EXISTS` - Label key exists
- `NOT_EXISTS` - Label key does not exist
- `GT` - Label value greater than specified value (for numeric comparison)
- `LT` - Label value less than specified value (for numeric comparison)

## Plugin Distribution Mechanism

Agent Service supports dynamic plugin distribution, supporting three plugin locations:

1. **SERVER** - Server filesystem
2. **STORAGE** - Object storage (S3/OSS/COS/GCS)
3. **REGISTRY** - Plugin registry

### Download Plugin Example

```go
req := &agentv1.DownloadPluginRequest{
    AgentId:  "agent-123",
    PluginId: "notify",
    Version:  "1.0.0", // Optional, download latest if not specified
}

resp, err := client.DownloadPlugin(context.Background(), req)
if err != nil {
    log.Fatalf("Failed to download plugin: %v", err)
}

// Verify checksum
if calculateSHA256(resp.PluginData) != resp.Checksum {
    log.Fatal("Plugin checksum mismatch")
}

// Save plugin
os.WriteFile("plugins/notify.so", resp.PluginData, 0755)
```

## Development Guide

### Modifying Proto Files

1. Modify the corresponding `.proto` file
2. Run `buf lint` to check code style
3. Run `buf breaking --against '.git#branch=main'` to check breaking changes
4. Run `buf generate` to generate new code
5. Commit code

### Adding New RPC Methods

```protobuf
service YourService {
  // Add new method
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

### Version Management

API uses semantic versioning, following these rules:

- **Major version** (`v1`, `v2`) - Incompatible API changes
- **Minor version** - Backward compatible feature additions
- **Patch version** - Backward compatible bug fixes

When introducing breaking changes, create a new version directory (e.g. `agent/v2/`).

## FAQ

### 1. How to handle large file transfers?

For large files (such as plugin binaries), recommend:
- Use streaming RPC for chunked transfer
- Or return pre-signed URL, let client download directly from object storage

### 2. How to handle long-running tasks?

Use Stream Service's streaming interface:
- Real-time push task status updates
- Real-time push log output
- Use bidirectional stream to maintain connection

### 3. How to implement task priority?

Add `priority` label in task's `labels`:
```go
labels: map[string]string{
    "priority": "high",
}
```

Agent can sort by priority when FetchTask.

### 4. How to handle Agent disconnect and reconnect?

Agent should:
1. Implement exponential backoff reconnection strategy
2. Re-register after reconnection
3. Report status of incomplete tasks

## Related Documentation

- [Plugin Development Guide](../docs/PLUGIN_DEVELOPMENT.md)
- [Plugin Distribution Guide](../docs/PLUGIN_DISTRIBUTION.md)
- [Implementation Guide](../docs/IMPLEMENTATION_GUIDE.md)
- [Buf Documentation](https://docs.buf.build/)
- [gRPC Documentation](https://grpc.io/docs/)
- [Protocol Buffers Documentation](https://protobuf.dev/)

## Contribution Guide

Contributions welcome! Before submitting PR, please ensure:

1. ✅ All proto files pass `buf lint` check
2. ✅ No breaking changes introduced (or in new version)
3. ✅ Added adequate comments
4. ✅ Generated code is updated
5. ✅ Related documentation is updated

## License

This project uses the license defined in the [LICENSE](../LICENSE) file.
