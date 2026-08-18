package server

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/handlers"
	"github.com/amyismebyme/the-village/apps/api/internal/health"
	"github.com/amyismebyme/the-village/apps/api/internal/model"
)

type routerContractCommunityService struct{}

func (routerContractCommunityService) Create(
	_ context.Context,
	community *model.Community,
) error {
	community.ID = 1

	return nil
}

func (routerContractCommunityService) Get(
	_ context.Context,
	_ int64,
) (*model.Community, error) {
	return nil, nil
}

func (routerContractCommunityService) List(
	_ context.Context,
) ([]*model.Community, error) {
	return []*model.Community{}, nil
}

func (routerContractCommunityService) Update(
	_ context.Context,
	_ *model.Community,
) error {
	return nil
}

func (routerContractCommunityService) Delete(
	_ context.Context,
	_ int64,
) error {
	return nil
}

func newRouterContractHandler() http.Handler {
	logger := slog.New(
		slog.NewTextHandler(
			httptest.NewRecorder(),
			nil,
		),
	)

	healthRegistry := health.NewRegistry()

	handler := handlers.NewHandler(
		routerContractCommunityService{},
	)

	return NewRouter(
		logger,
		healthRegistry,
		handler,
	)
}

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

			newRouterContractHandler().ServeHTTP(
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

			newRouterContractHandler().ServeHTTP(
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

			newRouterContractHandler().ServeHTTP(
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
