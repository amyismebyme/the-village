package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCommunityRouteRegistration(t *testing.T) {
	mux := http.NewServeMux()

	// Temporary handler just to prove the route pattern works.
	mux.HandleFunc(
		"GET /api/v1/communities/{id}",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/communities/1",
		nil,
	)

	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusTeapot {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusTeapot,
			rec.Code,
		)
	}
}
