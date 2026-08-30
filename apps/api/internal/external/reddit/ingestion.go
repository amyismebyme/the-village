package reddit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/amyismebyme/the-village/apps/api/internal/cache"
	"github.com/amyismebyme/the-village/apps/api/internal/external"
)

type CacheMode int

const (
	CacheModeDefault CacheMode = iota
	CacheModeBypass
	CacheModeRefresh
)

type IngestionOptions struct {
	CacheMode CacheMode
}

type IngestionService struct {
	client     *Client
	normalizer external.Normalizer[RedditPost]

	cache    cache.Cache
	cacheTTL time.Duration
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

// SetCache enables Reddit listing caching.
//
// The cache is optional. When no cache is configured, ingestion always
// fetches from Reddit.
func (s *IngestionService) SetCache(
	c cache.Cache,
	ttl time.Duration,
) error {
	if c == nil {
		return errors.New(
			"reddit ingestion: cache is required",
		)
	}

	if ttl <= 0 {
		return errors.New(
			"reddit ingestion: cache TTL must be greater than zero",
		)
	}

	s.cache = c
	s.cacheTTL = ttl

	return nil
}

// IngestListing uses the normal cache behavior.
func (s *IngestionService) IngestListing(
	ctx context.Context,
	accessToken string,
	subreddit string,
	limit int,
	after string,
) ([]external.Item, error) {
	return s.IngestListingWithOptions(
		ctx,
		accessToken,
		subreddit,
		limit,
		after,
		IngestionOptions{
			CacheMode: CacheModeDefault,
		},
	)
}

// IngestListingWithOptions fetches and normalizes a Reddit listing.
//
// CacheModeDefault:
//   - read cache
//   - fetch on miss
//   - populate cache after a successful fetch
//
// CacheModeBypass:
//   - skip cache read
//   - fetch from Reddit
//   - do not change the cache
//
// CacheModeRefresh:
//   - skip cache read
//   - fetch from Reddit
//   - replace the cache after a successful fetch
func (s *IngestionService) IngestListingWithOptions(
	ctx context.Context,
	accessToken string,
	subreddit string,
	limit int,
	after string,
	options IngestionOptions,
) ([]external.Item, error) {
	if err := validateSubreddit(
		subreddit,
	); err != nil {
		return nil, fmt.Errorf(
			"%w: subreddit is required",
			err,
		)
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = 25
	}

	switch options.CacheMode {
	case CacheModeDefault,
		CacheModeBypass,
		CacheModeRefresh:
		// valid

	default:
		return nil, errors.New(
			"reddit ingestion: invalid cache mode",
		)
	}

	useCache := s.cache != nil
	readCache := useCache &&
		options.CacheMode == CacheModeDefault
	writeCache := useCache &&
		options.CacheMode != CacheModeBypass

	var cacheKey string

	if useCache {
		var err error

		cacheKey, err = cache.RedditListingKey(
			subreddit,
			after,
			limit,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"reddit ingestion: create cache key: %w",
				err,
			)
		}
	}

	// ---------------------------------------------------------------
	// Cache lookup
	// ---------------------------------------------------------------

	if readCache {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		cached, ok, err := s.cache.Get(
			ctx,
			cacheKey,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"reddit ingestion: cache get: %w",
				err,
			)
		}

		if ok {
			var listing ListingResponse

			if err := DecodeCachedJSON(
				cached,
				&listing,
			); err == nil {
				return s.normalizeListing(
					ctx,
					listing,
				)
			}

			// Corrupt cache data is treated as a cache miss.
			// Remove it before fetching a fresh response.
			if err := ctx.Err(); err != nil {
				return nil, err
			}

			if err := s.cache.Delete(
				ctx,
				cacheKey,
			); err != nil {
				return nil, fmt.Errorf(
					"reddit ingestion: delete invalid cache entry: %w",
					err,
				)
			}
		}
	}

	// ---------------------------------------------------------------
	// Source request
	// ---------------------------------------------------------------

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

	// ---------------------------------------------------------------
	// Cache population
	// ---------------------------------------------------------------

	if writeCache {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		cached, err := json.Marshal(
			listing,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"reddit ingestion: encode cache value: %w",
				err,
			)
		}

		if err := s.cache.Set(
			ctx,
			cacheKey,
			cached,
			s.cacheTTL,
		); err != nil {
			return nil, fmt.Errorf(
				"reddit ingestion: cache set: %w",
				err,
			)
		}
	}

	return s.normalizeListing(
		ctx,
		listing,
	)
}

func (s *IngestionService) normalizeListing(
	ctx context.Context,
	listing ListingResponse,
) ([]external.Item, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
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

		if err := ctx.Err(); err != nil {
			return nil, err
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

func DecodeCachedJSON[T any](
	data []byte,
	target *T,
) error {
	if len(data) == 0 {
		return external.ErrInvalidPayload
	}

	if target == nil {
		return external.ErrInvalidPayload
	}

	if err := json.Unmarshal(
		data,
		target,
	); err != nil {
		return fmt.Errorf(
			"%w: decode cached Reddit response",
			external.ErrInvalidPayload,
		)
	}

	return nil
}
