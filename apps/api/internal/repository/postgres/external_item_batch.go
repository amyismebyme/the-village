package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/amyismebyme/the-village/apps/api/internal/external"
	"github.com/amyismebyme/the-village/apps/api/internal/repository"

	"github.com/jackc/pgx/v5"
)

func (r *ExternalItemRepository) UpsertBatch(
	ctx context.Context,
	items []external.Item,
) (err error) {
	if err := ctx.Err(); err != nil {
		return err
	}

	if len(items) == 0 {
		return nil
	}

	// Validate the complete batch before opening the transaction.
	// This guarantees that an invalid item cannot leave a partially
	// persisted batch.
	for index, item := range items {
		if err := item.Validate(); err != nil {
			return fmt.Errorf(
				"%w: validate external item %d: %w",
				repository.ErrInvalidInput,
				index,
				err,
			)
		}
	}

	ctx, cancel := r.withQueryTimeout(ctx)
	defer cancel()

	start := time.Now()

	defer func() {
		observeQuery(
			"external_item_upsert_batch",
			start,
			err,
		)
	}()

	tx, err := r.Pool().BeginTx(
		ctx,
		pgx.TxOptions{},
	)
	if err != nil {
		return translateError(err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

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
	OR external_items.url IS DISTINCT FROM EXCLUDED.url;
`

	for _, item := range items {
		if err := ctx.Err(); err != nil {
			return err
		}

		if _, execErr := tx.Exec(
			ctx,
			query,
			item.Source,
			item.ExternalID,
			item.Title,
			item.Description,
			item.URL,
		); execErr != nil {
			return translateError(execErr)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return translateError(err)
	}

	return nil
}
