package postgres

import (
	"github.com/amyismebyme/the-village/apps/api/internal/model"
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
