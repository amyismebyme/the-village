package postgres

import (
	"context"
	"fmt"

	"github.com/amyismebyme/the-village/apps/api/internal/model"
	"github.com/amyismebyme/the-village/apps/api/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

var _ repository.CommunityRepository = (*CommunityRepository)(nil)

type CommunityRepository struct {
	Repository
}

func NewCommunityRepository(pool *pgxpool.Pool) *CommunityRepository {

	return &CommunityRepository{
		Repository: New(pool),
	}
}

func (r *CommunityRepository) List(
	ctx context.Context,
) ([]model.Community, error) {

	_ = ctx

	// TODO:
	// Query PostgreSQL after migrations are created.
	return nil, fmt.Errorf("community repository: List not implemented")
}

func (r *CommunityRepository) FindByID(
	ctx context.Context,
	id int64,
) (*model.Community, error) {

	_ = ctx
	_ = id

	// TODO:
	// Query PostgreSQL after migrations are created.
	return nil, fmt.Errorf("community repository: FindByID not implemented")
}

func (r *CommunityRepository) Create(
	ctx context.Context,
	community *model.Community,
) error {

	_ = ctx
	_ = community

	// TODO:
	// Insert community after schema exists.
	return repository.ErrNotImplemented
}

func (r *CommunityRepository) Update(
	ctx context.Context,
	community *model.Community,
) error {

	_ = ctx
	_ = community

	// TODO:
	// Update community after schema exists.
	return repository.ErrNotImplemented
}

func (r *CommunityRepository) Delete(
	ctx context.Context,
	id int64,
) error {

	_ = ctx
	_ = id

	// TODO:
	// Delete community after schema exists.
	return repository.ErrNotImplemented
}

// DeleteAll removes every community and resets the identity sequence.
//
// This method is primarily intended for integration-test cleanup.
func (r *CommunityRepository) DeleteAll(
	ctx context.Context,
) error {
	const query = `
TRUNCATE TABLE communities
RESTART IDENTITY
`

	if _, err := r.pool.Exec(ctx, query); err != nil {
		return fmt.Errorf(
			"community repository: delete all communities: %w",
			err,
		)
	}

	return nil
}
