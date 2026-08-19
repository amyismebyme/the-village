//go:build integration

package integration

import (
	"context"
	"errors"
	"github.com/amyismebyme/the-village/apps/api/internal/model"
	"github.com/amyismebyme/the-village/apps/api/internal/repository"
	"testing"
	"time"
)

func newTestCommunity() *model.Community {
	return &model.Community{
		Name:           "Repository Test Community",
		Slug:           "repository-test-community",
		Description:    "Repository integration test",
		ExternalSource: "integration",
	}
}

func TestCommunityRepositoryCreate(t *testing.T) {
	app := newCommunityRepositoryTestApp(t)

	community := newTestCommunity()

	err := app.repo.Create(
		context.Background(),
		community,
	)
	if err != nil {
		t.Fatalf(
			"Create failed: %v",
			err,
		)
	}

	if community.ID <= 0 {
		t.Fatalf(
			"expected generated ID, got %d",
			community.ID,
		)
	}

	if community.CreatedAt.IsZero() {
		t.Fatal("expected CreatedAt to be populated")
	}

	if community.UpdatedAt.IsZero() {
		t.Fatal("expected UpdatedAt to be populated")
	}
}

func TestCommunityRepositoryCreateDuplicate(t *testing.T) {
	app := newCommunityRepositoryTestApp(t)

	first := newTestCommunity()

	if err := app.repo.Create(
		context.Background(),
		first,
	); err != nil {
		t.Fatalf(
			"first Create failed: %v",
			err,
		)
	}

	second := newTestCommunity()

	err := app.repo.Create(
		context.Background(),
		second,
	)

	if !errors.Is(
		err,
		repository.ErrAlreadyExists,
	) {
		t.Fatalf(
			"expected repository.ErrAlreadyExists, got %v",
			err,
		)
	}
}

func TestCommunityRepositoryCreateNil(t *testing.T) {
	app := newCommunityRepositoryTestApp(t)

	err := app.repo.Create(
		context.Background(),
		nil,
	)

	if err == nil {
		t.Fatal(
			"expected error creating nil community",
		)
	}
}

func TestCommunityRepositoryCreateCancelled(t *testing.T) {
	app := newCommunityRepositoryTestApp(t)

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	err := app.repo.Create(
		ctx,
		newTestCommunity(),
	)

	if !errors.Is(
		err,
		context.Canceled,
	) {
		t.Fatalf(
			"expected context.Canceled, got %v",
			err,
		)
	}
}

func TestCommunityRepositoryFindByID(t *testing.T) {
	app := newCommunityRepositoryTestApp(t)

	community := newTestCommunity()

	if err := app.repo.Create(
		context.Background(),
		community,
	); err != nil {
		t.Fatalf(
			"Create failed: %v",
			err,
		)
	}

	found, err := app.repo.FindByID(
		context.Background(),
		community.ID,
	)
	if err != nil {
		t.Fatalf(
			"FindByID failed: %v",
			err,
		)
	}

	if found == nil {
		t.Fatal("expected community, got nil")
	}

	if found.ID != community.ID {
		t.Fatalf(
			"expected ID %d, got %d",
			community.ID,
			found.ID,
		)
	}
}

func TestCommunityRepositoryFindByIDNotFound(t *testing.T) {
	app := newCommunityRepositoryTestApp(t)

	_, err := app.repo.FindByID(
		context.Background(),
		999999999,
	)

	if !errors.Is(
		err,
		repository.ErrNotFound,
	) {
		t.Fatalf(
			"expected repository.ErrNotFound, got %v",
			err,
		)
	}
}

func TestCommunityRepositoryFindByIDCancelled(t *testing.T) {
	app := newCommunityRepositoryTestApp(t)

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	_, err := app.repo.FindByID(
		ctx,
		1,
	)

	if !errors.Is(
		err,
		context.Canceled,
	) {
		t.Fatalf(
			"expected context.Canceled, got %v",
			err,
		)
	}
}

