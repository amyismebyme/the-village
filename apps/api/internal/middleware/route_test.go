package middleware

import (
	"net/http"
	"strings"
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

func TestRouteLabelWithoutPatternFallsBackToPath(t *testing.T) {
	req, err := http.NewRequest(
		http.MethodGet,
		"/api/v1/communities/123",
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	got := routeLabel(req)

	want := "/api/v1/communities/123"

	if got != want {
		t.Fatalf(
			"expected %q, got %q",
			want,
			got,
		)
	}
}

func TestRouteLabelDoesNotExposeMethod(t *testing.T) {
	req, err := http.NewRequest(
		http.MethodGet,
		"/api/v1/communities/123",
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	req.Pattern = "GET /api/v1/communities/{id}"

	got := routeLabel(req)

	if strings.Contains(got, "GET ") {
		t.Fatalf(
			"route label contains HTTP method: %q",
			got,
		)
	}
}
