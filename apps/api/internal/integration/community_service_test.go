//go:build integration

package integration

import (
	"context"
	"errors"
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/config"
	"github.com/amyismebyme/the-village/apps/api/internal/database"
	"github.com/amyismebyme/the-village/apps/api/internal/model"
	"github.com/amyismebyme/the-village/apps/api/internal/repository"
	"github.com/amyismebyme/the-village/apps/api/internal/repository/postgres"
	"github.com/amyismebyme/the-village/apps/api/internal/service"
)

// -----------------------------------------------------------------------------
// Test service setup
// -----------------------------------------------------------------------------

type communityServiceTestApp struct {
	service service.CommunityService
	repo    *postgres.CommunityRepository
	close   func()
}

func newCommunityServiceTestApp(
	t *testing.T,
) *communityServiceTestApp {
	t.Helper()

	cfg := config.Load()

	db, err := database.Open(
		context.Background(),
		cfg.Database,
	)
	if err != nil {
		t.Fatalf(
			"open database: %v",
			err,
		)
	}

	repo := postgres.NewCommunityRepository(
		db.Pool(),
	)

	if err := repo.DeleteAll(
		context.Background(),
	); err != nil {
		db.Close()

		t.Fatalf(
			"clean communities before test: %v",
			err,
		)
	}

	closeApp := func() {
		if err := repo.DeleteAll(
			context.Background(),
		); err != nil {
			t.Logf(
				"clean communities after test: %v",
				err,
			)
		}

		db.Close()
	}

	t.Cleanup(closeApp)

	communityService := service.NewCommunityService(
		repo,
	)

	return &communityServiceTestApp{
		service: communityService,
		repo:    repo,
		close:   closeApp,
	}
}

// -----------------------------------------------------------------------------
// Create
// -----------------------------------------------------------------------------

func TestCommunityService_Create(t *testing.T) {
	app := newCommunityServiceTestApp(t)

	ctx := context.Background()

	community := &model.Community{
		Name:           "Toronto Men",
		Slug:           "toronto-men-service-integration",
		Description:    "A service integration test community.",
		ExternalSource: "manual",
	}

	err := app.service.Create(
		ctx,
		community,
	)

	if err != nil {
		t.Fatalf(
			"create community: %v",
			err,
		)
	}

	if community.ID <= 0 {
		t.Fatalf(
			"expected generated ID > 0, got %d",
			community.ID,
		)
	}

	if community.CreatedAt.IsZero() {
		t.Fatal(
			"expected CreatedAt to be populated",
		)
	}

	if community.UpdatedAt.IsZero() {
		t.Fatal(
			"expected UpdatedAt to be populated",
		)
	}
}

// -----------------------------------------------------------------------------
// Create duplicate
// -----------------------------------------------------------------------------

func TestCommunityService_CreateDuplicate(t *testing.T) {
	app := newCommunityServiceTestApp(t)

	ctx := context.Background()

	first := &model.Community{
		Name:           "Toronto Men",
		Slug:           "duplicate-service-integration",
		Description:    "First community.",
		ExternalSource: "manual",
	}

	if err := app.service.Create(
		ctx,
		first,
	); err != nil {
		t.Fatalf(
			"create first community: %v",
			err,
		)
	}

	second := &model.Community{
		Name:           "Another Toronto Community",
		Slug:           "duplicate-service-integration",
		Description:    "Duplicate slug.",
		ExternalSource: "manual",
	}

	err := app.service.Create(
		ctx,
		second,
	)

	if !errors.Is(
		err,
		service.ErrCommunityAlreadyExists,
	) {
		t.Fatalf(
			"expected ErrCommunityAlreadyExists, got %v",
			err,
		)
	}
}

// -----------------------------------------------------------------------------
// Get existing
// -----------------------------------------------------------------------------

func TestCommunityService_Get(t *testing.T) {
	app := newCommunityServiceTestApp(t)

	ctx := context.Background()

	community := &model.Community{
		Name:           "Mississauga Men",
		Slug:           "mississauga-men-service-integration",
		Description:    "Service integration test.",
		ExternalSource: "manual",
	}

	if err := app.service.Create(
		ctx,
		community,
	); err != nil {
		t.Fatalf(
			"create community: %v",
			err,
		)
	}

	retrieved, err := app.service.Get(
		ctx,
		community.ID,
	)

	if err != nil {
		t.Fatalf(
			"get community: %v",
			err,
		)
	}

	if retrieved == nil {
		t.Fatal(
			"expected community, got nil",
		)
	}

	if retrieved.ID != community.ID {
		t.Fatalf(
			"expected ID %d, got %d",
			community.ID,
			retrieved.ID,
		)
	}

	if retrieved.Name != community.Name {
		t.Fatalf(
			"expected name %q, got %q",
			community.Name,
			retrieved.Name,
		)
	}

	if retrieved.Slug != community.Slug {
		t.Fatalf(
			"expected slug %q, got %q",
			community.Slug,
			retrieved.Slug,
		)
	}
}

// -----------------------------------------------------------------------------
// Get missing
// -----------------------------------------------------------------------------

