package config

import (
	"fmt"
	"net/url"
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

	if err := validateReddit(cfg); err != nil {
		return err
	}

	return nil
}

func validateReddit(
	cfg Config,
) error {
	if !cfg.External.Reddit.Enabled {
		return nil
	}

	if cfg.External.Reddit.ClientID == "" {
		return fmt.Errorf(
			"REDDIT_CLIENT_ID must be configured when Reddit is enabled",
		)
	}

	if cfg.External.Reddit.ClientSecret == "" {
		return fmt.Errorf(
			"REDDIT_CLIENT_SECRET must be configured when Reddit is enabled",
		)
	}

	if cfg.External.Reddit.UserAgent == "" {
		return fmt.Errorf(
			"REDDIT_USER_AGENT must be configured when Reddit is enabled",
		)
	}

	if cfg.External.Reddit.BaseURL == "" {
		return fmt.Errorf(
			"REDDIT_BASE_URL must be configured when Reddit is enabled",
		)
	}

	u, err := url.Parse(cfg.External.Reddit.BaseURL)
	if err != nil {
		return fmt.Errorf(
			"REDDIT_BASE_URL is invalid: %w",
			err,
		)
	}

	if u.Scheme != "https" {
		return fmt.Errorf(
			"REDDIT_BASE_URL must use https",
		)
	}

	if u.Host == "" {
		return fmt.Errorf(
			"REDDIT_BASE_URL must include a host",
		)
	}

	return nil
}
