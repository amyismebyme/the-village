package telemetry

import (
	"fmt"
	"net/url"
	"strings"
)

// Config controls OpenTelemetry tracing.
//
// The configuration intentionally remains independent of any particular
// tracing backend. OTLP is the transport used when tracing is enabled.
type Config struct {
	Enabled        bool
	ServiceName    string
	ServiceVersion string
	Endpoint       string
	Insecure       bool
	Sampler        string
	SamplerArg     float64
	Environment    string
}

func (c Config) Validate() error {
	if !c.Enabled {
		return nil
	}

	if strings.TrimSpace(c.ServiceName) == "" {
		return fmt.Errorf("OTEL_SERVICE_NAME must be configured when OpenTelemetry is enabled")
	}

	endpoint := strings.TrimSpace(c.Endpoint)
	if endpoint == "" {
		return fmt.Errorf("OTEL_EXPORTER_OTLP_ENDPOINT must be configured when OpenTelemetry is enabled")
	}

	parsed, err := url.Parse(endpoint)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return fmt.Errorf("OTEL_EXPORTER_OTLP_ENDPOINT must be a valid http or https URL")
	}

	if parsed.Scheme == "https" && c.Insecure {
		return fmt.Errorf("OTEL_EXPORTER_OTLP_INSECURE must be false for an https endpoint")
	}

	if c.SamplerArg < 0 || c.SamplerArg > 1 {
		return fmt.Errorf("OTEL_TRACES_SAMPLER_ARG must be between 0 and 1")
	}

	switch normalizeSampler(c.Sampler) {
	case "always_on", "always_off", "traceidratio", "parentbased_always_on", "parentbased_always_off", "parentbased_traceidratio":
		return nil
	default:
		return fmt.Errorf("unsupported OTEL_TRACES_SAMPLER %q", c.Sampler)
	}
}

func normalizeSampler(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	return strings.ReplaceAll(value, "-", "_")
}
