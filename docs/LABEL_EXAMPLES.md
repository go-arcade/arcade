# Label系统使用示例

本文档提供Label系统的实际使用示例，帮助您快速上手。

## 场景1：基于环境的任务路由

### Agent配置

```yaml
# 生产环境Agent
agent:
  labels:
    env: production
    region: us-west
    team: platform

# 测试环境Agent
agent:
  labels:
    env: staging
    region: cn-north
    team: qa
```

### 任务配置

```yaml
# 只在生产环境执行的部署任务
job:
  name: deploy-production
  label_selector:
    match_labels:
      env: production
      region: us-west
```

## 场景2：基于硬件能力的任务分配

### Agent配置

```yaml
# 高性能GPU服务器
agent:
  labels:
    gpu: "true"
    gpu-model: nvidia-a100
    cpu-cores: "32"
    memory-gb: "128"
    capability: ml-training

# 普通构建服务器
agent:
  labels:
    gpu: "false"
    cpu-cores: "8"
    memory-gb: "16"
    capability: build
```

### 任务配置

```yaml
# AI模型训练任务
job:
  name: train-model
  label_selector:
    match_expressions:
      - key: gpu
        operator: EXISTS
      - key: memory-gb
        operator: GT
        values: ["64"]

# 普通编译任务
job:
  name: build-project
  label_selector:
    match_labels:
      capability: build
```

## 场景3：多环境灵活部署

### Agent配置

```yaml
# Agent 1 - 支持多环境
agent:
  labels:
    env: production
    region: us-east
    platform: kubernetes
    deploy-type: blue-green

# Agent 2 - 专用于金丝雀部署
agent:
  labels:
    env: production
    region: us-west
    platform: kubernetes
    deploy-type: canary
```

### 任务配置

```yaml
# 蓝绿部署任务
job:
  name: deploy-blue-green
  label_selector:
    match_labels:
      env: production
      platform: kubernetes
      deploy-type: blue-green

# 金丝雀部署任务
job:
  name: deploy-canary
  label_selector:
    match_labels:
      env: production
      deploy-type: canary
```

## 场景4：复杂表达式匹配

### 任务配置

```yaml
# 在staging或production环境，且有GPU的Agent上执行
job:
  name: test-ml-model
  label_selector:
    match_expressions:
      - key: env
        operator: IN
        values: ["staging", "production"]
      - key: gpu
        operator: EXISTS
      - key: cpu-cores
        operator: GT
        values: ["16"]
      - key: region
        operator: NOT_IN
        values: ["deprecated-region"]
```

## 场景5：版本控制和滚动升级

### Agent配置

```yaml
# 旧版本Agent
agent:
  labels:
    agent-version: v1.0.0
    feature-flags: basic

# 新版本Agent
agent:
  labels:
    agent-version: v2.0.0
    feature-flags: advanced,streaming
```

### 任务配置

```yaml
# 使用新特性的任务
job:
  name: advanced-build
  label_selector:
    match_expressions:
      - key: agent-version
        operator: IN
        values: ["v2.0.0", "v2.1.0"]

# 向后兼容的任务
job:
  name: legacy-build
  label_selector:
    match_expressions:
      - key: agent-version
        operator: EXISTS
```

## 场景6：团队和项目隔离

### Agent配置

```yaml
# 前端团队Agent
agent:
  labels:
    team: frontend
    project: web-app
    language: nodejs
    environment: production

# 后端团队Agent
agent:
  labels:
    team: backend
    project: api-service
    language: golang
    environment: production
```

### 任务配置

```yaml
# 前端构建任务
job:
  name: build-frontend
  label_selector:
    match_labels:
      team: frontend
      language: nodejs

# 后端构建任务
job:
  name: build-backend
  label_selector:
    match_labels:
      team: backend
      language: golang
```

## 场景7：动态更新Agent标签

### 使用UpdateLabels RPC

```go
// 场景：Agent检测到新安装了GPU
client.UpdateLabels(ctx, &UpdateLabelsRequest{
    AgentId: "agent-123",
    Labels: map[string]string{
        "gpu": "true",
        "gpu-model": "nvidia-rtx-4090",
    },
    Merge: true,  // 合并模式，不会删除其他labels
})

// 场景：Agent升级完成，更新版本标签
client.UpdateLabels(ctx, &UpdateLabelsRequest{
    AgentId: "agent-123",
    Labels: map[string]string{
        "agent-version": "v2.1.0",
        "features": "streaming,caching,gpu",
    },
    Merge: true,
})
```

### 在心跳中更新标签

```go
// Agent在心跳时动态上报labels
client.Heartbeat(ctx, &HeartbeatRequest{
    AgentId: "agent-123",
    Status: AGENT_STATUS_IDLE,
    Labels: map[string]string{
        "current-load": "low",
        "available-memory": "80",  // 百分比
        "status": "healthy",
    },
})
```

## Label命名建议

### 推荐的Label Key格式

```yaml
# 环境相关
env: production|staging|dev
region: us-west|us-east|cn-north|cn-south
zone: az-1|az-2

# 能力相关
capability: build|deploy|test
platform: kubernetes|docker|vm
arch: amd64|arm64

# 硬件相关
gpu: true|false
gpu-model: nvidia-a100|nvidia-v100
cpu-cores: "8"|"16"|"32"
memory-gb: "16"|"32"|"64"|"128"

# 团队相关
team: frontend|backend|infra|data
project: web-app|api-service|ml-platform
owner: team-a|team-b

# 版本相关
agent-version: v1.0.0|v2.0.0
feature-flags: basic|advanced|experimental

# 状态相关（动态更新）
current-load: low|medium|high
health-status: healthy|degraded|unhealthy
```

## 高级技巧

### 1. 使用NOT_IN排除特定Agent

```yaml
# 排除维护中的Agent
label_selector:
  match_expressions:
    - key: status
      operator: NOT_IN
      values: ["maintenance", "upgrading"]
```

### 2. 组合多个条件

```yaml
# 必须在生产环境，必须有GPU，CPU核心数大于16
label_selector:
  match_labels:
    env: production
  match_expressions:
    - key: gpu
      operator: EXISTS
    - key: cpu-cores
      operator: GT
      values: ["16"]
    - key: status
      operator: NOT_IN
      values: ["maintenance"]
```

### 3. 使用EXISTS检查能力

```yaml
# 检查Agent是否支持特定能力
label_selector:
  match_expressions:
    - key: kubernetes-support
      operator: EXISTS
    - key: docker-support
      operator: EXISTS
```

## 常见问题

### Q1: match_labels和match_expressions的区别？
- `match_labels`: 精确匹配，多个条件是AND关系
- `match_expressions`: 支持更复杂的操作符（IN、NOT_IN、EXISTS等）

### Q2: 如何实现OR逻辑？
使用`LABEL_OPERATOR_IN`操作符：
```yaml
match_expressions:
  - key: env
    operator: IN
    values: ["staging", "production"]  # staging OR production
```

### Q3: Label值可以是数字吗？
Label值始终是字符串，但可以使用GT/LT操作符进行数值比较。

### Q4: 动态更新Label会立即生效吗？
是的，通过UpdateLabels或Heartbeat更新的Label会立即在Server端生效，影响后续的任务匹配。

### Q5: 一个任务如果没有匹配的Agent怎么办？
任务会保持在PENDING状态，直到有符合条件的Agent上线并拉取任务。

