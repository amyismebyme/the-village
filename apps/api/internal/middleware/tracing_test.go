package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestTracingCreatesServerSpanWithRouteAndStatus(t *testing.T) {
	previousProvider := otel.GetTracerProvider()
	previousPropagator := otel.GetTextMapPropagator()

	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(
			sdktrace.NewSimpleSpanProcessor(exporter),
		),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	t.Cleanup(func() {
		_ = provider.Shutdown(context.Background())
		otel.SetTracerProvider(previousProvider)
		otel.SetTextMapPropagator(previousPropagator)
	})

	mux := http.NewServeMux()
	mux.HandleFunc(
		"GET /api/v1/communities/{id}",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
	)

	handler := Tracing(mux)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/communities/123?token=secret",
		nil,
	)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf(
			"expected one span, got %d",
			len(spans),
		)
	}

	span := spans[0]

	if span.Name != "GET /api/v1/communities/{id}" {
		t.Fatalf(
			"unexpected span name %q",
			span.Name,
		)
	}

	if got := span.Attributes; len(got) == 0 {
		t.Fatal("expected HTTP span attributes")
	}

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status 200, got %d",
			rec.Code,
		)
	}

	if len(span.SpanContext.TraceID()) == 0 {
		t.Fatal("expected trace ID")
	}

	if len(span.SpanContext.SpanID()) == 0 {
		t.Fatal("expected span ID")
	}

}
