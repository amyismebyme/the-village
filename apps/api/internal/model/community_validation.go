package model

import (
	"fmt"

	"github.com/amyismebyme/the-village/apps/api/internal/validation"
)

// Validate validates the Community domain model.
func (c Community) Validate() error {

	if err := validation.Required(c.Name); err != nil {
		return fmt.Errorf("name: %w", err)
	}

	if err := validation.Length(c.Name, 3, 100); err != nil {
		return fmt.Errorf("name: %w", err)
	}

	if err := validation.Required(c.Slug); err != nil {
		return fmt.Errorf("slug: %w", err)
	}

	if err := validation.Slug(c.Slug); err != nil {
		return fmt.Errorf("slug: %w", err)
	}

	if err := validation.MaxLength(c.Description, 2000); err != nil {
		return fmt.Errorf("description: %w", err)
	}

	return nil
}