func TestCommunityRepositoryFindBySlug(t *testing.T) {
	app := newCommunityRepositoryTestApp(t)

	community := newTestCommunity()

	if err := app.repo.Create(
		context.Background(),
		community,
	); err != nil {
		t.Fatalf(
			"Create failed: %v",
			err,
		)
	}

	found, err := app.repo.FindBySlug(
		context.Background(),
		community.Slug,
	)
	if err != nil {
		t.Fatalf(
			"FindBySlug failed: %v",
			err,
		)
	}

	if found == nil {
		t.Fatal("expected community, got nil")
	}

	if found.ID != community.ID {
		t.Fatalf(
			"expected ID %d, got %d",
			community.ID,
			found.ID,
		)
	}
}

func TestCommunityRepositoryFindBySlugNotFound(t *testing.T) {
	app := newCommunityRepositoryTestApp(t)

	_, err := app.repo.FindBySlug(
		context.Background(),
		"does-not-exist",
	)

	if !errors.Is(
		err,
		repository.ErrNotFound,
	) {
		t.Fatalf(
			"expected repository.ErrNotFound, got %v",
			err,
		)
	}
}

func TestCommunityRepositoryFindBySlugCancelled(t *testing.T) {
	app := newCommunityRepositoryTestApp(t)

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	_, err := app.repo.FindBySlug(
		ctx,
		"does-not-exist",
	)

	if !errors.Is(
		err,
		context.Canceled,
	) {
		t.Fatalf(
			"expected context.Canceled, got %v",
			err,
		)
	}
}

func TestCommunityRepositoryList(t *testing.T) {
	app := newCommunityRepositoryTestApp(t)

	first := newTestCommunity()
	second := &model.Community{
		Name:           "Second Repository Community",
		Slug:           "second-repository-community",
		Description:    "Second repository test",
		ExternalSource: "integration",
	}

	if err := app.repo.Create(
		context.Background(),
		first,
	); err != nil {
		t.Fatalf(
			"Create first failed: %v",
			err,
		)
	}

	if err := app.repo.Create(
		context.Background(),
		second,
	); err != nil {
		t.Fatalf(
			"Create second failed: %v",
			err,
		)
	}

	communities, err := app.repo.List(
		context.Background(),
	)
	if err != nil {
		t.Fatalf(
			"List failed: %v",
			err,
		)
	}

	if communities == nil {
		t.Fatal(
			"expected non-nil community slice",
		)
	}

	if len(communities) != 2 {
		t.Fatalf(
			"expected 2 communities, got %d",
			len(communities),
		)
	}
}

func TestCommunityRepositoryListEmpty(t *testing.T) {
	app := newCommunityRepositoryTestApp(t)

	communities, err := app.repo.List(
		context.Background(),
	)
	if err != nil {
		t.Fatalf(
			"List failed: %v",
			err,
		)
	}

	if communities == nil {
		t.Fatal(
			"expected empty slice, got nil",
		)
	}

	if len(communities) != 0 {
		t.Fatalf(
			"expected 0 communities, got %d",
			len(communities),
		)
	}
}

func TestCommunityRepositoryListCancelled(t *testing.T) {
	app := newCommunityRepositoryTestApp(t)

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	_, err := app.repo.List(ctx)

	if !errors.Is(
		err,
		context.Canceled,
	) {
		t.Fatalf(
			"expected context.Canceled, got %v",
			err,
		)
	}
}

func TestCommunityRepositoryUpdate(t *testing.T) {
	app := newCommunityRepositoryTestApp(t)

	community := newTestCommunity()

	if err := app.repo.Create(
		context.Background(),
		community,
	); err != nil {
		t.Fatalf(
			"Create failed: %v",
			err,
		)
	}

	originalUpdatedAt := community.UpdatedAt

	community.Name = "Updated Repository Community"
	community.Description = "Updated repository description"

	if err := app.repo.Update(
		context.Background(),
		community,
	); err != nil {
		t.Fatalf(
			"Update failed: %v",
			err,
		)
	}

	if community.UpdatedAt.IsZero() {
		t.Fatal(
			"expected UpdatedAt to be populated",
		)
	}

	if community.UpdatedAt.Before(
		originalUpdatedAt,
	) {
		t.Fatalf(
			"UpdatedAt moved backwards: before=%v after=%v",
			originalUpdatedAt,
			community.UpdatedAt,
		)
	}

	found, err := app.repo.FindByID(
		context.Background(),
		community.ID,
	)
	if err != nil {
		t.Fatalf(
			"FindByID after Update failed: %v",
			err,
		)
	}

	if found.Name != community.Name {
		t.Fatalf(
			"expected name %q, got %q",
			community.Name,
			found.Name,
		)
	}
}

