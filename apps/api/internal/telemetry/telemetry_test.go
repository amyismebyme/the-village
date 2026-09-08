package telemetry

import (
	"context"
	"testing"
)

func TestConfigValidateDisabled(t *testing.T) {
	cfg := Config{}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("validate disabled config: %v", err)
	}
}

func TestConfigValidateEnabledRequiresEndpoint(t *testing.T) {
	cfg := Config{
		Enabled:     true,
		ServiceName: "village-api",
		Sampler:     "always_on",
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected endpoint validation error")
	}
}

func TestConfigValidateSampler(t *testing.T) {
	cfg := Config{
		Enabled:     true,
		ServiceName: "village-api",
		Endpoint:    "http://localhost:4318/v1/traces",
		Sampler:     "traceidratio",
		SamplerArg:  0.25,
		Environment: "test",
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("validate sampler config: %v", err)
	}
}

func TestNewProviderDisabled(t *testing.T) {
	provider, err := NewProvider(
		context.Background(),
		Config{Enabled: false},
	)
	if err != nil {
		t.Fatalf("create disabled provider: %v", err)
	}

	if provider.Enabled() {
		t.Fatal("expected provider to be disabled")
	}

	tracer := provider.Tracer()
	if tracer == nil {
		t.Fatal("expected non-nil no-op tracer")
	}

	_, span := tracer.Start(
		context.Background(),
		"test",
	)
	if span == nil {
		t.Fatal("expected non-nil span")
	}
	span.End()
}

func TestSamplerFromConfig(t *testing.T) {
	cases := []struct {
		name string
		cfg  Config
		want string
	}{
		{
			name: "always on",
			cfg: Config{
				Sampler: "always_on",
			},
			want: "AlwaysOnSampler",
		},
		{
			name: "always off",
			cfg: Config{
				Sampler: "always_off",
			},
			want: "AlwaysOffSampler",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			sampler := samplerFromConfig(tt.cfg)
			if sampler == nil {
				t.Fatal("expected sampler")
			}
		})
	}

}
