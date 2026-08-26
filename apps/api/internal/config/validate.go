package config

import (
	"fmt"
	"os"
)

// Validates params at application startup time and exits if not matching
func Validate(cfg Config) error {
	if cfg.Environment == "production" {
		required := []string{
			"DB_HOST",
			"DB_USER",
			"DB_PASSWORD",
			"DB_NAME",
			"DB_SSLMODE",
		}

		for _, key := range required {
			if os.Getenv(key) == "" {
				return fmt.Errorf("%s must be explicitly configured in production", key)
			}
		}
	}

	if cfg.Database.QueryTimeout >= cfg.RequestTimeout {
		return fmt.Errorf(
			"database query timeout must be less than request timeout",
		)
	}

	if cfg.RequestTimeout >= cfg.WriteTimeout {
		return fmt.Errorf(
			"request timeout must be less than write timeout",
		)
	}

	if cfg.External.RequestTimeout <= 0 {
		return fmt.Errorf(
			"external request timeout must be greater than zero",
		)
	}

	return nil
}
