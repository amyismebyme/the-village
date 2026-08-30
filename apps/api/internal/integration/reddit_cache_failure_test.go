package integration

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/amyismebyme/the-village/apps/api/internal/cache"
	"github.com/amyismebyme/the-village/apps/api/internal/external"
	"github.com/amyismebyme/the-village/apps/api/internal/external/reddit"
)

func TestRedditCacheExpiryFetchesFreshData(
	t *testing.T,
) {
	var requests int

	server := http.HandlerFunc(
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			if r.URL.Path != "/r/toronto/new" {
				http.NotFound(w, r)
				return
			}

			requests++

			w.Header().Set(
				"Content-Type",
				"application/json",
			)

			_, _ = w.Write(
				[]byte(`{
					"data":{
						"after":null,
						"before":null,
						"children":[
							{
								"kind":"t3",
								"data":{
									"id":"fresh-post",
									"name":"t3_fresh-post",
									"title":"Fresh post",
									"selftext":"",
									"url":"https://example.test/fresh",
									"permalink":"/r/toronto/comments/fresh-post/",
									"subreddit":"toronto",
									"author":"test-user",
									"created_utc":1234567890,
									"is_self":true
								}
							}
						]
					}
				}`),
			)
		},
	)

	testServer := httptest.NewServer(server)
	defer testServer.Close()

	client, err := reddit.NewClient(
		testServer.Client(),
		testServer.URL,
		"the-village/test",
		time.Second,
	)
	if err != nil {
		t.Fatalf(
			"create client: %v",
			err,
		)
	}

	ingestion := reddit.NewIngestionService(
		client,
		reddit.NewPostNormalizer(),
	)

	memoryCache, err := cache.NewMemory(10)
	if err != nil {
		t.Fatalf(
			"create cache: %v",
			err,
		)
	}

	if err := ingestion.SetCache(
		memoryCache,
		20*time.Millisecond,
	); err != nil {
		t.Fatalf(
			"configure cache: %v",
			err,
		)
	}

	ctx := context.Background()

	if _, err := ingestion.IngestListing(
		ctx,
		"test-token",
		"toronto",
		10,
		"",
	); err != nil {
		t.Fatalf(
			"first ingestion: %v",
			err,
		)
	}

	time.Sleep(
		30 * time.Millisecond,
	)

	items, err := ingestion.IngestListing(
		ctx,
		"test-token",
		"toronto",
		10,
		"",
	)
	if err != nil {
		t.Fatalf(
			"second ingestion: %v",
			err,
		)
	}

	if requests != 2 {
		t.Fatalf(
			"expected two source requests, got %d",
			requests,
		)
	}

	if len(items) != 1 ||
		items[0].ExternalID != "fresh-post" {
		t.Fatalf(
			"expected fresh post after expiry, got %+v",
			items,
		)
	}
}

