package postgres

import (
	"context"

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

	const query = `
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
`
	rows, err := r.Pool().Query(ctx, query)

	if err != nil {
		return nil, translateError(err)
	}

	return scanCommunities(rows)
}

func (r *CommunityRepository) FindByID(
	ctx context.Context,
	id int64,
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
WHERE id=$1;
`

	community, err := scanCommunity(
		r.Pool().QueryRow(ctx, query, id),
	)

	if err != nil {
		return nil, translateError(err)
	}

	return community, nil
}
func (r *CommunityRepository) Create(
	ctx context.Context,
	community *model.Community,
) error {

	query := `
INSERT INTO communities
(
	name,
	slug,
	description,
	external_source
)
VALUES
(
	$1,
	$2,
	$3,
	$4
)
RETURNING
	id,
	created_at,
	updated_at;
`

	err := r.Pool().
		QueryRow(
			ctx,
			query,
			community.Name,
			community.Slug,
			community.Description,
			community.ExternalSource,
		).
		Scan(
			&community.ID,
			&community.CreatedAt,
			&community.UpdatedAt,
		)

	if err != nil {
		return translateError(err)
	}

	return nil
}

func (r *CommunityRepository) Update(
	ctx context.Context,
	community *model.Community,
) error {

	query := `
UPDATE communities
SET

	name=$1,
	slug=$2,
	description=$3,
	external_source=$4,
	updated_at=NOW()

WHERE id=$5

RETURNING updated_at;
`

	err := r.Pool().
		QueryRow(
			ctx,
			query,
			community.Name,
			community.Slug,
			community.Description,
			community.ExternalSource,
			community.ID,
		).
		Scan(&community.UpdatedAt)

	if err != nil {
		return translateError(err)
	}

	return nil
}

func (r *CommunityRepository) Delete(
	ctx context.Context,
	id int64,
) error {

	const query = `
DELETE FROM communities
WHERE id=$1;
`

	return execOne(
		ctx,
		r.Repository,
		query,
		id,
	)
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
		return translateError(err)
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

	community, err := scanCommunity(
		r.Pool().QueryRow(ctx, query, slug),
	)

	if err != nil {
		return nil, translateError(err)
	}

	return community, nil
}
