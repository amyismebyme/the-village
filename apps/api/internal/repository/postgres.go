package repository

import (
	"context"
	"errors"

	"github.com/amyismebyme/the-village/apps/api/internal/model"
)

type PostgresRepository struct{}

func NewPostgresRepository() *PostgresRepository {
	return &PostgresRepository{}
}

func (p *PostgresRepository) Ping(ctx context.Context) error {
	return nil
}

func (p *PostgresRepository) CreateCommunity(
	ctx context.Context,
	community model.Community,
) error {
	return errors.New("not implemented")
}

func (p *PostgresRepository) ListCommunities(
	ctx context.Context,
) ([]model.Community, error) {
	return nil, errors.New("not implemented")
}
