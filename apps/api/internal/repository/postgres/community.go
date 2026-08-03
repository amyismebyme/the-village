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
) ([]*model.Community, error) {

	rows, err := r.Pool().Query(ctx, `
SELECT
    id,
    name,
    slug,
    description,
    external_source,
    created_at,
    updated_at
FROM communities
ORDER BY name;
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	communities := make([]*model.Community, 0)

	for rows.Next() {

		community := &model.Community{}

		err := rows.Scan(
			&community.ID,
			&community.Name,
			&community.Slug,
			&community.Description,
			&community.ExternalSource,
			&community.CreatedAt,
			&community.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		communities = append(communities, community)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return communities, nil
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

func (r *CommunityRepository) FindBySlug(
	ctx context.Context,
	slug string,
) (*model.Community, error) {

	query := `
SELECT
    id,
    name,
    slug,
    description,
    external_source,
    created_at,
    updated_at
FROM communities
WHERE slug=$1;
`

	var community model.Community

	err := r.Pool().
		QueryRow(ctx, query, slug).
		Scan(
			&community.ID,
			&community.Name,
			&community.Slug,
			&community.Description,
			&community.ExternalSource,
			&community.CreatedAt,
			&community.UpdatedAt,
		)

	if err != nil {
		return nil, err
	}

	return &community, nil
}
