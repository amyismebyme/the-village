package integration

import (
	"context"
	"fmt"
	"sync"

	"github.com/amyismebyme/the-village/apps/api/internal/external"
	"github.com/amyismebyme/the-village/apps/api/internal/repository"
)

var _ repository.ExternalItemRepository = (*integrationExternalItemRepository)(nil)

type integrationExternalItemRepository struct {
	mu    sync.Mutex
	items map[string]external.Item
}

func (r *integrationExternalItemRepository) Upsert(
	ctx context.Context,
	item external.Item,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if err := item.Validate(); err != nil {
		return fmt.Errorf(
			"%w: validate external item: %w",
			repository.ErrInvalidInput,
			err,
		)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.items == nil {
		r.items = make(map[string]external.Item)
	}

	r.items[item.Identity().Key()] = item

	return nil
}

func (r *integrationExternalItemRepository) UpsertBatch(
	ctx context.Context,
	items []external.Item,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	validated := make(
		[]external.Item,
		0,
		len(items),
	)

	for index, item := range items {
		if err := ctx.Err(); err != nil {
			return err
		}

		if err := item.Validate(); err != nil {
			return fmt.Errorf(
				"%w: validate external item %d: %w",
				repository.ErrInvalidInput,
				index,
				err,
			)
		}

		validated = append(
			validated,
			item,
		)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.items == nil {
		r.items = make(map[string]external.Item)
	}

	for _, item := range validated {
		r.items[item.Identity().Key()] = item
	}

	return nil
}

func (r *integrationExternalItemRepository) FindByIdentity(
	ctx context.Context,
	identity external.Identity,
) (*external.Item, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if err := identity.Validate(); err != nil {
		return nil, fmt.Errorf(
			"%w: validate identity: %w",
			repository.ErrInvalidInput,
			err,
		)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	item, ok := r.items[identity.Key()]
	if !ok {
		return nil, repository.ErrNotFound
	}

	copy := item

	return &copy, nil
}

func (r *integrationExternalItemRepository) DeleteAll(
	ctx context.Context,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	clear(r.items)

	return nil
}

func (r *integrationExternalItemRepository) Count() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	return len(r.items)
}
