package postgres

import (
	"context"
	"errors"
	"github.com/amyismebyme/the-village/apps/api/internal/model"
	"github.com/amyismebyme/the-village/apps/api/internal/repository"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var _ repository.CommunityRepository = (*CommunityRepository)(nil)

type CommunityRepository struct {
	Repository
}

func NewCommunityRepository(
	pool *pgxpool.Pool,
	queryTimeout ...time.Duration,
) *CommunityRepository {
	return &CommunityRepository{
		Repository: New(
			pool,
			queryTimeout...,
		),
	}
}

func (r *CommunityRepository) List(
	ctx context.Context,
) (communities []*model.Community, err error) {

	start := time.Now()

	defer func() {
		observeQuery("list", start, err)
	}()

	ctx, cancel := r.withQueryTimeout(ctx)
	defer cancel()

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
	defer rows.Close()

	communities, err = scanCommunities(rows)
	if err != nil {
		return nil, translateError(err)
	}

	if err := rows.Err(); err != nil {
		return nil, translateError(err)
	}

	if communities == nil {
		communities = []*model.Community{}
	}

	return communities, nil
}

func (r *CommunityRepository) FindByID(
	ctx context.Context,
	id int64,
) (community *model.Community, err error) {

	start := time.Now()

	defer func() {
		observeQuery("find_by_id", start, err)
	}()
	ctx, cancel := r.withQueryTimeout(ctx)
	defer cancel()
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
WHERE id=$1;
`

	community, err = scanCommunity(
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
) (err error) {
	start := time.Now()

	defer func() {
		observeQuery("create", start, err)
	}()

	if community == nil {
		return errors.New("community is nil")
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	ctx, cancel := r.withQueryTimeout(ctx)
	defer cancel()

	const query = `
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

	err = r.Pool().
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
) (err error) {
	start := time.Now()

	defer func() {
		observeQuery("update", start, err)
	}()

	if community == nil {
		return errors.New("community is nil")
	}

	if community.ID <= 0 {
		return errors.New("community ID must be greater than zero")
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	ctx, cancel := r.withQueryTimeout(ctx)
	defer cancel()

	const query = `
UPDATE communities
SET
	name = $1,
	slug = $2,
	description = $3,
	external_source = $4,
	updated_at = NOW()
WHERE id = $5
RETURNING updated_at;
`

	err = r.Pool().
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

// DeleteAll removes every community and resets the identity sequence.
//
// This method is primarily intended for integration-test cleanup.
func (r *CommunityRepository) DeleteAll(
	ctx context.Context,
) (err error) {

	start := time.Now()

	defer func() {
		observeQuery("delete_all", start, err)
	}()
	ctx, cancel := r.withQueryTimeout(ctx)
	defer cancel()
	const query = `
TRUNCATE TABLE communities
RESTART IDENTITY
`

	_, err = r.pool.Exec(ctx, query)
	if err != nil {
		return translateError(err)
	}

	return nil
}

func (r *CommunityRepository) FindBySlug(
	ctx context.Context,
	slug string,
) (community *model.Community, err error) {

	start := time.Now()

	defer func() {
		observeQuery("find_by_slug", start, err)
	}()
	ctx, cancel := r.withQueryTimeout(ctx)
	defer cancel()
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
WHERE slug=$1;
`

	community, err = scanCommunity(
		r.Pool().QueryRow(ctx, query, slug),
	)

	if err != nil {
		return nil, translateError(err)
	}

	return community, nil
}

func (r *CommunityRepository) Delete(
	ctx context.Context,
	id int64,
) (err error) {
	start := time.Now()

	defer func() {
		observeQuery("delete", start, err)
	}()

	if id <= 0 {
		return errors.New("community ID must be greater than zero")
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	ctx, cancel := r.withQueryTimeout(ctx)
	defer cancel()

	const query = `
DELETE FROM communities
WHERE id = $1;
`

	result, err := r.Pool().Exec(
		ctx,
		query,
		id,
	)
	if err != nil {
		return translateError(err)
	}

	if result.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}
