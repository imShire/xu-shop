// Package logger 封装 zap，提供全局 logger 和 context-aware logger。
package logger

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type contextKey string

const loggerKey contextKey = "logger"

var global *zap.Logger

// Init 初始化全局 logger，以 JSON 格式输出到 stdout。
func Init(level string) {
	lvl := zapcore.InfoLevel
	_ = lvl.UnmarshalText([]byte(level))

	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(lvl)
	cfg.EncoderConfig.EncodeTime = zapcore.RFC3339NanoTimeEncoder

	logger, err := cfg.Build(zap.AddCallerSkip(0))
	if err != nil {
		panic(err)
	}
	global = logger
}

// L 返回全局 logger；若未初始化则返回 nop logger。
func L() *zap.Logger {
	if global == nil {
		return zap.NewNop()
	}
	return global
}

// WithContext 将 logger 注入 context。
func WithContext(ctx context.Context, l *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

// Ctx 从 context 取 logger，并注入 request_id / trace_id 字段。
func Ctx(ctx context.Context) *zap.Logger {
	if l, ok := ctx.Value(loggerKey).(*zap.Logger); ok && l != nil {
		return l
	}
	return L()
}
