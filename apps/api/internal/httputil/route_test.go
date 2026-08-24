package httputil

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouteLabelWithMethodAwareServeMuxPattern(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/communities/{id}", func(w http.ResponseWriter, _ *http.Request) {})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/communities/123", nil)
	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, req)

	if got := RouteLabel(req); got != "/api/v1/communities/{id}" {
		t.Fatalf("expected normalized route, got %q", got)
	}
}

func TestRouteLabelFallsBackToPath(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	if got := RouteLabel(req); got != "/health" {
		t.Fatalf("expected /health, got %q", got)
	}
}

func TestRouteLabelDefaultsToRoot(t *testing.T) {
	req := &http.Request{}
	if got := RouteLabel(req); got != "/" {
		t.Fatalf("expected /, got %q", got)
	}
}
