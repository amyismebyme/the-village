package repository

import (
	"context"
	"github.com/amyismebyme/the-village/apps/api/internal/model"
)

type Repository interface {
	Ping(ctx context.Context) error

	CreateCommunity(
		ctx context.Context,
		community model.Community,
	) error

	ListCommunities(
		ctx context.Context,
	) ([]model.Community, error)
}
