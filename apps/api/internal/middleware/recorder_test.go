package middleware

import (
	"net/http"
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/httputil"
	"github.com/amyismebyme/the-village/apps/api/internal/testutil"
)

func TestResponseRecorderCapturesStatus(t *testing.T) {
	rec := httputil.NewResponseRecorder(
		testutil.NewRecorder(),
	)

	rec.WriteHeader(http.StatusCreated)

	if rec.Status != http.StatusCreated {
		t.Fatalf(
			"expected %d got %d",
			http.StatusCreated,
			rec.Status,
		)
	}
}

func TestResponseRecorderDefaultsTo200(t *testing.T) {
	rec := httputil.NewResponseRecorder(
		testutil.NewRecorder(),
	)

	_, err := rec.Write([]byte("hello"))
	if err != nil {
		t.Fatal(err)
	}

	if rec.Status != http.StatusOK {
		t.Fatalf(
			"expected %d got %d",
			http.StatusOK,
			rec.Status,
		)
	}
}
