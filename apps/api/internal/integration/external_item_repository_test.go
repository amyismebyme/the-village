//go:build integration

package integration

import (
	"context"
	"errors"
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/external"
	"github.com/amyismebyme/the-village/apps/api/internal/repository"
	"github.com/amyismebyme/the-village/apps/api/internal/repository/postgres"
)

func newExternalItemRepositoryTestRepository(
	t *testing.T,
) *postgres.ExternalItemRepository {
	t.Helper()

	db := OpenTestDatabase(t)

	repo := postgres.NewExternalItemRepository(
		db.Pool(),
	)

	if err := repo.DeleteAll(
		context.Background(),
	); err != nil {
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

func TestExternalItemRepositoryUpsertCreatesItem(
	t *testing.T,
) {
	repo := newExternalItemRepositoryTestRepository(t)

	item := external.Item{
		Source:      external.SourceReddit,
		ExternalID:  "abc123",
		Title:       "Original title",
		Description: "Original description",
		URL:         "https://example.com/abc123",
	}

	err := repo.Upsert(
		context.Background(),
		item,
	)
	if err != nil {
		t.Fatalf(
			"upsert external item: %v",
			err,
		)
	}

	got, err := repo.FindByIdentity(
		context.Background(),
		item.Identity(),
	)
	if err != nil {
		t.Fatalf(
			"find external item: %v",
			err,
		)
	}

	if got == nil {
		t.Fatal(
			"expected persisted external item",
		)
	}

	if got.Source != item.Source {
		t.Fatalf(
			"expected source %q, got %q",
			item.Source,
			got.Source,
		)
	}

	if got.ExternalID != item.ExternalID {
		t.Fatalf(
			"expected external ID %q, got %q",
			item.ExternalID,
			got.ExternalID,
		)
	}

	if got.Title != item.Title {
		t.Fatalf(
			"expected title %q, got %q",
			item.Title,
			got.Title,
		)
	}
}

func TestExternalItemRepositoryUpsertIsReplaySafe(
	t *testing.T,
) {
	repo := newExternalItemRepositoryTestRepository(t)

	item := external.Item{
		Source:      external.SourceReddit,
		ExternalID:  "replay-123",
		Title:       "Replay-safe post",
		Description: "Same payload",
		URL:         "https://example.com/replay-123",
	}

	ctx := context.Background()

	if err := repo.Upsert(ctx, item); err != nil {
		t.Fatalf(
			"first upsert: %v",
			err,
		)
	}

	if err := repo.Upsert(ctx, item); err != nil {
		t.Fatalf(
			"second upsert: %v",
			err,
		)
	}

	var count int

	err := repo.Pool().
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
		Scan(&count)

	if err != nil {
		t.Fatalf(
			"count persisted rows: %v",
			err,
		)
	}

	if count != 1 {
		t.Fatalf(
			"expected exactly one persisted row after replay, got %d",
			count,
		)
	}
}

func TestExternalItemRepositoryUpsertUpdatesExistingItem(
	t *testing.T,
) {
	repo := newExternalItemRepositoryTestRepository(t)

	ctx := context.Background()

	original := external.Item{
		Source:      external.SourceReddit,
		ExternalID:  "mutable-123",
		Title:       "Original title",
		Description: "Original description",
		URL:         "https://example.com/original",
	}

	if err := repo.Upsert(ctx, original); err != nil {
		t.Fatalf(
			"first upsert: %v",
			err,
		)
	}

	updated := original
	updated.Title = "Updated title"
	updated.Description = "Updated description"
	updated.URL = "https://example.com/updated"

	if err := repo.Upsert(ctx, updated); err != nil {
		t.Fatalf(
			"second upsert: %v",
			err,
		)
	}

	got, err := repo.FindByIdentity(
		ctx,
		updated.Identity(),
	)
	if err != nil {
		t.Fatalf(
			"find updated item: %v",
			err,
		)
	}

	if got.Title != "Updated title" {
		t.Fatalf(
			"expected updated title, got %q",
			got.Title,
		)
	}

	if got.Description != "Updated description" {
		t.Fatalf(
			"expected updated description, got %q",
			got.Description,
		)
	}

	if got.URL != "https://example.com/updated" {
		t.Fatalf(
			"expected updated URL, got %q",
			got.URL,
		)
	}

	if got.ExternalID != updated.ExternalID {
		t.Fatalf(
			"expected identity to remain unchanged, got %q",
			got.ExternalID,
		)
	}
}

func TestExternalItemRepositoryAllowsSameExternalIDFromDifferentSources(
	t *testing.T,
) {
	repo := newExternalItemRepositoryTestRepository(t)

	ctx := context.Background()

	redditItem := external.Item{
		Source:     external.SourceReddit,
		ExternalID: "123",
		Title:      "Reddit item",
	}

	otherItem := external.Item{
		Source:     external.Source("other"),
		ExternalID: "123",
		Title:      "Other item",
	}

	if err := repo.Upsert(ctx, redditItem); err != nil {
		t.Fatalf(
			"upsert Reddit item: %v",
			err,
		)
	}

	if err := repo.Upsert(ctx, otherItem); err != nil {
		t.Fatalf(
			"upsert other-source item: %v",
			err,
		)
	}

	var count int

	err := repo.Pool().
		QueryRow(
			ctx,
			`
SELECT COUNT(*)
FROM external_items
WHERE external_id = $1
`,
			"123",
		).
		Scan(&count)

	if err != nil {
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

func TestExternalItemRepositoryRejectsInvalidItem(
	t *testing.T,
) {
	repo := newExternalItemRepositoryTestRepository(t)

	err := repo.Upsert(
		context.Background(),
		external.Item{
			Source: external.SourceReddit,
		},
	)

	if err == nil {
		t.Fatal(
			"expected invalid input error",
		)
	}

	if !errors.Is(
		err,
		repository.ErrInvalidInput,
	) {
		t.Fatalf(
			"expected repository.ErrInvalidInput, got %v",
			err,
		)
	}
}
