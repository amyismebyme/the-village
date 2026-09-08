package service

import (
	"context"
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/model"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func TestCommunityServiceCreatesOperationSpan(t *testing.T) {
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

	repo := newMockCommunityRepository()
	svc := NewCommunityService(repo)

	err := svc.Create(
		context.Background(),
		&model.Community{
			Name: "Toronto Men",
			Slug: "toronto-men",
		},
	)
	if err != nil {
		t.Fatalf("create community: %v", err)
	}

	spans := exporter.GetSpans()
	found := false

	for _, span := range spans {
		if span.Name == "community.create" {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("expected community.create span, got %v spans", len(spans))
	}
}

func TestCommunityServiceSpanUsesCallerTraceContext(t *testing.T) {
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

	tracer := provider.Tracer("test")
	ctx, parent := tracer.Start(
		context.Background(),
		"parent",
	)

	repo := newMockCommunityRepository()
	svc := NewCommunityService(repo)

	err := svc.Create(
		ctx,
		&model.Community{
			Name: "Toronto Men",
			Slug: "toronto-men",
		},
	)
	parent.End()

	if err != nil {
		t.Fatalf("create community: %v", err)
	}

	var child trace.SpanContext

	for _, span := range exporter.GetSpans() {
		if span.Name == "community.create" {
			child = span.SpanContext
			break
		}
	}

	if !child.IsValid() {
		t.Fatal("expected valid child span context")
	}
}
