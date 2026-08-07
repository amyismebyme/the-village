package postgres

import (
	"errors"
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/repository"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
)

func TestTranslateNoRows(t *testing.T) {

	err := translateError(pgx.ErrNoRows)

	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatal("expected ErrNotFound")
	}
}

func TestTranslateDuplicateKey(t *testing.T) {

	err := translateError(&pgconn.PgError{
		Code: "23505",
	})

	if !errors.Is(err, repository.ErrAlreadyExists) {
		t.Fatal("expected ErrAlreadyExists")
	}
}

func TestTranslateForeignKey(t *testing.T) {

	err := translateError(&pgconn.PgError{
		Code: "23503",
	})

	if !errors.Is(err, repository.ErrConflict) {
		t.Fatal("expected ErrConflict")
	}
}
