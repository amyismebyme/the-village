package reddit

import (
	"context"
	"fmt"
	"strings"

	"github.com/amyismebyme/the-village/apps/api/internal/external"
)

type IngestionService struct {
	client     *Client
	normalizer external.Normalizer[RedditPost]
}

func NewIngestionService(
	client *Client,
	normalizer external.Normalizer[RedditPost],
) *IngestionService {
	if client == nil {
		panic(
			"reddit ingestion service: client is required",
		)
	}

	if normalizer == nil {
		panic(
			"reddit ingestion service: normalizer is required",
		)
	}

	return &IngestionService{
		client:     client,
		normalizer: normalizer,
	}
}

func (s *IngestionService) IngestListing(
	ctx context.Context,
	accessToken string,
	subreddit string,
	limit int,
	after string,
) ([]external.Item, error) {
	if strings.TrimSpace(subreddit) == "" {
		return nil, fmt.Errorf(
			"%w: subreddit is required",
			external.ErrInvalidConfig,
		)
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	listing, err := s.client.FetchListing(
		ctx,
		accessToken,
		subreddit,
		limit,
		after,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"reddit ingestion: fetch listing: %w",
			err,
		)
	}

	items := make(
		[]external.Item,
		0,
		len(listing.Data.Children),
	)

	for _, child := range listing.Data.Children {
		if child.Kind != "t3" {
			continue
		}

		item, err := s.normalizer.Normalize(
			ctx,
			child.Data,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"reddit ingestion: normalize %q: %w",
				child.Data.ID,
				err,
			)
		}

		if err := item.Validate(); err != nil {
			return nil, fmt.Errorf(
				"reddit ingestion: validate normalized item %q: %w",
				child.Data.ID,
				err,
			)
		}

		identity := item.Identity()

		if err := identity.Validate(); err != nil {
			return nil, fmt.Errorf(
				"reddit ingestion: validate identity %q: %w",
				child.Data.ID,
				err,
			)
		}

		items = append(
			items,
			item,
		)
	}

	return items, nil
}
