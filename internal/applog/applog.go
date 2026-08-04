package applog

import (
	"context"
	"log/slog"
	"os"
)

// 阶段 G：结构化日志（slog）+ 从 context 取 request_id。
// 中间件写入 context 后，handler/service 用 FromContext(ctx).Info(...)

type ctxKey struct{}

const RequestIDKey = "request_id"

// Setup 进程级默认 logger。调用一次即可（main）。
func Setup(env string) {
	// TODO(G): 按 env 选 JSONHandler / TextHandler，设默认 Level
	// slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))
	_ = env
	_ = os.Stdout
}

// WithRequestID 把 request_id 放进 context（middleware 调用）。
func WithRequestID(ctx context.Context, requestID string) context.Context {
	// TODO(G): return context.WithValue(ctx, ctxKey{}, requestID)
	_ = requestID
	return ctx
}

// RequestIDFrom 读取 request_id，没有则返回 ""。
func RequestIDFrom(ctx context.Context) string {
	// TODO(G): if v, ok := ctx.Value(ctxKey{}).(string); ok { return v }
	_ = ctx
	return ""
}

// FromContext 返回带 request_id 字段的 slog.Logger。
func FromContext(ctx context.Context) *slog.Logger {
	// TODO(G):
	//  id := RequestIDFrom(ctx)
	//  if id == "" { return slog.Default() }
	//  return slog.Default().With(RequestIDKey, id)
	_ = ctx
	return slog.Default()
}
