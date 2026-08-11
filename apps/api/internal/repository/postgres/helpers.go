package postgres

import (
	"context"
	"github.com/amyismebyme/the-village/apps/api/internal/model"
	"github.com/amyismebyme/the-village/apps/api/internal/repository"
	"github.com/jackc/pgx/v5"
)

func scanCommunity(
	row pgx.Row,
) (*model.Community, error) {

	community := &model.Community{}

	err := row.Scan(
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

	return community, nil
}

func execOne(
	ctx context.Context,
	r Repository,
	query string,
	args ...any,
) error {

	result, err := r.Pool().Exec(
		ctx,
		query,
		args...,
	)

	if err != nil {
		return translateError(err)
	}

	if result.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}
