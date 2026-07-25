package handlers

import (
	"net/http"
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/testutil"
)

func TestVersionHandler(t *testing.T) {

	req := testutil.NewRequest(http.MethodGet, "/version")
	rr := testutil.NewRecorder()

	VersionHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected %d got %d",
			http.StatusOK,
			rr.Code,
		)
	}

	var response VersionResponse

	testutil.DecodeJSON(
		t,
		rr.Body.Bytes(),
		&response,
	)

	if response.Version == "" {
		t.Error("version should not be empty")
	}

	if response.BuildDate == "" {
		t.Error("build date should not be empty")
	}

	if response.GitCommit == "" {
		t.Error("git commit should not be empty")
	}
}
