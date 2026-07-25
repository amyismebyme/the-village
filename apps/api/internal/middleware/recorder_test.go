package middleware

import (
	"net/http"
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/testutil"
)

func TestResponseRecorderCapturesStatus(t *testing.T) {

	rec := &responseRecorder{
		ResponseWriter: testutil.NewRecorder(),
	}

	rec.WriteHeader(http.StatusCreated)

	if rec.status != http.StatusCreated {
		t.Fatalf("expected %d got %d",
			http.StatusCreated,
			rec.status,
		)
	}
}

func TestResponseRecorderDefaultsTo200(t *testing.T) {

	rec := &responseRecorder{
		ResponseWriter: testutil.NewRecorder(),
	}

	_, err := rec.Write([]byte("hello"))
	if err != nil {
		t.Fatal(err)
	}

	if rec.status != http.StatusOK {
		t.Fatalf("expected %d got %d",
			http.StatusOK,
			rec.status,
		)
	}
}
