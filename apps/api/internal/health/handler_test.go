package health

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)



func TestHealthLivenessReturnsOKWithoutRegistry(t *testing.T) {
	logger := slog.Default()

	handler := NewHealthHandler(
		logger,
		nil,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/health",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected /health status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}

	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf(
			"expected Content-Type application/json, got %q",
			got,
		)
	}

	if !strings.Contains(
		rec.Body.String(),
		`"status":"healthy"`,
	) {
		t.Fatalf(
			"expected healthy response, got %q",
			rec.Body.String(),
		)
	}
}

func TestHealthDoesNotDependOnDatabase(t *testing.T) {
	// A nil registry represents no dependency checks being available.
	// Liveness must still succeed.
	handler := NewHealthHandler(
		slog.Default(),
		nil,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/health",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected liveness to remain healthy without dependency checks, got %d",
			rec.Code,
		)
	}
}

func TestReadyReturns503WhenRegistryUnavailable(t *testing.T) {
	handler := NewHealthHandler(
		slog.Default(),
		nil,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/ready",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf(
			"expected /ready status %d, got %d",
			http.StatusServiceUnavailable,
			rec.Code,
		)
	}

	if !strings.Contains(
		rec.Body.String(),
		`"status":"unhealthy"`,
	) {
		t.Fatalf(
			"expected unhealthy response, got %q",
			rec.Body.String(),
		)
	}
}

func TestHealthUnknownPathReturns404(t *testing.T) {
	handler := NewHealthHandler(
		slog.Default(),
		nil,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/does-not-exist",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			rec.Code,
		)
	}
}

// Keep this compile-time assertion close to the tests so changes to the
// health handler cannot accidentally stop implementing http.Handler.
var _ http.Handler = NewHealthHandler(
	slog.Default(),
	nil,
)

// Ensure context remains part of the health-check contract.
var _ = context.Background
