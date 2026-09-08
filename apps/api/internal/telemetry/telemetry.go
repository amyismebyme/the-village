package telemetry

import (
	"context"
	"fmt"
	"log/slog"
)

// Initialize creates the application's tracing provider.
//
// Telemetry is deliberately optional. When disabled, the OpenTelemetry API's
// no-op provider remains in effect and application behavior is unchanged.
func Initialize(
	ctx context.Context,
	cfg Config,
	logger *slog.Logger,
) (*Provider, error) {
	provider, err := NewProvider(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("initialize telemetry: %w", err)
	}

	if logger != nil {
		if provider.Enabled() {
			logger.Info(
				"OpenTelemetry tracing initialized",
				"service_name",
				cfg.ServiceName,
				"endpoint",
				cfg.Endpoint,
				"sampler",
				cfg.Sampler,
			)
		} else {
			logger.Info(
				"OpenTelemetry tracing disabled",
			)
		}
	}

	return provider, nil
}

// Shutdown gracefully flushes pending telemetry before application exit.
func Shutdown(
	ctx context.Context,
	provider *Provider,
	logger *slog.Logger,
) error {
	if provider == nil || !provider.Enabled() {
		return nil
	}

	if err := provider.Shutdown(ctx); err != nil {
		if logger != nil {
			logger.Error(
				"OpenTelemetry shutdown failed",
				"error",
				err,
			)
		}

		return err
	}

	return nil
}
