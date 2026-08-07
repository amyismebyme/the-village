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

	if errors.Is(err, pgx.ErrNoRows) {
		return repository.ErrNotFound
	}

	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {

		switch pgErr.Code {

		// unique violation
		case "23505":
			return repository.ErrAlreadyExists
		}
	}

	return err
}
