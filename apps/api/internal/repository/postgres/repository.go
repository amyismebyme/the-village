package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultQueryTimeout = 30 * time.Second

type Repository struct {
	pool         *pgxpool.Pool
	queryTimeout time.Duration
}

func New(
	pool *pgxpool.Pool,
	queryTimeout ...time.Duration,
) Repository {
	timeout := defaultQueryTimeout

	if len(queryTimeout) > 0 &&
		queryTimeout[0] > 0 {
		timeout = queryTimeout[0]
	}

	return Repository{
		pool:         pool,
		queryTimeout: timeout,
	}
}

func (r Repository) Pool() *pgxpool.Pool {
	return r.pool
}

func (r Repository) Ping(
	ctx context.Context,
) error {
	return r.pool.Ping(ctx)
}