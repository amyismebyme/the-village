package external

import (
	"errors"
	"fmt"
	"strings"
)

// Identity uniquely identifies an item from an external provider.
type Identity struct {
	Source     Source
	ExternalID string
}

func (i Identity) Validate() error {
	if strings.TrimSpace(string(i.Source)) == "" {
		return ErrInvalidConfig
	}

	if strings.TrimSpace(i.ExternalID) == "" {
		return ErrInvalidConfig
	}

	return nil
}

func (i Identity) Key() string {
	return fmt.Sprintf(
		"%s:%s",
		i.Source,
		i.ExternalID,
	)
}

func SameIdentity(
	left Identity,
	right Identity,
) bool {
	return left.Source == right.Source &&
		left.ExternalID == right.ExternalID
}

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
