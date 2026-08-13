package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/health"
	"github.com/amyismebyme/the-village/apps/api/internal/middleware"
	"github.com/amyismebyme/the-village/apps/api/internal/testutil"
)

type mockChecker struct {
	name string
	err  error
}

func (m mockChecker) Name() string {
	return m.name
}

func (m mockChecker) Check(context.Context) error {
	return m.err
}

func TestHealthHandlerHealthy(t *testing.T) {
	registry := health.NewRegistry()

	registry.Register(mockChecker{
		name: "database",
	})

	handler := NewHealthHandler(
		newDiscardLogger(),
		registry,
	)

	req := testutil.NewRequest(http.MethodGet, "/health")
	rr := testutil.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rr.Code,
		)
	}

	if contentType := rr.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf(
			"expected application/json, got %q",
			contentType,
		)
	}

	var response HealthResponse
	testutil.DecodeJSON(t, rr.Body.Bytes(), &response)

	if response.Status != "healthy" {
		t.Fatalf(
			"expected healthy, got %q",
			response.Status,
		)
	}

	if response.Checks["database"] != "healthy" {
		t.Fatalf(
			"expected database check to be healthy, got %q",
			response.Checks["database"],
		)
	}
}

func TestHealthHandlerUnhealthy(t *testing.T) {
	registry := health.NewRegistry()

	registry.Register(mockChecker{
		name: "database",
		err:  errors.New("database unavailable"),
	})

	handler := NewHealthHandler(
		newDiscardLogger(),
		registry,
	)

	req := testutil.NewRequest(http.MethodGet, "/health")
	rr := testutil.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusServiceUnavailable,
			rr.Code,
		)
	}

	var response HealthResponse
	testutil.DecodeJSON(t, rr.Body.Bytes(), &response)

	if response.Status != "unhealthy" {
		t.Fatalf(
			"expected unhealthy, got %q",
			response.Status,
		)
	}

	if response.Checks["database"] != "unhealthy" {
		t.Fatalf(
			"expected database check to be unhealthy, got %q",
			response.Checks["database"],
		)
	}
}

func TestHealthHandlerWithoutRegistry(t *testing.T) {
	handler := NewHealthHandler(
		newDiscardLogger(),
		nil,
	)

	req := testutil.NewRequest(http.MethodGet, "/health")
	rr := testutil.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusServiceUnavailable,
			rr.Code,
		)
	}

	var response HealthResponse
	testutil.DecodeJSON(t, rr.Body.Bytes(), &response)

	if response.Checks["registry"] != "unhealthy" {
		t.Fatalf(
			"expected registry check to be unhealthy, got %q",
			response.Checks["registry"],
		)
	}
}

func newDiscardLogger() *slog.Logger {
	return slog.New(
		slog.NewTextHandler(io.Discard, nil),
	)
}

func assertJSONError(
	t *testing.T,
	rec *httptest.ResponseRecorder,
	expectedStatus int,
	expectedCode string,
) {
	t.Helper()

	if rec.Code != expectedStatus {
		t.Fatalf(
			"expected status %d, got %d",
			expectedStatus,
			rec.Code,
		)
	}

	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf(
			"expected Content-Type application/json, got %q",
			got,
		)
	}

	var response errorResponse

	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf(
			"decode error response: %v",
			err,
		)
	}

	if response.Error.Code != expectedCode {
		t.Fatalf(
			"expected error code %q, got %q",
			expectedCode,
			response.Error.Code,
		)
	}

	if response.Error.Message == "" {
		t.Fatal("expected error message")
	}
}

func assertRequestID(
	t *testing.T,
	rec *httptest.ResponseRecorder,
) {
	t.Helper()

	if requestID := rec.Header().Get("X-Request-ID"); requestID == "" {
		t.Fatal("expected X-Request-ID header")
	}
}

func TestRecoveryReturnsJSONError(t *testing.T) {
	t.Parallel()

	next := http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		panic("test panic")
	})

	handler := middleware.Recovery(
		slog.Default(),
		next,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/test",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			rec.Code,
		)
	}

	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf(
			"expected application/json, got %q",
			got,
		)
	}

	var response struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(
		rec.Body,
	).Decode(&response); err != nil {
		t.Fatalf(
			"decode recovery response: %v",
			err,
		)
	}

	if response.Error.Code != "internal_error" {
		t.Fatalf(
			"expected internal_error, got %q",
			response.Error.Code,
		)
	}

	if response.Error.Message != "internal server error" {
		t.Fatalf(
			"unexpected message %q",
			response.Error.Message,
		)
	}
}
