package tracer

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel/trace/noop"
)

func TestInit_DisabledReturnsNoop(t *testing.T) {
	tp, shutdown, err := Init(context.Background(), Config{Enabled: false})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := tp.(noop.TracerProvider); !ok {
		t.Errorf("expected noop tracer provider, got %T", tp)
	}
	if err := shutdown(context.Background()); err != nil {
		t.Errorf("noop shutdown should not error: %v", err)
	}
}

func TestInit_EmptyEndpointReturnsNoop(t *testing.T) {
	tp, shutdown, err := Init(context.Background(), Config{Enabled: true, Endpoint: ""})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := tp.(noop.TracerProvider); !ok {
		t.Errorf("expected noop tracer provider, got %T", tp)
	}
	if err := shutdown(context.Background()); err != nil {
		t.Errorf("noop shutdown should not error: %v", err)
	}
}

// TestInit_SampleRatioBounds 验证采样比例边界值（0/1/0.5）下 Init 都能正常返回。
// 由于 endpoint 为空走 noop 分支，这里通过 Enabled+Endpoint 走真实分支但用本地不可达地址，
// gRPC 异步连接，Init 不会立即失败。
func TestInit_SampleRatioBounds(t *testing.T) {
	cases := []float64{0.0, 0.5, 1.0, -1, 2}
	for _, r := range cases {
		r := r
		t.Run("", func(t *testing.T) {
			tp, shutdown, err := Init(context.Background(), Config{
				Enabled:     true,
				Endpoint:    "127.0.0.1:0", // 任意 endpoint，gRPC 异步重连
				ServiceName: "xu-shop-test",
				Environment: "test",
				SampleRatio: r,
			})
			if err != nil {
				t.Fatalf("ratio=%v init error: %v", r, err)
			}
			if tp == nil {
				t.Fatalf("ratio=%v provider nil", r)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 0)
			cancel()
			// Shutdown 应该可以被调用，即便 ctx 已 cancel 也只会返回 ctx error 或 nil。
			_ = shutdown(ctx)
		})
	}
}
