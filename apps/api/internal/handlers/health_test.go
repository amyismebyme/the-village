package handlers

import (
	"net/http"
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/testutil"
)

func TestHealthHandler(t *testing.T) {

	req := testutil.NewRequest(http.MethodGet, "/health")
	rr := testutil.NewRecorder()

	HealthHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected %d got %d",
			http.StatusOK,
			rr.Code,
		)
	}

	var response HealthResponse

	testutil.DecodeJSON(
		t,
		rr.Body.Bytes(),
		&response,
	)

	if response.Status != "healthy" {
		t.Errorf("expected healthy got %s", response.Status)
	}
}
