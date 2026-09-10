package telemetry

import (
	"errors"
	"fmt"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func sampler(cfg Config) (sdktrace.Sampler, error) {
	if cfg.SamplerArgument < 0 || cfg.SamplerArgument > 1 {
		return nil, errors.New("telemetry: sampler argument must be between 0 and 1")
	}

	switch cfg.Sampler {
	case "always_on":
		return sdktrace.AlwaysSample(), nil
	case "always_off":
		return sdktrace.NeverSample(), nil
	case "traceidratio":
		return sdktrace.TraceIDRatioBased(cfg.SamplerArgument), nil
	case "parentbased_always_on":
		return sdktrace.ParentBased(sdktrace.AlwaysSample()), nil
	case "parentbased_always_off":
		return sdktrace.ParentBased(sdktrace.NeverSample()), nil
	case "parentbased_traceidratio":
		return sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.SamplerArgument)), nil
	default:
		return nil, fmt.Errorf("telemetry: unsupported sampler %q", cfg.Sampler)
	}
}
