package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) Repository {
	return Repository{
		pool: pool,
	}
}

func (r Repository) Pool() *pgxpool.Pool {
	return r.pool
}

func (r Repository) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}
