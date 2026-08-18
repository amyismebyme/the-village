package server

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/middleware"
)

func TestRouterContract(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
	}{
		{
			name:       "health",
			method:     http.MethodGet,
			path:       "/health",
			wantStatus: http.StatusOK,
		},
		{
			name:       "ready",
			method:     http.MethodGet,
			path:       "/ready",
			wantStatus: http.StatusServiceUnavailable,
		},
		{
			name:       "version",
			method:     http.MethodGet,
			path:       "/version",
			wantStatus: http.StatusOK,
		},
		{
			name:       "status",
			method:     http.MethodGet,
			path:       "/status",
			wantStatus: http.StatusOK,
		},
		{
			name:       "metrics",
			method:     http.MethodGet,
			path:       "/metrics",
			wantStatus: http.StatusOK,
		},
		{
			name:       "community collection GET",
			method:     http.MethodGet,
			path:       "/api/v1/communities",
			wantStatus: http.StatusOK,
		},
		{
			name:       "community resource GET",
			method:     http.MethodGet,
			path:       "/api/v1/communities/1",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "community resource DELETE",
			method:     http.MethodDelete,
			path:       "/api/v1/communities/1",
			wantStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(
				tt.method,
				tt.path,
				strings.NewReader(tt.body),
			)

			if tt.body != "" {
				req.Header.Set(
					"Content-Type",
					"application/json",
				)
			}

			rec := httptest.NewRecorder()

			newTestRouter().ServeHTTP(
				rec,
				req,
			)

			if rec.Code != tt.wantStatus {
				t.Fatalf(
					"%s %s: expected status %d, got %d: %s",
					tt.method,
					tt.path,
					tt.wantStatus,
					rec.Code,
					rec.Body.String(),
				)
			}

			if requestID := rec.Header().Get("X-Request-ID"); requestID == "" {
				t.Fatal("expected X-Request-ID header")
			}
		})
	}
}

func TestRouterContractRejectsUnknownPaths(t *testing.T) {
	t.Parallel()

	tests := []string{
		"/does-not-exist",
		"/api/v1/does-not-exist",
		"/api/v2/communities",
		"/api/v1/community",
		"/api/v1/communities/1/extra",
	}

	for _, path := range tests {
		path := path

		t.Run(path, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(
				http.MethodGet,
				path,
				nil,
			)

			rec := httptest.NewRecorder()

			newTestRouter().ServeHTTP(
				rec,
				req,
			)

			if rec.Code != http.StatusNotFound {
				t.Fatalf(
					"expected %s to return 404, got %d",
					path,
					rec.Code,
				)
			}
		})
	}
}

func TestRouterContractRejectsUnsupportedMethods(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{
			name:   "PUT collection",
			method: http.MethodPut,
			path:   "/api/v1/communities",
		},
		{
			name:   "DELETE collection",
			method: http.MethodDelete,
			path:   "/api/v1/communities",
		},
		{
			name:   "PATCH collection",
			method: http.MethodPatch,
			path:   "/api/v1/communities",
		},
		{
			name:   "POST resource",
			method: http.MethodPost,
			path:   "/api/v1/communities/1",
		},
		{
			name:   "PATCH resource",
			method: http.MethodPatch,
			path:   "/api/v1/communities/1",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(
				tt.method,
				tt.path,
				nil,
			)

			rec := httptest.NewRecorder()

			newTestRouter().ServeHTTP(
				rec,
				req,
			)

			if rec.Code != http.StatusMethodNotAllowed {
				t.Fatalf(
					"expected %s %s to return 405, got %d",
					tt.method,
					tt.path,
					rec.Code,
				)
			}
		})
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
