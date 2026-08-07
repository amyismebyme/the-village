package postgres

import (
	"errors"

	"github.com/amyismebyme/the-village/apps/api/internal/repository"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
)

func translateError(err error) error {

	if err == nil {
		return nil
	}

	//----------------------------------------
	// No rows
	//----------------------------------------

	if errors.Is(err, pgx.ErrNoRows) {
		return repository.ErrNotFound
	}

	//----------------------------------------
	// PostgreSQL errors
	//----------------------------------------

	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {

		switch pgErr.Code {

		// unique_violation
		case "23505":
			return repository.ErrAlreadyExists

		// foreign_key_violation
		case "23503":
			return repository.ErrConflict

		// check_violation
		case "23514":
			return repository.ErrInvalidInput
		}
	}

	return err
}
