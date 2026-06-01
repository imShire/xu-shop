// Package tracer 提供 asynq worker 的 OpenTelemetry tracing middleware。
package tracer

import (
	"context"

	"github.com/hibiken/asynq"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const asynqTracerName = "github.com/xushop/xu-shop/internal/pkg/tracer/asynq"

// AsynqMiddleware 为每个 asynq 任务创建一个 span，包含 task.type 属性。
// 失败时记录 status=Error 并 RecordError。
//
// 用法：mux.Use(tracer.AsynqMiddleware)
func AsynqMiddleware(next asynq.Handler) asynq.Handler {
	return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		tracer := otel.GetTracerProvider().Tracer(asynqTracerName)
		ctx, span := tracer.Start(ctx, "asynq.process "+t.Type(),
			trace.WithSpanKind(trace.SpanKindConsumer),
			trace.WithAttributes(attribute.String("asynq.task.type", t.Type())),
		)
		defer span.End()

		err := next.ProcessTask(ctx, t)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		return err
	})
}
