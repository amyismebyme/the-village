package postgres

import (
	"context"
	"fmt"

	"github.com/amyismebyme/the-village/apps/api/internal/model"
	"github.com/amyismebyme/the-village/apps/api/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

var _ repository.ResourceRepository = (*ResourceRepository)(nil)

type ResourceRepository struct {
	Repository
}

func NewResourceRepository(pool *pgxpool.Pool) *ResourceRepository {

	return &ResourceRepository{
		Repository: New(pool),
	}
}

func (r *ResourceRepository) List(
	ctx context.Context,
) ([]model.Resource, error) {

	_ = ctx

	// TODO:
	// Query PostgreSQL after migrations are created.
	return nil, fmt.Errorf("resource repository: List not implemented")
}

func (r *ResourceRepository) FindByID(
	ctx context.Context,
	id int64,
) (*model.Resource, error) {

	_ = ctx
	_ = id

	// TODO:
	// Query PostgreSQL after migrations are created.
	return nil, fmt.Errorf("resource repository: FindByID not implemented")
}

func (r *ResourceRepository) Create(
	ctx context.Context,
	resource *model.Resource,
) error {

	_ = ctx
	_ = resource

	// TODO:
	// Insert resource after schema exists.
	return repository.ErrNotImplemented
}

func (r *ResourceRepository) Update(
	ctx context.Context,
	resource *model.Resource,
) error {

	_ = ctx
	_ = resource

	// TODO:
	// Update resource after schema exists.
	return repository.ErrNotImplemented
}

func (r *ResourceRepository) Delete(
	ctx context.Context,
	id int64,
) error {

	_ = ctx
	_ = id

	// TODO:
	// Delete resource after schema exists.
	return repository.ErrNotImplemented
}
