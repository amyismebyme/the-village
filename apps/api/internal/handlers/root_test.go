package handlers

import (
	"net/http"
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/testutil"
)

func TestRootHandler(t *testing.T) {

	req := testutil.NewRequest(http.MethodGet, "/")
	rr := testutil.NewRecorder()

	RootHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected %d got %d",
			http.StatusOK,
			rr.Code,
		)
	}

	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected application/json got %s", ct)
	}

	var response RootResponse

	testutil.DecodeJSON(
		t,
		rr.Body.Bytes(),
		&response,
	)

	if response.Service != "village-api" {
		t.Errorf("unexpected service: %s", response.Service)
	}

	if response.Status != "running" {
		t.Errorf("unexpected status: %s", response.Status)
	}
}
