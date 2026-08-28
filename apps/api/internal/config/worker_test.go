package config

import (
	"testing"
	"time"
)

func TestWorkerConfigurationDefaults(
	t *testing.T,
) {
	t.Setenv(
		"WORKER_ENABLED",
		"",
	)
	t.Setenv(
		"REDDIT_INGEST_SUBREDDIT",
		"",
	)
	t.Setenv(
		"REDDIT_INGEST_LIMIT",
		"",
	)
	t.Setenv(
		"REDDIT_INGEST_INTERVAL",
		"",
	)

	cfg := Load()

	if cfg.Worker.Enabled {
		t.Fatal(
			"expected workers to be disabled by default",
		)
	}

	if cfg.Worker.Reddit.Subreddit !=
		"toronto" {
		t.Fatalf(
			"expected default subreddit toronto, got %q",
			cfg.Worker.Reddit.Subreddit,
		)
	}

	if cfg.Worker.Reddit.Limit != 25 {
		t.Fatalf(
			"expected default Reddit limit 25, got %d",
			cfg.Worker.Reddit.Limit,
		)
	}

	if cfg.Worker.Reddit.IngestInterval !=
		5*time.Minute {
		t.Fatalf(
			"expected default ingest interval 5m, got %s",
			cfg.Worker.Reddit.IngestInterval,
		)
	}
}

func TestWorkerConfigurationIsConfigurable(
	t *testing.T,
) {
	t.Setenv(
		"WORKER_ENABLED",
		"true",
	)
	t.Setenv(
		"REDDIT_INGEST_SUBREDDIT",
		"niagara",
	)
	t.Setenv(
		"REDDIT_INGEST_LIMIT",
		"50",
	)
	t.Setenv(
		"REDDIT_INGEST_INTERVAL",
		"120",
	)

	cfg := Load()

	if !cfg.Worker.Enabled {
		t.Fatal(
			"expected workers enabled",
		)
	}

	if cfg.Worker.Reddit.Subreddit !=
		"niagara" {
		t.Fatalf(
			"unexpected subreddit %q",
			cfg.Worker.Reddit.Subreddit,
		)
	}

	if cfg.Worker.Reddit.Limit != 50 {
		t.Fatalf(
			"unexpected Reddit limit %d",
			cfg.Worker.Reddit.Limit,
		)
	}

	if cfg.Worker.Reddit.IngestInterval !=
		120*time.Second {
		t.Fatalf(
			"unexpected ingest interval %s",
			cfg.Worker.Reddit.IngestInterval,
		)
	}
}

func TestDisabledWorkerDoesNotRequireWorkerValidation(
	t *testing.T,
) {
	t.Setenv(
		"WORKER_ENABLED",
		"false",
	)
	t.Setenv(
		"REDDIT_INGEST_SUBREDDIT",
		"",
	)
	t.Setenv(
		"REDDIT_INGEST_LIMIT",
		"0",
	)
	t.Setenv(
		"REDDIT_INGEST_INTERVAL",
		"0",
	)

	cfg := Load()

	if err := Validate(cfg); err != nil {
		t.Fatalf(
			"expected disabled worker configuration to validate: %v",
			err,
		)
	}
}

func TestEnabledWorkerRejectsInvalidLimit(
	t *testing.T,
) {
	t.Setenv(
		"WORKER_ENABLED",
		"true",
	)
	t.Setenv(
		"REDDIT_INGEST_LIMIT",
		"101",
	)

	cfg := Load()

	if err := Validate(cfg); err == nil {
		t.Fatal(
			"expected invalid worker limit error",
		)
	}
}
