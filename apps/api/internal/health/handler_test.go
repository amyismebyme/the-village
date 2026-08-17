package health

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type testChecker struct {
	name string
	err  error
}

func (c testChecker) Name() string {
	return c.name
}

func (c testChecker) Check(context.Context) error {
	return c.err
}

func TestHealthLivenessIgnoresDependencyFailure(t *testing.T) {
	registry := NewRegistry()

	registry.Register(testChecker{
		name: "database",
		err:  errors.New("database unavailable"),
	})

	handler := NewHealthHandler(
		slog.Default(),
		registry,
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

	if !strings.Contains(
		rec.Body.String(),
		`"status":"healthy"`,
	) {
		t.Fatalf(
			"expected healthy liveness response, got %q",
			rec.Body.String(),
		)
	}

	// Liveness must not expose dependency checks.
	if strings.Contains(rec.Body.String(), "database") {
		t.Fatalf(
			"liveness response unexpectedly contains dependency data: %s",
			rec.Body.String(),
		)
	}
}

func TestReadinessReturns503WhenDependencyFails(t *testing.T) {
	registry := NewRegistry()

	registry.Register(testChecker{
		name: "database",
		err:  errors.New("database unavailable"),
	})

	handler := NewHealthHandler(
		slog.Default(),
		registry,
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

	body := rec.Body.String()

	if !strings.Contains(
		body,
		`"status":"unhealthy"`,
	) {
		t.Fatalf(
			"expected unhealthy readiness response, got %q",
			body,
		)
	}

	if !strings.Contains(
		body,
		`"name":"database"`,
	) {
		t.Fatalf(
			"expected database check in readiness response, got %q",
			body,
		)
	}
}
