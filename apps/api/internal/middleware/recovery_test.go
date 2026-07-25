package middleware

import (
	"net/http"
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/testutil"
)

func TestRecoveryRecoversFromPanic(t *testing.T) {

	logger := testutil.NewDiscardLogger()

	handler := Recovery(
		logger,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			panic("boom")

		}),
	)

	req := testutil.NewRequest(http.MethodGet, "/panic")
	rr := testutil.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected %d got %d",
			http.StatusInternalServerError,
			rr.Code,
		)
	}
}

func TestRecoveryPassesThrough(t *testing.T) {

	logger := testutil.NewDiscardLogger()

	called := false

	handler := Recovery(
		logger,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			called = true

			w.WriteHeader(http.StatusNoContent)

		}),
	)

	req := testutil.NewRequest(http.MethodGet, "/")
	rr := testutil.NewRecorder()

	handler.ServeHTTP(rr, req)

	if !called {
		t.Fatal("handler not called")
	}

	if rr.Code != http.StatusNoContent {
		t.Fatalf(
			"expected %d got %d",
			http.StatusNoContent,
			rr.Code,
		)
	}
}
