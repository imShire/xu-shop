// Package tracer 封装 OpenTelemetry TracerProvider 装配与关闭。
//
// 行为：
//   - 当 cfg.Enabled=false 或 cfg.Endpoint 为空：返回 noop TracerProvider，Shutdown 为 no-op；
//   - 否则使用 otlptracegrpc exporter（不安全连接，适配内网 Collector）上报；
//   - 采样：ParentBased(TraceIDRatioBased(SampleRatio))；
//   - 全局 propagator：TraceContext + Baggage。
//
// 用法：
//
//	_, shutdown, _ := tracer.Init(ctx, tracer.Config{
//	    Enabled: true, Endpoint: "otel-collector:4317",
//	    ServiceName: "xu-shop-api", SampleRatio: 0.1,
//	})
//	defer shutdown(context.Background())
package tracer

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

// Config 是 tracer 初始化配置。
type Config struct {
	Enabled     bool
	Endpoint    string  // OTLP gRPC endpoint，如 "otel-collector:4317"
	ServiceName string  // resource service.name
	Environment string  // deployment.environment（dev/staging/prod）
	SampleRatio float64 // 采样率 [0,1]，<=0 视为 0（不采样），>=1 视为全采样
	ServiceVer  string  // service.version，可选
}

// ShutdownFunc 优雅关闭 TracerProvider；noop 模式下为空函数。
type ShutdownFunc func(ctx context.Context) error

func noopShutdown(context.Context) error { return nil }

func setupPropagator() {
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{},
	))
}

// Init 初始化 TracerProvider 并设置为全局；返回 provider、shutdown、error。
//
// Endpoint 无效时由 gRPC 在后台异步重连，Init 本身不会因 endpoint 拒绝而失败；
// 调用方可通过 cfg.Enabled=false 显式降级。
func Init(ctx context.Context, cfg Config) (trace.TracerProvider, ShutdownFunc, error) {
	if !cfg.Enabled || cfg.Endpoint == "" {
		tp := noop.NewTracerProvider()
		otel.SetTracerProvider(tp)
		setupPropagator()
		return tp, noopShutdown, nil
	}

	exp, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(cfg.Endpoint),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("tracer: init otlp grpc exporter: %w", err)
	}

	svcName := cfg.ServiceName
	if svcName == "" {
		svcName = "xu-shop"
	}
	attrs := []attribute.KeyValue{semconv.ServiceName(svcName)}
	if cfg.ServiceVer != "" {
		attrs = append(attrs, semconv.ServiceVersion(cfg.ServiceVer))
	}
	if cfg.Environment != "" {
		attrs = append(attrs, semconv.DeploymentEnvironment(cfg.Environment))
	}

	res, err := resource.New(ctx, resource.WithAttributes(attrs...))
	if err != nil {
		_ = exp.Shutdown(ctx)
		return nil, nil, fmt.Errorf("tracer: build resource: %w", err)
	}

	ratio := cfg.SampleRatio
	if ratio < 0 {
		ratio = 0
	} else if ratio > 1 {
		ratio = 1
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(ratio))),
	)
	otel.SetTracerProvider(tp)
	setupPropagator()

	return tp, tp.Shutdown, nil
}
