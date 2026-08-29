package reddit

import (
	"context"
	"errors"
	"github.com/amyismebyme/the-village/apps/api/internal/external"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestNewClientRejectsMissingHTTPClient(t *testing.T) {
	_, err := NewClient(
		nil,
		"https://oauth.reddit.com",
		"the-village/test",
		time.Second,
	)

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewClientRejectsMissingUserAgent(t *testing.T) {
	_, err := NewClient(
		http.DefaultClient,
		"https://oauth.reddit.com",
		"",
		time.Second,
	)

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewClientUsesDefaultBaseURL(t *testing.T) {
	client, err := NewClient(
		http.DefaultClient,
		"",
		"the-village/test",
		time.Second,
	)
	if err != nil {
		t.Fatalf(
			"create client: %v",
			err,
		)
	}

	if client.baseURL.String() != defaultBaseURL {
		t.Fatalf(
			"expected default base URL %q, got %q",
			defaultBaseURL,
			client.baseURL.String(),
		)
	}
}

func TestNewRequest(t *testing.T) {
	client, err := NewClient(
		http.DefaultClient,
		"https://oauth.reddit.com",
		"the-village/test",
		time.Second,
	)
	if err != nil {
		t.Fatalf(
			"create client: %v",
			err,
		)
	}

	req, err := client.NewRequest(
		context.Background(),
		http.MethodGet,
		"/r/test/new",
		url.Values{
			"limit": {"10"},
		},
	)
	if err != nil {
		t.Fatalf(
			"create request: %v",
			err,
		)
	}

	if req.Method != http.MethodGet {
		t.Fatalf(
			"expected GET, got %s",
			req.Method,
		)
	}

	if req.URL.Path != "/r/test/new" {
		t.Fatalf(
			"unexpected path %q",
			req.URL.Path,
		)
	}

	if req.URL.Query().Get("limit") != "10" {
		t.Fatalf(
			"expected limit=10, got %q",
			req.URL.Query().Get("limit"),
		)
	}

	if got := req.Header.Get("User-Agent"); got !=
		"the-village/test" {
		t.Fatalf(
			"unexpected user agent %q",
			got,
		)
	}
}

func TestDoAuthenticatedSetsBearerToken(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				if got := r.Header.Get(
					"Authorization",
				); got != "Bearer test-token" {
					t.Fatalf(
						"unexpected authorization header %q",
						got,
					)
				}

				if got := r.Header.Get(
					"User-Agent",
				); got != "the-village/test" {
					t.Fatalf(
						"unexpected user agent %q",
						got,
					)
				}

				w.Header().Set(
					"Content-Type",
					"application/json",
				)

				w.WriteHeader(http.StatusOK)

				_, _ = w.Write(
					[]byte(`{"ok":true}`),
				)
			},
		),
	)
	defer server.Close()

	client, err := NewClient(
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

	req, err := client.NewRequest(
		context.Background(),
		http.MethodGet,
		"/test",
		nil,
	)
	if err != nil {
		t.Fatalf(
			"create request: %v",
			err,
		)
	}

	response, err := client.DoAuthenticated(
		context.Background(),
		req,
		"test-token",
	)
	if err != nil {
		t.Fatalf(
			"authenticated request: %v",
			err,
		)
	}

	if err := response.Body.Close(); err != nil {
		t.Fatalf(
			"close response body: %v",
			err,
		)
	}

	if response.StatusCode != http.StatusOK {
		t.Fatalf(
			"expected status 200, got %d",
			response.StatusCode,
		)
	}
}

func TestDoAuthenticatedRejectsMissingToken(t *testing.T) {
	client, err := NewClient(
		http.DefaultClient,
		"https://oauth.reddit.com",
		"the-village/test",
		time.Second,
	)
	if err != nil {
		t.Fatalf(
			"create client: %v",
			err,
		)
	}

	req, err := client.NewRequest(
		context.Background(),
		http.MethodGet,
		"/test",
		nil,
	)
	if err != nil {
		t.Fatalf(
			"create request: %v",
			err,
		)
	}

	_, err = client.DoAuthenticated(
		context.Background(),
		req,
		"",
	)

	if !errors.Is(
		err,
		external.ErrUnauthorized,
	) {
		t.Fatalf(
			"expected ErrUnauthorized, got %v",
			err,
		)
	}
}

func TestFetchListing(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				if r.URL.Path !=
					"/r/toronto/new" {
					t.Fatalf(
						"unexpected path %q",
						r.URL.Path,
					)
				}

				if r.URL.Query().Get("limit") != "10" {
					t.Fatalf(
						"expected limit=10, got %q",
						r.URL.Query().Get("limit"),
					)
				}

				if got := r.Header.Get(
					"Authorization",
				); got != "Bearer test-token" {
					t.Fatalf(
						"unexpected authorization %q",
						got,
					)
				}

				w.Header().Set(
					"Content-Type",
					"application/json",
				)

				w.WriteHeader(http.StatusOK)

				_, _ = w.Write(
					[]byte(`{
						"data": {
							"after": null,
							"before": null,
							"children": [
								{
									"kind": "t3",
									"data": {
										"id": "abc123",
										"name": "t3_abc123",
										"title": "Toronto Community",
										"selftext": "Example post",
										"url": "https://example.test/post",
										"permalink": "/r/toronto/comments/abc123/",
										"subreddit": "toronto",
										"author": "example-user",
										"created_utc": 1234567890
									}
								}
							]
						}
					}`),
				)
			},
		),
	)
	defer server.Close()

	client, err := NewClient(
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

	result, err := client.FetchListing(
		context.Background(),
		"test-token",
		"toronto",
		10,
		"",
	)
	if err != nil {
		t.Fatalf(
			"FetchListing failed: %v",
			err,
		)
	}

	if len(result.Data.Children) != 1 {
		t.Fatalf(
			"expected 1 child, got %d",
			len(result.Data.Children),
		)
	}

	post := result.Data.Children[0].Data

	if post.ID != "abc123" {
		t.Fatalf(
			"expected ID abc123, got %q",
			post.ID,
		)
	}

	if post.Title != "Toronto Community" {
		t.Fatalf(
			"unexpected title %q",
			post.Title,
		)
	}
}

