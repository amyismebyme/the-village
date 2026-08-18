package server

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/handlers"
	"github.com/amyismebyme/the-village/apps/api/internal/health"
	"github.com/amyismebyme/the-village/apps/api/internal/middleware"
)

func newTestRouter() http.Handler {
	logger := slog.New(
		slog.NewTextHandler(
			httptest.NewRecorder(),
			nil,
		),
	)

	healthRegistry := health.NewRegistry()

	handler := handlers.NewHandler(
		routerCommunityServiceMock{},
	)

	return NewRouter(
		logger,
		healthRegistry,
		handler,
	)
}

func TestRouterUnknownPathReturnsNotFound(t *testing.T) {
	t.Parallel()

	router := newTestRouter()

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/does-not-exist",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			rec.Code,
		)
	}
}

func TestRouterHealth(t *testing.T) {
	t.Parallel()

	router := newTestRouter()

	req := httptest.NewRequest(
		http.MethodGet,
		"/health",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected /health status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}
}

func TestRouterReady(t *testing.T) {
	t.Parallel()

	router := newTestRouter()

	req := httptest.NewRequest(
		http.MethodGet,
		"/ready",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// The test router has no dependency registered in the health registry.
	// Therefore readiness should fail closed with 503.
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf(
			"expected /ready status %d, got %d",
			http.StatusServiceUnavailable,
			rec.Code,
		)
	}
}

func TestRouterUnknownSystemPathReturnsNotFound(t *testing.T) {
	t.Parallel()

	router := newTestRouter()

	req := httptest.NewRequest(
		http.MethodGet,
		"/this-does-not-exist",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			rec.Code,
		)
	}
}

func TestRouterLogsNormalizedCommunityRoute(t *testing.T) {
	t.Parallel()

	var logs bytes.Buffer

	logger := slog.New(
		slog.NewTextHandler(
			&logs,
			&slog.HandlerOptions{
				Level: slog.LevelDebug,
			},
		),
	)

	mux := http.NewServeMux()

	mux.HandleFunc(
		"GET /api/v1/communities/{id}",
		func(
			w http.ResponseWriter,
			_ *http.Request,
		) {
			w.WriteHeader(http.StatusOK)
		},
	)

	handler := middleware.RequestID(
		middleware.Logging(
			logger,
			mux,
		),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/communities/839274",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	output := logs.String()

	if !strings.Contains(
		output,
		"route=/api/v1/communities/{id}",
	) {
		t.Fatalf(
			"expected normalized community route; got:\n%s",
			output,
		)
	}

	if strings.Contains(
		output,
		"route=/api/v1/communities/839274",
	) {
		t.Fatalf(
			"unexpected concrete route in log; got:\n%s",
			output,
		)
	}

	if !strings.Contains(
		output,
		"status=200",
	) {
		t.Fatalf(
			"expected status=200; got:\n%s",
			output,
		)
	}

	if !strings.Contains(
		output,
		"request_id=",
	) {
		t.Fatalf(
			"expected request_id; got:\n%s",
			output,
		)
	}

	if !strings.Contains(
		output,
		"duration_ms=",
	) {
		t.Fatalf(
			"expected duration_ms; got:\n%s",
			output,
		)
	}
}

func TestRouterCreateCommunityReachesHandler(t *testing.T) {
	t.Parallel()

	router := newTestRouter()

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/communities",
		strings.NewReader(`{
			"name": "Toronto Men",
			"slug": "toronto-men",
			"description": "Toronto community",
			"external_source": "test"
		}`),
	)

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf(
			"expected POST /api/v1/communities to return %d, got %d: %s",
			http.StatusCreated,
			rec.Code,
			rec.Body.String(),
		)
	}

	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf(
			"expected Content-Type application/json, got %q",
			got,
		)
	}

	if requestID := rec.Header().Get("X-Request-ID"); requestID == "" {
		t.Fatal("expected X-Request-ID header")
	}
}
