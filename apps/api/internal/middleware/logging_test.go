package middleware

import (
	"net/http"
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/testutil"
)

func TestLoggingCallsNextHandler(t *testing.T) {

	logger := testutil.NewDiscardLogger()

	called := false

	handler := Logging(
		logger,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			called = true

			w.WriteHeader(http.StatusAccepted)
		}),
	)

	req := testutil.NewRequest(http.MethodGet, "/health")
	rr := testutil.NewRecorder()

	handler.ServeHTTP(rr, req)

	if !called {
		t.Fatal("next handler not called")
	}

	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected %d got %d",
			http.StatusAccepted,
			rr.Code,
		)
	}
}
