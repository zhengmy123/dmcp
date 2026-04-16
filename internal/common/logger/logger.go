package logger

import (
	"context"

	"go.uber.org/zap"
)

// Logger 统一日志接口，解耦具体实现
type Logger interface {
	// 基础日志
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)

	// 格式化日志
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})

	// With 返回带有预设字段的子 Logger
	With(fields ...Field) Logger

	// Ctx 从 context 自动提取 trace_id，返回携带该字段的子 Logger
	// 如果 ctx 中没有 trace_id，返回自身（不创建子 Logger）
	Ctx(ctx context.Context) Logger

	// Sync 刷新缓冲
	Sync() error
}

// Field 日志字段，底层映射到 zap.Field
type Field = zap.Field

// 便捷构造 Field 的函数（直接代理 zap）
var (
	String   = zap.String
	Int      = zap.Int
	Int64    = zap.Int64
	Uint     = zap.Uint
	Float64  = zap.Float64
	Bool     = zap.Bool
	Error    = zap.Error
	Any      = zap.Any
	Duration = zap.Duration
	Time     = zap.Time
)
