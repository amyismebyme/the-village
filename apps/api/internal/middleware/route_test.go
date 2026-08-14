package middleware

import (
	"net/http"
	"testing"
)

func TestRouteLabel(t *testing.T) {

	req, _ := http.NewRequest(
		http.MethodGet,
		"/api/v1/communities/123",
		nil,
	)

	req.Pattern = "GET /api/v1/communities/{id}"

	got := routeLabel(req)

	want := "/api/v1/communities/{id}"

	if got != want {
		t.Fatalf(
			"expected %q, got %q",
			want,
			got,
		)
	}
}
