// Package tracer 封装 OpenTelemetry TracerProvider 装配与关闭。
//
// 行为：
//   - 若环境变量 OTEL_EXPORTER_OTLP_ENDPOINT 为空，返回 noop TracerProvider，Shutdown 为 no-op；
//   - 否则使用 otlptracehttp exporter 上报到该端点（默认 path /v1/traces），采样 AlwaysSample。
//
// 用法：
//
//	tp, shutdown, _ := tracer.Init(ctx, "xu-shop-api")
//	defer shutdown(context.Background())
//	otel.SetTracerProvider(tp)
package tracer

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

// ShutdownFunc 优雅关闭 TracerProvider；noop 模式下为空函数。
type ShutdownFunc func(ctx context.Context) error

// Init 初始化 TracerProvider 并设置为全局；返回 provider、shutdown、error。
// serviceName 写入 resource attribute service.name。
func Init(ctx context.Context, serviceName string) (trace.TracerProvider, ShutdownFunc, error) {
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		tp := noop.NewTracerProvider()
		otel.SetTracerProvider(tp)
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{}, propagation.Baggage{},
		))
		return tp, func(context.Context) error { return nil }, nil
	}

	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpointURL(endpoint),
		otlptracehttp.WithTimeout(5 * time.Second),
	}
	exp, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("tracer: init otlp exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(semconv.ServiceName(serviceName)),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("tracer: build resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{},
	))

	return tp, tp.Shutdown, nil
}
