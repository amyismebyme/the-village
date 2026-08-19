package postgres

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

const defaultQueryTimeout = 30 * time.Second

type Repository struct {
	pool         *pgxpool.Pool
	queryTimeout time.Duration
}

func New(pool *pgxpool.Pool) Repository {
	return Repository{
		pool:         pool,
		queryTimeout: defaultQueryTimeout,
	}
}

func (r Repository) Pool() *pgxpool.Pool {
	return r.pool
}

func (r Repository) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}
