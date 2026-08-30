package config

import (
	"fmt"
	"github.com/amyismebyme/the-village/apps/api/internal/httputil"
	"os"
	"strings"
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

	if err := validateWorker(cfg); err != nil {
		return err
	}

	if err := validateRetry(cfg); err != nil {
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

	if err := validateHTTPSURL(
		"REDDIT_BASE_URL",
		cfg.External.Reddit.BaseURL,
	); err != nil {
		return err
	}

	if err := validateHTTPSURL(
		"REDDIT_AUTH_BASE_URL",
		cfg.External.Reddit.AuthBaseURL,
	); err != nil {
		return err
	}

	if cfg.External.Reddit.RequestInterval <= 0 {
		return fmt.Errorf(
			"REDDIT_REQUEST_INTERVAL must be greater than zero",
		)
	}

	if err := validateCache(cfg); err != nil {
		return err
	}

	return nil
}
func validateWorker(cfg Config) error {
	if !cfg.Worker.Enabled {
		return nil
	}

	if cfg.Worker.Reddit.IngestInterval <= 0 {
		return fmt.Errorf(
			"REDDIT_INGEST_INTERVAL must be greater than zero",
		)
	}

	if cfg.Worker.Reddit.Limit < 1 ||
		cfg.Worker.Reddit.Limit > 100 {
		return fmt.Errorf(
			"REDDIT_INGEST_LIMIT must be between 1 and 100",
		)
	}

	if strings.TrimSpace(
		cfg.Worker.Reddit.Subreddit,
	) == "" {
		return fmt.Errorf(
			"REDDIT_INGEST_SUBREDDIT must be configured",
		)
	}

	return nil
}

func validateHTTPSURL(
	name string,
	value string,
) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf(
			"%s must be configured",
			name,
		)
	}

	if _, err := httputil.ParseBaseURL(
		value,
		false,
	); err != nil {
		return fmt.Errorf(
			"REDDIT_BASE_URL is invalid: %w",
			err,
		)
	}

	return nil
}

func validateRetry(
	cfg Config,
) error {
	retry := cfg.External.Retry

	if retry.MaxAttempts < 1 {
		return fmt.Errorf(
			"EXTERNAL_RETRY_MAX_ATTEMPTS must be at least 1",
		)
	}

	if retry.InitialBackoff <= 0 {
		return fmt.Errorf(
			"EXTERNAL_RETRY_INITIAL_BACKOFF must be greater than zero",
		)
	}

	if retry.MaxBackoff < retry.InitialBackoff {
		return fmt.Errorf(
			"EXTERNAL_RETRY_MAX_BACKOFF must be greater than or equal to initial backoff",
		)
	}

	if retry.BackoffMultiplier < 1 {
		return fmt.Errorf(
			"EXTERNAL_RETRY_BACKOFF_MULTIPLIER must be at least 1",
		)
	}

	if retry.Jitter < 0 ||
		retry.Jitter > 1 {
		return fmt.Errorf(
			"EXTERNAL_RETRY_JITTER must be between 0 and 1",
		)
	}

	return nil
}

func validateCache(
	cfg Config,
) error {
	if !cfg.Cache.Enabled {
		return nil
	}

	if cfg.Cache.MaxEntries < 1 {
		return fmt.Errorf(
			"CACHE_MAX_ENTRIES must be greater than zero",
		)
	}

	if cfg.Cache.RedditListingTTL <= 0 {
		return fmt.Errorf(
			"CACHE_REDDIT_LISTING_TTL must be greater than zero",
		)
	}

	if cfg.Cache.CommunityTTL <= 0 {
		return fmt.Errorf(
			"CACHE_COMMUNITY_TTL must be greater than zero",
		)
	}

	return nil
}
