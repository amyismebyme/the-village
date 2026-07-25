package handlers

import (
	"net/http"
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/testutil"
)

func TestReadyHandler(t *testing.T) {

	req := testutil.NewRequest(http.MethodGet, "/ready")
	rr := testutil.NewRecorder()

	ReadyHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected %d got %d",
			http.StatusOK,
			rr.Code,
		)
	}

	var response ReadyResponse

	testutil.DecodeJSON(
		t,
		rr.Body.Bytes(),
		&response,
	)

	if response.Status != "ready" {
		t.Errorf("expected ready got %s", response.Status)
	}
}
