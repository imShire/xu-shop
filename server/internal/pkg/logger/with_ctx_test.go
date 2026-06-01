package logger

import (
	"context"
	"strings"
	"testing"

	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func newObservedLogger() (*zap.Logger, *observer.ObservedLogs) {
	core, obs := observer.New(zapcore.DebugLevel)
	return zap.New(core), obs
}

func TestWithCtx_NoSpanReturnsBaseLogger(t *testing.T) {
	base, obs := newObservedLogger()
	ctx := WithContext(context.Background(), base)

	WithCtx(ctx).Info("hi")
	logs := obs.All()
	if len(logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(logs))
	}
	for _, f := range logs[0].Context {
		if f.Key == "trace_id" || f.Key == "span_id" {
			t.Errorf("did not expect trace fields, got key=%s", f.Key)
		}
	}
}

func TestWithCtx_WithSpanInjectsTraceID(t *testing.T) {
	base, obs := newObservedLogger()
	exp := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exp),
		trace.WithSampler(trace.AlwaysSample()),
	)
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })

	ctx := WithContext(context.Background(), base)
	ctx, span := tp.Tracer("test").Start(ctx, "op")
	defer span.End()

	WithCtx(ctx).Info("hi")

	logs := obs.All()
	if len(logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(logs))
	}
	var hasTrace, hasSpan bool
	var traceID, spanID string
	for _, f := range logs[0].Context {
		switch f.Key {
		case "trace_id":
			hasTrace = true
			traceID = f.String
		case "span_id":
			hasSpan = true
			spanID = f.String
		}
	}
	if !hasTrace || !hasSpan {
		t.Fatalf("expected trace_id and span_id in fields, got %+v", logs[0].Context)
	}
	if strings.Trim(traceID, "0") == "" || strings.Trim(spanID, "0") == "" {
		t.Errorf("trace_id/span_id should be non-zero, got trace=%s span=%s", traceID, spanID)
	}
}
