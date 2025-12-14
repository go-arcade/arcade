# Trace 配置说明

## 配置位置

OpenTelemetry 的上报目标配置在 `conf.d/config.toml` 文件的 `[trace]` 部分。

## 配置示例

```toml
[trace]
# 是否启用 trace
enabled = true
# OTLP 端点地址
# gRPC 协议: "localhost:4317" 或 "otel-collector:4317"
# HTTP 协议: "http://localhost:4318" 或 "http://otel-collector:4318"
endpoint = "localhost:4317"
# 协议类型: grpc 或 http
protocol = "grpc"
# 服务名称
serviceName = "arcade"
# 服务版本
serviceVersion = "1.0.0"
# 是否使用不安全连接（不使用 TLS）
insecure = true
# 批量发送超时时间（秒）
batchTimeout = 5
# 导出超时时间（秒）
exportTimeout = 30
# 最大批量大小
maxExportBatchSize = 512

# HTTP 协议的额外头信息（仅用于 HTTP 协议）
[trace.headers]
Authorization = "Bearer your-token"
```

## 常见配置场景

### 1. 使用 gRPC 协议连接 OTel Collector

```toml
[trace]
enabled = true
endpoint = "otel-collector:4317"
protocol = "grpc"
insecure = true
```

### 2. 使用 HTTP 协议连接 OTel Collector

```toml
[trace]
enabled = true
endpoint = "http://otel-collector:4318"
protocol = "http"
insecure = true
```

### 3. 使用 TLS 连接（安全连接）

```toml
[trace]
enabled = true
endpoint = "otel-collector:4317"
protocol = "grpc"
insecure = false  # 使用 TLS
```

### 4. HTTP 协议带认证

```toml
[trace]
enabled = true
endpoint = "http://otel-collector:4318"
protocol = "http"
insecure = true

[trace.headers]
Authorization = "Bearer your-api-token"
```

## 依赖安装

在使用前，需要安装 OpenTelemetry OTLP exporter 依赖：

```bash
go get go.opentelemetry.io/otel/exporters/otlp/otlptrace
go get go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc
go get go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp
go get go.opentelemetry.io/otel/sdk/resource
go get go.opentelemetry.io/otel/semconv/v1.24.0
```

安装依赖后，正常编译即可：

```bash
go build ./cmd/arcade
```

## 初始化位置

TracerProvider 在 `internal/engine/bootstrap/bootstrap.go` 的 `Bootstrap` 函数中初始化，应用启动时会自动读取配置并初始化。

## 工作原理

1. 应用启动时，`Bootstrap` 函数会读取 `[trace]` 配置
2. 如果 `enabled = true`，会创建 OTLP exporter 并连接到指定的端点
3. 所有通过 `pkg/trace/inject` 包创建的 span 会自动上报到配置的端点
4. 应用关闭时，会自动刷新并关闭 TracerProvider
