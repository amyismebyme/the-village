package postgres

import (
	"github.com/amyismebyme/the-village/apps/api/internal/model"

	"github.com/jackc/pgx/v5"
)

const resourceColumns = `
	id,
	title,
	description,
	url,
	category,
	created_at,
	updated_at
`

func scanResource(
	row pgx.Row,
) (*model.Resource, error) {
	resource := &model.Resource{}

	if err := row.Scan(
		&resource.ID,
		&resource.Title,
		&resource.Description,
		&resource.URL,
		&resource.Category,
		&resource.CreatedAt,
		&resource.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return resource, nil
}

func scanResources(
	rows pgx.Rows,
) ([]model.Resource, error) {
	defer rows.Close()

	resources := make(
		[]model.Resource,
		0,
	)

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

	return resources, nil
}
