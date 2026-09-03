package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/amyismebyme/the-village/apps/api/internal/external"
	"github.com/amyismebyme/the-village/apps/api/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ repository.ExternalItemRepository = (*ExternalItemRepository)(nil)

type ExternalItemRepository struct {
	Repository
}

func NewExternalItemRepository(
	pool *pgxpool.Pool,
	queryTimeout ...time.Duration,
) *ExternalItemRepository {
	return &ExternalItemRepository{
		Repository: New(
			pool,
			queryTimeout...,
		),
	}
}

func (r *ExternalItemRepository) Upsert(
	ctx context.Context,
	item external.Item,
) (err error) {
	start := time.Now()

	defer func() {
		observeQuery(
			"external_item_upsert",
			start,
			err,
		)
	}()

	if err := ctx.Err(); err != nil {
		return err
	}

	if err := item.Validate(); err != nil {
		return fmt.Errorf(
			"%w: validate external item: %w",
			repository.ErrInvalidInput,
			err,
		)
	}

	ctx, cancel := r.withQueryTimeout(ctx)
	defer cancel()

	const query = `
INSERT INTO external_items
(
	source,
	external_id,
	title,
	description,
	url
)
VALUES
(
	$1,
	$2,
	$3,
	$4,
	$5
)
ON CONFLICT
(
	source,
	external_id
)
DO UPDATE SET
	title = EXCLUDED.title,
	description = EXCLUDED.description,
	url = EXCLUDED.url,
	updated_at = NOW()
WHERE
	external_items.title IS DISTINCT FROM EXCLUDED.title
	OR external_items.description IS DISTINCT FROM EXCLUDED.description
	OR external_items.url IS DISTINCT FROM EXCLUDED.url
RETURNING id;
`

	var id int64

	err = r.Pool().
		QueryRow(
			ctx,
			query,
			item.Source,
			item.ExternalID,
			item.Title,
			item.Description,
			item.URL,
		).
		Scan(&id)

	// An existing identical row produces no UPDATE because of the
	// IS DISTINCT FROM predicate. That is a successful replay.
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}

	if err != nil {
		return translateError(err)
	}

	return nil
}

func (r *ExternalItemRepository) FindByIdentity(
	ctx context.Context,
	identity external.Identity,
) (item *external.Item, err error) {
	start := time.Now()

	defer func() {
		observeQuery(
			"external_item_find_by_identity",
			start,
			err,
		)
	}()

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if err := identity.Validate(); err != nil {
		return nil, fmt.Errorf(
			"%w: validate identity: %w",
			repository.ErrInvalidInput,
			err,
		)
	}

	ctx, cancel := r.withQueryTimeout(ctx)
	defer cancel()

	const query = `
SELECT
	source,
	external_id,
	title,
	description,
	url
FROM external_items
WHERE source = $1
  AND external_id = $2;
`

	var source external.Source
	var externalID string
	var title string
	var description string
	var url string

	err = r.Pool().
		QueryRow(
			ctx,
			query,
			identity.Source,
			identity.ExternalID,
		).
		Scan(
			&source,
			&externalID,
			&title,
			&description,
			&url,
		)

	if err != nil {
		return nil, translateError(err)
	}

	return &external.Item{
		Source:      source,
		ExternalID:  externalID,
		Title:       title,
		Description: description,
		URL:         url,
	}, nil
}

func (r *ExternalItemRepository) DeleteAll(
	ctx context.Context,
) (err error) {
	start := time.Now()

	defer func() {
		observeQuery(
			"external_item_delete_all",
			start,
			err,
		)
	}()

	ctx, cancel := r.withQueryTimeout(ctx)
	defer cancel()

	const query = `
TRUNCATE TABLE external_items
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
