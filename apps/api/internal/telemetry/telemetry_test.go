package telemetry

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

func TestLoadFromEnvDefaults(t *testing.T) {
	cfg := LoadFromEnv(func(string) string { return "" })
	if cfg.Enabled {
		t.Fatal("expected telemetry disabled by default")
	}
	if cfg.ServiceName != "village-api" {
		t.Fatalf("unexpected service name %q", cfg.ServiceName)
	}
	if cfg.Endpoint != "http://localhost:4318/v1/traces" {
		t.Fatalf("unexpected endpoint %q", cfg.Endpoint)
	}
}

func TestSetupDisabledUsesNoopProvider(t *testing.T) {
	shutdown, err := Setup(context.Background(), Config{Enabled: false}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("setup telemetry: %v", err)
	}
	if shutdown == nil {
		t.Fatal("expected shutdown function")
	}

	span := otel.Tracer("test").Start
	ctx, s := span(context.Background(), "disabled")
	_ = ctx
	if s.SpanContext().IsValid() {
		t.Fatal("expected noop span context")
	}
	s.End()
}

func TestSetupFailsOpenForInvalidExporterConfig(t *testing.T) {
	logs := new(strings.Builder)
	logger := slog.New(slog.NewTextHandler(logs, nil))

	shutdown, err := Setup(context.Background(), Config{
		Enabled:     true,
		ServiceName: "test",
		Endpoint:    "://bad",
	}, logger)
	if err != nil {
		t.Fatalf("expected fail-open setup, got error: %v", err)
	}
	if shutdown == nil {
		t.Fatal("expected shutdown function")
	}
	if !strings.Contains(logs.String(), "tracing disabled") {
		t.Fatal("expected exporter failure to be logged")
	}
}

var _ trace.Tracer = otel.Tracer("compile-check")
