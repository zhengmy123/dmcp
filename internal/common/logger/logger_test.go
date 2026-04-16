package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newTestLogger() (*ZapLogger, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.AddSync(buf),
		zap.DebugLevel,
	)
	z := zap.New(core)
	return &ZapLogger{zap: z}, buf
}

func TestGetTraceID_Empty(t *testing.T) {
	ctx := context.Background()
	if id := GetTraceID(ctx); id != "" {
		t.Fatalf("expected empty, got %q", id)
	}
}

func TestWithTraceID_RoundTrip(t *testing.T) {
	ctx := context.Background()
	ctx = WithTraceID(ctx, "test-trace-123")
	if id := GetTraceID(ctx); id != "test-trace-123" {
		t.Fatalf("expected test-trace-123, got %q", id)
	}
}

func TestCtx_NoTraceID(t *testing.T) {
	log, buf := newTestLogger()
	ctx := context.Background()

	subLog := log.Ctx(ctx)
	// Should return same logger (no trace_id)
	subLog.Info("test")

	var entry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatal(err)
	}
	if _, exists := entry["trace_id"]; exists {
		t.Fatal("trace_id should not exist when ctx has no trace_id")
	}
}

func TestCtx_WithTraceID(t *testing.T) {
	log, buf := newTestLogger()
	ctx := context.Background()
	ctx = WithTraceID(ctx, "abc-123-def")

	subLog := log.Ctx(ctx)
	subLog.Info("test message")

	var entry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatal(err)
	}
	if id, ok := entry["trace_id"].(string); !ok || id != "abc-123-def" {
		t.Fatalf("expected trace_id=abc-123-def, got %v", entry["trace_id"])
	}
}

func TestCtx_WithTraceID_MultipleLogs(t *testing.T) {
	log, buf := newTestLogger()
	ctx := context.Background()
	ctx = WithTraceID(ctx, "multi-trace-id")

	subLog := log.Ctx(ctx)
	subLog.Info("first")
	subLog.Warn("second")
	subLog.Error("third")

	// All 3 lines should have trace_id
	lines := bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte("\n"))
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	for i, line := range lines {
		var entry map[string]interface{}
		if err := json.Unmarshal(line, &entry); err != nil {
			t.Fatal(err)
		}
		if id, ok := entry["trace_id"].(string); !ok || id != "multi-trace-id" {
			t.Fatalf("line %d: expected trace_id=multi-trace-id, got %v", i, entry["trace_id"])
		}
	}
}
