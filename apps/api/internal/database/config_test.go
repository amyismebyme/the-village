package database

import (
	"testing"
	"time"
)

func TestConfigQueryTimeout(t *testing.T) {
	cfg := Config{
		Host:              "localhost",
		Port:              5432,
		User:              "village",
		Name:              "village",
		MaxConns:          10,
		MinConns:          1,
		MaxConnLifetime:   time.Hour,
		MaxConnIdleTime:   5 * time.Minute,
		HealthCheckPeriod: time.Minute,
		QueryTimeout:      30 * time.Second,
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf(
			"expected valid database config, got %v",
			err,
		)
	}
}

func TestConfigRejectsInvalidQueryTimeout(t *testing.T) {
	cfg := Config{
		Host:              "localhost",
		Port:              5432,
		User:              "village",
		Name:              "village",
		MaxConns:          10,
		MinConns:          1,
		MaxConnLifetime:   time.Hour,
		MaxConnIdleTime:   5 * time.Minute,
		HealthCheckPeriod: time.Minute,
		QueryTimeout:      0,
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal(
			"expected invalid query timeout to fail validation",
		)
	}
}
