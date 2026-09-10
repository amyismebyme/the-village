package telemetry

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

// Setup configures tracing and intentionally fails open. Tempo/exporter
// availability must never determine API startup availability.
func Setup(ctx context.Context, cfg Config, logger *slog.Logger) (func(context.Context) error, error) {
	if ctx == nil {
		return nil, errors.New("telemetry: context is required")
	}
	if logger == nil {
		logger = slog.Default()
	}

	otel.SetTextMapPropagator(propagation.TraceContext{})

	if !cfg.Enabled {
		otel.SetTracerProvider(noop.NewTracerProvider())
		return func(context.Context) error { return nil }, nil
	}

	exporter, err := newExporter(ctx, cfg)
	if err != nil {
		logger.ErrorContext(ctx, "OpenTelemetry exporter initialization failed; tracing disabled", "error", err)
		otel.SetTracerProvider(noop.NewTracerProvider())
		return func(context.Context) error { return nil }, nil
	}

	s, err := sampler(cfg)
	if err != nil {
		logger.ErrorContext(ctx, "OpenTelemetry sampler configuration failed; tracing disabled", "error", err)
		_ = exporter.Shutdown(ctx)
		otel.SetTracerProvider(noop.NewTracerProvider())
		return func(context.Context) error { return nil }, nil
	}

	resourceAttrs := []attribute.KeyValue{
		attribute.String("service.name", cfg.ServiceName),
		attribute.String("service.version", cfg.ServiceVersion),
		attribute.String("deployment.environment.name", cfg.Environment),
	}

	res, err := sdkresource.Merge(
		sdkresource.Default(),
		sdkresource.NewSchemaless(resourceAttrs...),
	)
	if err != nil {
		logger.ErrorContext(ctx, "OpenTelemetry resource initialization failed; tracing disabled", "error", err)
		_ = exporter.Shutdown(ctx)
		otel.SetTracerProvider(noop.NewTracerProvider())
		return func(context.Context) error { return nil }, nil
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(&loggingExporter{exporter: exporter, logger: logger}),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(s),
	)
	otel.SetTracerProvider(provider)

	logger.InfoContext(ctx, "OpenTelemetry tracing initialized", "service_name", cfg.ServiceName, "endpoint", sanitizeEndpoint(cfg.Endpoint), "sampler", cfg.Sampler)

	return func(shutdownCtx context.Context) error {
		if shutdownCtx == nil {
			return errors.New("telemetry: shutdown context is required")
		}
		if err := provider.Shutdown(shutdownCtx); err != nil {
			logger.ErrorContext(shutdownCtx, "OpenTelemetry shutdown failed", "error", err)
			return err
		}
		return nil
	}, nil
}

func newExporter(ctx context.Context, cfg Config) (sdktrace.SpanExporter, error) {
	parsed, err := url.Parse(cfg.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("parse OTLP endpoint: %w", err)
	}
	if parsed.Host == "" {
		return nil, errors.New("OTLP endpoint host is required")
	}

	options := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(parsed.Host),
	}
	if path := parsed.EscapedPath(); path != "" && path != "/" {
		options = append(options, otlptracehttp.WithURLPath(path))
	}
	if cfg.Insecure || strings.EqualFold(parsed.Scheme, "http") {
		options = append(options, otlptracehttp.WithInsecure())
	}

	exporter, err := otlptracehttp.New(ctx, options...)
	if err != nil {
		return nil, fmt.Errorf("create OTLP/HTTP exporter: %w", err)
	}
	return exporter, nil
}

func sanitizeEndpoint(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return "invalid"
	}
	parsed.User = nil
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String()
}

type loggingExporter struct {
	exporter sdktrace.SpanExporter
	logger   *slog.Logger
}

func (e *loggingExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	err := e.exporter.ExportSpans(ctx, spans)
	if err != nil {
		e.logger.ErrorContext(ctx, "OpenTelemetry trace export failed", "error", err, "span_count", len(spans))
	}
	return err
}

func (e *loggingExporter) Shutdown(ctx context.Context) error {
	return e.exporter.Shutdown(ctx)
}
