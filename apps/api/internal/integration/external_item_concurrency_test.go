//go:build integration

package integration

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/external"
	"github.com/amyismebyme/the-village/apps/api/internal/repository/postgres"
)

func TestExternalItemRepositoryConcurrentUpsertIsSafe(
	t *testing.T,
) {
	db := OpenTestDatabase(t)

	repo := postgres.NewExternalItemRepository(
		db.Pool(),
	)

	ctx := context.Background()

	if err := repo.DeleteAll(ctx); err != nil {
		t.Fatalf(
			"clean external items: %v",
			err,
		)
	}

	t.Cleanup(func() {
		if err := repo.DeleteAll(
			context.Background(),
		); err != nil {
			t.Logf(
				"cleanup external items: %v",
				err,
			)
		}
	})

	item := external.Item{
		Source:     external.SourceReddit,
		ExternalID: "concurrent-17-7",
		Title:      "Concurrent item",
	}

	const workers = 20
	const writesPerWorker = 5

	var wg sync.WaitGroup

	errCh := make(chan error, workers)

	for workerID := 0; workerID < workers; workerID++ {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			for attempt := 0; attempt < writesPerWorker; attempt++ {
				value := item

				value.Title = fmt.Sprintf(
					"Concurrent item %d-%d",
					id,
					attempt,
				)

				if err := repo.Upsert(
					ctx,
					value,
				); err != nil {
					errCh <- err
					return
				}
			}
		}(workerID)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Fatalf(
			"concurrent upsert failed: %v",
			err,
		)
	}

	var count int

	if err := db.Pool().
		QueryRow(
			ctx,
			`
SELECT COUNT(*)
FROM external_items
WHERE source = $1
  AND external_id = $2
`,
			item.Source,
			item.ExternalID,
		).
		Scan(&count); err != nil {
		t.Fatalf(
			"count external items: %v",
			err,
		)
	}

	if count != 1 {
		t.Fatalf(
			"expected exactly one row after concurrent upserts, got %d",
			count,
		)
	}
}

func TestExternalItemRepositoryUpsertBatchIsAtomic(
	t *testing.T,
) {
	db := OpenTestDatabase(t)

	repo := postgres.NewExternalItemRepository(
		db.Pool(),
	)

	ctx := context.Background()

	if err := repo.DeleteAll(ctx); err != nil {
		t.Fatalf(
			"clean external items: %v",
			err,
		)
	}

	t.Cleanup(func() {
		if err := repo.DeleteAll(
			context.Background(),
		); err != nil {
			t.Logf(
				"cleanup external items: %v",
				err,
			)
		}
	})

	err := repo.UpsertBatch(
		ctx,
		[]external.Item{
			{
				Source:     external.SourceReddit,
				ExternalID: "atomic-valid",
				Title:      "Should not persist",
			},
			{
				Source:     external.SourceReddit,
				ExternalID: "",
				Title:      "Invalid",
			},
		},
	)

	if err == nil {
		t.Fatal(
			"expected invalid batch to fail",
		)
	}

	var count int

	if err := db.Pool().
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

	if count != 0 {
		t.Fatalf(
			"expected atomic batch failure to persist zero rows, got %d",
			count,
		)
	}
}
