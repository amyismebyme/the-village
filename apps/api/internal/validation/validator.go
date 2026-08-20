package validation

import (
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"
)

var slugPattern = regexp.MustCompile(
	`^[a-z0-9]+(?:-[a-z0-9]+)*$`,
)

var (
	ErrRequired      = errors.New("value is required")
	ErrTooShort      = errors.New("value is too short")
	ErrTooLong       = errors.New("value is too long")
	ErrInvalidSlug   = errors.New("invalid slug")
	ErrInvalidLength = errors.New("invalid length bounds")
)

func Required(value string) error {
	if strings.TrimSpace(value) == "" {
		return ErrRequired
	}

	return nil
}

func Length(
	value string,
	min int,
	max int,
) error {
	if min < 0 || max < min {
		return ErrInvalidLength
	}

	value = strings.TrimSpace(value)

	length := utf8.RuneCountInString(value)

	if length < min {
		return ErrTooShort
	}

	if length > max {
		return ErrTooLong
	}

	return nil
}

func MaxLength(
	value string,
	max int,
) error {
	if max < 0 {
		return ErrInvalidLength
	}

	value = strings.TrimSpace(value)

	length := utf8.RuneCountInString(value)

	if length > max {
		return ErrTooLong
	}

	return nil
}

func Slug(value string) error {
	if value == "" {
		return ErrInvalidSlug
	}

	if value != strings.ToLower(value) {
		return ErrInvalidSlug
	}

	if !slugPattern.MatchString(value) {
		return ErrInvalidSlug
	}

	return nil
}
