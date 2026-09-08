package postgres

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "github.com/amyismebyme/the-village/apps/api/internal/repository/postgres"

func startDBSpan(ctx context.Context, operation string) (context.Context, trace.Span) {
	ctx, span := otel.Tracer(tracerName).Start(
		ctx,
		"db.query",
		trace.WithSpanKind(trace.SpanKindClient),
	)

	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", operation),
	)

	return ctx, span
}

func finishDBSpan(span trace.Span, err error) {
	if span == nil {
		return
	}

	if err != nil {
		span.RecordError(err)
		span.SetStatus(
			codes.Error,
			"database operation failed",
		)
	}

	span.End()
}
