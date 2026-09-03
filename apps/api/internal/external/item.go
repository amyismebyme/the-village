package external

import (
	"context"
	"fmt"
)

type Item struct {
	Source      Source
	ExternalID  string
	Title       string
	Description string
	URL         string
}

func (i Item) Identity() Identity {
	return Identity{
		Source:     i.Source,
		ExternalID: i.ExternalID,
	}
}

func (i Item) Validate() error {
	identity := i.Identity()

	// Preserve Item-level payload semantics.
	if identity.Source == "" ||
		identity.ExternalID == "" {
		return ErrInvalidPayload
	}

	if err := identity.Validate(); err != nil {
		return fmt.Errorf(
			"validate identity: %w",
			err,
		)
	}

	return nil
}

type Normalizer[T any] interface {
	Normalize(
		ctx context.Context,
		input T,
	) (Item, error)
}
