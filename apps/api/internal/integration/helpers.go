//go:build integration
// +build integration

package integration

import (
	"context"
	"fmt"
	"github.com/amyismebyme/the-village/apps/api/internal/config"
	"github.com/amyismebyme/the-village/apps/api/internal/database"
	"github.com/joho/godotenv"
	"os"
	"sync"
	"testing"
	"time"
)

const (
	waitTimeout = 30 * time.Second
	retryDelay  = 1 * time.Second
)

var loadEnvOnce sync.Once

// loadIntegrationEnv loads the integration environment file once.
func loadIntegrationEnv() {

	loadEnvOnce.Do(func() {

		if err := godotenv.Overload(".env.integration"); err != nil {
			fmt.Printf(
				"warning: unable to load .env.integration: %v\n",
				err,
			)
		}

	})
}

// OpenTestDatabase opens a real PostgreSQL database
// using the application's configuration.
func OpenTestDatabase(t *testing.T) *database.Database {

	t.Helper()

	loadIntegrationEnv()

	cfg := config.Load()

	ctx := context.Background()

	db, err := database.Open(ctx, cfg.Database)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	if err := WaitForDatabase(ctx, db); err != nil {
		db.Close()
		t.Fatalf("database never became ready: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

// WaitForDatabase retries Ping until PostgreSQL becomes available.
func WaitForDatabase(
	ctx context.Context,
	db *database.Database,
) error {

	timeout := time.NewTimer(waitTimeout)
	defer timeout.Stop()

	ticker := time.NewTicker(retryDelay)
	defer ticker.Stop()

	var lastErr error

	for {

		select {

		case <-ctx.Done():
			return ctx.Err()

		case <-timeout.C:

			if lastErr != nil {
				return fmt.Errorf(
					"timed out waiting for database: %w",
					lastErr,
				)
			}

			return fmt.Errorf(
				"timed out waiting for database",
			)

		case <-ticker.C:

			err := db.Ping(ctx)

			if err == nil {
				return nil
			}

			lastErr = err
		}
	}
}

// MustGetEnv returns an environment variable or panics.
func MustGetEnv(key string) string {

	value := os.Getenv(key)

	if value == "" {
		panic(fmt.Sprintf(
			"environment variable %s not set",
			key,
		))
	}

	return value
}