func TestRedditCacheFailureDoesNotServeExpiredData(
	t *testing.T,
) {
	var requests int

	server := http.HandlerFunc(
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			if r.URL.Path != "/r/toronto/new" {
				http.NotFound(w, r)
				return
			}

			requests++

			http.Error(
				w,
				"temporary source failure",
				http.StatusBadGateway,
			)
		},
	)

	testServer := httptest.NewServer(server)
	defer testServer.Close()

	client, err := reddit.NewClient(
		testServer.Client(),
		testServer.URL,
		"the-village/test",
		time.Second,
	)
	if err != nil {
		t.Fatalf(
			"create client: %v",
			err,
		)
	}

	ingestion := reddit.NewIngestionService(
		client,
		reddit.NewPostNormalizer(),
	)

	memoryCache, err := cache.NewMemory(10)
	if err != nil {
		t.Fatalf(
			"create cache: %v",
			err,
		)
	}

	if err := ingestion.SetCache(
		memoryCache,
		20*time.Millisecond,
	); err != nil {
		t.Fatalf(
			"configure cache: %v",
			err,
		)
	}

	key, err := cache.RedditListingKey(
		"toronto",
		"",
		10,
	)
	if err != nil {
		t.Fatalf(
			"create cache key: %v",
			err,
		)
	}

	staleListing := reddit.ListingResponse{
		Data: reddit.ListingData{
			Children: []reddit.ListingChild{
				{
					Kind: "t3",
					Data: reddit.RedditPost{
						ID:        "stale-post",
						Name:      "t3_stale-post",
						Title:     "Stale post",
						Subreddit: "toronto",
					},
				},
			},
		},
	}

	staleValue, err := json.Marshal(
		staleListing,
	)
	if err != nil {
		t.Fatalf(
			"encode stale value: %v",
			err,
		)
	}

	if err := memoryCache.Set(
		context.Background(),
		key,
		staleValue,
		20*time.Millisecond,
	); err != nil {
		t.Fatalf(
			"seed stale cache: %v",
			err,
		)
	}

	time.Sleep(
		30 * time.Millisecond,
	)

	_, err = ingestion.IngestListing(
		context.Background(),
		"test-token",
		"toronto",
		10,
		"",
	)

	if err == nil {
		t.Fatal(
			"expected source failure",
		)
	}

	if !errors.Is(
		err,
		external.ErrUpstream,
	) {
		t.Fatalf(
			"expected ErrUpstream, got %v",
			err,
		)
	}

	if requests != 1 {
		t.Fatalf(
			"expected one source request, got %d",
			requests,
		)
	}
}

func TestRedditCacheCorruptionRecovers(
	t *testing.T,
) {
	var (
		mu       sync.Mutex
		requests int
	)

	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				if r.URL.Path != "/r/toronto/new" {
					http.NotFound(w, r)
					return
				}

				mu.Lock()
				requests++
				mu.Unlock()

				w.Header().Set(
					"Content-Type",
					"application/json",
				)

				_, _ = w.Write(
					[]byte(`{
						"data":{
							"after":null,
							"before":null,
							"children":[]
						}
					}`),
				)
			},
		),
	)
	defer server.Close()

	client, err := reddit.NewClient(
		server.Client(),
		server.URL,
		"the-village/test",
		time.Second,
	)
	if err != nil {
		t.Fatalf(
			"create client: %v",
			err,
		)
	}

	ingestion := reddit.NewIngestionService(
		client,
		reddit.NewPostNormalizer(),
	)

	memoryCache, err := cache.NewMemory(10)
	if err != nil {
		t.Fatalf(
			"create cache: %v",
			err,
		)
	}

	if err := ingestion.SetCache(
		memoryCache,
		time.Minute,
	); err != nil {
		t.Fatalf(
			"configure cache: %v",
			err,
		)
	}

	key, err := cache.RedditListingKey(
		"toronto",
		"",
		10,
	)
	if err != nil {
		t.Fatalf(
			"create cache key: %v",
			err,
		)
	}

	if err := memoryCache.Set(
		context.Background(),
		key,
		[]byte("{ definitely not valid json"),
		time.Minute,
	); err != nil {
		t.Fatalf(
			"seed corrupt cache: %v",
			err,
		)
	}

	items, err := ingestion.IngestListing(
		context.Background(),
		"test-token",
		"toronto",
		10,
		"",
	)
	if err != nil {
		t.Fatalf(
			"ingestion: %v",
			err,
		)
	}

	if len(items) != 0 {
		t.Fatalf(
			"expected empty fresh listing, got %d items",
			len(items),
		)
	}

	mu.Lock()
	gotRequests := requests
	mu.Unlock()

	if gotRequests != 1 {
		t.Fatalf(
			"expected one source request, got %d",
			gotRequests,
		)
	}

	cached, ok, err := memoryCache.Get(
		context.Background(),
		key,
	)
	if err != nil {
		t.Fatalf(
			"get repaired cache: %v",
			err,
		)
	}

	if !ok {
		t.Fatal(
			"expected cache to be repaired",
		)
	}

	var repaired reddit.ListingResponse

	if err := reddit.DecodeCachedJSON(
		cached,
		&repaired,
	); err != nil {
		t.Fatalf(
			"decode repaired cache: %v",
			err,
		)
	}
}
