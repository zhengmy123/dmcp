package logger

import "context"

// traceKey 用于 context 存取 trace_id 的键
type traceKey struct{}

// WithTraceID 向 context 注入 trace_id
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceKey{}, traceID)
}

// GetTraceID 从 context 提取 trace_id，不存在返回空字符串
func GetTraceID(ctx context.Context) string {
	if v, ok := ctx.Value(traceKey{}).(string); ok {
		return v
	}
	return ""
}
