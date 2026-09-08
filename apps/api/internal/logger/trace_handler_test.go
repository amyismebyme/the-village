package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"go.opentelemetry.io/otel/trace"
)

func TestTraceHandlerAddsTraceAndSpanIDs(t *testing.T) {
	var buffer bytes.Buffer

	base := slog.NewJSONHandler(&buffer, nil)
	logger := slog.New(newTraceHandler(base))

	spanCtx := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    trace.TraceID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		SpanID:     trace.SpanID{1, 2, 3, 4, 5, 6, 7, 8},
		TraceFlags: trace.FlagsSampled,
	})

	ctx := trace.ContextWithSpanContext(
		context.Background(),
		spanCtx,
	)

	logger.InfoContext(ctx, "correlated log")

	var payload map[string]any
	if err := json.Unmarshal(buffer.Bytes(), &payload); err != nil {
		t.Fatalf("decode log: %v", err)
	}

	if payload["trace_id"] != spanCtx.TraceID().String() {
		t.Fatalf("unexpected trace_id: %v", payload["trace_id"])
	}

	if payload["span_id"] != spanCtx.SpanID().String() {
		t.Fatalf("unexpected span_id: %v", payload["span_id"])
	}
}

func TestTraceHandlerDoesNotAddIDsWithoutValidSpan(t *testing.T) {
	var buffer bytes.Buffer

	base := slog.NewJSONHandler(&buffer, nil)
	logger := slog.New(newTraceHandler(base))

	logger.InfoContext(context.Background(), "uncorrelated log")

	var payload map[string]any
	if err := json.Unmarshal(buffer.Bytes(), &payload); err != nil {
		t.Fatalf("decode log: %v", err)
	}

	if _, ok := payload["trace_id"]; ok {
		t.Fatal("unexpected trace_id without valid span")
	}

	if _, ok := payload["span_id"]; ok {
		t.Fatal("unexpected span_id without valid span")
	}
}