func TestCommunityService_GetMissing(t *testing.T) {
	app := newCommunityServiceTestApp(t)

	_, err := app.service.Get(
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

// -----------------------------------------------------------------------------
// List
// -----------------------------------------------------------------------------

func TestCommunityService_List(t *testing.T) {
	app := newCommunityServiceTestApp(t)

	ctx := context.Background()

	communities := []*model.Community{
		{
			Name:           "Toronto Men",
			Slug:           "toronto-men-list-integration",
			Description:    "Toronto community.",
			ExternalSource: "manual",
		},
		{
			Name:           "Hamilton Men",
			Slug:           "hamilton-men-list-integration",
			Description:    "Hamilton community.",
			ExternalSource: "manual",
		},
	}

	for _, community := range communities {
		if err := app.service.Create(
			ctx,
			community,
		); err != nil {
			t.Fatalf(
				"create community %q: %v",
				community.Name,
				err,
			)
		}
	}

	result, err := app.service.List(ctx, 20, 0)

	if err != nil {
		t.Fatalf(
			"list communities: %v",
			err,
		)
	}

	if len(result.Communities) != 2 {
		t.Fatalf(
			"expected 2 communities, got %d",
			len(result.Communities),
		)
	}
}

// -----------------------------------------------------------------------------
// Update existing
// -----------------------------------------------------------------------------

func TestCommunityService_Update(t *testing.T) {
	app := newCommunityServiceTestApp(t)

	ctx := context.Background()

	community := &model.Community{
		Name:           "Hamilton Men",
		Slug:           "hamilton-men-update-integration",
		Description:    "Original description.",
		ExternalSource: "manual",
	}

	if err := app.service.Create(
		ctx,
		community,
	); err != nil {
		t.Fatalf(
			"create community: %v",
			err,
		)
	}

	community.Name = "Hamilton Men's Community"
	community.Description = "Updated description."

	if err := app.service.Update(
		ctx,
		community,
	); err != nil {
		t.Fatalf(
			"update community: %v",
			err,
		)
	}

	updated, err := app.service.Get(
		ctx,
		community.ID,
	)

	if err != nil {
		t.Fatalf(
			"get updated community: %v",
			err,
		)
	}

	if updated.Name != "Hamilton Men's Community" {
		t.Fatalf(
			"expected updated name %q, got %q",
			"Hamilton Men's Community",
			updated.Name,
		)
	}

	if updated.Description != "Updated description." {
		t.Fatalf(
			"expected updated description %q, got %q",
			"Updated description.",
			updated.Description,
		)
	}
}

// -----------------------------------------------------------------------------
// Update missing
// -----------------------------------------------------------------------------

func TestCommunityService_UpdateMissing(t *testing.T) {
	app := newCommunityServiceTestApp(t)

	community := &model.Community{
		ID:             999999999,
		Name:           "Missing Community",
		Slug:           "missing-community-update",
		Description:    "Should not exist.",
		ExternalSource: "manual",
	}

	err := app.service.Update(
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

// -----------------------------------------------------------------------------
// Update duplicate slug
// -----------------------------------------------------------------------------

func TestCommunityService_UpdateDuplicateSlug(t *testing.T) {
	app := newCommunityServiceTestApp(t)

	ctx := context.Background()

	first := &model.Community{
		Name:           "Toronto Men",
		Slug:           "toronto-update-duplicate",
		Description:    "First community.",
		ExternalSource: "manual",
	}

	second := &model.Community{
		Name:           "Hamilton Men",
		Slug:           "hamilton-update-duplicate",
		Description:    "Second community.",
		ExternalSource: "manual",
	}

	if err := app.service.Create(
		ctx,
		first,
	); err != nil {
		t.Fatalf(
			"create first community: %v",
			err,
		)
	}

	if err := app.service.Create(
		ctx,
		second,
	); err != nil {
		t.Fatalf(
			"create second community: %v",
			err,
		)
	}

	second.Slug = first.Slug

	err := app.service.Update(
		ctx,
		second,
	)

	if !errors.Is(
		err,
		service.ErrCommunityAlreadyExists,
	) {
		t.Fatalf(
			"expected ErrCommunityAlreadyExists, got %v",
			err,
		)
	}
}

// -----------------------------------------------------------------------------
// Delete existing
// -----------------------------------------------------------------------------

func TestCommunityService_Delete(t *testing.T) {
	app := newCommunityServiceTestApp(t)

	ctx := context.Background()

	community := &model.Community{
		Name:           "Burlington Men",
		Slug:           "burlington-men-delete-integration",
		Description:    "Delete integration test.",
		ExternalSource: "manual",
	}

	if err := app.service.Create(
		ctx,
		community,
	); err != nil {
		t.Fatalf(
			"create community: %v",
			err,
		)
	}

	if err := app.service.Delete(
		ctx,
		community.ID,
	); err != nil {
		t.Fatalf(
			"delete community: %v",
			err,
		)
	}

	_, err := app.service.Get(
		ctx,
		community.ID,
	)

	if !errors.Is(
		err,
		repository.ErrNotFound,
	) {
		t.Fatalf(
			"expected repository.ErrNotFound after delete, got %v",
			err,
		)
	}
}

// -----------------------------------------------------------------------------
// Delete missing
// -----------------------------------------------------------------------------

func TestCommunityService_DeleteMissing(t *testing.T) {
	app := newCommunityServiceTestApp(t)

	err := app.service.Delete(
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
