package config

import (
	"testing"
	"time"
)

func TestRedditConfigurationDefaults(t *testing.T) {
	t.Setenv(
		"REDDIT_ENABLED",
		"",
	)
	t.Setenv(
		"REDDIT_CLIENT_ID",
		"",
	)
	t.Setenv(
		"REDDIT_CLIENT_SECRET",
		"",
	)
	t.Setenv(
		"REDDIT_USER_AGENT",
		"",
	)
	t.Setenv(
		"REDDIT_BASE_URL",
		"",
	)
	t.Setenv(
		"REDDIT_AUTH_BASE_URL",
		"",
	)

	cfg := Load()

	if cfg.External.Reddit.Enabled {
		t.Fatal("expected Reddit to be disabled by default")
	}

	if cfg.External.Reddit.ClientID != "" {
		t.Fatalf(
			"expected empty Reddit client ID, got %q",
			cfg.External.Reddit.ClientID,
		)
	}

	if cfg.External.Reddit.ClientSecret != "" {
		t.Fatalf(
			"expected empty Reddit client secret, got %q",
			cfg.External.Reddit.ClientSecret,
		)
	}

	if cfg.External.Reddit.UserAgent != "" {
		t.Fatalf(
			"expected empty Reddit user agent, got %q",
			cfg.External.Reddit.UserAgent,
		)
	}

	if cfg.External.Reddit.BaseURL !=
		"https://oauth.reddit.com" {
		t.Fatalf(
			"expected default Reddit base URL, got %q",
			cfg.External.Reddit.BaseURL,
		)
	}

	if cfg.External.Reddit.AuthBaseURL !=
		"https://www.reddit.com" {
		t.Fatalf(
			"expected default Reddit auth base URL, got %q",
			cfg.External.Reddit.AuthBaseURL,
		)
	}
}

func TestRedditConfigurationCanBeConfigured(t *testing.T) {
	t.Setenv(
		"REDDIT_ENABLED",
		"true",
	)
	t.Setenv(
		"REDDIT_CLIENT_ID",
		"client-id",
	)
	t.Setenv(
		"REDDIT_CLIENT_SECRET",
		"client-secret",
	)
	t.Setenv(
		"REDDIT_USER_AGENT",
		"the-village/1.0",
	)
	t.Setenv(
		"REDDIT_BASE_URL",
		"https://reddit.example.test",
	)

	cfg := Load()

	if !cfg.External.Reddit.Enabled {
		t.Fatal("expected Reddit to be enabled")
	}

	if cfg.External.Reddit.ClientID != "client-id" {
		t.Fatalf(
			"unexpected client ID %q",
			cfg.External.Reddit.ClientID,
		)
	}

	if cfg.External.Reddit.ClientSecret !=
		"client-secret" {
		t.Fatalf(
			"unexpected client secret %q",
			cfg.External.Reddit.ClientSecret,
		)
	}

	if cfg.External.Reddit.UserAgent !=
		"the-village/1.0" {
		t.Fatalf(
			"unexpected user agent %q",
			cfg.External.Reddit.UserAgent,
		)
	}

	if cfg.External.Reddit.BaseURL !=
		"https://reddit.example.test" {
		t.Fatalf(
			"unexpected base URL %q",
			cfg.External.Reddit.BaseURL,
		)
	}
}

func TestRedditDisabledDoesNotRequireCredentials(t *testing.T) {
	t.Setenv(
		"REDDIT_ENABLED",
		"false",
	)
	t.Setenv(
		"REDDIT_CLIENT_ID",
		"",
	)
	t.Setenv(
		"REDDIT_CLIENT_SECRET",
		"",
	)
	t.Setenv(
		"REDDIT_USER_AGENT",
		"",
	)

	cfg := Load()

	if err := Validate(cfg); err != nil {
		t.Fatalf(
			"expected disabled Reddit to require no credentials: %v",
			err,
		)
	}
}

func TestRedditEnabledRequiresClientID(t *testing.T) {
	t.Setenv(
		"REDDIT_ENABLED",
		"true",
	)
	t.Setenv(
		"REDDIT_CLIENT_ID",
		"",
	)
	t.Setenv(
		"REDDIT_CLIENT_SECRET",
		"client-secret",
	)
	t.Setenv(
		"REDDIT_USER_AGENT",
		"the-village/test",
	)

	cfg := Load()

	if err := Validate(cfg); err == nil {
		t.Fatal(
			"expected missing Reddit client ID validation error",
		)
	}
}

func TestRedditEnabledRequiresClientSecret(t *testing.T) {
	t.Setenv(
		"REDDIT_ENABLED",
		"true",
	)
	t.Setenv(
		"REDDIT_CLIENT_ID",
		"client-id",
	)
	t.Setenv(
		"REDDIT_CLIENT_SECRET",
		"",
	)
	t.Setenv(
		"REDDIT_USER_AGENT",
		"the-village/test",
	)

	cfg := Load()

	if err := Validate(cfg); err == nil {
		t.Fatal(
			"expected missing Reddit client secret validation error",
		)
	}
}

func TestRedditEnabledRequiresUserAgent(t *testing.T) {
	t.Setenv(
		"REDDIT_ENABLED",
		"true",
	)
	t.Setenv(
		"REDDIT_CLIENT_ID",
		"client-id",
	)
	t.Setenv(
		"REDDIT_CLIENT_SECRET",
		"client-secret",
	)
	t.Setenv(
		"REDDIT_USER_AGENT",
		"",
	)

	cfg := Load()

	if err := Validate(cfg); err == nil {
		t.Fatal(
			"expected missing Reddit user agent validation error",
		)
	}
}

func TestRedditConfigurationDoesNotChangeExternalTimeout(
	t *testing.T,
) {
	t.Setenv(
		"EXTERNAL_REQUEST_TIMEOUT",
		"12",
	)

	cfg := Load()

	if cfg.External.RequestTimeout !=
		12*time.Second {
		t.Fatalf(
			"expected external request timeout 12s, got %s",
			cfg.External.RequestTimeout,
		)
	}
}
