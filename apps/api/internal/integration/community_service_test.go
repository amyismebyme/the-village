//go:build integration
// +build integration

package integration

import (
	"context"
	"errors"
	"github.com/amyismebyme/the-village/apps/api/internal/model"
	"github.com/amyismebyme/the-village/apps/api/internal/repository"
	"github.com/amyismebyme/the-village/apps/api/internal/repository/postgres"
	"github.com/amyismebyme/the-village/apps/api/internal/service"
	"testing"
	"time"
)

func newCommunityService(
	t *testing.T,
) (
	service.CommunityService,
	*postgres.CommunityRepository,
) {

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

	t.Cleanup(func() {
		db.Close()
	})

	repo := postgres.NewCommunityRepository(
		db.Pool(),
	)

	if err := repo.DeleteAll(context.Background()); err != nil {
		t.Fatal(err)
	}

	svc := service.NewCommunityService(repo)

	return svc, repo
}

func TestCommunityServiceCreate(
	t *testing.T,
) {

	svc, _ := newCommunityService(t)

	community := &model.Community{
		Name: "Toronto Men",
		Slug: "toronto-men",
	}

	err := svc.Create(
		context.Background(),
		community,
	)

	if err != nil {
		t.Fatal(err)
	}

	if community.ID == 0 {
		t.Fatal("expected ID")
	}
}

func TestCommunityServiceDuplicateSlug(
	t *testing.T,
) {

	svc, _ := newCommunityService(t)

	first := &model.Community{
		Name: "First",
		Slug: "toronto",
	}

	second := &model.Community{
		Name: "Second",
		Slug: "toronto",
	}

	if err := svc.Create(
		context.Background(),
		first,
	); err != nil {
		t.Fatal(err)
	}

	err := svc.Create(
		context.Background(),
		second,
	)

	if !errors.Is(
		err,
		service.ErrCommunityAlreadyExists,
	) {

		t.Fatalf(
			"expected duplicate error got %v",
			err,
		)
	}
}

func TestCommunityServiceGet(
	t *testing.T,
) {

	svc, _ := newCommunityService(t)

	created := &model.Community{
		Name: "Ottawa",
		Slug: "ottawa",
	}

	if err := svc.Create(
		context.Background(),
		created,
	); err != nil {
		t.Fatal(err)
	}

	found, err := svc.Get(
		context.Background(),
		created.ID,
	)

	if err != nil {
		t.Fatal(err)
	}

	if found.Name != created.Name {
		t.Fatal("wrong community")
	}
}

func TestCommunityServiceUpdate(
	t *testing.T,
) {

	svc, _ := newCommunityService(t)

	community := &model.Community{
		Name: "Old",
		Slug: "old",
	}

	if err := svc.Create(
		context.Background(),
		community,
	); err != nil {
		t.Fatal(err)
	}

	community.Name = "New"

	if err := svc.Update(
		context.Background(),
		community,
	); err != nil {
		t.Fatal(err)
	}

	updated, err := svc.Get(
		context.Background(),
		community.ID,
	)

	if err != nil {
		t.Fatal(err)
	}

	if updated.Name != "New" {
		t.Fatal("update failed")
	}
}

func TestCommunityServiceList(
	t *testing.T,
) {

	svc, _ := newCommunityService(t)

	if err := svc.Create(context.Background(), &model.Community{
		Name: "Aaaa",
		Slug: "aaaa",
	}); err != nil {
		t.Fatal(err)
	}

	if err := svc.Create(context.Background(), &model.Community{
		Name: "Bbbb",
		Slug: "bbbb",
	}); err != nil {
		t.Fatal(err)
	}

	communities, err := svc.List(
		context.Background(),
	)

	if err != nil {
		t.Fatal(err)
	}

	if len(communities) != 2 {
		t.Fatalf(
			"expected 2 got %d",
			len(communities),
		)
	}
}

func TestCommunityServiceDelete(
	t *testing.T,
) {

	svc, _ := newCommunityService(t)

	community := &model.Community{
		Name: "Delete",
		Slug: "delete",
	}

	if err := svc.Create(
		context.Background(),
		community,
	); err != nil {
		t.Fatal(err)
	}

	if err := svc.Delete(
		context.Background(),
		community.ID,
	); err != nil {
		t.Fatal(err)
	}

	_, err := svc.Get(
		context.Background(),
		community.ID,
	)

	if !errors.Is(
		err,
		repository.ErrNotFound,
	) {

		t.Fatalf(
			"expected not found got %v",
			err,
		)
	}
}

func TestCommunityValidation(
	t *testing.T,
) {

	svc, _ := newCommunityService(t)

	community := &model.Community{
		Name: "",
		Slug: "",
	}

	err := svc.Create(
		context.Background(),
		community,
	)

	if err == nil {
		t.Fatal("expected validation error")
	}
}