func TestFetchListingMapsHTTPErrors(t *testing.T) {
	tests := []struct {
		name   string
		status int
		want   error
	}{
		{
			name:   "unauthorized",
			status: http.StatusUnauthorized,
			want:   external.ErrUnauthorized,
		},
		{
			name:   "forbidden",
			status: http.StatusForbidden,
			want:   external.ErrForbidden,
		},
		{
			name:   "not found",
			status: http.StatusNotFound,
			want:   external.ErrNotFound,
		},
		{
			name:   "rate limited",
			status: http.StatusTooManyRequests,
			want:   external.ErrRateLimited,
		},
		{
			name:   "upstream",
			status: http.StatusBadGateway,
			want:   external.ErrUpstream,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(
				http.HandlerFunc(
					func(
						w http.ResponseWriter,
						r *http.Request,
					) {
						w.WriteHeader(tt.status)
					},
				),
			)
			defer server.Close()

			client, err := NewClient(
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

			req, err := client.NewRequest(
				context.Background(),
				http.MethodGet,
				"/test",
				nil,
			)
			if err != nil {
				t.Fatalf(
					"create request: %v",
					err,
				)
			}

			_, err = client.DoAuthenticated(
				context.Background(),
				req,
				"test-token",
			)

			if !errors.Is(err, tt.want) {
				t.Fatalf(
					"expected %v, got %v",
					tt.want,
					err,
				)
			}
		})
	}
}

func TestFetchListingPropagatesCancellation(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				<-r.Context().Done()
			},
		),
	)
	defer server.Close()

	client, err := NewClient(
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

	ctx, cancel := context.WithCancel(
		context.Background(),
	)

	cancel()

	_, err = client.FetchListing(
		ctx,
		"test-token",
		"toronto",
		10,
		"",
	)

	if !errors.Is(
		err,
		context.Canceled,
	) {
		t.Fatalf(
			"expected context.Canceled, got %v",
			err,
		)
	}
}

func TestFetchListingPreservesTypedErrors(
	t *testing.T,
) {
	tests := []struct {
		name   string
		status int
		want   error
	}{
		{
			name:   "unauthorized",
			status: http.StatusUnauthorized,
			want:   external.ErrUnauthorized,
		},
		{
			name:   "forbidden",
			status: http.StatusForbidden,
			want:   external.ErrForbidden,
		},
		{
			name:   "not found",
			status: http.StatusNotFound,
			want:   external.ErrNotFound,
		},
		{
			name:   "rate limited",
			status: http.StatusTooManyRequests,
			want:   external.ErrRateLimited,
		},
		{
			name:   "internal server error",
			status: http.StatusInternalServerError,
			want:   external.ErrUpstream,
		},
		{
			name:   "bad gateway",
			status: http.StatusBadGateway,
			want:   external.ErrUpstream,
		},
		{
			name:   "service unavailable",
			status: http.StatusServiceUnavailable,
			want:   external.ErrUpstream,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(
				http.HandlerFunc(
					func(
						w http.ResponseWriter,
						r *http.Request,
					) {
						w.WriteHeader(tt.status)
					},
				),
			)
			defer server.Close()

			client, err := NewClient(
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

			_, err = client.FetchListing(
				context.Background(),
				"test-token",
				"toronto",
				25,
				"",
			)

			if !errors.Is(err, tt.want) {
				t.Fatalf(
					"expected errors.Is(%v), got %v",
					tt.want,
					err,
				)
			}

			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestFetchListingPreservesTimeoutError(
	t *testing.T,
) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				<-r.Context().Done()
			},
		),
	)
	defer server.Close()

	client, err := NewClient(
		server.Client(),
		server.URL,
		"the-village/test",
		25*time.Millisecond,
	)
	if err != nil {
		t.Fatalf(
			"create client: %v",
			err,
		)
	}

	_, err = client.FetchListing(
		context.Background(),
		"test-token",
		"toronto",
		25,
		"",
	)

	if !errors.Is(
		err,
		external.ErrTimeout,
	) {
		t.Fatalf(
			"expected ErrTimeout, got %v",
			err,
		)
	}
}

