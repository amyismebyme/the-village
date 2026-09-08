package config

import (
	"testing"
)

func TestTelemetryDefaultsDisabled(t *testing.T) {
	cfg := Load()

	if cfg.Telemetry.Enabled {
		t.Fatal("expected OpenTelemetry to be disabled by default")
	}

	if cfg.Telemetry.ServiceName != "village-api" {
		t.Fatalf(
			"expected service name village-api, got %q",
			cfg.Telemetry.ServiceName,
		)
	}
}

func TestTelemetryEnabledRequiresServiceName(t *testing.T) {
	cfg := Load()
	cfg.Telemetry.Enabled = true
	cfg.Telemetry.ServiceName = ""

	if err := Validate(cfg); err == nil {
		t.Fatal("expected telemetry service-name validation error")
	}
}

func TestTelemetryValidationAcceptsTraceIDRatio(t *testing.T) {
	cfg := Load()
	cfg.Telemetry.Enabled = true
	cfg.Telemetry.ServiceName = "village-api"
	cfg.Telemetry.Endpoint = "http://localhost:4318/v1/traces"
	cfg.Telemetry.Sampler = "traceidratio"
	cfg.Telemetry.SamplerArg = 0.25

	if err := Validate(cfg); err != nil {
		t.Fatalf("validate telemetry config: %v", err)
	}
}
