package trace

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

// 默认 sampler：总是采样，确保所有 span 都有有效的 trace ID
var defaultSampler = sdktrace.AlwaysSample()

// Conf Trace 配置
type Conf struct {
	// Enabled 是否启用 trace
	Enabled bool `mapstructure:"enabled"`
	// Endpoint OTLP 端点地址（如：http://localhost:4318 或 localhost:4317）
	Endpoint string `mapstructure:"endpoint"`
	// Protocol 协议类型：grpc 或 http
	Protocol string `mapstructure:"protocol"`
	// ServiceName 服务名称
	ServiceName string `mapstructure:"serviceName"`
	// ServiceVersion 服务版本
	ServiceVersion string `mapstructure:"serviceVersion"`
	// Insecure 是否使用不安全连接（TLS）
	Insecure bool `mapstructure:"insecure"`
	// Headers 额外的 HTTP 头（仅用于 HTTP 协议）
	Headers map[string]string `mapstructure:"headers"`
	// BatchTimeout 批量发送超时时间（秒）
	BatchTimeout int `mapstructure:"batchTimeout"`
	// ExportTimeout 导出超时时间（秒）
	ExportTimeout int `mapstructure:"exportTimeout"`
	// MaxExportBatchSize 最大批量大小
	MaxExportBatchSize int `mapstructure:"maxExportBatchSize"`
}

// SetDefaults 设置默认值
func (c *Conf) SetDefaults() {
	if c.ServiceName == "" {
		c.ServiceName = "arcade"
	}
	if c.ServiceVersion == "" {
		c.ServiceVersion = "1.0.0"
	}
	if c.Protocol == "" {
		c.Protocol = "grpc"
	}
	if c.BatchTimeout == 0 {
		c.BatchTimeout = 5
	}
	if c.ExportTimeout == 0 {
		c.ExportTimeout = 30
	}
	if c.MaxExportBatchSize == 0 {
		c.MaxExportBatchSize = 512
	}
	if c.Endpoint == "" {
		if c.Protocol == "grpc" {
			c.Endpoint = "localhost:4317"
		} else {
			c.Endpoint = "http://localhost:4318"
		}
	}
}

// InitTracerProvider 初始化 TracerProvider
// 如果 conf.Enabled = false，返回 NoOp TracerProvider，确保能创建有效的 span
func InitTracerProvider(ctx context.Context, conf Conf) (*sdktrace.TracerProvider, func(), error) {
	if !conf.Enabled {
		// 如果未启用，创建 NoOp TracerProvider，确保能创建有效的 span（虽然不会上报）
		tp := sdktrace.NewTracerProvider(
			sdktrace.WithSampler(defaultSampler), // 使用默认 sampler，确保所有 span 都有有效的 trace ID
		)
		otel.SetTracerProvider(tp)
		return tp, func() {}, nil
	}

	conf.SetDefaults()

	// 创建资源
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(conf.ServiceName),
			semconv.ServiceVersionKey.String(conf.ServiceVersion),
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// 创建 exporter
	var exporter sdktrace.SpanExporter
	exporter, err = createExporter(ctx, conf)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	// 创建 TracerProvider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(time.Duration(conf.BatchTimeout)*time.Second),
			sdktrace.WithExportTimeout(time.Duration(conf.ExportTimeout)*time.Second),
			sdktrace.WithMaxExportBatchSize(conf.MaxExportBatchSize),
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(defaultSampler), // 使用默认 sampler，确保所有 span 都有有效的 trace ID
	)

	// 设置全局 TracerProvider
	otel.SetTracerProvider(tp)

	// 设置全局传播器（使用 W3C TraceContext 和 Baggage）
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	// 返回清理函数
	cleanup := func() {
		// 使用 ExportTimeout + 5秒缓冲时间，但至少10秒，最多30秒
		shutdownTimeout := min(max(time.Duration(conf.ExportTimeout)*time.Second+5*time.Second, 10*time.Second), 30 * time.Second)

		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := tp.Shutdown(shutdownCtx); err != nil {
			// 对于超时错误，只记录警告，因为这是预期的行为（特别是在网络不好的情况下）
			if err == context.DeadlineExceeded {
				fmt.Printf("TracerProvider shutdown timeout after %v (this is expected if exporter is still sending data)\n", shutdownTimeout)
			} else {
				// 使用标准库日志，因为此时可能 logger 已经关闭
				fmt.Printf("failed to shutdown TracerProvider: %v\n", err)
			}
		}
	}

	return tp, cleanup, nil
}

// createExporter 创建 exporter
func createExporter(ctx context.Context, conf Conf) (sdktrace.SpanExporter, error) {
	switch conf.Protocol {
	case "grpc":
		return createGRPCExporter(ctx, conf)
	case "http":
		return createHTTPExporter(ctx, conf)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", conf.Protocol)
	}
}

// createGRPCExporter 创建 gRPC exporter
func createGRPCExporter(ctx context.Context, conf Conf) (sdktrace.SpanExporter, error) {
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(conf.Endpoint),
	}

	if conf.Insecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}

	if conf.ExportTimeout > 0 {
		opts = append(opts, otlptracegrpc.WithTimeout(time.Duration(conf.ExportTimeout)*time.Second))
	}

	client := otlptracegrpc.NewClient(opts...)
	return otlptrace.New(ctx, client)
}

// createHTTPExporter 创建 HTTP exporter
func createHTTPExporter(ctx context.Context, conf Conf) (sdktrace.SpanExporter, error) {
	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(conf.Endpoint),
	}

	if conf.Insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	if len(conf.Headers) > 0 {
		opts = append(opts, otlptracehttp.WithHeaders(conf.Headers))
	}

	if conf.ExportTimeout > 0 {
		opts = append(opts, otlptracehttp.WithTimeout(time.Duration(conf.ExportTimeout)*time.Second))
	}

	client := otlptracehttp.NewClient(opts...)
	return otlptrace.New(ctx, client)
}

// GetTracer 获取 Tracer（便捷方法）
func GetTracer(name string) trace.Tracer {
	return otel.Tracer(name)
}