func TestFetchListingPreservesCancellation(
	t *testing.T,
) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				<-r.Context().Done()
			},
		),
	)
	defer server.Close()

	client, err := NewClient(
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

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	_, err = client.FetchListing(
		ctx,
		"test-token",
		"toronto",
		25,
		"",
	)

	if !errors.Is(
		err,
		context.Canceled,
	) {
		t.Fatalf(
			"expected context.Canceled, got %v",
			err,
		)
	}
}

func TestFetchListingRejectsMalformedJSON(
	t *testing.T,
) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				w.Header().Set(
					"Content-Type",
					"application/json",
				)

				_, _ = w.Write(
					[]byte(`{"data":`),
				)
			},
		),
	)
	defer server.Close()

	client, err := NewClient(
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

	_, err = client.FetchListing(
		context.Background(),
		"test-token",
		"toronto",
		25,
		"",
	)

	if !errors.Is(
		err,
		external.ErrInvalidPayload,
	) {
		t.Fatalf(
			"expected ErrInvalidPayload, got %v",
			err,
		)
	}
}

func TestFetchListingRejectsEmptyResponse(
	t *testing.T,
) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer server.Close()

	client, err := NewClient(
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

	_, err = client.FetchListing(
		context.Background(),
		"test-token",
		"toronto",
		25,
		"",
	)

	if !errors.Is(
		err,
		external.ErrInvalidPayload,
	) {
		t.Fatalf(
			"expected ErrInvalidPayload, got %v",
			err,
		)
	}
}

func TestRedditErrorsHaveCorrectRetryClassification(
	t *testing.T,
) {
	tests := []struct {
		name      string
		err       error
		retry     bool
		permanent bool
	}{
		{
			name:  "rate limited",
			err:   external.ErrRateLimited,
			retry: true,
		},
		{
			name:  "timeout",
			err:   external.ErrTimeout,
			retry: true,
		},
		{
			name:  "upstream",
			err:   external.ErrUpstream,
			retry: true,
		},
		{
			name:      "unauthorized",
			err:       external.ErrUnauthorized,
			permanent: true,
		},
		{
			name:      "forbidden",
			err:       external.ErrForbidden,
			permanent: true,
		},
		{
			name:      "invalid payload",
			err:       external.ErrInvalidPayload,
			permanent: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := external.IsRetryable(tt.err); got != tt.retry {
				t.Fatalf(
					"expected retryable=%v, got %v",
					tt.retry,
					got,
				)
			}

			if got := external.IsPermanent(tt.err); got != tt.permanent {
				t.Fatalf(
					"expected permanent=%v, got %v",
					tt.permanent,
					got,
				)
			}
		})
	}
}

type recordingRateLimiter struct {
	calls int
	err   error
}

func (r *recordingRateLimiter) Wait(
	context.Context,
) error {
	r.calls++
	return r.err
}

func TestFetchListingUsesRateLimiter(
	t *testing.T,
) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {
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

	client, err := NewClient(
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

	limiter := &recordingRateLimiter{}

	client.SetRateLimiter(limiter)

	_, err = client.FetchListing(
		context.Background(),
		"test-token",
		"toronto",
		10,
		"",
	)
	if err != nil {
		t.Fatalf(
			"FetchListing failed: %v",
			err,
		)
	}

	if limiter.calls != 1 {
		t.Fatalf(
			"expected limiter to be called once, got %d",
			limiter.calls,
		)
	}
}

func TestFetchListingRateLimitCancellationPreventsRequest(
	t *testing.T,
) {
	var requests int

	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				requests++

				w.WriteHeader(
					http.StatusOK,
				)
			},
		),
	)
	defer server.Close()

	client, err := NewClient(
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

	limiter := &recordingRateLimiter{
		err: context.Canceled,
	}

	client.SetRateLimiter(limiter)

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	_, err = client.FetchListing(
		ctx,
		"test-token",
		"toronto",
		10,
		"",
	)

	if !errors.Is(
		err,
		context.Canceled,
	) {
		t.Fatalf(
			"expected context.Canceled, got %v",
			err,
		)
	}

	if requests != 0 {
		t.Fatalf(
			"expected zero outbound requests, got %d",
			requests,
		)
	}
}
