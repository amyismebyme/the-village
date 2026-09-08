package telemetry

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

const instrumentationName = "github.com/amyismebyme/the-village/apps/api"

// Provider owns the application tracer provider and its exporter lifecycle.
type Provider struct {
	tracerProvider *sdktrace.TracerProvider
}

func NewProvider(
	ctx context.Context,
	cfg Config,
) (*Provider, error) {
	if !cfg.Enabled {
		return &Provider{}, nil
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	res, err := sdkresource.New(
		ctx,
		sdkresource.WithAttributes(
			attribute.String("service.name", cfg.ServiceName),
			attribute.String("service.version", cfg.ServiceVersion),
			attribute.String("deployment.environment.name", cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("create telemetry resource: %w", err)
	}

	exporterOptions := []otlptracehttp.Option{
		otlptracehttp.WithEndpointURL(normalizeEndpointURL(cfg.Endpoint)),
	}

	if cfg.Insecure {
		exporterOptions = append(
			exporterOptions,
			otlptracehttp.WithInsecure(),
		)
	}

	exporter, err := otlptracehttp.New(
		ctx,
		exporterOptions...,
	)
	if err != nil {
		return nil, fmt.Errorf("create OTLP trace exporter: %w", err)
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(
			exporter,
			sdktrace.WithBatchTimeout(5*time.Second),
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(samplerFromConfig(cfg)),
	)

	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	return &Provider{
		tracerProvider: tracerProvider,
	}, nil
}

func (p *Provider) Enabled() bool {
	return p != nil && p.tracerProvider != nil
}

func (p *Provider) Tracer() trace.Tracer {
	if p == nil || p.tracerProvider == nil {
		return otel.Tracer(instrumentationName)
	}

	return p.tracerProvider.Tracer(instrumentationName)
}

func (p *Provider) Shutdown(ctx context.Context) error {
	if p == nil || p.tracerProvider == nil {
		return nil
	}

	return p.tracerProvider.Shutdown(ctx)
}

func samplerFromConfig(cfg Config) sdktrace.Sampler {
	switch normalizeSampler(cfg.Sampler) {
	case "always_off":
		return sdktrace.NeverSample()

	case "traceidratio":
		return sdktrace.TraceIDRatioBased(cfg.SamplerArg)

	case "parentbased_always_off":
		return sdktrace.ParentBased(
			sdktrace.NeverSample(),
		)

	case "parentbased_traceidratio":
		return sdktrace.ParentBased(
			sdktrace.TraceIDRatioBased(cfg.SamplerArg),
		)

	case "parentbased_always_on":
		return sdktrace.ParentBased(
			sdktrace.AlwaysSample(),
		)

	case "always_on", "":
		return sdktrace.AlwaysSample()

	default:
		return sdktrace.ParentBased(
			sdktrace.TraceIDRatioBased(cfg.SamplerArg),
		)
	}
}

func normalizeEndpointURL(endpoint string) string {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return endpoint
	}

	if parsed, err := url.Parse(endpoint); err == nil && parsed.Scheme != "" {
		return parsed.String()
	}

	return "http://" + endpoint
}
