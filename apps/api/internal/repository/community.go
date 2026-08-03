package repository

import (
	"context"

	"github.com/amyismebyme/the-village/apps/api/internal/model"
)

type CommunityRepository interface {
	Create(
		ctx context.Context,
		community *model.Community,
	) error

	FindByID(
		ctx context.Context,
		id int64,
	) (*model.Community, error)

	FindBySlug(
		ctx context.Context,
		slug string,
	) (*model.Community, error)

	List(
		ctx context.Context,
	) ([]*model.Community, error)

	Update(
		ctx context.Context,
		community *model.Community,
	) error

	Delete(
		ctx context.Context,
		id int64,
	) error

	DeleteAll(
		ctx context.Context,
	) error
}
