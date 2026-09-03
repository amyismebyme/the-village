//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/external"
	"github.com/amyismebyme/the-village/apps/api/internal/repository/postgres"
)

func newExternalItemRepositoryTestApp(
	t *testing.T,
) *postgres.ExternalItemRepository {
	t.Helper()

	db := OpenTestDatabase(t)

	repo := postgres.NewExternalItemRepository(
		db.Pool(),
	)

	ctx := context.Background()

	if err := repo.DeleteAll(ctx); err != nil {
		t.Fatalf(
			"clean external items before test: %v",
			err,
		)
	}

	t.Cleanup(func() {
		if err := repo.DeleteAll(
			context.Background(),
		); err != nil {
			t.Logf(
				"clean external items after test: %v",
				err,
			)
		}
	})

	return repo
}

func TestExternalItemIdempotencyAcrossReplayedBatches(
	t *testing.T,
) {
	repo := newExternalItemRepositoryTestApp(t)

	ctx := context.Background()

	batchWithDuplicate := []external.Item{
		{
			Source:     external.SourceReddit,
			ExternalID: "integration-duplicate",
			Title:      "First occurrence",
		},
		{
			Source:     external.SourceReddit,
			ExternalID: "integration-unique",
			Title:      "Unique occurrence",
		},
		{
			Source:     external.SourceReddit,
			ExternalID: "integration-duplicate",
			Title:      "Duplicate occurrence",
		},
	}

	unique, err := external.DeduplicateItems(
		ctx,
		batchWithDuplicate,
	)
	if err != nil {
		t.Fatalf(
			"deduplicate batch: %v",
			err,
		)
	}

	if len(unique) != 2 {
		t.Fatalf(
			"expected two unique items, got %d",
			len(unique),
		)
	}

	if unique[0].Title != "First occurrence" {
		t.Fatalf(
			"expected first occurrence to win, got %q",
			unique[0].Title,
		)
	}

	if err := repo.UpsertBatch(
		ctx,
		unique,
	); err != nil {
		t.Fatalf(
			"first batch upsert: %v",
			err,
		)
	}

	// Replay the exact same batch.
	if err := repo.UpsertBatch(
		ctx,
		unique,
	); err != nil {
		t.Fatalf(
			"replayed batch upsert: %v",
			err,
		)
	}

	var count int

	if err := repo.Pool().
		QueryRow(
			ctx,
			`SELECT COUNT(*) FROM external_items`,
		).
		Scan(&count); err != nil {
		t.Fatalf(
			"count external items: %v",
			err,
		)
	}

	if count != 2 {
		t.Fatalf(
			"expected exactly two rows after replay, got %d",
			count,
		)
	}
}

func TestExternalItemIdempotencySameExternalIDDifferentSources(
	t *testing.T,
) {
	repo := newExternalItemRepositoryTestApp(t)

	ctx := context.Background()

	items := []external.Item{
		{
			Source:     external.SourceReddit,
			ExternalID: "same-id",
			Title:      "Reddit item",
		},
		{
			Source:     external.Source("other"),
			ExternalID: "same-id",
			Title:      "Other-source item",
		},
	}

	if err := repo.UpsertBatch(
		ctx,
		items,
	); err != nil {
		t.Fatalf(
			"upsert source-distinct items: %v",
			err,
		)
	}

	var count int

	if err := repo.Pool().
		QueryRow(
			ctx,
			`
SELECT COUNT(*)
FROM external_items
WHERE external_id = $1
`,
			"same-id",
		).
		Scan(&count); err != nil {
		t.Fatalf(
			"count source variants: %v",
			err,
		)
	}

	if count != 2 {
		t.Fatalf(
			"expected two source-specific rows, got %d",
			count,
		)
	}
}
