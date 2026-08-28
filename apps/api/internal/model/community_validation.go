package model

import "github.com/amyismebyme/the-village/apps/api/internal/validation"

// Validate validates the Community domain model.
func (c Community) Validate() error {
	if err := validation.Required(c.Name); err != nil {
		return validation.NewFieldError(
			"name",
			err,
		)
	}

	if err := validation.Length(c.Name, 3, 100); err != nil {
		return validation.NewFieldError(
			"name",
			err,
		)
	}

	if err := validation.Required(c.Slug); err != nil {
		return validation.NewFieldError(
			"slug",
			err,
		)
	}

	if err := validation.Slug(c.Slug); err != nil {
		return validation.NewFieldError(
			"slug",
			err,
		)
	}

	if err := validation.MaxLength(
		c.Description,
		2000,
	); err != nil {
		return validation.NewFieldError(
			"description",
			err,
		)
	}

	return nil
}
