package postgres

import (
	"github.com/amyismebyme/the-village/apps/api/internal/model"

	"github.com/jackc/pgx/v5"
)

func scanCommunities(
	rows pgx.Rows,
) ([]*model.Community, error) {

	defer rows.Close()

	var communities []*model.Community

	for rows.Next() {

		var community model.Community

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
			return nil, translateError(err)
		}

		communities = append(
			communities,
			&community,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, translateError(err)
	}

	return communities, nil
}
