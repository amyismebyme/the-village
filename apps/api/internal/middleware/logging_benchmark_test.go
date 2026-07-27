package middleware

import (
	"net/http"
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/testutil"
)

func BenchmarkLoggingMiddleware(b *testing.B) {

	logger := testutil.NewDiscardLogger()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := Logging(logger, next)

	req := testutil.NewRequest(http.MethodGet, "/health")

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		rr := testutil.NewRecorder()
		handler.ServeHTTP(rr, req)
	}
}
