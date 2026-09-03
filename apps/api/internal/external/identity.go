package external

import (
	"errors"
	"fmt"
	"strings"
)

// Identity uniquely identifies an item from an external provider.
//
// An external identity is the combination of Source and ExternalID.
// The same ExternalID from two different sources is therefore valid.
type Identity struct {
	Source     Source
	ExternalID string
}

// Normalize returns the canonical representation of an identity.
func (i Identity) Normalize() Identity {
	return Identity{
		Source:     Source(strings.TrimSpace(string(i.Source))),
		ExternalID: strings.TrimSpace(i.ExternalID),
	}
}

func (i Identity) Validate() error {
	normalized := i.Normalize()

	if normalized.Source == "" {
		return ErrInvalidConfig
	}

	if normalized.ExternalID == "" {
		return ErrInvalidConfig
	}

	return nil
}

// Key returns the deterministic provider-neutral identity key.
func (i Identity) Key() string {
	normalized := i.Normalize()

	return fmt.Sprintf(
		"%s:%s",
		normalized.Source,
		normalized.ExternalID,
	)
}

func SameIdentity(
	left Identity,
	right Identity,
) bool {
	left = left.Normalize()
	right = right.Normalize()

	return left.Source == right.Source &&
		left.ExternalID == right.ExternalID
}

// ValidateUniqueIdentities performs strict duplicate detection.
//
// This is useful when a caller wants duplicates to be treated as an error.
// Normal ingestion uses DeduplicateItems when duplicates should be skipped.
func ValidateUniqueIdentities(
	identities []Identity,
) error {
	seen := make(map[string]struct{}, len(identities))

	for _, identity := range identities {
		if err := identity.Validate(); err != nil {
			return err
		}

		key := identity.Key()

		if _, exists := seen[key]; exists {
			return errors.New(
				"duplicate external identity",
			)
		}

		seen[key] = struct{}{}
	}

	return nil
}
