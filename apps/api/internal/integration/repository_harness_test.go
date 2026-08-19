//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/database"
	"github.com/amyismebyme/the-village/apps/api/internal/repository/postgres"
)

// communityRepositoryTestApp contains the shared database and repository
// dependencies used by repository integration tests.
type communityRepositoryTestApp struct {
	db   *database.Database
	repo *postgres.CommunityRepository
}

// newCommunityRepositoryTestApp creates a clean CommunityRepository test
// environment.
//
// The communities table is cleared before the test and again during cleanup.
// Keeping this setup in one shared helper prevents individual repository test
// files from maintaining duplicate database lifecycle logic.
func newCommunityRepositoryTestApp(
	t *testing.T,
) *communityRepositoryTestApp {
	t.Helper()

	db := OpenTestDatabase(t)

	repo := postgres.NewCommunityRepository(
		db.Pool(),
	)

	ctx := context.Background()

	if err := repo.DeleteAll(ctx); err != nil {
		t.Fatalf(
			"clean communities before test: %v",
			err,
		)
	}

	t.Cleanup(func() {
		if err := repo.DeleteAll(
			context.Background(),
		); err != nil {
			t.Logf(
				"clean communities after test: %v",
				err,
			)
		}
	})

	return &communityRepositoryTestApp{
		db:   db,
		repo: repo,
	}
}
