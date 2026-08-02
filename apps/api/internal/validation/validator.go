package validation

import (
	"errors"
	"regexp"
	"strings"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func Required(value string) error {

	if strings.TrimSpace(value) == "" {
		return errors.New("value is required")
	}

	return nil
}

func Length(
	value string,
	min int,
	max int,
) error {

	length := len(strings.TrimSpace(value))

	if length < min {
		return errors.New("value is too short")
	}

	if length > max {
		return errors.New("value is too long")
	}

	return nil
}

func MaxLength(
	value string,
	max int,
) error {

	length := len(strings.TrimSpace(value))

	if length > max {
		return errors.New("value is too long")
	}

	return nil
}

func Slug(value string) error {

	if value != strings.ToLower(value) {
		return errors.New("slug must be lowercase")
	}

	if !slugPattern.MatchString(value) {
		return errors.New("invalid slug")
	}

	return nil
}
