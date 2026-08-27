package config

import (
	"testing"
	"time"
)

func TestProductionConfigurationRequiresExplicitDatabaseSettings(t *testing.T) {
	keys := []string{"DB_HOST", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSLMODE"}
	for _, key := range keys {
		t.Setenv(key, "")
	}

	cfg := Load()
	cfg.Environment = "production"
	cfg.Database.SSLMode = "require"

	if err := Validate(cfg); err == nil {
		t.Fatal("expected production database configuration validation to fail")
	}
}

func TestProductionConfigurationAcceptsExplicitDatabaseSettings(t *testing.T) {
	t.Setenv("DB_HOST", "db.example.internal")
	t.Setenv("DB_USER", "village-prod")
	t.Setenv("DB_PASSWORD", "test-password")
	t.Setenv("DB_NAME", "village")
	t.Setenv("DB_SSLMODE", "require")

	cfg := Load()
	cfg.Environment = "production"

	if err := Validate(cfg); err != nil {
		t.Fatalf("expected valid production configuration, got %v", err)
	}
}

func TestTimeoutPolicyRelationships(t *testing.T) {
	cfg := Load()
	if cfg.Database.QueryTimeout >= cfg.RequestTimeout ||
		cfg.RequestTimeout >= cfg.WriteTimeout {
		t.Fatalf("expected DB query < request < write timeout, got %v < %v < %v", cfg.Database.QueryTimeout, cfg.RequestTimeout, cfg.WriteTimeout)
	}
}

func TestTimeoutPolicyRejectsDatabaseTimeoutAtOrAboveRequestTimeout(t *testing.T) {
	cfg := Load()
	cfg.Database.QueryTimeout = 35 * time.Second
	cfg.RequestTimeout = 35 * time.Second
	if err := Validate(cfg); err == nil {
		t.Fatal("expected invalid timeout relationship to fail validation")
	}
}

func TestProductionRedditConfigurationRequiresCredentials(
	t *testing.T,
) {
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("DB_HOST", "db")
	t.Setenv("DB_USER", "village")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("DB_NAME", "village")
	t.Setenv("DB_SSLMODE", "require")

	t.Setenv("REDDIT_ENABLED", "true")
	t.Setenv("REDDIT_CLIENT_ID", "")
	t.Setenv("REDDIT_CLIENT_SECRET", "")
	t.Setenv("REDDIT_USER_AGENT", "")

	cfg := Load()

	if err := Validate(cfg); err == nil {
		t.Fatal(
			"expected production Reddit configuration validation error",
		)
	}
}
