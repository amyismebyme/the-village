//go:build integration

package integration

import (
	"context"
	"errors"
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/model"
	"github.com/amyismebyme/the-village/apps/api/internal/repository"
	"github.com/amyismebyme/the-village/apps/api/internal/repository/postgres"
)

func newResourceRepositoryTestApp(
	t *testing.T,
) *postgres.ResourceRepository {
	t.Helper()

	db := OpenTestDatabase(t)
	repo := postgres.NewResourceRepository(
		db.Pool(),
	)

	ctx := context.Background()
	if err := repo.DeleteAll(ctx); err != nil {
		t.Fatalf(
			"clean resources before test: %v",
			err,
		)
	}

	t.Cleanup(func() {
		if err := repo.DeleteAll(
			context.Background(),
		); err != nil {
			t.Logf(
				"clean resources after test: %v",
				err,
			)
		}
	})

	return repo
}

func TestResourceRepositoryCRUD(t *testing.T) {
	repo := newResourceRepositoryTestApp(t)
	ctx := context.Background()

	resource := &model.Resource{
		Title:       "988 Crisis Lifeline",
		Description: "24/7 mental health support",
		URL:         "https://988.ca/",
		Category:    "Mental Health",
	}

	if err := repo.Create(ctx, resource); err != nil {
		t.Fatalf("create resource: %v", err)
	}

	if resource.ID <= 0 {
		t.Fatalf("expected created ID, got %d", resource.ID)
	}

	found, err := repo.FindByID(ctx, resource.ID)
	if err != nil {
		t.Fatalf("find resource: %v", err)
	}

	if found.URL != resource.URL {
		t.Fatalf("expected URL %q, got %q", resource.URL, found.URL)
	}

	resource.Title = "988 Canada"
	if err := repo.Update(ctx, resource); err != nil {
		t.Fatalf("update resource: %v", err)
	}

	updated, err := repo.FindByID(ctx, resource.ID)
	if err != nil {
		t.Fatalf("find updated resource: %v", err)
	}

	if updated.Title != "988 Canada" {
		t.Fatalf("expected updated title, got %q", updated.Title)
	}

	items, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("list resources: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected one resource, got %d", len(items))
	}

	if err := repo.Delete(ctx, resource.ID); err != nil {
		t.Fatalf("delete resource: %v", err)
	}

	_, err = repo.FindByID(ctx, resource.ID)
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected repository.ErrNotFound after delete, got %v", err)
	}
}

func TestResourceRepositoryRejectsInvalidInput(t *testing.T) {
	repo := newResourceRepositoryTestApp(t)
	ctx := context.Background()

	if _, err := repo.FindByID(ctx, 0); !errors.Is(err, repository.ErrInvalidInput) {
		t.Fatalf("expected invalid input for ID 0, got %v", err)
	}

	if err := repo.Create(ctx, nil); !errors.Is(err, repository.ErrInvalidInput) {
		t.Fatalf("expected invalid input for nil resource, got %v", err)
	}
}
