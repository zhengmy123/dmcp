package logger

import (
	"context"

	"go.uber.org/zap"
)

// ZapLogger 基于 zap 的 Logger 实现
type ZapLogger struct {
	zap *zap.Logger
}

// NewZapLogger 从 zap.Logger 创建 Logger
func NewZapLogger(z *zap.Logger) Logger {
	return &ZapLogger{zap: z}
}

func (l *ZapLogger) Debug(msg string, fields ...Field) {
	l.zap.Debug(msg, fields...)
}

func (l *ZapLogger) Info(msg string, fields ...Field) {
	l.zap.Info(msg, fields...)
}

func (l *ZapLogger) Warn(msg string, fields ...Field) {
	l.zap.Warn(msg, fields...)
}

func (l *ZapLogger) Error(msg string, fields ...Field) {
	l.zap.Error(msg, fields...)
}

func (l *ZapLogger) Fatal(msg string, fields ...Field) {
	l.zap.Fatal(msg, fields...)
}

func (l *ZapLogger) Debugf(format string, args ...interface{}) {
	l.zap.Sugar().Debugf(format, args...)
}

func (l *ZapLogger) Infof(format string, args ...interface{}) {
	l.zap.Sugar().Infof(format, args...)
}

func (l *ZapLogger) Warnf(format string, args ...interface{}) {
	l.zap.Sugar().Warnf(format, args...)
}

func (l *ZapLogger) Errorf(format string, args ...interface{}) {
	l.zap.Sugar().Errorf(format, args...)
}

func (l *ZapLogger) With(fields ...Field) Logger {
	return &ZapLogger{zap: l.zap.With(fields...)}
}

// Ctx 从 context 自动提取 trace_id，返回携带该字段的子 Logger
func (l *ZapLogger) Ctx(ctx context.Context) Logger {
	if traceID := GetTraceID(ctx); traceID != "" {
		return &ZapLogger{zap: l.zap.With(zap.String("trace_id", traceID))}
	}
	return l
}

func (l *ZapLogger) Sync() error {
	return l.zap.Sync()
}

// Unwrap 返回底层 *zap.Logger（用于兼容旧代码渐进迁移）
func (l *ZapLogger) Unwrap() *zap.Logger {
	return l.zap
}