func TestCommunityRepositoryUpdateNotFound(t *testing.T) {
	app := newCommunityRepositoryTestApp(t)

	community := &model.Community{
		ID:   999999999,
		Name: "Missing Community",
		Slug: "missing-community",
	}

	err := app.repo.Update(
		context.Background(),
		community,
	)

	if !errors.Is(
		err,
		repository.ErrNotFound,
	) {
		t.Fatalf(
			"expected repository.ErrNotFound, got %v",
			err,
		)
	}
}

func TestCommunityRepositoryUpdateDuplicate(t *testing.T) {
	app := newCommunityRepositoryTestApp(t)

	first := newTestCommunity()
	second := &model.Community{
		Name:           "Second Community",
		Slug:           "second-community",
		Description:    "Second test",
		ExternalSource: "integration",
	}

	if err := app.repo.Create(
		context.Background(),
		first,
	); err != nil {
		t.Fatalf(
			"Create first failed: %v",
			err,
		)
	}

	if err := app.repo.Create(
		context.Background(),
		second,
	); err != nil {
		t.Fatalf(
			"Create second failed: %v",
			err,
		)
	}

	second.Slug = first.Slug

	err := app.repo.Update(
		context.Background(),
		second,
	)

	if !errors.Is(
		err,
		repository.ErrAlreadyExists,
	) {
		t.Fatalf(
			"expected repository.ErrAlreadyExists, got %v",
			err,
		)
	}
}

func TestCommunityRepositoryUpdateNil(t *testing.T) {
	app := newCommunityRepositoryTestApp(t)

	err := app.repo.Update(
		context.Background(),
		nil,
	)

	if err == nil {
		t.Fatal(
			"expected error updating nil community",
		)
	}
}

func TestCommunityRepositoryUpdateCancelled(t *testing.T) {
	app := newCommunityRepositoryTestApp(t)

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	err := app.repo.Update(
		ctx,
		&model.Community{
			ID:   1,
			Name: "Cancelled Update",
			Slug: "cancelled-update",
		},
	)

	if !errors.Is(
		err,
		context.Canceled,
	) {
		t.Fatalf(
			"expected context.Canceled, got %v",
			err,
		)
	}
}

func TestCommunityRepositoryDelete(t *testing.T) {
	app := newCommunityRepositoryTestApp(t)

	community := newTestCommunity()

	if err := app.repo.Create(
		context.Background(),
		community,
	); err != nil {
		t.Fatalf(
			"Create failed: %v",
			err,
		)
	}

	if err := app.repo.Delete(
		context.Background(),
		community.ID,
	); err != nil {
		t.Fatalf(
			"Delete failed: %v",
			err,
		)
	}

	_, err := app.repo.FindByID(
		context.Background(),
		community.ID,
	)

	if !errors.Is(
		err,
		repository.ErrNotFound,
	) {
		t.Fatalf(
			"expected deleted community to be absent, got %v",
			err,
		)
	}
}

func TestCommunityRepositoryDeleteNotFound(t *testing.T) {
	app := newCommunityRepositoryTestApp(t)

	err := app.repo.Delete(
		context.Background(),
		999999999,
	)

	if !errors.Is(
		err,
		repository.ErrNotFound,
	) {
		t.Fatalf(
			"expected repository.ErrNotFound, got %v",
			err,
		)
	}
}

func TestCommunityRepositoryDeleteCancelled(t *testing.T) {
	app := newCommunityRepositoryTestApp(t)

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	err := app.repo.Delete(
		ctx,
		1,
	)

	if !errors.Is(
		err,
		context.Canceled,
	) {
		t.Fatalf(
			"expected context.Canceled, got %v",
			err,
		)
	}
}

// Keep time imported here because repository tests deliberately create
// explicit request/deadline contexts when timeout behavior is expanded.
var _ = time.Second
