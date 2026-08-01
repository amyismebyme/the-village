package handlers

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/health"
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
