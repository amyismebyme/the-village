package external

import (
	"context"
	"fmt"
)

// DeduplicateItems removes duplicate external items from a batch.
//
// Idempotency is defined by Item.Identity(), which is the pair:
//
//	(Source, ExternalID)
//
// The first valid occurrence wins. Later occurrences with the same
// identity are discarded.
//
// Items from different sources remain distinct even if their ExternalID
// values are identical.
func DeduplicateItems(
	ctx context.Context,
	items []Item,
) ([]Item, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	unique := make(
		[]Item,
		0,
		len(items),
	)

	seen := make(
		map[string]struct{},
		len(items),
	)

	for index, item := range items {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		if err := item.Validate(); err != nil {
			return nil, fmt.Errorf(
				"validate item %d: %w",
				index,
				err,
			)
		}

		identity := item.Identity()
		key := identity.Key()

		if _, exists := seen[key]; exists {
			// First occurrence wins.
			continue
		}

		seen[key] = struct{}{}

		unique = append(
			unique,
			item,
		)
	}

	return unique, nil
}
