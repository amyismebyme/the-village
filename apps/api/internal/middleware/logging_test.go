package middleware

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestLogger(buffer *bytes.Buffer) *slog.Logger {
	return slog.New(
		slog.NewTextHandler(
			buffer,
			&slog.HandlerOptions{
				Level: slog.LevelDebug,
			},
		),
	)
}

func TestLoggingRequestContract(t *testing.T) {
	var logs bytes.Buffer

	logger := newTestLogger(&logs)

	handler := Logging(
		logger,
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		}),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/communities/123",
		nil,
	)

	req = req.WithContext(
		withRequestID(
			req.Context(),
			"test-request-id",
		),
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	output := logs.String()

	for _, expected := range []string{
		"request_id=test-request-id",
		"method=GET",
		"status=200",
		"duration_ms=",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf(
				"expected log to contain %q; got:\n%s",
				expected,
				output,
			)
		}
	}
}

func TestLoggingUsesServeMuxRoutePattern(t *testing.T) {
	var logs bytes.Buffer

	logger := newTestLogger(&logs)

	mux := http.NewServeMux()

	mux.HandleFunc(
		"GET /api/v1/communities/{id}",
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			w.WriteHeader(http.StatusOK)
		},
	)

	handler := Logging(logger, mux)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/communities/123",
		nil,
	)

	req = req.WithContext(
		withRequestID(
			req.Context(),
			"route-test-id",
		),
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	output := logs.String()

	if !strings.Contains(
		output,
		"route=/api/v1/communities/{id}",
	) {
		t.Fatalf(
			"expected normalized route in log; got:\n%s",
			output,
		)
	}

	if strings.Contains(
		output,
		"route=/api/v1/communities/123",
	) {
		t.Fatalf(
			"log contains unnormalized route:\n%s",
			output,
		)
	}
}

func TestLoggingCapturesStatus(t *testing.T) {
	var logs bytes.Buffer

	logger := newTestLogger(&logs)

	handler := Logging(
		logger,
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			http.Error(
				w,
				"not found",
				http.StatusNotFound,
			)
		}),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/test",
		nil,
	)

	req = req.WithContext(
		withRequestID(
			context.Background(),
			"status-test-id",
		),
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	output := logs.String()

	if !strings.Contains(
		output,
		"status=404",
	) {
		t.Fatalf(
			"expected status=404; got:\n%s",
			output,
		)
	}
}

func TestLoggingDoesNotLogSensitiveRequestData(t *testing.T) {
	var logs bytes.Buffer

	logger := newTestLogger(&logs)

	handler := Logging(
		logger,
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/test",
		strings.NewReader(
			`{"password":"super-secret","token":"secret-token"}`,
		),
	)

	req.Header.Set(
		"Authorization",
		"Bearer very-secret-token",
	)

	req = req.WithContext(
		withRequestID(
			req.Context(),
			"security-test-id",
		),
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	output := logs.String()

	for _, sensitive := range []string{
		"Authorization",
		"Bearer",
		"super-secret",
		"secret-token",
		"password",
		"token",
	} {
		if strings.Contains(output, sensitive) {
			t.Fatalf(
				"log contains sensitive value %q:\n%s",
				sensitive,
				output,
			)
		}
	}
}
