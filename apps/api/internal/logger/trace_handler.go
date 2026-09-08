package logger

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

// traceHandler adds trace/span identifiers to structured logs whenever the
// logging context contains a valid OpenTelemetry span. It deliberately uses
// the OpenTelemetry API package only; the logger remains independent of any
// exporter/backend such as Tempo.
type traceHandler struct {
	next slog.Handler
}

func newTraceHandler(next slog.Handler) slog.Handler {
	return &traceHandler{next: next}
}

func (h *traceHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *traceHandler) Handle(
	ctx context.Context,
	record slog.Record,
) error {
	sc := trace.SpanContextFromContext(ctx)
	if sc.IsValid() {
		record = record.Clone()
		record.AddAttrs(
			slog.String("trace_id", sc.TraceID().String()),
			slog.String("span_id", sc.SpanID().String()),
		)
	}

	return h.next.Handle(ctx, record)
}

func (h *traceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &traceHandler{
		next: h.next.WithAttrs(attrs),
	}
}

func (h *traceHandler) WithGroup(name string) slog.Handler {
	return &traceHandler{
		next: h.next.WithGroup(name),
	}
}

var _ slog.Handler = (*traceHandler)(nil)
