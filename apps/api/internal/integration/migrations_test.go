package integration

import (
	"context"
	"github.com/amyismebyme/the-village/apps/api/internal/database"
	"testing"
	"time"
)

func TestMigrationsApplied(t *testing.T) {

	db := OpenTestDatabase(t)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	t.Run("schema_migrations exists", func(t *testing.T) {
		assertTableExists(
			t,
			ctx,
			db,
			"schema_migrations",
		)
	})

	t.Run("communities exists", func(t *testing.T) {
		assertTableExists(
			t,
			ctx,
			db,
			"communities",
		)
	})

	t.Run("resources exists", func(t *testing.T) {
		assertTableExists(
			t,
			ctx,
			db,
			"resources",
		)
	})

	// Uncomment these once the indexes exist.
	/*
		t.Run("communities_name_idx exists", func(t *testing.T) {
			assertIndexExists(
				t,
				ctx,
				db,
				"communities_name_idx",
			)
		})

		t.Run("resources_community_id_idx exists", func(t *testing.T) {
			assertIndexExists(
				t,
				ctx,
				db,
				"resources_community_id_idx",
			)
		})
	*/
}

func assertTableExists(
	t *testing.T,
	ctx context.Context,
	db *database.Database,
	table string,
) {

	t.Helper()

	const query = `
SELECT EXISTS (
    SELECT 1
    FROM information_schema.tables
    WHERE table_schema='public'
      AND table_name=$1
)
`

	var exists bool

	err := db.Pool().
		QueryRow(ctx, query, table).
		Scan(&exists)

	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	if !exists {
		t.Fatalf("table %q does not exist", table)
	}
}

func assertIndexExists(
	t *testing.T,
	ctx context.Context,
	db *database.Database,
	index string,
) {

	t.Helper()

	const query = `
SELECT EXISTS (
    SELECT 1
    FROM pg_indexes
    WHERE schemaname='public'
      AND indexname=$1
)
`

	var exists bool

	err := db.Pool().
		QueryRow(ctx, query, index).
		Scan(&exists)

	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	if !exists {
		t.Fatalf("index %q does not exist", index)
	}
}
