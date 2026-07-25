package handlers

import (
	"net/http"
	"strings"
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/testutil"
)

func TestStatusHandler(t *testing.T) {

	req := testutil.NewRequest(http.MethodGet, "/status")
	rr := testutil.NewRecorder()

	StatusHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected %d got %d",
			http.StatusOK,
			rr.Code,
		)
	}

	var response StatusResponse

	testutil.DecodeJSON(
		t,
		rr.Body.Bytes(),
		&response,
	)

	if response.Status != "running" {
		t.Errorf("unexpected status %s", response.Status)
	}

	if response.Version == "" {
		t.Error("version should not be empty")
	}

	if response.Uptime == "" {
		t.Error("uptime should not be empty")
	}

	if !strings.HasPrefix(response.GoVersion, "go") {
		t.Errorf("unexpected go version %s", response.GoVersion)
	}

	if response.StartedAt == "" {
		t.Error("started_at should not be empty")
	}

	if response.GitCommit == "" {
		t.Error("git commit should not be empty")
	}
}
