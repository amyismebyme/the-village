package external

import (
	"context"
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
	if i.Source == "" {
		return ErrInvalidPayload
	}

	if i.ExternalID == "" {
		return ErrInvalidPayload
	}

	return nil
}

type Normalizer[T any] interface {
	Normalize(
		ctx context.Context,
		input T,
	) (Item, error)
}
