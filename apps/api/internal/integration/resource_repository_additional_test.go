//go:build integration

package integration

import (
	"context"
	"errors"
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/model"
	"github.com/amyismebyme/the-village/apps/api/internal/repository"
)

func TestResourceRepositoryUpdateAndDeleteNotFound(
	t *testing.T,
) {
	repo := newResourceRepositoryTestApp(t)
	ctx := context.Background()

	missing := &model.Resource{
		ID:          999999,
		Title:       "Missing resource",
		Description: "Should not exist",
		URL:         "https://example.com/missing",
		Category:    "Test",
	}

	if err := repo.Update(ctx, missing); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf(
			"expected ErrNotFound from updating missing resource, got %v",
			err,
		)
	}

	if err := repo.Delete(ctx, missing.ID); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf(
			"expected ErrNotFound from deleting missing resource, got %v",
			err,
		)
	}
}
