package reddit

import (
	"context"
	"errors"
	"github.com/amyismebyme/the-village/apps/api/internal/external"
	"github.com/amyismebyme/the-village/apps/api/internal/external/testutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAuthenticatorFetchesApplicationToken(
	t *testing.T,
) {
	requests := 0

	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				requests++

				if r.Method != http.MethodPost {
					t.Fatalf(
						"expected POST, got %s",
						r.Method,
					)
				}

				if r.URL.Path != accessTokenPath {
					t.Fatalf(
						"unexpected path %q",
						r.URL.Path,
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

				username, password, ok :=
					r.BasicAuth()

				if !ok {
					t.Fatal(
						"expected basic authentication",
					)
				}

				if username != "client-id" {
					t.Fatalf(
						"unexpected client ID %q",
						username,
					)
				}

				if password != "client-secret" {
					t.Fatalf(
						"unexpected client secret",
					)
				}

				w.Header().Set(
					"Content-Type",
					"application/json",
				)

				_, _ = w.Write(
					[]byte(`{
						"access_token": "test-access-token",
						"token_type": "bearer",
						"expires_in": 3600,
						"scope": "*"
					}`),
				)
			},
		),
	)
	defer server.Close()

	authenticator, err := NewAuthenticator(
		server.Client(),
		server.URL,
		"client-id",
		"client-secret",
		"the-village/test",
		time.Second,
	)
	if err != nil {
		t.Fatalf(
			"create authenticator: %v",
			err,
		)
	}

	token, err := authenticator.Token(
		context.Background(),
	)
	if err != nil {
		t.Fatalf(
			"get token: %v",
			err,
		)
	}

	if token != "test-access-token" {
		t.Fatalf(
			"unexpected token %q",
			token,
		)
	}

	if requests != 1 {
		t.Fatalf(
			"expected 1 token request, got %d",
			requests,
		)
	}

	// The cached token should be reused.
	token, err = authenticator.Token(
		context.Background(),
	)
	if err != nil {
		t.Fatalf(
			"get cached token: %v",
			err,
		)
	}

	if token != "test-access-token" {
		t.Fatalf(
			"unexpected cached token %q",
			token,
		)
	}

	if requests != 1 {
		t.Fatalf(
			"expected cached token to avoid second request, got %d requests",
			requests,
		)
	}
}

func TestAuthenticatorMapsUnauthorized(
	t *testing.T,
) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				w.WriteHeader(
					http.StatusUnauthorized,
				)
			},
		),
	)
	defer server.Close()

	authenticator, err := NewAuthenticator(
		server.Client(),
		server.URL,
		"client-id",
		"client-secret",
		"the-village/test",
		time.Second,
	)
	if err != nil {
		t.Fatalf(
			"create authenticator: %v",
			err,
		)
	}

	_, err = authenticator.Token(
		context.Background(),
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

func TestAuthenticatorPropagatesCancellation(
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

	authenticator, err := NewAuthenticator(
		server.Client(),
		server.URL,
		"client-id",
		"client-secret",
		"the-village/test",
		time.Second,
	)
	if err != nil {
		t.Fatalf(
			"create authenticator: %v",
			err,
		)
	}

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	_, err = authenticator.Token(
		ctx,
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

func TestRedditListingRequestContract(
	t *testing.T,
) {
	server := testutil.NewRouteServer(
		map[string]testutil.Response{
			"/r/toronto/new": {
				StatusCode: http.StatusOK,
				Body: `{
					"data": {
						"after": null,
						"before": null,
						"children": []
					}
				}`,
			},
		},
		testutil.Response{
			StatusCode: http.StatusNotFound,
		},
	)
	defer server.Close()

	client, err := NewClient(
		server.Server.Client(),
		server.URL(),
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
		10,
		"",
	)
	if err != nil {
		t.Fatalf(
			"fetch listing: %v",
			err,
		)
	}

	_, method, path, headers :=
		server.Snapshot()

	if method != http.MethodGet {
		t.Fatalf(
			"expected GET, got %s",
			method,
		)
	}

	if path != "/r/toronto/new" {
		t.Fatalf(
			"unexpected path %q",
			path,
		)
	}

	if headers.Get("User-Agent") !=
		"the-village/test" {
		t.Fatalf(
			"unexpected user agent %q",
			headers.Get("User-Agent"),
		)
	}

	if headers.Get("Authorization") !=
		"Bearer test-token" {
		t.Fatalf(
			"unexpected authorization header",
		)
	}
}

func TestRedditAuthenticationAndClientShareRateLimiter(
	t *testing.T,
) {
	limiter := &recordingRateLimiter{}

	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				switch r.URL.Path {
				case accessTokenPath:
					w.Header().Set(
						"Content-Type",
						"application/json",
					)

					_, _ = w.Write(
						[]byte(`{
							"access_token": "test-access-token",
							"token_type": "bearer",
							"expires_in": 3600,
							"scope": "*"
						}`),
					)

				case "/r/toronto/new":
					w.Header().Set(
						"Content-Type",
						"application/json",
					)

					_, _ = w.Write(
						[]byte(`{
							"data": {
								"after": null,
								"before": null,
								"children": []
							}
						}`),
					)

				default:
					http.NotFound(
						w,
						r,
					)
				}
			},
		),
	)
	defer server.Close()

	authenticator, err := NewAuthenticator(
		server.Client(),
		server.URL,
		"client-id",
		"client-secret",
		"the-village/test",
		time.Second,
	)
	if err != nil {
		t.Fatalf(
			"create authenticator: %v",
			err,
		)
	}

	client, err := NewClient(
		server.Client(),
		server.URL,
		"the-village/test",
		time.Second,
	)
	if err != nil {
		t.Fatalf(
			"create Reddit client: %v",
			err,
		)
	}

	authenticator.SetRateLimiter(limiter)
	client.SetRateLimiter(limiter)

	// First operation: OAuth token request.
	token, err := authenticator.Token(
		context.Background(),
	)
	if err != nil {
		t.Fatalf(
			"authenticate: %v",
			err,
		)
	}

	if token != "test-access-token" {
		t.Fatalf(
			"unexpected token %q",
			token,
		)
	}

	// Second operation: Reddit listing request.
	_, err = client.FetchListing(
		context.Background(),
		token,
		"toronto",
		10,
		"",
	)
	if err != nil {
		t.Fatalf(
			"fetch listing: %v",
			err,
		)
	}

	if limiter.calls != 2 {
		t.Fatalf(
			"expected limiter to be called twice, got %d",
			limiter.calls,
		)
	}
}
