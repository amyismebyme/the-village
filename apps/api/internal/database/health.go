package database

import (
	"context"
	"time"
)

func (db *Database) Health(ctx context.Context) error {

	if db.pool == nil {
		return ErrDatabaseNotInitialized
	}

	return db.pool.Ping(ctx)
}

func (db *Database) PingLatency(ctx context.Context) (time.Duration, error) {

	start := time.Now()

	err := db.Health(ctx)

	return time.Since(start), err
}
