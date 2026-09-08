package postgres

import (
	"context"
	"errors"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestDBSpanContainsSafeAttributes(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(
			sdktrace.NewSimpleSpanProcessor(exporter),
		),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	original := otel.GetTracerProvider()
	otel.SetTracerProvider(provider)
	t.Cleanup(func() {
		otel.SetTracerProvider(original)
		_ = provider.Shutdown(context.Background())
	})

	_, span := startDBSpan(
		context.Background(),
		"external_item_upsert_batch",
	)

	finishDBSpan(
		span,
		errors.New("database unavailable"),
	)

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected one database span, got %d", len(spans))
	}

	recorded := spans[0]

	if recorded.Name != "db.query" {
		t.Fatalf("expected db.query, got %q", recorded.Name)
	}

	attrs := recorded.Attributes
	values := make(map[attribute.Key]string)

	for _, attr := range attrs {
		values[attr.Key] = attr.Value.AsString()
	}

	if values[attribute.Key("db.system")] != "postgresql" {
		t.Fatalf("expected PostgreSQL db.system, got %q", values[attribute.Key("db.system")])
	}

	if values[attribute.Key("db.operation")] != "external_item_upsert_batch" {
		t.Fatalf("unexpected db.operation: %q", values[attribute.Key("db.operation")])
	}
}
