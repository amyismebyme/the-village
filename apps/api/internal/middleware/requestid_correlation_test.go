package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestIDCorrelation(t *testing.T) {
	var logs bytes.Buffer

	logger := slog.New(
		slog.NewTextHandler(
			&logs,
			&slog.HandlerOptions{
				Level: slog.LevelDebug,
			},
		),
	)

	const requestID = "task7-test-request-id"

	handler := RequestID(
		Logging(
			logger,
			http.HandlerFunc(func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				if got := GetRequestID(r.Context()); got != requestID {
					t.Fatalf(
						"expected request ID %q in context, got %q",
						requestID,
						got,
					)
				}

				w.WriteHeader(http.StatusOK)
			}),
		),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/communities/123",
		nil,
	)

	req.Header.Set(
		"X-Request-ID",
		requestID,
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Request-ID"); got != requestID {
		t.Fatalf(
			"expected X-Request-ID %q, got %q",
			requestID,
			got,
		)
	}

	output := logs.String()

	if !strings.Contains(
		output,
		"request_id="+requestID,
	) {
		t.Fatalf(
			"expected request ID in logs; got:\n%s",
			output,
		)
	}
}

func TestRequestIDGeneratedWhenMissing(t *testing.T) {
	handler := RequestID(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			requestID := GetRequestID(r.Context())

			if requestID == "" || requestID == "unknown" {
				t.Fatalf(
					"expected generated request ID, got %q",
					requestID,
				)
			}

			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/health",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	requestID := rec.Header().Get("X-Request-ID")

	if requestID == "" {
		t.Fatal("expected X-Request-ID response header")
	}

	if requestID == "unknown" {
		t.Fatal("expected generated request ID, got unknown")
	}
}
