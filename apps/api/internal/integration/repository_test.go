//go:build integration
// +build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/amyismebyme/the-village/apps/api/internal/model"
	postgresrepo "github.com/amyismebyme/the-village/apps/api/internal/repository/postgres"
)

func TestCommunityRepositoryCRUD(t *testing.T) {
	t.Skip("CommunityRepository CRUD not implemented yet")

	db := OpenTestDatabase(t)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	repo := postgresrepo.NewCommunityRepository(db.Pool())

	if err := repo.DeleteAll(ctx); err != nil {
		t.Fatalf("clean communities table: %v", err)
	}

	community := model.Community{
		Name:           "Integration Test Community",
		Description:    "Created during integration testing",
		ExternalSource: "internal",
	}

	t.Run("Create", func(t *testing.T) {

		err := repo.Create(ctx, &community)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	})

	t.Run("FindByID", func(t *testing.T) {

		found, err := repo.FindByID(ctx, community.ID)
		if err != nil {
			t.Fatalf("FindByID failed: %v", err)
		}

		if found.ID != community.ID {
			t.Fatalf("expected ID %v got %v", community.ID, found.ID)
		}

		if found.Name != community.Name {
			t.Fatalf("expected name %q got %q", community.Name, found.Name)
		}
	})

	t.Run("Update", func(t *testing.T) {

		community.Name = "Updated Integration Community"

		err := repo.Update(ctx, &community)
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}

		found, err := repo.FindByID(ctx, community.ID)
		if err != nil {
			t.Fatalf("FindByID failed: %v", err)
		}

		if found.Name != "Updated Integration Community" {
			t.Fatalf("update did not persist")
		}
	})

	t.Run("List", func(t *testing.T) {

		list, err := repo.List(ctx)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}

		found := false

		for _, c := range list {
			if c.ID == community.ID {
				found = true
				break
			}
		}

		if !found {
			t.Fatal("community not returned by List")
		}
	})

	t.Run("Delete", func(t *testing.T) {

		err := repo.Delete(ctx, community.ID)
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}
	})

	t.Run("VerifyDeleted", func(t *testing.T) {

		_, err := repo.FindByID(ctx, community.ID)

		if err == nil {
			t.Fatal("expected record to be deleted")
		}
	})
}
