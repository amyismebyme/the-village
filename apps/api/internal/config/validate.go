package config

import (
	"fmt"
	"os"
	"strings"
)

// Validates params at application startup time and exits if not matching
func Validate(cfg Config) error {

	if cfg.Port == "" {
		return fmt.Errorf("PORT cannot be empty")
	}

	switch strings.ToLower(cfg.Environment) {
	case "development", "staging", "production":
	default:
		return fmt.Errorf("invalid environment: %s", cfg.Environment)
	}

	switch strings.ToLower(cfg.LogLevel) {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("invalid log level: %s", cfg.LogLevel)
	}

	switch strings.ToLower(cfg.LogFormat) {
	case "text", "json":
	default:
		return fmt.Errorf("invalid log format: %s", cfg.LogFormat)
	}

	// Production requires database settings to be explicitly configured.
	if strings.ToLower(cfg.Environment) == "production" {
		requiredDBVars := []string{
			"DB_HOST",
			"DB_USER",
			"DB_PASSWORD",
			"DB_NAME",
			"DB_SSLMODE",
		}

		for _, key := range requiredDBVars {
			if os.Getenv(key) == "" {
				return fmt.Errorf("%s must be explicitly configured in production", key)
			}
		}
	}

	// Database queries must complete before the request timeout.
	if cfg.Database.QueryTimeout >= cfg.RequestTimeout {
		return fmt.Errorf(
			"database query timeout must be less than request timeout",
		)
	}

	// Requests must complete before the write timeout.
	if cfg.RequestTimeout >= cfg.WriteTimeout {
		return fmt.Errorf(
			"request timeout must be less than write timeout",
		)
	}

	return nil
}
