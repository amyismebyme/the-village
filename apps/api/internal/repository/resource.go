package repository

import (
	"context"

	"github.com/amyismebyme/the-village/apps/api/internal/model"
)

type ResourceRepository interface {
	List(ctx context.Context) ([]model.Resource, error)

	FindByID(
		ctx context.Context,
		id int64,
	) (*model.Resource, error)

	Create(
		ctx context.Context,
		resource *model.Resource,
	) error

	Update(
		ctx context.Context,
		resource *model.Resource,
	) error

	Delete(
		ctx context.Context,
		id int64,
	) error
}
