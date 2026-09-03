package repository

import (
	"context"

	"github.com/amyismebyme/the-village/apps/api/internal/external"
)

type ExternalItemRepository interface {
	Upsert(
		ctx context.Context,
		item external.Item,
	) error
    	UpsertBatch(
    		ctx context.Context,
    		items []external.Item,
    	) error
	FindByIdentity(
		ctx context.Context,
		identity external.Identity,
	) (*external.Item, error)

	DeleteAll(
		ctx context.Context,
	) error
}
