package postgres

import (
	"github.com/amyismebyme/the-village/apps/api/internal/model"

	"github.com/jackc/pgx/v5"
)

const communityColumns = `
	id,
	name,
	slug,
	description,
	external_source,
	created_at,
	updated_at
`

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

func scanCommunities(
	rows pgx.Rows,
) ([]*model.Community, error) {
	defer rows.Close()

	communities := make(
		[]*model.Community,
		0,
	)

	for rows.Next() {
		community, err := scanCommunity(rows)
		if err != nil {
			return nil, translateError(err)
		}

		communities = append(
			communities,
			community,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, translateError(err)
	}

	return communities, nil
}
