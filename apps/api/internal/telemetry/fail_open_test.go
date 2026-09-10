package telemetry

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type failingExporter struct {
	err error
}

func (e failingExporter) ExportSpans(_ context.Context, _ []sdktrace.ReadOnlySpan) error {
	return e.err
}

func (e failingExporter) Shutdown(context.Context) error {
	return nil
}

func TestLoggingExporterObservesExportFailureWithoutPanicking(t *testing.T) {
	logs := new(strings.Builder)
	logger := slog.New(slog.NewTextHandler(logs, nil))
	wantErr := errors.New("tempo unavailable")

	exporter := &loggingExporter{
		exporter: failingExporter{err: wantErr},
		logger:   logger,
	}

	err := exporter.ExportSpans(
		context.Background(),
		nil,
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected exporter error to be returned, got %v", err)
	}
	if !strings.Contains(logs.String(), "OpenTelemetry trace export failed") {
		t.Fatal("expected export failure to be logged")
	}
}

func TestSetupRemainsSuccessfulWhenTempoEndpointIsUnavailable(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	shutdown, err := Setup(context.Background(), Config{
		Enabled:         true,
		ServiceName:     "village-api-test",
		ServiceVersion:  "test",
		Environment:     "test",
		Endpoint:        "http://127.0.0.1:1/v1/traces",
		Insecure:        true,
		Sampler:         "always_on",
		SamplerArgument: 1,
	}, logger)
	if err != nil {
		t.Fatalf("telemetry setup must fail open: %v", err)
	}
	if shutdown == nil {
		t.Fatal("expected shutdown function")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	if err := shutdown(ctx); err != nil {
		t.Fatalf("unexpected telemetry shutdown error: %v", err)
	}
}
