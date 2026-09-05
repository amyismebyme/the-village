package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/amyismebyme/the-village/apps/api/internal/model"
	"github.com/amyismebyme/the-village/apps/api/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

var _ repository.ResourceRepository = (*ResourceRepository)(nil)

type ResourceRepository struct {
	Repository
}

func NewResourceRepository(
	pool *pgxpool.Pool,
	queryTimeout ...time.Duration,
) *ResourceRepository {
	return &ResourceRepository{
		Repository: New(
			pool,
			queryTimeout...,
		),
	}
}

func (r *ResourceRepository) List(
	ctx context.Context,
) (resources []model.Resource, err error) {
	start := time.Now()

	defer func() {
		observeQuery(
			"resource_list",
			start,
			err,
		)
	}()

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	ctx, cancel := r.withQueryTimeout(ctx)
	defer cancel()

	const query = `
SELECT
	id,
	title,
	description,
	url,
	category,
	created_at,
	updated_at
FROM resources
ORDER BY id;
`

	rows, err := r.Pool().Query(
		ctx,
		query,
	)
	if err != nil {
		return nil, translateError(err)
	}
	defer rows.Close()

	for rows.Next() {
		var resource model.Resource

		if err := rows.Scan(
			&resource.ID,
			&resource.Title,
			&resource.Description,
			&resource.URL,
			&resource.Category,
			&resource.CreatedAt,
			&resource.UpdatedAt,
		); err != nil {
			return nil, translateError(err)
		}

		resources = append(
			resources,
			resource,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, translateError(err)
	}

	if resources == nil {
		resources = []model.Resource{}
	}

	return resources, nil
}

func (r *ResourceRepository) FindByID(
	ctx context.Context,
	id int64,
) (resource *model.Resource, err error) {
	start := time.Now()

	defer func() {
		observeQuery(
			"resource_find_by_id",
			start,
			err,
		)
	}()

	if id <= 0 {
		return nil, fmt.Errorf(
			"%w: resource ID must be greater than zero",
			repository.ErrInvalidInput,
		)
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	ctx, cancel := r.withQueryTimeout(ctx)
	defer cancel()

	const query = `
SELECT
	id,
	title,
	description,
	url,
	category,
	created_at,
	updated_at
FROM resources
WHERE id = $1;
`

	resource = &model.Resource{}

	err = r.Pool().
		QueryRow(
			ctx,
			query,
			id,
		).
		Scan(
			&resource.ID,
			&resource.Title,
			&resource.Description,
			&resource.URL,
			&resource.Category,
			&resource.CreatedAt,
			&resource.UpdatedAt,
		)

	if err != nil {
		return nil, translateError(err)
	}

	return resource, nil
}

func (r *ResourceRepository) Create(
	ctx context.Context,
	resource *model.Resource,
) (err error) {
	start := time.Now()

	defer func() {
		observeQuery(
			"resource_create",
			start,
			err,
		)
	}()

	if resource == nil {
		return repository.ErrInvalidInput
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	ctx, cancel := r.withQueryTimeout(ctx)
	defer cancel()

	const query = `
INSERT INTO resources
(
	title,
	description,
	url,
	category
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
			resource.Title,
			resource.Description,
			resource.URL,
			resource.Category,
		).
		Scan(
			&resource.ID,
			&resource.CreatedAt,
			&resource.UpdatedAt,
		)

	if err != nil {
		return translateError(err)
	}

	return nil
}

func (r *ResourceRepository) Update(
	ctx context.Context,
	resource *model.Resource,
) (err error) {
	start := time.Now()

	defer func() {
		observeQuery(
			"resource_update",
			start,
			err,
		)
	}()

	if resource == nil {
		return repository.ErrInvalidInput
	}

	if resource.ID <= 0 {
		return fmt.Errorf(
			"%w: resource ID must be greater than zero",
			repository.ErrInvalidInput,
		)
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	ctx, cancel := r.withQueryTimeout(ctx)
	defer cancel()

	const query = `
UPDATE resources
SET
	title = $1,
	description = $2,
	url = $3,
	category = $4,
	updated_at = NOW()
WHERE id = $5
RETURNING updated_at;
`

	err = r.Pool().
		QueryRow(
			ctx,
			query,
			resource.Title,
			resource.Description,
			resource.URL,
			resource.Category,
			resource.ID,
		).
		Scan(&resource.UpdatedAt)

	if err != nil {
		return translateError(err)
	}

	return nil
}

func (r *ResourceRepository) Delete(
	ctx context.Context,
	id int64,
) (err error) {
	start := time.Now()

	defer func() {
		observeQuery(
			"resource_delete",
			start,
			err,
		)
	}()

	if id <= 0 {
		return fmt.Errorf(
			"%w: resource ID must be greater than zero",
			repository.ErrInvalidInput,
		)
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	ctx, cancel := r.withQueryTimeout(ctx)
	defer cancel()

	const query = `
DELETE FROM resources
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

func (r *ResourceRepository) DeleteAll(
	ctx context.Context,
) (err error) {
	start := time.Now()

	defer func() {
		observeQuery(
			"resource_delete_all",
			start,
			err,
		)
	}()

	if err := ctx.Err(); err != nil {
		return err
	}

	ctx, cancel := r.withQueryTimeout(ctx)
	defer cancel()

	const query = `
TRUNCATE TABLE resources
RESTART IDENTITY
`

	if _, err := r.Pool().Exec(
		ctx,
		query,
	); err != nil {
		return translateError(err)
	}

	return nil
}
