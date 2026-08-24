package config

import (
	"testing"
	"time"
)

func TestDefaultTimeoutPolicy(t *testing.T) {
	cfg := Load()

	if cfg.ReadTimeout != 10*time.Second {
		t.Fatalf("expected read timeout 10s, got %v", cfg.ReadTimeout)
	}

	if cfg.RequestTimeout != 35*time.Second {
		t.Fatalf("expected request timeout 35s, got %v", cfg.RequestTimeout)
	}

	if cfg.Database.QueryTimeout != 30*time.Second {
		t.Fatalf("expected DB query timeout 30s, got %v", cfg.Database.QueryTimeout)
	}

	if cfg.WriteTimeout != 40*time.Second {
		t.Fatalf("expected write timeout 40s, got %v", cfg.WriteTimeout)
	}

	if cfg.ShutdownTimeout != 15*time.Second {
		t.Fatalf("expected shutdown timeout 15s, got %v", cfg.ShutdownTimeout)
	}

	if cfg.Database.QueryTimeout >= cfg.RequestTimeout ||
        cfg.RequestTimeout >= cfg.WriteTimeout {
		t.Fatalf("expected DB query < request < write timeout")
	}
}
