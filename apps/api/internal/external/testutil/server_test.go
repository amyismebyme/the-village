package testutil

import (
	"net/http"
	"testing"
)

func TestServerCapturesRequests(t *testing.T) {
	server := NewServer(
		Response{
			StatusCode: http.StatusTooManyRequests,
			Body:       `{"error":"rate limited"}`,
		},
	)
	defer server.Close()

	response, err := http.Get(
		server.URL() + "/rate-limit",
	)
	if err != nil {
		t.Fatalf(
			"request test server: %v",
			err,
		)
	}

	defer func() {
		if err := response.Body.Close(); err != nil {
			t.Errorf(
				"close response body: %v",
				err,
			)
		}
	}()

	if response.StatusCode != http.StatusTooManyRequests {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusTooManyRequests,
			response.StatusCode,
		)
	}

	count, method, path, _ := server.Snapshot()

	if count != 1 {
		t.Fatalf(
			"expected request count 1, got %d",
			count,
		)
	}

	if method != http.MethodGet {
		t.Fatalf(
			"expected GET, got %s",
			method,
		)
	}

	if path != "/rate-limit" {
		t.Fatalf(
			"expected /rate-limit, got %s",
			path,
		)
	}
}

func TestNewRouteServer(t *testing.T) {
	server := NewRouteServer(
		map[string]Response{
			"/token": {
				StatusCode: http.StatusOK,
				Body:       `{"access_token":"test"}`,
			},
			"/items": {
				StatusCode: http.StatusOK,
				Body:       `{"items":[]}`,
			},
		},
		Response{
			StatusCode: http.StatusNotFound,
		},
	)
	defer server.Close()

	for _, path := range []string{
		"/token",
		"/items",
	} {
		response, err := http.Get(
			server.URL() + path,
		)
		if err != nil {
			t.Fatalf(
				"GET %s: %v",
				path,
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
				"GET %s: expected 200, got %d",
				path,
				response.StatusCode,
			)
		}
	}

	response, err := http.Get(
		server.URL() + "/unknown",
	)
	if err != nil {
		t.Fatalf(
			"GET /unknown: %v",
			err,
		)
	}
	if err := response.Body.Close(); err != nil {
		t.Fatalf(
			"close response body: %v",
			err,
		)
	}

	if response.StatusCode != http.StatusNotFound {
		t.Fatalf(
			"expected 404 for unknown route, got %d",
			response.StatusCode,
		)
	}
}
