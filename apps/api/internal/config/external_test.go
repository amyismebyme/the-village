package config

// test EXTERNAL_REQUEST_TIMEOUT
import (
	"testing"
	"time"
)

func TestExternalRequestTimeoutDefault(t *testing.T) {
	t.Setenv(
		"EXTERNAL_REQUEST_TIMEOUT",
		"",
	)

	cfg := Load()

	if cfg.External.RequestTimeout != 15*time.Second {
		t.Fatalf(
			"expected external request timeout 15s, got %s",
			cfg.External.RequestTimeout,
		)
	}
}

func TestExternalRequestTimeoutConfigurable(t *testing.T) {
	t.Setenv(
		"EXTERNAL_REQUEST_TIMEOUT",
		"7",
	)

	cfg := Load()

	if cfg.External.RequestTimeout != 7*time.Second {
		t.Fatalf(
			"expected external request timeout 7s, got %s",
			cfg.External.RequestTimeout,
		)
	}
}
