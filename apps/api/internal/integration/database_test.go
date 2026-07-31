//go:build integration
// +build integration

package integration

import (
	"context"
	"testing"
	"time"
)

func TestDatabaseConnection(t *testing.T) {

	db := OpenTestDatabase(t)

	if db == nil {
		t.Fatal("expected database instance")
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	// Verify PostgreSQL is reachable.
	if err := db.Ping(ctx); err != nil {
		t.Fatalf("failed to ping database: %v", err)
	}

	// Verify the pool exists.
	if db.Pool() == nil {
		t.Fatal("expected connection pool to be initialized")
	}

	// Verify statistics are available.
	stats := db.Stats()

	if stats.MaxConnections <= 0 {
		t.Fatalf(
			"expected max connections > 0, got %d",
			stats.MaxConnections,
		)
	}

	if stats.TotalConnections < 0 {
		t.Fatalf(
			"expected total connections >= 0, got %d",
			stats.TotalConnections,
		)
	}

	// Close should not panic or error.
	db.Close()
}
