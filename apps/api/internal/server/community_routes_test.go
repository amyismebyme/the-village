package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/handlers"
)

func newCommunityRouteTestMux() *http.ServeMux {
	mux := http.NewServeMux()

	handler := handlers.NewHandler(
		routerCommunityServiceMock{},
	)

	registerCommunityRoutes(
		mux,
		handler,
	)

	return mux
}

func TestRegisterCommunityRoutes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{
			name:   "GET collection",
			method: http.MethodGet,
			path:   "/api/v1/communities",
		},
		{
			name:   "POST collection",
			method: http.MethodPost,
			path:   "/api/v1/communities",
		},
		{
			name:   "GET resource",
			method: http.MethodGet,
			path:   "/api/v1/communities/1",
		},
		{
			name:   "PUT resource",
			method: http.MethodPut,
			path:   "/api/v1/communities/1",
		},
		{
			name:   "DELETE resource",
			method: http.MethodDelete,
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

			if tt.method == http.MethodPost {
				req = httptest.NewRequest(
					tt.method,
					tt.path,
					nil,
				)
			}

			rec := httptest.NewRecorder()

			newCommunityRouteTestMux().ServeHTTP(
				rec,
				req,
			)

			// A correctly registered method-aware route must not
			// fall through to ServeMux's default 404 handler.
			if rec.Code == http.StatusNotFound &&
				rec.Body.String() == "404 page not found\n" {
				t.Fatalf(
					"route not registered: %s %s",
					tt.method,
					tt.path,
				)
			}
		})
	}
}

func TestRegisterCommunityRoutesRejectsUnsupportedMethods(
	t *testing.T,
) {
	t.Parallel()

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{
			name:   "PATCH collection",
			method: http.MethodPatch,
			path:   "/api/v1/communities",
		},
		{
			name:   "DELETE collection",
			method: http.MethodDelete,
			path:   "/api/v1/communities",
		},
		{
			name:   "PATCH resource",
			method: http.MethodPatch,
			path:   "/api/v1/communities/1",
		},
		{
			name:   "POST resource",
			method: http.MethodPost,
			path:   "/api/v1/communities/1",
		},
		{
			name:   "OPTIONS collection",
			method: http.MethodOptions,
			path:   "/api/v1/communities",
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

			newCommunityRouteTestMux().ServeHTTP(
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

func TestRegisterCommunityRoutesRejectsInvalidPaths(
	t *testing.T,
) {
	t.Parallel()

	tests := []struct {
		name string
		path string
	}{
		{
			name: "missing collection path",
			path: "/api/v1",
		},
		{
			name: "singular community path",
			path: "/api/v1/community",
		},
		{
			name: "extra resource path",
			path: "/api/v1/communities/1/extra",
		},
		{
			name: "wrong API version",
			path: "/api/v2/communities",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(
				http.MethodGet,
				tt.path,
				nil,
			)

			rec := httptest.NewRecorder()

			newCommunityRouteTestMux().ServeHTTP(
				rec,
				req,
			)

			if rec.Code != http.StatusNotFound {
				t.Fatalf(
					"expected %s to return 404, got %d",
					tt.path,
					rec.Code,
				)
			}
		})
	}
}
