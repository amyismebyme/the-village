package postgres

import (
	"context"
	"errors"

	"github.com/amyismebyme/the-village/apps/api/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func translateError(err error) error {
	if err == nil {
		return nil
	}

	// Preserve context semantics.
	if errors.Is(err, context.Canceled) {
		return context.Canceled
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return context.DeadlineExceeded
	}

	// pgx represents PostgreSQL errors as *pgconn.PgError.
	//
	// Use errors.As rather than a direct type assertion because the
	// database error may be wrapped by another layer.
	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return repository.ErrAlreadyExists

		case "23503":
			return repository.ErrConflict
		}
	}

	// pgx.ErrNoRows is the canonical "record not found" result for
	// QueryRow(...).Scan(...).
	if errors.Is(err, pgx.ErrNoRows) {
		return repository.ErrNotFound
	}

	return err
}
