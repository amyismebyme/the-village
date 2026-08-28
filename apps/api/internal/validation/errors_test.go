package validation

import (
	"errors"
	"testing"
)

func TestFieldErrorPreservesCause(t *testing.T) {
	err := NewFieldError(
		"name",
		ErrTooShort,
	)

	if !errors.Is(
		err,
		ErrTooShort,
	) {
		t.Fatal(
			"expected original validation error to be preserved",
		)
	}
}

func TestFieldErrorExposesField(t *testing.T) {
	err := NewFieldError(
		"slug",
		ErrInvalidSlug,
	)

	var fieldErr FieldError

	if !errors.As(err, &fieldErr) {
		t.Fatal(
			"expected FieldError",
		)
	}

	if fieldErr.Field != "slug" {
		t.Fatalf(
			"expected field slug, got %q",
			fieldErr.Field,
		)
	}
}

func TestFieldErrorMessageRemainsUseful(t *testing.T) {
	err := NewFieldError(
		"name",
		ErrTooShort,
	)

	if err.Error() != "name: value is too short" {
		t.Fatalf(
			"unexpected error message %q",
			err.Error(),
		)
	}
}
