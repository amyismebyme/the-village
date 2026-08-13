package server

import (
	"context"
	"github.com/amyismebyme/the-village/apps/api/internal/handlers"
	"github.com/amyismebyme/the-village/apps/api/internal/health"
	"github.com/amyismebyme/the-village/apps/api/internal/model"
	"github.com/amyismebyme/the-village/apps/api/internal/service"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// -----------------------------------------------------------------------------
// Mock Community Service
// -----------------------------------------------------------------------------

type routerCommunityServiceMock struct{}

func (routerCommunityServiceMock) Create(
	_ context.Context,
	_ *model.Community,
) error {
	return nil
}

func assertAllowMethods(
	t *testing.T,
	resp *http.Response,
	expected ...string,
) {
	t.Helper()

	allow := resp.Header.Get("Allow")

	if allow == "" {
		t.Fatal("expected Allow header to be present")
	}

	actual := map[string]bool{}

	for _, method := range strings.Split(allow, ",") {
		actual[strings.TrimSpace(method)] = true
	}

	for _, method := range expected {
		if !actual[method] {
			t.Fatalf(
				"expected Allow header to contain %q, got %q",
				method,
				allow,
			)
		}
	}
}

func (routerCommunityServiceMock) Get(
	_ context.Context,
	_ int64,
) (*model.Community, error) {
	return nil, nil
}

func (routerCommunityServiceMock) List(
	_ context.Context,
) ([]*model.Community, error) {
	return nil, nil
}

func (routerCommunityServiceMock) Update(
	_ context.Context,
	_ *model.Community,
) error {
	return nil
}

func (routerCommunityServiceMock) Delete(
	_ context.Context,
	_ int64,
) error {
	return nil
}

var _ service.CommunityService = (*routerCommunityServiceMock)(nil)

// -----------------------------------------------------------------------------
// Test router construction
// -----------------------------------------------------------------------------

func newTestRouter() http.Handler {
	logger := slog.New(slog.NewTextHandler(
		httptest.NewRecorder(),
		nil,
	))

	healthRegistry := health.NewRegistry()

	communityService := routerCommunityServiceMock{}
	handler := handlers.NewHandler(communityService)

	return NewRouter(
		logger,
		healthRegistry,
		handler,
	)
}

// -----------------------------------------------------------------------------
// 404 - unknown route
// -----------------------------------------------------------------------------

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

// -----------------------------------------------------------------------------
// Community collection route
// -----------------------------------------------------------------------------

func TestRouterCommunityCollectionMethods(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		method     string
		wantStatus int
		wantAllow  string
	}{
		{
			name:       "POST allowed",
			method:     http.MethodPost,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "GET allowed",
			method:     http.MethodGet,
			wantStatus: http.StatusOK,
		},
		{
			name:       "PUT rejected",
			method:     http.MethodPut,
			wantStatus: http.StatusMethodNotAllowed,
			wantAllow:  http.MethodGet + ", " + http.MethodPost,
		},
		{
			name:       "DELETE rejected",
			method:     http.MethodDelete,
			wantStatus: http.StatusMethodNotAllowed,
			wantAllow:  http.MethodGet + ", " + http.MethodPost,
		},
		{
			name:       "PATCH rejected",
			method:     http.MethodPatch,
			wantStatus: http.StatusMethodNotAllowed,
			wantAllow:  http.MethodGet + ", " + http.MethodPost,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := newTestRouter()

			var body *httptest.ResponseRecorder

			if tt.method == http.MethodPost {
				body = httptest.NewRecorder()

				req := httptest.NewRequest(
					tt.method,
					"/api/v1/communities",
					nil,
				)

				router.ServeHTTP(body, req)

				// POST reaches the handler. With an empty body,
				// the handler is expected to reject the request.
				if body.Code != http.StatusBadRequest {
					t.Fatalf(
						"expected status %d, got %d",
						http.StatusBadRequest,
						body.Code,
					)
				}

				return
			}

			req := httptest.NewRequest(
				tt.method,
				"/api/v1/communities",
				nil,
			)

			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf(
					"expected status %d, got %d",
					tt.wantStatus,
					rec.Code,
				)
			}

			if tt.wantAllow != "" {
				assertAllowMethods(
					t,
					rec.Result(),
					strings.Split(tt.wantAllow, ", ")...,
				)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// Community resource route
// -----------------------------------------------------------------------------

func TestRouterCommunityResourceMethods(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		method     string
		wantStatus int
		wantAllow  string
	}{
		{
			name:       "GET allowed",
			method:     http.MethodGet,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "PUT allowed",
			method:     http.MethodPut,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "POST rejected",
			method:     http.MethodPost,
			wantStatus: http.StatusMethodNotAllowed,
			wantAllow: http.MethodGet + ", " +
				http.MethodPut + ", " +
				http.MethodDelete,
		},
		{
			name:       "DELETE allowed",
			method:     http.MethodDelete,
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "PATCH rejected",
			method:     http.MethodPatch,
			wantStatus: http.StatusMethodNotAllowed,
			wantAllow: http.MethodGet + ", " +
				http.MethodPut + ", " +
				http.MethodDelete,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := newTestRouter()

			req := httptest.NewRequest(
				tt.method,
				"/api/v1/communities/1",
				nil,
			)

			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf(
					"expected status %d, got %d",
					tt.wantStatus,
					rec.Code,
				)
			}

			if tt.wantAllow != "" {
				assertAllowMethods(
					t,
					rec.Result(),
					strings.Split(tt.wantAllow, ", ")...,
				)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// Health routes
// -----------------------------------------------------------------------------

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

	// The test registry has no database dependency configured,
	// so the exact readiness status is determined by the existing
	// health implementation. The important router contract here
	// is that /ready is registered and does not return 404.
	if rec.Code == http.StatusNotFound {
		t.Fatalf(
			"expected /ready to be registered, got 404",
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
